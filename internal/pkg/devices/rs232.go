package devices

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tarm/serial"
	"time"
)

type RS232Device struct {
	name         string
	port         *serial.Port
	dataChannel  chan []byte
	dataHandlers []func([]byte) // 添加一个用于存储数据处理器的切片
}

// NewRS485Device 创建一个新的 RS485 设备实例
func NewRS232Device(name string) *RS232Device {
	return &RS232Device{
		name:         name,
		dataChannel:  make(chan []byte),
		dataHandlers: make([]func([]byte), 0), // 初始化 dataHandlers
	}
}

// Start 启动设备
func (r *RS232Device) Start() error {
	log.Printf("Starting RS232 device: %s", r.name)

	go r.readFromSerial()
	// 开始监听 dataChannel 并分发数据到所有注册的处理器
	go func() {
		for data := range r.dataChannel {
			for _, handler := range r.dataHandlers {
				handler(data)
			}
		}
	}()

	return nil
}

// Halt 停止设备
func (r *RS232Device) Halt() error {
	// 停止设备时关闭 Modbus 客户端连接
	log.Printf("Halting RS232 device: %s", r.name)
	return nil
}

// Connection returns the Driver Connection
//func (k *RS232Device) Connection() gobot.Connection { return nil }

// Name 返回设备的名称
func (r *RS232Device) Name() string {
	return r.name
}

// Name 返回设备的名称
func (r *RS232Device) SetName(s string) {
	r.name = s
}

// Connect 连接到 RS485 设备
func (r *RS232Device) Connect(port string) error {
	// 串口连接逻辑
	config := &serial.Config{
		Name:        port,
		Baud:        9600,
		Size:        8,
		StopBits:    1, // 1 停止位
		Parity:      0, // 无校验
		ReadTimeout: time.Microsecond * 100,
	}

	portObj, err := serial.OpenPort(config)
	if err != nil {
		//log.Error(" open serial port error: " + port)
		return err
	}
	r.port = portObj

	log.Printf("Connected to RS485 device %s on port %s", r.name, port)
	return nil
}

// Disconnect 断开与 RS485 设备的连接
func (r *RS232Device) Disconnect() error {
	if r.port != nil {
		if err := r.port.Close(); err != nil {
			return err
		}
	}
	//log.Printf("Disconnected from RS485 device %s", r.name)
	return nil
}

// RegisterDataHandler 注册一个新的数据处理器
func (r *RS232Device) RegisterDataHandler(handler func([]byte)) {
	r.dataHandlers = append(r.dataHandlers, handler)
}

func (r *RS232Device) readFromSerial() {
	buf := make([]byte, 128)
	for {
		n, _ := r.port.Read(buf)
		//if err != nil {
		//	log.Warn("-------- error： " + err.Error())
		//}
		if n > 0 {
			data := make([]byte, n)
			copy(data, buf[:n])
			r.dataChannel <- data // 将读取到的数据发送到channel
		}
	}
}

func (r *RS232Device) writeToSerial(buf []byte) {
	n, err := r.port.Write(buf)
	if err != nil {
		log.Warn(err)
	}
	fmt.Printf("Wrote %d bytes to serial port.\n", n)
}
