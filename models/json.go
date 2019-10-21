package models

import (
	"encoding/json"
	"io"
)

type JSONResponse struct {
	Status  string      `json:"status"`
	Message interface{} `json:"message"`
}

func (jr JSONResponse) OkWriter(w io.Writer, message interface{}) error {
	return json.NewEncoder(w).Encode(JSONResponse{Status: "ok", Message: message})
}

func (jr JSONResponse) ErrorWriter(w io.Writer, err error) error {
	return json.NewEncoder(w).Encode(JSONResponse{Status: "error", Message: err.Error()})
}