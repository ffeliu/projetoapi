package model

// swagger:model
type EvaluationFullView struct {
	Id     int    `json:"Id"`
	Rating string `json:"Rating"`
	Note   string `json:"Note"`
}
