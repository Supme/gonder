package models

import (
	"encoding/json"
	"encoding/base64"
	"database/sql"
	"text/template"
	"strings"
	"regexp"
	"bytes"
	"fmt"
	"log"
)

var (
	StatUrl string
	Db *sql.DB
)

type (
	Json struct {
		Campaign    string `json:"c"`
		Recipient   string `json:"r"`
		Url         string `json:"u"`
		Webver      string `json:"w"`
		Opened      string `json:"o"`
		Unsubscribe string `json:"s"`
	}
	message struct {
		Subject string
		Body    string
	}
)

func webUrl(campaignId string, recipientId string) string {
	dat := Json{
		Campaign:    campaignId,
		Recipient: recipientId,
		Url:         "",
		Webver:      "y",
		Opened:      "",
		Unsubscribe: "",
	}
	j, err := json.Marshal(dat)
	checkErr(err)
	return StatUrl + "/data/" + base64.URLEncoding.EncodeToString(j) + ".html"
}

func UnsubscribeUrl(campaignId string, recipientId string) string {
	return unsubscribeUrl(campaignId, recipientId)
}

func unsubscribeUrl(campaignId string, recipientId string) string {
	dat := Json{
		Campaign:    campaignId,
		Recipient:   recipientId,
		Url:         "",
		Webver:      "",
		Opened:      "",
		Unsubscribe: "y",
	}
	j, err := json.Marshal(dat)
	checkErr(err)
	return StatUrl + "/data/" + base64.URLEncoding.EncodeToString(j) + ".html"
}

func statUrl(campaignId string, recipientId string, url string) string {
	d := Json{
		Campaign:    campaignId,
		Recipient:   recipientId,
		Url:         url,
		Webver:      "",
		Opened:      "",
		Unsubscribe: "",
	}
	j, err := json.Marshal(d)
	checkErr(err)
	return StatUrl + "/data/" + base64.URLEncoding.EncodeToString(j)
}

func statPngUrl(campaignId string, recipientId string) string {
	dat := Json{
		Campaign:    campaignId,
		Recipient:   recipientId,
		Url:         "",
		Webver:      "",
		Opened:      "y",
		Unsubscribe: "",
	}
	j, err := json.Marshal(dat)
	checkErr(err)
	return StatUrl + "/data/" + base64.URLEncoding.EncodeToString(j) + ".png"
}

func MailMessage(campaignId, recipientId, subject, body string ) (message, error) {
	return getMessage(campaignId, recipientId, subject, body)
}

func WebMessage(campaignId string, recipientId string) string {
	subject := ""
	body := ""
	m, err := getMessage(campaignId, recipientId, subject, body)
	checkErr(err)
	return string(m.Body)
}

func getMessage(campaignId, recipientId, subject, body string) (message, error) {

	var err error
	var web bool

	if subject == "" && body == "" {
		err = Db.QueryRow("SELECT `subject`, `body` FROM campaign WHERE `id`=?", campaignId).Scan(&subject, &body)
		if err == sql.ErrNoRows {
			return message{Subject: "Error", Body: "Message not found"}, nil
		}
		web = true
	} else {
		web = false
	}

	weburl := webUrl(campaignId, recipientId)

	people := recipientParam(recipientId)
	people["UnsubscribeUrl"] = unsubscribeUrl(campaignId, recipientId)
	people["StatPng"] = statPngUrl(campaignId, recipientId)

	if !web {
		people["WebUrl"] = weburl

		// add statistic png
		if strings.Index(body, "{{.StatPng}}") == -1 {
			if strings.Index(body, "</body>") == -1 {
				body = body + "<img src='{{.StatPng}}' border='0px' width='10px' height='10px'/>"
			} else {
				body = strings.Replace(body, "</body>", "\n<img src='{{.StatPng}}' border='0px' width='10px' height='10px'/>\n</body>", -1)
			}
		}
	}

	// Replace links for statistic
	re := regexp.MustCompile(`href=["'](\bhttp:\/\/\b|\bhttps:\/\/\b)(.*?)["']`)
	body = re.ReplaceAllStringFunc(body, func(str string) string {
		// get only url
		s := strings.Replace(str, "'", "", -1)
		s = strings.Replace(s, `"`, "", -1)
		s = strings.Replace(s, "href=", "", 1)

		switch s {
		case "{{.WebUrl}}":
			if web {
				return `href=""`
			} else {
				return `href="` + weburl + `"`
			}
		case "{{.UnsubscribeUrl}}":
			return `href="` + people["UnsubscribeUrl"] + `"`
		default:
			// template parameter in url
			urlt := template.New("url")
			urlt, err = urlt.Parse(s)
			if err != nil {
				s = fmt.Sprintf("Error parse url params: %v", err)
			}
			u := bytes.NewBufferString("")
			urlt.Execute(u, people)
			s = u.String()

			return `href="` + statUrl(campaignId, recipientId, s) + `"`
		}
	})


	//replace static url to absolute
	body = strings.Replace(body, "\"/static/", "\"" + StatUrl + "/static/", -1)
	body = strings.Replace(body, "'/static/", "'" + StatUrl + "'/static/", -1)

	tmpl := template.New("mail")

	tmpl, err = tmpl.Parse(body)
	if err != nil {
		e := fmt.Sprintf("Error parse template: %v", err)
		return message{Subject: "Error", Body: e}, err
	}

	t := bytes.NewBufferString("")
	tmpl.Execute(t, people)
	return message{Subject: subject, Body: t.String()}, nil
}

func recipientParam(id string) map[string]string {
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

func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}