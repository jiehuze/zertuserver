package services

import (
	"encoding/hex"
	"fmt"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
	"zertuserver/internal/pkg/devices"
)

var (
	rtuImpl *rtu
	rtuOnce sync.Once
)

// Action 定义设备的动作枚举
type Action int

const (
	Forward  Action = iota // 前进
	Backward               // 后退
	Stop                   // 停止
	Up                     // 上升
	Down                   // 下降
)
const (
	IDLE    = iota //空闲状态
	RUNNING        //运行状态
	STOP           //停止状态
	PAUSE          //暂停状态

)

type rtu struct {
	name        string
	rs485Device *devices.RS485Device
	mqttDevice  *devices.MqttDevice
	status      int
}

func RtuService() *rtu {
	rtuOnce.Do(func() {
		rtuImpl = &rtu{
			name:        "rtuServer",
			rs485Device: devices.NewRS485Device("RS485 Sensor"),
			mqttDevice:  devices.NewMqttDevice("", ""),
			status:      IDLE,
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

// 调整方向
func (r *rtu) Direction(direction string) error {
	switch direction {
	case "forward":
		break
	case "backward":
		break
	case "stop":
		break
	case "up":
		break
	case "down":
		break
	case "reset":
		break
	default:
		fmt.Printf("Unknown command: %s\n", direction)
		break
	}

	address := uint16(0x0001) // Modbus 地址
	quantity := uint16(1)     // 读取数量
	data, err := r.rs485Device.ReadHoldingRegisters(address, quantity)
	if err != nil {
		log.Error("Error reading registers: ", err)
	} else {
		log.Info("Received Modbus hex data: ", hex.EncodeToString(data))
	}

	// 在这里可以进一步处理数据
	time.Sleep(1 * time.Second) // 假设设备每 2 秒发送一次数据

	log.Infoln("direction action :", direction)
	return nil
}

func (r *rtu) MqttWork(msg devices.Message) {

}

// 工作函数
func (rs *rtu) work() {
	// 连接设备
	if err := rs.rs485Device.Connect("/dev/tty485_2"); err != nil {
		log.Fatalf("Failed to connect RS485 device: %v", err)
		return
	}

	//if err := rs.mqttDevice.Connect(); err != nil {
	//	log.Fatalf("Failed to connect mqtt device: %v", err)
	//	return
	//}
	//_, _ = rs.mqttDevice.Subscribe("deviceInfo", 2, MqttWork)

	// 模拟设备数据读取
	//go func() {
	//	for {
	//		// 读取保持寄存器
	//		// 模拟读取 Modbus 数据
	//		address := uint16(0x0001) // Modbus 地址
	//		quantity := uint16(1)     // 读取数量
	//		data, err := rs.rs485Device.ReadHoldingRegisters(address, quantity)
	//		if err != nil {
	//			log.Error("Error reading registers: ", err)
	//		} else {
	//			log.Info("Received Modbus data: ", data)
	//			log.Info("Received Modbus hex data: ", hex.EncodeToString(data))
	//			//log2.Println("  Received Modbus hex data: ", hex.EncodeToString(data))
	//			//fmt.Println(util.CurrentTimeFormat(), "  Received Modbus hex data: ", hex.EncodeToString(data))
	//		}
	//
	//		// 在这里可以进一步处理数据
	//		time.Sleep(1 * time.Second) // 假设设备每 2 秒发送一次数据
	//	}
	//}()
}
