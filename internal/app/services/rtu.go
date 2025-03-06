package services

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math"
	"sort"
	"sync"
	"time"
	"zertuserver/internal/app/models"
	"zertuserver/internal/pkg/devices"
	"zertuserver/internal/pkg/task"
	"zertuserver/pkg/util"
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

const (
	PositionStatusIdle    = iota
	PositionStatusInWater //入水
	PositionStatusDepth   //测试水深
	PositionStatusSpeed   //测试流速
)

type rtu struct {
	name           string
	motor          *devices.RS485Device
	newMeter       *devices.RS485Device
	oldMeter       *devices.RS232Device
	mqttDevice     *devices.MqttDevice
	motorData      models.MotorData  //电机数据
	sensorData     models.SensorData //镜像数据
	ticker         *time.Ticker
	taskConfig     *models.TaskConfig
	positionStatus int
}

func RtuService() *rtu {
	rtuOnce.Do(func() {
		rtuImpl = &rtu{
			name:           "rtuServer",
			motor:          devices.NewRS485Device("RS485_motor"),
			newMeter:       devices.NewRS485Device("RS485_meter"),
			oldMeter:       devices.NewRS232Device("RS232_meter"),
			mqttDevice:     devices.NewMqttDevice("tcp://39.107.116.95:1883", "mactest"),
			motorData:      models.MotorData{},
			ticker:         nil,
			positionStatus: PositionStatusIdle,
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

	//r.test()
}

func (r *rtu) test() {
	go func() {
		//for data := range r.buf {
		//	fmt.Printf("Received data: %s\n", data)
		//	// 在这里处理接收到的数据
		r.drive(devices.DirectForward, devices.FrequencyLevel0)
		//}
		//for {
		//r.readNewSpeed()
		//r.readNewDepth()
		//r.readNewDepthAverage(20)
		//r.readNewSpeedAverage(20)
		//time.Sleep(200 * time.Millisecond) // 假设设备每 2 秒发送一次数据
		//}
		//time.Sleep(1 * time.Second) // 假设设备每 2 秒发送一次数据
	}()
}

// 定时发送状到mqtt,sensor data
func (r *rtu) startMqttTimerSender(interval time.Duration) {
	r.ticker = time.NewTicker(interval)
	go func() {
		defer r.ticker.Stop()
		for range r.ticker.C {
			// 将状态数据序列化为 JSON
			r.sensorData.TS = time.Now().UnixNano() / int64(time.Millisecond)
			jsonData, err := json.Marshal(r.sensorData)
			if err != nil {
				fmt.Printf("Failed to marshal status: %v\n", err)
				continue
			}

			// 发布消息
			success := r.mqttDevice.Publish("/sys/{device_id}/status/upload", jsonData)
			if success == false {
				fmt.Printf("Message sent error: %s\n", r.sensorData)
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

/**
* 读取缆车数据：
* 1. 缆车的起点距，高度
* 2。老设备，可以读取水面，水底，流速
* 3。新设备需要从另一个485口读取水深和流速
 */
func (r *rtu) rs232ReadData(data []byte) {
	log.Infoln("---------RS232 read: " + hex.EncodeToString(data))
	motorData, err := models.ParseData(data)
	if err != nil {
		log.Warnf("prase data error: " + err.Error())
		return
	}
	log.Info("----- parse data: ", motorData)
	r.sensorData.Data.Distance = motorData.Width
	r.sensorData.Data.Height = motorData.Height

	if r.sensorData.Data.TaskStatus != task.StatusIdle {
		if r.taskConfig.Device.DeviceType != 4 {
			//允许入水或者测水深了，才开始
			if r.positionStatus == PositionStatusInWater || r.positionStatus == PositionStatusDepth {
				//需要采集水面，水深，水底，水速信号
				r.readOldInWater(motorData)
			}

			if r.positionStatus == PositionStatusSpeed {
				r.startReadOldSpeed(motorData)
			}
		}
	}

	if motorData.Status == 5 { //到河底一定要停车
		r.drive(devices.DirectStop, devices.FrequencyLevel0) //停车
	}
}

// mqtt接收的数据
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

// 接收到的参数,配置的任务参数
func (r *rtu) setMessage(msg devices.Message) {
	log.Infoln("message handler topic : " + msg.Topic())
	log.Infoln("message handler message : " + string(msg.Payload()))
	var err error
	r.taskConfig, err = models.ParseTaskConfig(msg.Payload())
	if err != nil {
		log.Warn("It is error to parse task config")
		return
	}
	log.Info("taskConfig: ", r.taskConfig)
	r.sensorData.Data.TaskStatus = task.StatusReady //任务准备完成

	ack := models.Ack{ID: r.taskConfig.ID, Ack: 1, Ts: time.Now().UnixMilli()}
	// 将结构体实例转换为JSON字符串
	jsonData, err := json.Marshal(ack)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}
	r.mqttDevice.Publish("/sys/rtu/config/set_ack", jsonData)
}

// mqtt或者 前端发送的指令
func (r *rtu) exectCmd(command models.Command) {
	switch command.Cmd {
	case models.CommandMoveForward: //前进
		fmt.Printf("command: %s\n", models.CommandStrings[command.Cmd])
		r.drive(devices.DirectForward, devices.FrequencyLevel0)
	case models.CommandMoveBackward: //后退
		fmt.Printf("command: %s\n", models.CommandStrings[command.Cmd])
		r.drive(devices.DirectBackward, devices.FrequencyLevel0)
	case models.CommandStop:
		fmt.Printf("command: %s\n", models.CommandStrings[command.Cmd])
		r.drive(devices.DirectStop, devices.FrequencyLevel0)
	case models.CommandMoveUp: //上
		fmt.Printf("command: %s\n", models.CommandStrings[command.Cmd])
		r.drive(devices.DirectUp, devices.FrequencyLevel0)
	case models.CommandMoveDown: //下
		fmt.Printf("command: %s\n", models.CommandStrings[command.Cmd])
		r.drive(devices.DirectDown, devices.FrequencyLevel0)
	case models.CommandReset: //复位
		fmt.Printf("command: %s\n", models.CommandStrings[command.Cmd])
		r.reset()
	case models.CommandStartTask: //前进
		fmt.Printf("command: %s\n", models.CommandStrings[command.Cmd])
	case models.CommandStopTask: //前进
		fmt.Printf("command: %s\n", models.CommandStrings[command.Cmd])
	case models.CommandSpeedTestStart: //测试水流速
		fmt.Printf("command: %s\n", models.CommandStrings[command.Cmd])
		if r.taskConfig.Device.DeviceType != 4 {
			r.positionStatus = PositionStatusSpeed //开始测水速
		} else {
			r.readNewSpeedAverage(20)
		}
	case models.CommandContinueExecution: //继续执行
		fmt.Printf("command: %s\n", models.CommandStrings[command.Cmd])
	default:
		fmt.Printf("Unknown command: %s\n", models.CommandStrings[command.Cmd])
	}
}

/**
 * 开动缆车进行前进后退，设置预定位置开车
 */
func (r *rtu) drive(direct uint16, speed uint16) {
	r.sensorData.Data.MotorStatus = int(direct)
	r.motor.SetHandlerSlaveId(0x07)
	quantity := (direct << 8) | speed
	data, err := r.motor.ReadHoldingRegisters(uint16(0x0001), quantity)
	if err != nil {
		log.Error("Error reading registers: ", err)
	} else {
		log.Info("Received Modbus hex data: ", hex.EncodeToString(data))
	}
}

// 读取新的流速
func (r *rtu) readNewSpeed() (int, error) {
	r.motor.SetHandlerSlaveId(0x01) //从地址为0x01
	data, err := r.motor.ReadHoldingRegisters(uint16(0x0000), uint16(0x0001))
	if err != nil {
		log.Error("Error reading registers: ", err)
	} else {
		log.Info("Received Speed hex data: ", hex.EncodeToString(data))
		log.Info("Received Speed Dex data: ", binary.BigEndian.Uint16(data))
		return int(binary.BigEndian.Uint16(data)), nil
	}

	return 0, errors.New("Read speed error!")
}

// 读取新的水深
func (r *rtu) readNewDepth() (float64, error) {
	r.motor.SetHandlerSlaveId(0x08) //从地址为0x01
	data, err := r.motor.ReadInputRegisters(uint16(0x0065), uint16(0x0002))
	if err != nil {
		log.Error("Error reading registers: ", err)
	} else {
		log.Info("Received Depth hex data: ", hex.EncodeToString(data))
		float, _ := util.ParseIEEE754Float(data)
		log.Info("Received Depth float data: ", float)

		return float, nil
	}
	return 0.0, errors.New("read depth error!")
}

// 读取新设备的水流速
func (r *rtu) readNewSpeedAverage(times int) (float64, error) {
	var results []int
	for i := 0; i < times; i++ {
		speed, err := r.readNewSpeed()
		if err != nil {
			continue
		}
		results = append(results, speed)
		time.Sleep(200 * time.Millisecond) // 假设设备每 2 秒发送一次数据
	}
	// 排序
	sort.Ints(results)

	// 取出中间数据（去掉最大 2 个和最小 2 个）
	validValues := results[2 : len(results)-2]

	// 计算剩余数据的平均值
	var sum int
	for _, v := range validValues {
		sum += v
	}

	// 计算平均值并取整
	average := sum / len(validValues)

	// 转换为 m（单位为米），并保留 2 位小数
	averageInM := float64(average) / 1000.0
	roundedAverage := math.Round(averageInM*100) / 100 // 保留 2 位小数
	log.Info("Received Depth average data: ", roundedAverage)
	return roundedAverage, nil
}

// 读取新设备测试的水深
func (r *rtu) readNewDepthAverage(times int) (float64, error) {
	var results []float64
	for i := 0; i < times; i++ {
		depth, err := r.readNewDepth()
		if err != nil {
			continue
		}
		results = append(results, depth)
		time.Sleep(200 * time.Millisecond) // 假设设备每 2 秒发送一次数据
	}
	// 排序，保证有序存储
	sort.Slice(results, func(i, j int) bool {
		return results[i] < results[j] // 升序排列
	})

	// 取出中间数据
	validValues := results[2 : len(results)-2]

	// 计算平均值
	var sum float64
	for _, v := range validValues {
		sum += v
	}
	average := sum / float64(len(validValues))
	log.Info("Received Depth average data: ", average)
	return average, nil
}

// 采集旧设备的入水信号
func (r *rtu) readOldInWater(motorData *models.MotorData) {
	// 记录水平状态
	if motorData.Status == 1 && r.motorData.StateFlag.WaterSurface == false {
		r.motorData.StateFlag.WaterSurface = true    //入水
		r.sensorData.Data.InWater = 1                //入水
		r.sensorData.Data.Surface = motorData.Height //记录水面高度
		fmt.Println("水平状态已记录")
	}

	// 记录水底状态
	if motorData.Status == 2 && r.motorData.StateFlag.Underwater == false {
		r.motorData.StateFlag.WaterSurface = true                                           //入水
		r.sensorData.Data.Bottom = motorData.Width                                          //记录水底高度
		r.sensorData.Data.WaterDepth = r.sensorData.Data.Bottom - r.sensorData.Data.Surface //水深
		fmt.Println("水底状态已记录")

		//需要停车
		r.drive(devices.DirectStop, devices.FrequencyLevel0) //停车
	}
}

// 读取旧设备的水流速
func (r *rtu) startReadOldSpeed(motorData *models.MotorData) {
	if motorData.Status == 5 {
		if r.motorData.StateFlag.FlowRate == false {
			r.motorData.StateFlag.StartSpeedTime = time.Now()
			r.motorData.StateFlag.FlowRate = true
		} else {
			r.motorData.StateFlag.SpeedCount++
			if time.Since(r.motorData.StateFlag.StartSpeedTime) > 30*time.Second {
				r.motorData.StateFlag.EndSpeedTime = time.Now()
				//计算速度
				r.sensorData.Data.Speed = 1.2 //先默认写了
			}
		}
	}
}

/*********************************任务********************************/
//移动到目标距离
func (r *rtu) moveToTargetDistance(target float64) bool {
	//先启动缆车
	r.drive(devices.DirectForward, devices.FrequencyLevel3)
	for {
		//相差0.2，停车
		if math.Abs(r.motorData.Width-target) <= 0.2 {
			r.drive(devices.DirectStop, devices.FrequencyLevel0) //停车
			return true
		}
	}

	if r.motorData.Width > target {
		if r.motorData.Width-target > 5.0 {
			r.drive(devices.DirectBackward, devices.FrequencyLevel7) //快速运动
		} else {
			r.drive(devices.DirectBackward, devices.FrequencyLevel1) //降速运动
		}

	} else if r.motorData.Width < target {
		if r.motorData.Width-target > 5.0 {
			r.drive(devices.DirectForward, devices.FrequencyLevel7)
		} else {
			r.drive(devices.DirectForward, devices.FrequencyLevel1)
		}
	}

	return false
}

// 移动到测速深度
func (r *rtu) moveToTestDepth() {
	r.drive(devices.DirectDown, devices.FrequencyLevel4) //放缆绳
	//如果是新设备
	if r.taskConfig.Device.DeviceType == 4 {
		//测试水面和水深
		for {
			depth, _ := r.readNewDepth()
			if r.motorData.StateFlag.WaterSurface == false && depth > 0 {
				r.motorData.StateFlag.WaterSurface = true //入水
				r.sensorData.Data.InWater = 1             //入水
				//需要停车，测试水深
				r.drive(devices.DirectStop, devices.FrequencyLevel0)
				break
			}
			time.Sleep(200 * time.Millisecond) // 假设设备每 2 秒发送一次数据
		}
		//测试水深
		depth, _ := r.readNewDepthAverage(20)
		r.sensorData.Data.WaterDepth = depth //获取到水深

	} else {
		//旧设备，直接启动了，
	}
}

// 移动到目标高度
func (r *rtu) moveToTargetHeight(target float64) bool {
	for {
		if math.Abs(r.motorData.Height-target) <= 0.2 {
			//需要停车，测试流速
			r.drive(devices.DirectStop, devices.FrequencyLevel0)
			return true
		}

		if r.motorData.Height-target > 0 { //向上
			r.drive(devices.DirectUp, devices.FrequencyLevel3)
		}

		if r.motorData.Height-target < 0 { //向下
			r.drive(devices.DirectDown, devices.FrequencyLevel3)
		}
	}
}

// 重置
func (r *rtu) reset() {
	r.moveToTargetHeight(r.taskConfig.Params.MaxHeightAboveWater)
	r.moveToTargetDistance(r.taskConfig.Params.MinDistanceFromStart)
}

func (r *rtu) startTestSpeed() {
	if r.taskConfig.Device.DeviceType == 4 {
		average, _ := r.readNewSpeedAverage(20)
		r.sensorData.Data.Speed = average
	} else {
		//通知232测试水流速度,需要定义一些状态
	}
}

func (r *rtu) createReport() {

}

// 手动，只配置基础参数即可
func (r *rtu) startManualTask() {
	r.motorData.StateFlag = models.State{
		UnderVoltage: false, //欠压
		WaterSurface: false, //水面
		Underwater:   false, // 水底
		FlowRate:     false, //流速
		SpeedCount:   0,
	}
	//如果是新设备
	r.sensorData.Data.Distance = r.motorData.Width //测点距
	r.startMqttTimerSender(1 * time.Second)        //1秒发送一次数据

	//需要知道去哪个位置测试

	if r.taskConfig.Device.DeviceType == 4 {
		//测试水面和水深
		for {
			depth, _ := r.readNewDepth()
			if r.motorData.StateFlag.WaterSurface == false && depth > 0 {
				r.motorData.StateFlag.WaterSurface = true //入水
				r.sensorData.Data.InWater = 1             //入水
				//需要停车，测试水深
				r.drive(devices.DirectStop, devices.FrequencyLevel0)
				break
			}
			time.Sleep(200 * time.Millisecond) // 假设设备每 2 秒发送一次数据
		}
		//测试水深
		depth, _ := r.readNewDepthAverage(20)
		r.sensorData.Data.WaterDepth = depth //获取到水深

	} else {
		r.positionStatus = PositionStatusInWater
		//旧设备，直接启动了，
	}
}

// 半自动
func (r *rtu) startSemiTask() {
	position := r.taskConfig.Position
	if len(position) <= 0 { //直接结束
		return
	}
	for i := 0; i < len(position); i++ {
		targetPosition := position[i]
		r.moveToTargetDistance(targetPosition.DistanceFromStart)
		r.moveToTargetHeight(r.sensorData.Data.WaterDepth*0.6 + r.sensorData.Data.Surface)
		r.startTestSpeed()
		r.moveToTargetHeight(r.sensorData.Data.Surface - 1.0)
		r.createReport() //生成报告，阶段完成
	}
}

// 自动
func (r *rtu) startAutoTask() {
	position := r.taskConfig.Position
	if len(position) <= 0 { //直接结束
		return
	}
	for i := 0; i < len(position); i++ {
		targetPosition := position[i]
		r.moveToTargetDistance(targetPosition.DistanceFromStart)
		r.moveToTargetHeight(r.sensorData.Data.WaterDepth*0.6 + r.sensorData.Data.Surface) //到底测速点
		r.startTestSpeed()
		r.moveToTargetHeight(r.sensorData.Data.Surface - 1.0) //到达回收点
		r.createReport()                                      //生成报告，阶段完成
	}
	r.reset() //复位
}
