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

type campaign struct {
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
	Stop chan struct{}
	Finish chan struct{}
}

func getCampaign(id string) (campaign, error) {
	var subject, html string
	c := campaign{ID: id}
	c.Stop = make(chan struct{})
	c.Finish = make(chan struct{})

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

	// ToDo QR code
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

func (c *campaign) send() {
	pipe, err := models.EmailPool.Get(c.ProfileID)
	checkErr(err)

	camplog.Printf("Start campaign id %s.", c.ID)
	c.streamSend(pipe)
	select {
	case <-c.Stop:
		camplog.Printf("Campaign %s stoped", c.ID)
		goto End
	default:
		resend := c.HasResend()
		if resend > 0 && c.ResendCount > 0 {
			camplog.Printf("Done stream send campaign id %s but need %d resend.", c.ID, resend)
		}
		if resend > 0 {
			for i := 1; i <= c.ResendCount; i++ {
				select {
				case <-c.Stop:
					goto Done
				default:
					resend := c.HasResend()
					if resend == 0 {
						goto Finish
					}

					timer := time.NewTimer(time.Duration(c.ResendDelay) * time.Second)
					select {
					case <-c.Stop:
						goto Done
					case <-timer.C:
						timer.Stop()
					}

					camplog.Printf("Start %s resend by %d email from campaign id %s ", models.Conv1st2nd(i), resend, c.ID)
					c.resend(pipe)
				}

				select {
				case <-c.Stop:
					goto Done
				default:
					camplog.Printf("Done %s resend campaign id %s", models.Conv1st2nd(i), c.ID)
				}
			}
		}
		Finish:
			camplog.Printf("Finish campaign id %s", c.ID)
	}
	Done:
		select {
		case <-c.Stop:
			camplog.Printf("Campaign %s stoped", c.ID)
		default:
			close(c.Stop)
		}
	End:
		close(c.Finish)
}

func (c *campaign) streamSend(pipe *smtpSender.Pipe) {
	query, err := models.Db.Query("SELECT `id` FROM recipient WHERE campaign_id=? AND removed=0 AND status IS NULL", c.ID)
	checkErr(err)
	defer query.Close()

	wg := &sync.WaitGroup{}
	for query.Next() {
		select {
		case <-c.Stop:
			camplog.Printf("Stop signal for campaign %s recieved", c.ID)
			goto Done
		default:
			var rID string
			err = query.Scan(&rID)
			checkErr(err)
			r, err := getRecipient(rID)
			checkErr(err)

			if r.unsubscribed() && !c.SendUnsubscribe {
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
				camplog.Printf("Campaign %s for recipient id %s email %s is %s send time %s", c.ID, r.ID, r.Email, res, wait.String())
				wg.Done()
				continue
			}

			bldr := new(smtpSender.Builder)
			bldr.SetFrom(c.FromName, c.FromEmail)
			bldr.SetTo(r.Name, r.Email)
			bldr.AddHeader(
				"List-Unsubscribe: " + models.EncodeUTM("unsubscribe", "mail", r.Params) + "\nPrecedence: bulk",
				"Message-ID: <"+strconv.FormatInt(time.Now().Unix(), 10)+c.ID+"."+r.ID+"@"+"gonder"+">", // ToDo hostname
				"X-Postmaster-Msgtype: campaign"+c.ID,
			)
			bldr.AddSubjectFunc(c.subjectTemplFunc(r))
			bldr.AddHTMLFunc(c.htmlTemplFunc(r))
			bldr.AddAttachment(c.Attachments...)
			email := bldr.Email(r.ID, func(result smtpSender.Result) {
				var res string
				if result.Err == nil {
					res = "Ok"
				} else {
					res = result.Err.Error()
				}
				_, err := models.Db.Exec("UPDATE recipient SET status=?, date=NOW() WHERE id=?", res, result.ID)
				checkErr(err)
				camplog.Printf("Campaign %s for recipient id %s email %s is %s send time %s", c.ID, r.ID, r.Email, res, result.Duration.String())
				wg.Done()
			})
			err = pipe.Send(email)
			if err != nil {
				break
			}
		}
	}

	Done:
		wg.Wait()
}

const softBounceWhere = "LOWER(`status`) REGEXP '^((4[0-9]{2})|(dial tcp)|(read tcp)|(proxy)|(eof)).+'"

func (c *campaign) HasResend() int {
	var count int
	err := models.Db.QueryRow("SELECT COUNT(`id`) FROM `recipient` WHERE `campaign_id`=? AND removed=0 AND "+softBounceWhere, c.ID).Scan(&count)
	checkErr(err)
	return count
}

func (c *campaign) resend(pipe *smtpSender.Pipe) {
	query, err := models.Db.Query("SELECT `id` FROM recipient WHERE campaign_id=? AND removed=0 AND " + softBounceWhere, c.ID)
	checkErr(err)
	defer query.Close()

	oneEmail := make(chan struct{})
	for query.Next() {
		select {
		case <-c.Stop:
			camplog.Printf("Stop signal for campaign %s recieved", c.ID)
			return
		default:
			var rID string
			err = query.Scan(&rID)
			checkErr(err)
			r, err := getRecipient(rID)
			checkErr(err)

			if r.unsubscribed() && !c.SendUnsubscribe {
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
				camplog.Printf("Campaign %s for recipient id %s email %s is %s send time %s", c.ID, r.ID, r.Email, res, wait.String())
				continue
			}

			oneEmail <- struct{}{}
			bldr := new(smtpSender.Builder)
			bldr.SetFrom(c.FromName, c.FromEmail)
			bldr.SetTo(r.Name, r.Email)
			bldr.AddHeader(
				"List-Unsubscribe: "+r.unsubscribeEmailHeaderURL()+"\nPrecedence: bulk",
				"Message-ID: <"+strconv.FormatInt(time.Now().Unix(), 10)+c.ID+"."+r.ID+"@"+"gonder"+">", // ToDo hostname
				"X-Postmaster-Msgtype: campaign"+c.ID,
			)
			bldr.AddSubjectFunc(c.subjectTemplFunc(r))
			bldr.AddHTMLFunc(c.htmlTemplFunc(r))
			bldr.AddAttachment(c.Attachments...)
			email := bldr.Email(r.ID, func(result smtpSender.Result) {
				var res string
				if result.Err == nil {
					res = "Ok"
				} else {
					res = result.Err.Error()
				}
				_, err := models.Db.Exec("UPDATE recipient SET status=?, date=NOW() WHERE id=?", res, result.ID)
				checkErr(err)
				camplog.Printf("Campaign %s for recipient id %s email %s is %s send time %s", c.ID, r.ID, r.Email, res, result.Duration.String())
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

func (c *campaign) subjectTemplFunc(r recipient) func(io.Writer) error {
	return func(w io.Writer) error {
		r.Params["RecipientId"] = r.ID
		r.Params["CampaignId"] = r.CampaignID
		r.Params["RecipientEmail"] = r.Email
		r.Params["RecipientName"] = r.Name
		return c.subjectTmpl.Execute(w, r.Params)
	}
}

func (c *campaign) htmlTemplFunc(r recipient) func(io.Writer) error {
	return func(w io.Writer) error {
		r.Params["RecipientId"] = r.ID
		r.Params["CampaignId"] = r.CampaignID
		r.Params["RecipientEmail"] = r.Email
		r.Params["RecipientName"] = r.Name
		r.Params["StatPng"] = models.EncodeUTM("open", "", r.Params)
		r.Params["UnsubscribeUrl"] = models.EncodeUTM("unsubscribe", "web", r.Params)
		r.Params["WebUrl"] = models.EncodeUTM("web", "", r.Params)
		return c.htmlTmpl.Execute(w, r.Params)
	}
}

func (c *campaign) webHTMLTemplFunc(r recipient) func(io.Writer) error {
	return func(w io.Writer) error {
		r.Params["RecipientId"] = r.ID
		r.Params["CampaignId"] = r.CampaignID
		r.Params["RecipientEmail"] = r.Email
		r.Params["RecipientName"] = r.Name
		r.Params["StatPng"] = models.EncodeUTM("open", "", r.Params)
		r.Params["UnsubscribeUrl"] = models.EncodeUTM("unsubscribe", "web", r.Params)
		return c.htmlTmpl.Execute(w, r.Params)
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

func (c *campaign) tmplFuncRedirectUrl(p map[string]interface{}, u string) string {
	return models.EncodeUTM("redirect", u, p)
}

