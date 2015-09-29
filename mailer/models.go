package mailer

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"html/template"
	"regexp"
	"strings"
)

func statOpened(campaignId string, recipientId string) {
	_ = Db.QueryRow("UPDATE recipient SET opened=1 WHERE id=? AND campaign_id=?", recipientId, campaignId)
}

func statJump(campaignId string, recipientId string, url string) {
	row, err := Db.Query("INSERT INTO jumping (campaign_id, recipient_id, url) VALUES (?, ?, ?)", campaignId, recipientId, url)
	checkErr(err)
	defer row.Close()
}

func postUnsubscribe(campaignId string, recipientId string) {
	_, err := Db.Query("INSERT INTO unsubscribe (`group_id`, `email`) VALUE ((SELECT group_id FROM campaign WHERE id=?), (SELECT email FROM recipient WHERE id=?))", campaignId, recipientId)
	checkErr(err)
}

func getWebMessage(campaignId string, recipientId string) string {
	m, err := getMessage(campaignId, recipientId)
	checkErr(err)
	return string(m.Body)
}

func getMailMessage(campaignId string, recipientId string) message {
	m, err := getMessage(campaignId, recipientId)
	checkErr(err)
	return m
}

func getWebUrl(campaignId string, recipientId string) string {
	dat := pJson{
		Campaign:    campaignId,
		Recipient: recipientId,
		Url:         "",
		Webver:      "y",
		Opened:      "",
		Unsubscribe: "",
	}
	j, err := json.Marshal(dat)
	checkErr(err)
	return HostName + "/data/" + base64.URLEncoding.EncodeToString(j) + ".html"
}

func getUnsubscribeUrl(campaignId string, recipientId string) string {
	dat := pJson{
		Campaign:    campaignId,
		Recipient:   recipientId,
		Url:         "",
		Webver:      "",
		Opened:      "",
		Unsubscribe: "y",
	}
	j, err := json.Marshal(dat)
	checkErr(err)
	return HostName + "/data/" + base64.URLEncoding.EncodeToString(j) + ".html"
}

func getStatUrl(campaignId string, recipientId string, url string) string {
	d := pJson{
		Campaign:    campaignId,
		Recipient:   recipientId,
		Url:         url,
		Webver:      "",
		Opened:      "",
		Unsubscribe: "",
	}
	j, err := json.Marshal(d)
	checkErr(err)
	return HostName + "/data/" + base64.URLEncoding.EncodeToString(j)
}

func getStatPngUrl(campaignId string, recipientId string) string {
	dat := pJson{
		Campaign:    campaignId,
		Recipient:   recipientId,
		Url:         "",
		Webver:      "",
		Opened:      "y",
		Unsubscribe: "",
	}
	j, err := json.Marshal(dat)
	checkErr(err)
	return HostName + "/data/" + base64.URLEncoding.EncodeToString(j) + ".png"
}

func getMessage(campaignId string, recipientId string) (message, error) {
	var subject, body string

	err := Db.QueryRow("SELECT `subject`, `message` FROM campaign WHERE `id`=?", campaignId).Scan(&subject, &body)
	if err != nil {
		return message{Subject: "Error", Body: "Message not found"}, nil
	} else {
		// Replace links for statistic
		rx := regexp.MustCompile(`href=["'](.*?)["']`)
		body = rx.ReplaceAllStringFunc(body, func(str string) string {
			s := strings.Replace(str, "'", "", -1)
			s = strings.Replace(s, `"`, "", -1)
			s = strings.Replace(s, "href=", "", 1)
			switch s {
			case "{{.WebUrl}}":
				return `href="` + getWebUrl(campaignId, recipientId) + `"`
			case "{{.UnsubscribeUrl}}":
				return `href="` + getUnsubscribeUrl(campaignId, recipientId) + `"`
			default:
				return `href="` + getStatUrl(campaignId, recipientId, s) + `"`
			}
		})
		tmpl := template.New("mail")
		tmpl, err = tmpl.Parse(body)
		if err != nil {
			return message{Subject: "Error", Body: "Error parse template"}, nil
		}
		people := getRecipientParam(recipientId)
		people["StatPng"] = getStatPngUrl(campaignId, recipientId)
		t := bytes.NewBufferString("")
		tmpl.Execute(t, people)
		return message{Subject: subject, Body: t.String()}, nil
	}
}

func getRecipientParam(id string) map[string]string {
	var paramKey, paramValue string
	recipient := make(map[string]string)
	param, err := Db.Query("SELECT `key`, `value` FROM parameter WHERE recipient_id=?", id)
	checkErr(err)
	defer param.Close()
	for param.Next() {
		err = param.Scan(&paramKey, &paramValue)
		checkErr(err)
		recipient[string(paramKey)] = string(paramValue)
	}
	return recipient
}
