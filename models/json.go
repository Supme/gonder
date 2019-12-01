package models

import (
	"encoding/json"
	"net/http"
)

type JSONResponse struct {
	Status  string      `json:"status"`
	Message interface{} `json:"message"`
}

func (jr JSONResponse) OkWriter(w http.ResponseWriter, message interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(JSONResponse{Status: "ok", Message: message})
}

func (jr JSONResponse) ErrorWriter(w http.ResponseWriter, err error) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: err.Error()})
}
