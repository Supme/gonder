package models

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"text/template"
)

type utm struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Data  string `json:"data"`
}

func EncodeUTM(cmd, data string, params map[string]interface{}) string {
	if _, ok := params["RecipientId"]; !ok {
		return "Parameters don`t have CampaignId"
	}
	if _, ok := params["RecipientEmail"]; !ok {
		return "Parameters don`t have RecipientEmail"
	}

	if data != "" {
		tmp := bytes.NewBufferString("")
		dataTmpl, err := template.New("url").Parse(data)
		if err != nil {
			return fmt.Sprintf("Error parse data params: %s", err)
		}
		dataTmpl.Execute(tmp, params)
		data = tmp.String()
	}

	j, _ := json.Marshal(
		utm{
			ID:    params["RecipientId"].(string),
			Email: params["RecipientEmail"].(string),
			Data:  data,
		})
	return Config.URL + "/" + cmd + "/" + base64.URLEncoding.EncodeToString(j)
}

// Decode utm data string and return Message with the pre-filled id and email
func DecodeUTM(base64data string) (message Message, data string, err error) {
	var param utm

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
