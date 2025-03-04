package util

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
)

func GetJson(v interface{}) string {
	marshal, _ := json.Marshal(v)
	return string(marshal)
}

// ParseIEEE754Float 解析小端存储的 IEEE 754 单精度浮点数
func ParseIEEE754Float(data []byte) (float64, error) {
	if len(data) != 4 {
		return 0, fmt.Errorf("数据长度错误，期望 4 字节，实际 %d 字节", len(data))
	}

	// 扭转字节序（小端转大端）
	reversed := []byte{data[2], data[3], data[0], data[1]}

	// 转换为 uint32
	bits := binary.BigEndian.Uint64(reversed)

	// 转换为 IEEE 754 浮点数
	return math.Float64frombits(bits), nil
}
