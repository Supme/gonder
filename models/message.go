package models

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"
)

type Message struct {
	RecipientID      string // ToDo int?
	RecipientEmail   string
	RecipientName    string
	RecipientParam   map[string]string
	CampaignID       string // ToDo int?
	CampaignSubject  string
	CampaignTemplate string
}

func (m *Message) New(recipientID string) error {
	m.RecipientID = recipientID
	err := Db.QueryRow("SELECT `campaign_id`,`email`,`name` FROM `recipient` WHERE `id`=?", m.RecipientID).Scan(&m.CampaignID, &m.RecipientEmail, &m.RecipientName)
	if err == sql.ErrNoRows {
		return errors.New("The recipient does not exist")
	}
	return nil
}

// Unsubscribe recipient from group
func (m *Message) Unsubscribe(extra map[string]string) error {
	r, err := Db.Exec("INSERT INTO unsubscribe (`group_id`, `campaign_id`, `email`) VALUE ((SELECT group_id FROM campaign WHERE id=?), ?, ?)", m.CampaignID, m.CampaignID, m.RecipientEmail)
	if err != nil {
		log.Print(err)
	}
	id, e := r.LastInsertId()
	if e != nil {
		log.Print(e)
	}
	for name, value := range extra {
		_, err = Db.Exec("INSERT INTO unsubscribe_extra (`unsubscribe_id`, `name`, `value`) VALUE (?, ?, ?)", id, name, value)
		if err != nil {
			log.Print(err)
		}
	}
	return err
}

func (m *Message) UnsubscribeTemplateDir() (name string) {
	err := Db.QueryRow("SELECT `group`.`template` FROM `campaign` INNER JOIN `group` ON `campaign`.`group_id`=`group`.`id` WHERE `group`.`template` IS NOT NULL AND `campaign`.`id`=?", m.CampaignID).Scan(&name)
	if err != nil {
		log.Print(err)
	}
	if name == "" {
		name = "default"
	} else {
		if _, err := os.Stat(FromRootDir("templates/" + name + "/accept.html")); err != nil {
			name = "default"
		}
		if _, err := os.Stat(FromRootDir("templates/" + name + "/success.html")); err != nil {
			name = "default"
		}
	}
	name = FromRootDir("templates/" + name)
	return
}

func (m *Message) makeLink(cmd, data string) string {
	return Config.URL + "/" + cmd + "/" + encodeUTM(m, data)
}

// Make unsubscribe utm link for web version
func (m *Message) UnsubscribeWebLink() string {
	return m.makeLink("unsubscribe", "web")
}

// Make unsubscribe utm link for mail version
func (m *Message) UnsubscribeMailLink() string {
	return m.makeLink("unsubscribe", "mail")
}

// Make utm link from real link
func (m *Message) RedirectLink(url string) string {
	return m.makeLink("redirect", url)
}

//// Make utm link to web version
func (m *Message) WebLink() string {
	return m.makeLink("web", "")
}

// Make utm link for check open mail
func (m *Message) StatPngLink() string {
	return m.makeLink("open", "")
}

// Regexp for replace all http and https in message to link on utm service
var reReplaceLink = regexp.MustCompile(`[hH][rR][eE][fF]\s*?=\s*?["']\s*?(\[.*?\])?\s*?(\b[hH][tT]{2}[pP][sS]?\b:\/\/\b)(.*?)["']`)

// Render message body
func (m *Message) RenderMessage() (string, error) {

	var (
		err error
		web bool
	)

	if m.CampaignSubject == "" && m.CampaignTemplate == "" {
		err := Db.QueryRow("SELECT `subject`,`body` FROM `campaign` WHERE `id`=?", m.CampaignID).Scan(&m.CampaignSubject, &m.CampaignTemplate)
		if err == sql.ErrNoRows {
			return "", err
		}
		web = true
	} else {
		web = false
	}

	m.RecipientParam = map[string]string{}
	var paramKey, paramValue string
	q, err := Db.Query("SELECT `key`, `value` FROM parameter WHERE recipient_id=?", m.RecipientID)
	if err != nil {
		return "", err
	}
	defer q.Close()
	for q.Next() {
		err = q.Scan(&paramKey, &paramValue)
		if err != nil {
			return "", err
		}
		m.RecipientParam[paramKey] = paramValue
	}

	m.RecipientParam["UnsubscribeUrl"] = m.UnsubscribeWebLink()
	m.RecipientParam["StatPng"] = m.StatPngLink()
	m.RecipientParam["RecipientEmail"] = m.RecipientEmail
	m.RecipientParam["RecipientName"] = m.RecipientName
	m.RecipientParam["CampaignId"] = m.CampaignID

	// render subject
	subj := template.New("subject" + m.RecipientID)
	subj, err = subj.Parse(m.CampaignSubject)
	if err != nil {
		e := fmt.Sprintf("Error parse subject: %v", err)
		return e, err
	}
	tSubj := bytes.NewBufferString("")
	subj.Execute(tSubj, m.RecipientParam)
	m.CampaignSubject = tSubj.String()

	if !web {
		m.RecipientParam["WebUrl"] = m.WebLink()

		// add statistic png
		if strings.Index(m.CampaignTemplate, "{{.StatPng}}") == -1 {
			if strings.Index(m.CampaignTemplate, "</body>") == -1 {
				m.CampaignTemplate = m.CampaignTemplate + "<img src='{{.StatPng}}' border='0px' width='10px' height='10px'/>"
			} else {
				m.CampaignTemplate = strings.Replace(m.CampaignTemplate, "</body>", "\n<img src='{{.StatPng}}' border='0px' width='10px' height='10px'/>\n</body>", -1)
			}
		}
	}

	// Replace links for statistic
	m.CampaignTemplate = reReplaceLink.ReplaceAllStringFunc(m.CampaignTemplate, func(str string) string {
		// get only url
		s := strings.Replace(str, `'`, "", -1)
		s = strings.Replace(s, `"`, "", -1)
		s = strings.Replace(s, "href=", "", 1)

		switch s {
		case "{{.WebUrl}}":
			return `href="` + m.RecipientParam["WebUrl"] + `"`
		case "{{.UnsubscribeUrl}}":
			return `href="` + m.RecipientParam["UnsubscribeUrl"] + `"`
		default:
			// template parameter in url
			urlt := template.New("url" + m.RecipientID)
			urlt, err = urlt.Parse(s)
			if err != nil {
				s = fmt.Sprintf("Error parse url params: %v", err)
			}
			u := bytes.NewBufferString("")
			urlt.Execute(u, m.RecipientParam)
			s = u.String()

			return `href="` + m.RedirectLink(s) + `"`
		}
	})

	//replace static url to absolute
	m.CampaignTemplate = strings.Replace(m.CampaignTemplate, "\"/files/", "\""+Config.URL+"/files/", -1)
	m.CampaignTemplate = strings.Replace(m.CampaignTemplate, "'/files/", "'"+Config.URL+"'/files/", -1)

	// render template
	tmpl := template.New("mail" + m.RecipientID)
	tmpl, err = tmpl.Parse(m.CampaignTemplate)
	if err != nil {
		e := fmt.Sprintf("Error parse template: %v", err)
		return e, err
	}
	tTempl := bytes.NewBufferString("")
	tmpl.Execute(tTempl, m.RecipientParam)
	return tTempl.String(), nil
}
