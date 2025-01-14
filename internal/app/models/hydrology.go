package models

type HydrologyBaseData struct {
	Distance int `json:"distance"`
	Depth    int `json:"depth"`
	Surface  int `json:"surface"`
	Bottom   int `json:"bottom"`
	Speed    int `json:"speed"`
}
