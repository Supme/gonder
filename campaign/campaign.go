package campaign

import (
	"database/sql"
	"fmt"
	"github.com/supme/gonder/models"
	"github.com/supme/smtpSender"
	"io"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"
)

type Campaign struct {
	ID              string
	FromEmail       string
	FromName        string
	DkimSelector    string
	DkimPrivateKey  []byte
	DkimUse         bool
	SendUnsubscribe bool
	ProfileID       string
	ResendDelay     int
	ResendCount     int
	Attachments     []string

	subjectTmpl *template.Template
	htmlTmpl    *template.Template
	stopSend chan struct{}
	stopResend chan struct{}
	finish chan struct{}
}

func GetCampaign(id string) (Campaign, error) {
	var subject, html string
	c := Campaign{ID: id}

	err := models.Db.QueryRow("SELECT t3.`email`,t3.`name`,t3.`dkim_selector`,t3.`dkim_key`,t3.`dkim_use`,t1.`subject`,t1.`body`,t2.`id`,t1.`send_unsubscribe`,t2.`resend_delay`,t2.`resend_count` FROM `campaign` t1 INNER JOIN `profile` t2 ON t2.`id`=t1.`profile_id` INNER JOIN `sender` t3 ON t3.`id`=t1.`sender_id` WHERE t1.`id`=?", id).Scan(
		&c.FromEmail,
		&c.FromName,
		&c.DkimSelector,
		&c.DkimPrivateKey,
		&c.DkimUse,
		&subject,
		&html,
		&c.ProfileID,
		&c.SendUnsubscribe,
		&c.ResendDelay,
		&c.ResendCount,
	)
	if err != nil {
		return c, err
	}

	c.subjectTmpl, err = template.New("Subject").Parse(subject)
	if err != nil {
		return c, fmt.Errorf("error parse campaign '%s' subject template: %s", c.ID, err)
	}

	html = replaceLinks(html)

	c.htmlTmpl, err = template.New("HTML").
		Funcs(template.FuncMap{
			"RedirectUrl": c.tmplFuncRedirectUrl,
		}).
		Parse(html)
	if err != nil {
		return c, fmt.Errorf("error parse campaign '%s' html template: %s", c.ID, err)
	}

	var attachments *sql.Rows
	attachments, err = models.Db.Query("SELECT `path` FROM attachment WHERE campaign_id=?", c.ID)
	if err != nil {
		return c, err
	}
	defer attachments.Close()

	c.Attachments = nil
	var location string
	for attachments.Next() {
		err = attachments.Scan(&location)
		if err != nil {
			return c, err
		}
		c.Attachments = append(c.Attachments, location)
	}

	return c, nil
}

func (campaign *Campaign) Send() {
	campaign.stopSend = make(chan struct{})
	campaign.stopResend = make(chan struct{})
	campaign.finish = make(chan struct{}, 1)

	pipe, err := models.EmailPool.Get(campaign.ProfileID)
	checkErr(err)

	camplog.Printf("Start campaign id %s.", campaign.ID)
	campaign.send(pipe)
	resend := campaign.HasResend()
	if resend > 0 && campaign.ResendCount > 0 {
		camplog.Printf("Done stream send campaign id %s but need %d resend.", campaign.ID, resend)
	}
	if campaign.HasResend() > 0 {
		for i := 1; i <= campaign.ResendCount; i++ {
			resend := campaign.HasResend()
			if resend == 0 {
				break
			}
			time.Sleep(time.Duration(campaign.ResendDelay) * time.Second)
			camplog.Printf("Start %s resend by %d email from campaign id %s ", models.Conv1st2nd(i), resend, campaign.ID)
			campaign.resend(pipe)
			camplog.Printf("Done %s resend campaign id %s", models.Conv1st2nd(i), campaign.ID)
		}
	}

	close(campaign.finish)
	camplog.Printf("Finish campaign id %s", campaign.ID)
}

func (campaign *Campaign) Stop() {
	campaign.stopSend <-struct{}{}
	campaign.stopResend <-struct{}{}
	<-campaign.finish
}

