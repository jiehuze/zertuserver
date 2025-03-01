package models

import (
	"encoding/json"
	"fmt"
)

// TaskConfig 表示任务配置信息
type TaskConfig struct {
	ID       string     `json:"id"`       // 指令ID
	Task     Task       `json:"task"`     // 任务信息
	Position []Position `json:"position"` // 点位信息集合
	Device   Device     `json:"device"`   // 设备信息
	Params   Params     `json:"params"`   // 通用参数
	TS       int64      `json:"ts"`       // 时间戳
}

// Task 表示任务信息
type Task struct {
	TaskID   string `json:"taskId"`   // 任务ID
	TaskType int    `json:"taskType"` // 任务类型：1 - 手动；2 - 半自动；3 - 全自动
}

// Position 表示点位信息
type Position struct {
	DistanceFromStart  float64 `json:"distanceFromStart"`  // 起点距，单位为米
	EstimatedDepth     float64 `json:"estimatedDepth"`     // 预估水深，单位为米
	IsVelocityMeasured bool    `json:"isVelocityMeasured"` // 是否测流，布尔值
	IsDepthMeasured    bool    `json:"isDepthMeasured"`    // 是否测量水深，布尔值
}

// Device 表示设备信息
type Device struct {
	DeviceType          int     `json:"deviceType"`          // 设备类型：1 - 旧设备1；2 - 旧设备2；3 - 旧设备3；4 - 新设备
	FrequencyDivision   int     `json:"frequencyDivision"`   // 分频
	FilterTime          float64 `json:"filterTime"`          // 滤波时间，单位为秒
	MeasurementDepth    float64 `json:"measurementDepth"`    // 测点深，单位为米
	MeasurementDuration int     `json:"measurementDuration"` // 测速历时，单位为秒
}

// Params 表示通用参数
type Params struct {
	MaxVerticalMotorSpeed     float64 `json:"maxVerticalMotorSpeed"`     // 垂直电机最大速度，单位为米/秒
	MaxHorizontalMotorSpeed   float64 `json:"maxHorizontalMotorSpeed"`   // 水平电机最大速度，单位为米/秒
	HorizontalPreStopDistance float64 `json:"horizontalPreStopDistance"` // 水平预停车距离，单位为米
	VerticalPreStopDistance   float64 `json:"verticalPreStopDistance"`   // 垂直预停车距离，单位为米
}

func ParseTaskConfig(data []byte) (*TaskConfig, error) {
	var taskConfig TaskConfig
	err := json.Unmarshal(data, &taskConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}
	return &taskConfig, nil
}
