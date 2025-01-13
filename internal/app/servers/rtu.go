package servers

import (
	"encoding/hex"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
	"zertuserver/internal/pkg/devices"
)

var (
	rtuImpl *rtu
	rtuOnce sync.Once
)

type rtu struct {
	name        string
	rs485Device *devices.RS485Device
}

func RtuServer() IServer {
	rtuOnce.Do(func() {
		rtuImpl = &rtu{
			name:        "rtuServer",
			rs485Device: devices.NewRS485Device("RS485 Sensor"),
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
func (rs *rtu) work() {
	// 连接设备
	if err := rs.rs485Device.Connect("/dev/tty485_2"); err != nil {
		log.Fatalf("Failed to connect RS485 device: %v", err)
		return
	}

	// 模拟设备数据读取
	go func() {
		for {
			// 读取保持寄存器
			// 模拟读取 Modbus 数据
			address := uint16(0x0001) // Modbus 地址
			quantity := uint16(1)     // 读取数量
			data, err := rs.rs485Device.ReadHoldingRegisters(address, quantity)
			if err != nil {
				log.Error("Error reading registers: ", err)
			} else {
				log.Info("Received Modbus data: ", data)
				log.Info("Received Modbus hex data: ", hex.EncodeToString(data))
				//log2.Println("  Received Modbus hex data: ", hex.EncodeToString(data))
				//fmt.Println(util.CurrentTimeFormat(), "  Received Modbus hex data: ", hex.EncodeToString(data))
			}

			// 在这里可以进一步处理数据
			time.Sleep(1 * time.Second) // 假设设备每 2 秒发送一次数据
		}
	}()
}
