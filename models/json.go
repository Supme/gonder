package models

type JSONResponse struct {
	Status  string      `json:"status"`
	Message interface{} `json:"message"`
}
