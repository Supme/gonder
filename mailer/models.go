package mailer

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"text/template"
	"regexp"
	"strings"
	"fmt"
	"database/sql"
)

func statOpened(campaignId string, recipientId string, userAgent string) {
	Db.Exec("INSERT INTO jumping (campaign_id, recipient_id, url) VALUES (?, ?, ?)", campaignId, recipientId, "open_trace")
	Db.Exec("UPDATE `recipient` SET `client_agent`= ? WHERE `id`=? AND `client_agent` IS NULL", userAgent, recipientId)
}

func statJump(campaignId string, recipientId string, url string, userAgent string) {
	Db.Exec("INSERT INTO jumping (campaign_id, recipient_id, url) VALUES (?, ?, ?)", campaignId, recipientId, url)
	Db.Exec("UPDATE `recipient` SET `web_agent`= ? WHERE `id`=? AND `web_agent` IS NULL", userAgent, recipientId)
}

func statWebVersion(campaignId string, recipientId string, userAgent string)  {
	Db.Exec("INSERT INTO jumping (campaign_id, recipient_id, url) VALUES (?, ?, ?)", campaignId, recipientId, "web_version")
	Db.Exec("UPDATE `recipient` SET `web_agent`= ? WHERE `id`=? AND `web_agent` IS NULL", userAgent, recipientId)
}

func statSend(id, result string) {
	Db.Exec("UPDATE recipient SET status=?, date=NOW() WHERE id=?", result, id)
}

func postUnsubscribe(campaignId string, recipientId string) {
	Db.Exec("INSERT INTO unsubscribe (`group_id`, campaign_id, `email`) VALUE ((SELECT group_id FROM campaign WHERE id=?), ?, (SELECT email FROM recipient WHERE id=?))", campaignId, campaignId, recipientId)
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

func getWebMessage(campaignId string, recipientId string) string {
	subject := ""
	body := ""
	m, err := getMessage(campaignId, recipientId, subject, body)
	checkErr(err)
	return string(m.Body)
}

func getMailMessage(campaignId, recipientId, subject, body string ) (message, error) {
	return getMessage(campaignId, recipientId, subject, body)
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

	weburl := getWebUrl(campaignId, recipientId)

	people := getRecipientParam(recipientId)
	people["UnsubscribeUrl"] = getUnsubscribeUrl(campaignId, recipientId)
	people["StatPng"] = getStatPngUrl(campaignId, recipientId)

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

			return `href="` + getStatUrl(campaignId, recipientId, s) + `"`
		}
	})


	//replace static url to absolute
	body = strings.Replace(body, "\"/static/", "\"" + HostName + "/static/", -1)
	body = strings.Replace(body, "'/static/", "'" + HostName + "'/static/", -1)

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