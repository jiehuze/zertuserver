package services

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
	"zertuserver/internal/app/models"
	"zertuserver/internal/pkg/code"
	"zertuserver/internal/pkg/devices"
	"zertuserver/internal/pkg/task"
)

var (
	rtuImpl *rtu
	rtuOnce sync.Once
)

const (
	DeviceTypeIdle = iota
	DeviceTypeOld
	DeviceTypeNew
)

type rtu struct {
	name       string
	motor      *devices.RS485Device
	newMeter   *devices.RS485Device
	oldMeter   *devices.RS232Device
	mqttDevice *devices.MqttDevice
	status     int
	Task       task.Task
	buf        chan string
	deviceType int
	swDate     code.SWData
	ticker     *time.Ticker
	taskConfig *models.TaskConfig
}

func RtuService() *rtu {
	rtuOnce.Do(func() {
		rtuImpl = &rtu{
			name:       "rtuServer",
			motor:      devices.NewRS485Device("RS485_motor"),
			newMeter:   devices.NewRS485Device("RS485_meter"),
			oldMeter:   devices.NewRS232Device("RS232_meter"),
			mqttDevice: devices.NewMqttDevice("tcp://39.107.116.95:1883", "mactest"),
			status:     task.StatusIdle,
			buf:        make(chan string),
			deviceType: DeviceTypeIdle,
			swDate:     code.SWData{},
			ticker:     nil,
		}

		log.Infoln("rtuServer init finish")
	})
	return rtuImpl
}

func (r *rtu) Start() error {
	log.Infoln("rtu sever start")
	r.work()
	return nil
}

func (r *rtu) Stop() error {
	log.Infoln("rtu server stop")
	return nil
}

func (r *rtu) SendBuf(buf string) {
	r.buf <- buf
}

// 工作函数
func (r *rtu) work() {
	// 连接设备
	if err := r.motor.Connect("/dev/tty485_2"); err != nil {
		log.Warnf("Failed to connect RS485 device: %v", err)
	}

	if err := r.newMeter.Connect("/dev/tty485_2"); err != nil {
		log.Warnf("Failed to connect RS485 device: %v", err)
	}

	if err := r.oldMeter.Connect("/dev/tty232_1"); err != nil {
		log.Warnf("Failed to connect RS485 device: %v", err)
	} else {
		log.Infoln("Success to connect RS232 device: /dev/tty232_1")
		r.oldMeter.RegisterDataHandler(r.rs232ReadData)
		r.oldMeter.Start()
	}

	if err := r.mqttDevice.Connect(); err != nil {
		log.Warnf("Failed to connect mqtt device: %v", err)
	} else {
		_, _ = r.mqttDevice.Subscribe("/sys/rtu/config/set", 2, r.setMessage)
		_, _ = r.mqttDevice.Subscribe("/sys/rtu/control/set", 2, r.cmdMessage)
	}
	//模拟设备数据读取
	go func() {
		for data := range r.buf {
			fmt.Printf("Received data: %s\n", data)
			// 在这里处理接收到的数据
			r.drive(devices.DirectForward, devices.FrequencyLevel0)
		}
	}()
}

// 定时发送状到mqtt
func (r *rtu) startMqttTimerSender(interval time.Duration) {
	r.ticker = time.NewTicker(interval)
	go func() {
		defer r.ticker.Stop()
		for range r.ticker.C {
			// 模拟数据,这个地方应该是发送状态数据，json格式
			message := fmt.Sprintf("Hello MQTT at %s", time.Now().Format(time.RFC3339))

			// 将状态数据序列化为 JSON
			jsonData, err := json.Marshal(message)
			if err != nil {
				fmt.Printf("Failed to marshal status: %v\n", err)
				continue
			}

			// 发布消息
			success := r.mqttDevice.Publish("/sys/{device_id}/status/upload", jsonData)
			if success == false {
				fmt.Printf("Message sent error: %s\n", message)
			}
		}
	}()
}

// 停止定时器
func (r *rtu) StopMqttTimerSender() {
	if r.ticker != nil {
		r.ticker.Stop()
	}
}

func (r *rtu) rs232ReadData(data []byte) {
	log.Infoln("---------RS232 read: " + hex.EncodeToString(data))
	parseData, err := code.ParseData(data)
	if err != nil {
		log.Warnf("prase data error: " + err.Error())
		return
	}
	if r.deviceType == DeviceTypeNew || r.deviceType == DeviceTypeIdle {
		return
	}

	log.Info("----- parse data: ", parseData)
}

// 调整方向
func (r *rtu) cmdMessage(msg devices.Message) {
	log.Infoln("message handler topic : " + msg.Topic())
	log.Infoln("message handler message : " + string(msg.Payload()))
	command, err := models.ParseCommand(string(msg.Payload()))
	if err != nil {
		log.Error("Cmd data error ! error: " + err.Error())
		return
	}
	r.exectCmd(command)
}

func (r *rtu) setMessage(msg devices.Message) {
	log.Infoln("message handler topic : " + msg.Topic())
	log.Infoln("message handler message : " + string(msg.Payload()))
	var err error
	r.taskConfig, err = models.ParseTaskConfig(msg.Payload())
	if err != nil {
		log.Warn("It is error to parse task config")
		return
	}
	r.status = task.StatusReady //任务准备完成
}

func (r *rtu) exectCmd(command models.Command) {
	switch command.Cmd {
	case models.CommandStartTask: //前进
		fmt.Printf("command: %s\n", models.CommandStrings[command.Cmd])
		r.drive(devices.DirectForward, devices.FrequencyLevel0)
	case models.CommandStopTask: //前进
		fmt.Printf("command: %s\n", models.CommandStrings[command.Cmd])
	case models.CommandMoveForward: //前进
		fmt.Printf("command: %s\n", models.CommandStrings[command.Cmd])
	case models.CommandMoveBackward: //后退
		fmt.Printf("command: %s\n", models.CommandStrings[command.Cmd])
	case models.CommandStopCurrentAction:
		fmt.Printf("command: %s\n", models.CommandStrings[command.Cmd])
	case models.CommandMoveUp:
		fmt.Printf("command: %s\n", models.CommandStrings[command.Cmd])
	case models.CommandMoveDown:
		fmt.Printf("command: %s\n", models.CommandStrings[command.Cmd])
	case models.CommandReset:
		fmt.Printf("command: %s\n", models.CommandStrings[command.Cmd])
	case models.CommandSpeedTestStart:
		fmt.Printf("command: %s\n", models.CommandStrings[command.Cmd])
	case models.CommandContinueExecution:
		fmt.Printf("command: %s\n", models.CommandStrings[command.Cmd])
	default:
		fmt.Printf("Unknown command: %s\n", models.CommandStrings[command.Cmd])
	}
}

/**
 * 开动缆车进行前进后退，设置预定位置开车
 */
func (r *rtu) drive(direct uint16, speed uint16) {
	quantity := (direct << 8) | speed
	data, err := r.motor.ReadHoldingRegisters(uint16(0x0001), quantity)
	if err != nil {
		log.Error("Error reading registers: ", err)
	} else {
		log.Info("Received Modbus hex data: ", hex.EncodeToString(data))
	}
	// 在这里可以进一步处理数据
	//time.Sleep(1 * time.Second) // 假设设备每 2 秒发送一次数据
}

// 读取旧的流速
func (r *rtu) readOldSpeed(data code.SWData) {

}
