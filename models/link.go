package models

import (
	"encoding/base64"
	"encoding/json"
	"errors"
)

type jsonData struct {
	ID    string `json:"id"` // ToDo int?
	Email string `json:"email"`
	Data  string `json:"data"`
}

// Decode utm data string and return Message whis prefilled id and email
func DecodeUTM(base64data string) (message Message, data string, err error) {
	var param jsonData

	decode, err := base64.URLEncoding.DecodeString(base64data)
	if err != nil {
		return message, data, err
	}
	err = json.Unmarshal([]byte(decode), &param) // ToDo decode whithout reflect
	if err != nil {
		return message, data, err
	}
	data = param.Data
	err = message.New(param.ID)
	if err != nil {
		return message, data, err
	}
	if param.Email != message.RecipientEmail {
		return message, data, errors.New("Not valid recipient")
	}
	return message, data, nil
}

func encodeUTM(message *Message, data string) string {
	j, _ := json.Marshal(
		jsonData{
			ID:    message.RecipientID,
			Email: message.RecipientEmail,
			Data:  data,
		})
	return base64.URLEncoding.EncodeToString(j)
}
