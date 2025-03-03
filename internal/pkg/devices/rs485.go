package devices

import (
	"fmt"
	"github.com/goburrow/modbus"
	log "github.com/sirupsen/logrus"
	"time"
)

// 方向
const (
	DirectForward  uint16 = 0x05 // 前进
	DirectBackward        = 0x09 // 后退
	DirectUp              = 0x0A // 上升
	DirectDown            = 0x11 // 下降
	DirectStop            = 0x13 // 停止
)

// Hz
const (
	FrequencyLevel0 uint16 = iota // 0 -> 0 Hz
	FrequencyLevel1               // 1 -> 5 Hz
	FrequencyLevel2               // 2 -> 10 Hz
	FrequencyLevel3               // 3 -> 15 Hz
	FrequencyLevel4               // 4 -> 20 Hz
	FrequencyLevel5               // 5 -> 25 Hz
	FrequencyLevel6               // 6 -> 37.5 Hz
	FrequencyLevel7               // 7 -> 50 Hz
)

// RS485Device 代表一个 RS485 设备
type RS485Device struct {
	name string
	//port         *serial.Port
	dataChannel  chan []byte
	handler      *modbus.RTUClientHandler
	modbusClient *client
}

// NewRS485Device 创建一个新的 RS485 设备实例
func NewRS485Device(name string) *RS485Device {
	return &RS485Device{name: name}
}

// Start 启动设备
func (r *RS485Device) Start() error {
	// 启动设备时初始化 Modbus 客户端
	log.Printf("Starting RS485 device: %s", r.name)
	return nil
}

// Halt 停止设备
func (r *RS485Device) Halt() error {
	// 停止设备时关闭 Modbus 客户端连接
	if r.modbusClient != nil {
		// 关闭 Modbus 连接
	}
	log.Printf("Halting RS485 device: %s", r.name)
	return nil
}

// Connection returns the Driver Connection
//func (k *RS485Device) Connection() gobot.Connection { return nil }

// Name 返回设备的名称
func (r *RS485Device) Name() string {
	return r.name
}

// Name 返回设备的名称
func (r *RS485Device) SetName(s string) {
	r.name = s
}

func (r *RS485Device) SetHandlerSlaveId(slaveId byte) {
	r.handler.SlaveId = slaveId
}

// Connect 连接到 RS485 设备
func (r *RS485Device) Connect(port string) error {
	// 串口连接逻辑
	//config := &serial.Config{
	//	Name:        port,
	//	Baud:        9600,
	//	Size:        8,
	//	StopBits:    1, // 1 停止位
	//	Parity:      0, // 无校验
	//	ReadTimeout: time.Second,
	//}
	//
	//portObj, err := serial.OpenPort(config)
	//if err != nil {
	//	//log.Error(" open serial port error: " + port)
	//	return err
	//}
	//r.port = portObj

	// 设置 Modbus RTU 客户端
	r.handler = modbus.NewRTUClientHandler(port) // 传入串口配置
	r.handler.BaudRate = 9600
	r.handler.DataBits = 8
	r.handler.Parity = "N" // 无校验
	r.handler.StopBits = 1
	r.handler.Timeout = time.Second

	r.handler.SlaveId = 0x07

	// 设置 Modbus RTU 客户端
	r.modbusClient = NewClient(r.handler)

	log.Printf("Connected to RS485 device %s on port %s", r.name, port)
	return nil
}

// Disconnect 断开与 RS485 设备的连接
func (r *RS485Device) Disconnect() error {
	//if r.port != nil {
	//	if err := r.port.Close(); err != nil {
	//		return err
	//	}
	//}
	//log.Printf("Disconnected from RS485 device %s", r.name)
	return nil
}

// ReadHoldingRegisters 读取 Modbus Holding Registers（例如传感器数据）
func (r *RS485Device) ReadHoldingRegisters(address uint16, quantity uint16) ([]byte, error) {
	// 读取指定地址的 Holding Registers 数据
	results, err := r.modbusClient.ReadHoldingRegisters(address, quantity)
	if err != nil {
		return nil, fmt.Errorf("failed to read holding registers: %v", err)
	}
	log.Info("Read Holding Registers: ", results)
	return results, nil
}

func (r *RS485Device) ReadInputRegisters(address uint16, quantity uint16) ([]byte, error) {
	// 读取指定地址的 Holding Registers 数据
	results, err := r.modbusClient.ReadInputRegisters(address, quantity)
	if err != nil {
		return nil, fmt.Errorf("failed to read holding registers: %v", err)
	}
	log.Info("Read Input Registers: ", results)
	return results, nil
}

// WriteSingleRegister 向 Modbus 设备写入一个单独的寄存器（例如控制命令）
func (r *RS485Device) WriteSingleRegister(address uint16, value uint16) error {
	// 向指定地址写入一个单独的寄存器
	_, err := r.modbusClient.WriteSingleRegister(address, value)
	if err != nil {
		return fmt.Errorf("failed to write single register: %v", err)
	}
	log.Info("Written value %v to register ", value, address)
	return nil
}

// WriteSingleRegister 向 Modbus 设备写入一个单独的寄存器（例如控制命令）
func (r *RS485Device) WriteSingleExRegister(address uint16, value uint16) error {
	// 向指定地址写入一个单独的寄存器
	_, err := r.modbusClient.WriteSingleExRegister(address, value)
	if err != nil {
		return fmt.Errorf("failed to write single register: %v", err)
	}
	log.Info("Written value %v to register ", value, address)
	return nil
}
