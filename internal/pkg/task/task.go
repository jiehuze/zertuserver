package task

import (
	"fmt"
	"time"
)

const (
	StatusIdle     int = iota // 任务空闲
	StatusReady               // 任务启动
	StatusStarted             // 任务启动
	StatusRunning             // 任务执行中
	StatusStopped             // 任务停止
	StatusFinished            // 任务结束
	StatusReset               // 任务复位
)

var statusStrings = map[int]string{
	StatusIdle:     "任务空闲",
	StatusStarted:  "任务启动",
	StatusRunning:  "任务执行中",
	StatusStopped:  "任务停止",
	StatusFinished: "任务结束",
	StatusReset:    "任务复位",
}

type Task struct {
	TaskId    int
	Type      int   //类型, 1: 手动；2：半自动化；3：自动化
	Status    int   //状态 0.任务空闲，1. 任务启动；2：任务执行中；3:任务停止；4：任务结束；5：任务复位；
	Stage     int   //任务阶段：分为3个阶段；1：定位位置；2：上下测试水速；3：测试流速
	Positions []int //测点位置

}

/**
* 配置参数
 */
func (t *Task) SetParam() {

}

// 调整定位位置
func (t *Task) Position() {

}

// 上下伸缩
func (t *Task) Retract() {

}

func (t *Task) FlowDetetion() {

}

func (t *Task) Run() {
	//根据参数判断是启动哪种任务
	switch t.Type {
	case 1:
		t.Manual()
	case 2:
		t.SemiAuto()
	case 3:
		t.Auto()
	default:
		t.Manual()
	}
	t.Status = 1
	go t.startRealTimeDataCollection()
}

// 手动
func (t *Task) Manual() {

}

// 半自动化
func (t *Task) SemiAuto() {

}

// 自动化
func (t *Task) Auto() {

}

// 定发送状态
func (t *Task) startRealTimeDataCollection() {
	ticker := time.NewTicker(1 * time.Second) // 每秒触发一次
	defer ticker.Stop()

	for range ticker.C {
		//data, err := s.rs485Client.ReadData(1, 0, 10) // 读取传感器数据
		//if err != nil {
		//	log.Printf("Failed to read sensor data: %v", err)
		//	continue
		//}
		//
		//// 发送实时数据到云端
		//payload := map[string]interface{}{
		//	"distance": distance,
		//	"data":     data,
		//	"timestamp": time.Now().Unix(), // 添加时间戳
		//}
		//s.mqttClient.Publish("sensor/realtime", payload)
		fmt.Printf("Sent real-time data: %+v\n")
	}
}
