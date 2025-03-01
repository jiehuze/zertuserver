package code

import "fmt"

/*
*

	欠压、	水面		水底		流速		编码

全有		1		1		1		1		0F
全无		0		0		0		0		00
欠压流速	1		0		0		1		09
水面流速	0		1		0		1		05
每组数据加02开始，0D结束。开始握手信号只加0D结束符。
*/
type State struct {
	UnderVoltage bool //欠压
	WaterSurface bool //水面
	Underwater   bool // 水底
	FlowRate     bool //流速
}

type SWData struct {
	Status    int     // 状态码
	Width     float64 // 距离（宽度）
	Height    float64 // 高度
	WidthDir  bool    // 宽度方向，true表示反向，false表示正向
	HeightDir bool    // 高度方向，true表示反向，false表示正向
}

// 计算校验位
func calculateChecksum(data []byte) byte {
	var sum byte = 0
	for _, b := range data[:len(data)] { // 排除最后一个字节（校验位）
		sum += b
	}
	return sum // 返回累加和的低 8 位
}

// 解析数值字段
func parseNum(data []byte, decimalShift int) float64 {
	value := int32(uint16(data[0])<<8 | uint16(data[1]))

	// 应用小数点偏移
	factor := 1.0
	for i := 0; i < decimalShift; i++ {
		factor *= 10
	}
	return float64(value) / factor
}

// 解析报文并返回ParsedData结构体
func ParseData(message []byte) (*SWData, error) {
	// 校验报文长度和开头结尾
	if len(message) != 10 || message[0] != 0x02 || message[len(message)-1] != 0x0D {
		return nil, fmt.Errorf("invalid message format")
	}

	// 状态码
	status := int(message[1])

	// 第一组数据：符号+数值
	firstSymbol := message[2]
	firstValueBytes := message[3:5]
	firstDecimalShift := 1 // 小数点左移一位
	width := parseNum(firstValueBytes, firstDecimalShift)
	widthDir := firstSymbol == 0x00 // FF为反向，00为正向

	// 第二组数据：符号+数值
	secondSymbol := message[5]
	secondValueBytes := message[6:8]
	secondDecimalShift := 2 // 小数点左移两位
	height := parseNum(secondValueBytes, secondDecimalShift)
	heightDir := secondSymbol == 0x00 // FF为反向，00为正向

	// 计算校验位
	calculatedChecksum := calculateChecksum(message[:len(message)-2]) // 不包括校验位和结尾

	if calculatedChecksum != message[len(message)-2] {
		return nil, fmt.Errorf("checksum mismatch: got %v, expected %v", calculatedChecksum, message[len(message)-2])
	}

	// 返回解析结果
	return &SWData{
		Status:    status,
		Width:     width,
		Height:    height,
		WidthDir:  widthDir,
		HeightDir: heightDir,
	}, nil
}