func (campaign *Campaign) send(pipe *smtpSender.Pipe) {
	query, err := models.Db.Query("SELECT `id` FROM recipient WHERE campaign_id=? AND removed=0 AND status IS NULL", campaign.ID)
	checkErr(err)
	defer query.Close()

	wg := &sync.WaitGroup{}
	for query.Next() {
		select {
		case <-campaign.stopSend:
			break

		default:
			var rID string
			err = query.Scan(&rID)
			checkErr(err)
			r, err := GetRecipient(rID)
			checkErr(err)

			if r.Unsubscribed() && !campaign.SendUnsubscribe {
				_, err := models.Db.Exec("UPDATE recipient SET status='Unsubscribe', date=NOW() WHERE id=?", r.ID)
				checkErr(err)
				camplog.Printf("Recipient id %s email %s is unsubscribed", r.ID, r.Email)
				continue
			}
			wg.Add(1)

			if !models.Config.RealSend {
				var res string
				wait := time.Duration(rand.Int()/10000000000) * time.Nanosecond
				time.Sleep(wait)
				if rand.Intn(2) == 0 {
					res = "421 Test send"
				} else {
					res = "Ok Test send"
				}
				_, err := models.Db.Exec("UPDATE recipient SET status=?, date=NOW() WHERE id=?", res, r.ID)
				checkErr(err)
				camplog.Printf("Campaign %s for recipient id %s email %s is %s send time %s", campaign.ID, r.ID, r.Email, res, wait.String())
				wg.Done()
				continue
			}

			bldr := new(smtpSender.Builder)
			bldr.SetFrom(campaign.FromName, campaign.FromEmail)
			bldr.SetTo(r.Name, r.Email)
			bldr.AddHeader(
				"List-Unsubscribe: " + models.EncodeUTM("unsubscribe", "mail", r.Params) + "\nPrecedence: bulk",
				"Message-ID: <"+strconv.FormatInt(time.Now().Unix(), 10)+campaign.ID+"."+r.ID+"@"+"gonder"+">", // ToDo hostname
				"X-Postmaster-Msgtype: campaign"+campaign.ID,
			)
			bldr.AddSubjectFunc(campaign.SubjectTemplFunc(r))
			bldr.AddHTMLFunc(campaign.HTMLTemplFunc(r))
			bldr.AddAttachment(campaign.Attachments...)
			email := bldr.Email(r.ID, func(result smtpSender.Result) {
				var res string
				if result.Err == nil {
					res = "Ok"
				} else {
					res = result.Err.Error()
				}
				_, err := models.Db.Exec("UPDATE recipient SET status=?, date=NOW() WHERE id=?", res, result.ID)
				checkErr(err)
				camplog.Printf("Campaign %s for recipient id %s email %s is %s send time %s", campaign.ID, r.ID, r.Email, res, result.Duration.String())
				wg.Done()
			})
			err = pipe.Send(email)
			if err != nil {
				break
			}
		}
	}
	wg.Wait()
}

const softBounceWhere = "LOWER(`status`) REGEXP '^((4[0-9]{2})|(dial tcp)|(read tcp)|(proxy)|(eof)).+'"

func (campaign *Campaign) HasResend() int {
	var count int
	err := models.Db.QueryRow("SELECT COUNT(`id`) FROM `recipient` WHERE `campaign_id`=? AND removed=0 AND "+softBounceWhere, campaign.ID).Scan(&count)
	checkErr(err)
	return count
}

func (campaign *Campaign) resend(pipe *smtpSender.Pipe) {
	query, err := models.Db.Query("SELECT `id` FROM recipient WHERE campaign_id=? AND removed=0 AND " + softBounceWhere, campaign.ID)
	checkErr(err)
	defer query.Close()

	oneEmail := make(chan struct{})
	for query.Next() {
		select {
		case <-campaign.stopResend:
			break

		default:
			var rID string
			err = query.Scan(&rID)
			checkErr(err)
			r, err := GetRecipient(rID)
			checkErr(err)

			if r.Unsubscribed() && !campaign.SendUnsubscribe {
				_, err := models.Db.Exec("UPDATE recipient SET status='Unsubscribe', date=NOW() WHERE id=?", r.ID)
				checkErr(err)
				camplog.Printf("Recipient id %s email %s is unsubscribed", r.ID, r.Email)
				continue
			}

			if !models.Config.RealSend {
				var res string
				wait := time.Duration(rand.Int()/10000000000) * time.Nanosecond
				time.Sleep(wait)
				if rand.Intn(2) == 0 {
					res = "421 Test send"
				} else {
					res = "Ok Test send"
				}
				_, err := models.Db.Exec("UPDATE recipient SET status=?, date=NOW() WHERE id=?", res, r.ID)
				checkErr(err)
				camplog.Printf("Campaign %s for recipient id %s email %s is %s send time %s", campaign.ID, r.ID, r.Email, res, wait.String())
				continue
			}

			oneEmail <- struct{}{}
			bldr := new(smtpSender.Builder)
			bldr.SetFrom(campaign.FromName, campaign.FromEmail)
			bldr.SetTo(r.Name, r.Email)
			bldr.AddHeader(
				"List-Unsubscribe: "+r.UnsubscribeEmailHeaderURL()+"\nPrecedence: bulk",
				"Message-ID: <"+strconv.FormatInt(time.Now().Unix(), 10)+campaign.ID+"."+r.ID+"@"+"gonder"+">", // ToDo hostname
				"X-Postmaster-Msgtype: campaign"+campaign.ID,
			)
			bldr.AddSubjectFunc(campaign.SubjectTemplFunc(r))
			bldr.AddHTMLFunc(campaign.HTMLTemplFunc(r))
			bldr.AddAttachment(campaign.Attachments...)
			email := bldr.Email(r.ID, func(result smtpSender.Result) {
				var res string
				if result.Err == nil {
					res = "Ok"
				} else {
					res = result.Err.Error()
				}
				_, err := models.Db.Exec("UPDATE recipient SET status=?, date=NOW() WHERE id=?", res, result.ID)
				checkErr(err)
				camplog.Printf("Campaign %s for recipient id %s email %s is %s send time %s", campaign.ID, r.ID, r.Email, res, result.Duration.String())
				<-oneEmail
			})
			err = pipe.Send(email)
			if err != nil {
				camplog.Println(err)
				break
			}
		}
	}
}

