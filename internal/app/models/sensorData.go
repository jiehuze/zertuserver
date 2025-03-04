package models

type Data struct {
	Distance          float64 `json:"distance"`          // 测点距距离，单位为米，浮点数，保留小数点后2位
	WaterDepth        float64 `json:"waterDepth"`        //水深
	Height            float64 `json:"height"`            // 高度，单位为米，浮点数，保留小数点后2位
	MeasurementHeight float64 `json:"measurementHeight"` // 测点高度，单位为米，浮点数，保留小数点后2位
	Status            int     `json:"status"`            // 状态，1：前进，2：后退，3：下，4：上
	Surface           float64 `json:"surface"`           // 水面高度，单位为米，浮点数，保留小数点后2位
	Bottom            float64 `json:"bottom"`            // 水底高度，单位为米，浮点数，保留小数点后2位
	InWater           int     `json:"inWater"`           // 设备状态，1：在水中，0：不在水中
	Speed             float64 `json:"speed"`             //水的流速
	TaskStatus        int     `json:"taskStatus"`        //任务状态 0.任务空闲，1. 任务启动；2：任务执行中；3:任务停止；4：任务结束；5：任务复位；
}

// 状态数据
type SensorData struct {
	Data Data  `json:"data"`
	TS   int64 `json:"ts"` // 时间戳，表示数据记录的时间
}
