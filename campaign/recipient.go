package campaign

import (
	"github.com/supme/gonder/models"
	"time"
	"bytes"
	"strconv"
	"errors"
	"math/rand"
)

type recipient struct {
	id, to_email, to_name	string
}

// Check recipient for unsubscribe
func (r *recipient) unsubscribe(campaignId string) bool {
	var unsubscribeCount int
	models.Db.QueryRow("SELECT COUNT(*) FROM `unsubscribe` t1 INNER JOIN `campaign` t2 ON t1.group_id = t2.group_id WHERE t2.id = ? AND t1.email = ?", campaignId, r.to_email).Scan(&unsubscribeCount)
	if unsubscribeCount == 0 {
		return false
	}
	return true
}

// Send mail for recipient this campaign data
func (r recipient) send(c *campaign, iface, host string) string {
	start := time.Now()

	data := new(MailData)
	data.Iface, data.Host = iface, host
	data.From_email = c.from_email
	data.From_name = c.from_name
	data.Attachments = c.attachments
	data.To_email = r.to_email
	data.To_name = r.to_name

	message := new(models.Message)
	message.CampaignId = c.id
	message.RecipientId = r.id
	message.CampaignSubject = c.subject
	message.CampaignTemplate = c.body
	message.RecipientEmail = r.to_email
	message.RecipientName = r.to_name

	var rs string
	m, e := message.RenderMessage()
	if e == nil {
		data.Subject = message.CampaignSubject
		data.Html = m
		var extraHeader bytes.Buffer
		extraHeader.WriteString("List-Unsubscribe: " + message.UnsubscribeMailLink() + "\r\nPrecedence: bulk\r\n")
		extraHeader.WriteString("Message-ID: <" + strconv.FormatInt(time.Now().Unix(), 10) + c.id + "." + r.id +"@" + data.Host + ">" + "\r\n")
		extraHeader.WriteString("X-Postmaster-Msgtype: campaign" + c.id + "\r\n")
		data.Extra_header = extraHeader.String()

		var res error
		if models.Config.RealSend {
			// Send mail
			res = data.Send()
		} else {
			wait := time.Duration(rand.Int()/10000000000) * time.Nanosecond
			time.Sleep(wait)
			res = errors.New("Test send")
		}

		if res == nil {
			rs = "Ok"
		} else {
			rs = res.Error()
		}
	} else {
		rs = e.Error()
	}

	camplog.Printf("Campaign %s for recipient id %s email %s is %s send time %s", c.id, r.id, data.To_email, rs, time.Since(start))
	return rs
}