func (campaign *Campaign) SubjectTemplFunc(recipient Recipient) func(io.Writer) error {
	return func(w io.Writer) error {
		recipient.Params["RecipientId"] = recipient.ID
		recipient.Params["CampaignId"] = recipient.CampaignID
		recipient.Params["RecipientEmail"] = recipient.Email
		recipient.Params["RecipientName"] = recipient.Name
		return campaign.subjectTmpl.Execute(w, recipient.Params)
	}
}

func (campaign *Campaign) HTMLTemplFunc(recipient Recipient) func(io.Writer) error {
	return func(w io.Writer) error {
		recipient.Params["RecipientId"] = recipient.ID
		recipient.Params["CampaignId"] = recipient.CampaignID
		recipient.Params["RecipientEmail"] = recipient.Email
		recipient.Params["RecipientName"] = recipient.Name
		recipient.Params["StatPng"] = models.EncodeUTM("open", "", recipient.Params)
		recipient.Params["UnsubscribeUrl"] = models.EncodeUTM("unsubscribe", "web", recipient.Params)
		recipient.Params["WebUrl"] = models.EncodeUTM("web", "", recipient.Params)
		return campaign.htmlTmpl.Execute(w, recipient.Params)
	}
}

func (campaign *Campaign) WebHTMLTemplFunc(recipient Recipient) func(io.Writer) error {
	return func(w io.Writer) error {
		recipient.Params["RecipientId"] = recipient.ID
		recipient.Params["CampaignId"] = recipient.CampaignID
		recipient.Params["RecipientEmail"] = recipient.Email
		recipient.Params["RecipientName"] = recipient.Name
		recipient.Params["StatPng"] = models.EncodeUTM("open", "", recipient.Params)
		recipient.Params["UnsubscribeUrl"] = models.EncodeUTM("unsubscribe", "web", recipient.Params)
		return campaign.htmlTmpl.Execute(w, recipient.Params)
	}
}

// Regexp for replace all http and https in message to link on utm service
var reReplaceLink = regexp.MustCompile(`[hH][rR][eE][fF]\s*?=\s*?["']\s*?(\[.*?\])?\s*?(\b[hH][tT]{2}[pP][sS]?\b:\/\/\b)(.*?)["']`)

func replaceLinks(tmpl string) string {
	if strings.Index(tmpl, "{{.StatPng}}") == -1 {
		if strings.Index(tmpl, "</body>") == -1 {
			tmpl = tmpl + "<img src='{{.StatPng}}' border='0px' width='10px' height='10px'/>"
		} else {
			tmpl = strings.Replace(tmpl, "</body>", "\n<img src='{{.StatPng}}' border='0px' width='10px' height='10px'/>\n</body>", -1)
		}
	}
	return reReplaceLink.ReplaceAllStringFunc(tmpl, func(str string) string {
		// get only url
		s := strings.Replace(str, `'`, "", -1)
		s = strings.Replace(s, `"`, "", -1)
		s = strings.Replace(s, "href=", "", 1)
		if str != "{{.WebUrl}}" || str != "{{.UnsubscribeUrl}}" || strings.HasPrefix(str, "{{RedirectUrl") {
			return `href="{{RedirectUrl . "` + s + `"}}"`
		}
		return `href="` + s + `"`
	})

}

func (campaign *Campaign) tmplFuncRedirectUrl(p map[string]interface{}, u string) string {
	return models.EncodeUTM("redirect", u, p)
}

