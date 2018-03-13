package models

import (
	"encoding/base64"
	"encoding/json"
)

type jsonData struct {
	ID    string `json:"id"` // ToDo int?
	Email string `json:"email"`
	Data  string `json:"data"`
}

// ToDo Remove it
func encodeUTM(message *Message, data string) string {
	j, _ := json.Marshal(
		jsonData{
			ID:    message.RecipientID,
			Email: message.RecipientEmail,
			Data:  data,
		})
	return base64.URLEncoding.EncodeToString(j)
}
