package campaign

import (
	"github.com/supme/directEmail"
	"github.com/supme/gonder/models"
	"math/rand"
	"strconv"
	"time"
)

type recipient struct {
	id, toEmail, toName string
}

// Check recipient for unsubscribe
func (r *recipient) checkUnsubscribe(campaignID string) bool {
	var unsubscribeCount int
	models.Db.QueryRow("SELECT COUNT(*) FROM `unsubscribe` t1 INNER JOIN `campaign` t2 ON t1.group_id = t2.group_id WHERE t2.id = ? AND t1.email = ?", campaignID, r.toEmail).Scan(&unsubscribeCount)
	if unsubscribeCount == 0 {
		return false
	}
	return true
}

// Send mail for recipient this campaign data
func (r *recipient) send(c *campaign, iface, host *string) (res string) {
	start := time.Now()

	email := directEmail.New()
	// ToDo to config for NAT
	//email.MapIp = map[string]string {
	//	"192.168.0.10": "31.33.34.35",
	//}
	email.Ip, email.Host = *iface, *host
	email.FromEmail = c.fromEmail
	email.FromName = c.fromName
	email.Attachment(c.attachments...)
	email.ToEmail = r.toEmail
	email.ToName = r.toName

	message := new(models.Message)
	message.CampaignID = c.id
	message.RecipientID = r.id
	message.CampaignSubject = c.subject
	message.CampaignTemplate = c.body
	message.RecipientEmail = r.toEmail
	message.RecipientName = r.toName

	m, err := message.RenderMessage()
	if err != nil {
		res = err.Error()
		return
	}
	email.Subject = message.CampaignSubject
	email.TextHtml(m)
	email.Header(
		"List-Unsubscribe: "+message.UnsubscribeMailLink()+"\nPrecedence: bulk",
		"Message-ID: <"+strconv.FormatInt(time.Now().Unix(), 10)+c.id+"."+r.id+"@"+*host+">",
		"X-Postmaster-Msgtype: campaign"+c.id,
	)

	var dkimSelector string
	if c.dkimUse {
		dkimSelector = c.dkimSelector
	}
	err = email.RenderWithDkim(dkimSelector, c.dkimPrivateKey) // ToDo DKIM to database sender
	if err != nil {
		res = err.Error()
		return
	}

	if models.Config.RealSend {
		// Send mail
		err = email.Send()
		if err != nil {
			res = err.Error()
			logResult(c.id, r.id, r.toEmail, res, time.Since(start))
			return
		}
		res = "Ok"
		logResult(c.id, r.id, r.toEmail, res, time.Since(start))
	} else {
		wait := time.Duration(rand.Int()/10000000000) * time.Nanosecond
		time.Sleep(wait)
		res = "Test send"
		logResult(c.id, r.id, r.toEmail, res, time.Since(start))
	}

	return
}

func logResult(params ...interface{}) {
	camplog.Printf("Campaign %s for recipient id %s email %s is %s send time %s", params...)

}
