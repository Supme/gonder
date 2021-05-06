package models

import (
	"bytes"
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

// EncodeUTM encode command, data with params templating and return unique link
func EncodeUTM(cmd, url, data string, params map[string]interface{}) string {
	if _, ok := params["RecipientId"]; !ok {
		return "Parameters don`t have CampaignId"
	}
	if _, ok := params["RecipientEmail"]; !ok {
		return "Parameters don`t have RecipientEmail"
	}

	if url == "" {
		url = Config.UTMDefaultURL
	}

	if data != "" {
		tmp := bytes.NewBufferString("")
		dataTmpl, err := template.New("url").Parse(data)
		if err != nil {
			return fmt.Sprintf("Error parse data params: %s", err)
		}
		err = dataTmpl.Execute(tmp, params)
		if err != nil {
			return fmt.Sprintf("Error execute template: %s", err)
		}
		data = tmp.String()
	}

	j, _ := json.Marshal(
		utm{
			ID:    params["RecipientId"].(string),
			Email: params["RecipientEmail"].(string),
			Data:  data,
		})
	encode, err := Encrypt(j)
	if err != nil {
		return fmt.Sprintf("Error encrypt: %s", err)
	}
	return url + "/" + cmd + "/" + encode
}

// DecodeUTM decode utm data string and return Message with the pre-filled id and email
func DecodeUTM(base64data string) (Message, string, error) {
	var param utm

	decode, err := Decrypt(base64data)
	if err != nil {
		return Message{}, "", err
	}

	err = json.Unmarshal(decode, &param)
	if err != nil {
		return Message{}, "", err
	}

	var message Message
	data := param.Data
	err = message.New(param.ID)
	if err != nil {
		return message, data, err
	}

	if param.Email != message.RecipientEmail {
		return message, data, errors.New("not valid recipient")
	}

	return message, data, nil
}
