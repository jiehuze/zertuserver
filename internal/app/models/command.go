package models

import (
	"encoding/json"
	"fmt"
)

// 定义操作命令的“枚举”

const (
	CommandStartTask         int = iota + 1 // 开始任务
	CommandStopTask                         // 终止任务
	CommandMoveForward                      // 前进
	CommandMoveBackward                     // 后退
	CommandMoveUp                           // 上移
	CommandMoveDown                         // 下移
	CommandStopCurrentAction                // 停止当前动作
	CommandReset                            // 复位
	CommandSpeedTestStart                   // 测速开始（设备入水后自动开始，无用）
	CommandContinueExecution                // 继续执行
)

var CommandStrings = map[int]string{
	CommandStartTask:         "开始任务",
	CommandStopTask:          "终止任务",
	CommandMoveForward:       "前进",
	CommandMoveBackward:      "后退",
	CommandMoveUp:            "上移",
	CommandMoveDown:          "下移",
	CommandStopCurrentAction: "停止当前动作",
	CommandReset:             "复位",
	CommandSpeedTestStart:    "测速开始（设备入水后自动开始，无用）",
	CommandContinueExecution: "继续执行",
}

// 定义结构体以匹配给定的JSON格式
type Command struct {
	ID  string `json:"id"`
	Cmd int    `json:"cmd"`
	Ts  int64  `json:"ts"` // 使用int64来存储时间戳
}

// 解析JSON字符串并返回Message结构体实例
func ParseCommand(jsonData string) (Command, error) {
	var cmd Command
	err := json.Unmarshal([]byte(jsonData), &cmd)
	if err != nil {
		return Command{}, fmt.Errorf("error parsing JSON: %v", err)
	}
	return cmd, nil
}
