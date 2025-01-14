package models

const (
	MODE_AP  int = iota
	MODE_STA     // 后退

)

type WifiInfos struct {
	Mode   int    `json:"mode"`
	Ssid   string `json:"ssid"`
	Pwd    string `json:"pwd"`
	Status int    `json:"status"`
}
