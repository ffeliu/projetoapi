package model

// swagger:model
type Evaluation struct {
	Id     int    `json:"Id"`
	Rating int    `json:"Rating"`
	Note   string `json:"Note"`
}
