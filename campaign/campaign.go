package campaign

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/supme/gonder/models"
	"github.com/supme/smtpSender"
	"io"
	"math/rand"
	"net/url"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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

	subjectTmplFunc *template.Template
	htmlTmpl        string
	htmlTmplFunc    *template.Template
	Stop            chan struct{}
	Finish          chan struct{}
}

func getCampaign(id string) (campaign, error) {
	var (
		subject string
		//html    htmlString
	)
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
		&c.htmlTmpl,
		&c.ProfileID,
		&c.SendUnsubscribe,
		&c.ResendDelay,
		&c.ResendCount,
	)
	if err != nil {
		return c, err
	}

	c.subjectTmplFunc, err = template.New("Subject").Parse(subject)
	if err != nil {
		return c, fmt.Errorf("error parse campaign '%s' subject template: %s", c.ID, err)
	}

	prepareHTMLTemplate(&c.htmlTmpl)
	//fmt.Println(c.htmlTmpl)
	c.htmlTmplFunc = template.New("HTML")

	var attachments *sql.Rows
	attachments, err = models.Db.Query("SELECT `path` FROM attachment WHERE campaign_id=?", c.ID)
	if err != nil {
		return c, err
	}
	defer attachments.Close()

	c.Attachments = nil
	for attachments.Next() {
		var location string
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

func (c *campaign) prepareQueryUpdateRecipientStatus() (*sql.Stmt, error) {
	return models.Db.Prepare("UPDATE recipient SET status=?, date=NOW() WHERE id=?")
}

func (c *campaign) streamSend(pipe *smtpSender.Pipe) {
	query, err := models.Db.Query("SELECT `id` FROM recipient WHERE campaign_id=? AND removed=0 AND status IS NULL", c.ID)
	checkErr(err)
	defer query.Close()

	updateRecipientStatus, err := c.prepareQueryUpdateRecipientStatus()
	checkErr(err)

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
			r, err := GetRecipient(rID)
			checkErr(err)

			if r.unsubscribed() && !c.SendUnsubscribe {
				_, err := updateRecipientStatus.Exec(models.StatusUnsubscribe, r.ID)
				checkErr(err)
				camplog.Printf("Campaign %s recipient id %s email %s is unsubscribed", c.ID, r.ID, r.Email)
				continue
			}

			wg.Add(1)
			_, err = updateRecipientStatus.Exec(models.StatusSending, r.ID)
			checkErr(err)

			if !models.Config.RealSend {
				var res string
				wait := time.Duration(rand.Int()/10000000000) * time.Nanosecond
				time.Sleep(wait)
				if rand.Intn(2) == 0 {
					res = "421 Test send"
				} else {
					res = "Ok Test send"
				}
				_, err := updateRecipientStatus.Exec(res, r.ID)
				checkErr(err)
				camplog.Printf("Campaign %s for recipient id %s email %s is %s send time %s", c.ID, r.ID, r.Email, res, wait.String())
				wg.Done()
				continue
			}

			bldr := new(smtpSender.Builder)
			bldr.SetFrom(c.FromName, c.FromEmail)
			bldr.SetTo(r.Name, r.Email)
			bldr.AddHeader(
				"List-Unsubscribe: "+models.EncodeUTM("unsubscribe", "mail", r.Params)+"\nPrecedence: bulk",
				"Message-ID: <"+strconv.FormatInt(time.Now().Unix(), 10)+c.ID+"."+r.ID+"@"+"gonder"+">", // ToDo hostname
				"X-Postmaster-Msgtype: campaign"+c.ID,
			)
			bldr.AddSubjectFunc(c.subjectTemplFunc(r))
			bldr.AddHTMLFunc(c.htmlTemplFunc(r, false, false))
			bldr.AddAttachment(c.Attachments...)
			email := bldr.Email(r.ID, func(result smtpSender.Result) {
				var res string
				if result.Err == nil {
					res = "Ok"
				} else {
					res = result.Err.Error()
				}
				_, err := updateRecipientStatus.Exec(res, result.ID)
				checkErr(err)
				camplog.Printf("Campaign %s for recipient id %s email %s is %s send time %s", c.ID, r.ID, r.Email, res, result.Duration.String())
				wg.Done()
			})
			if !models.Config.RealSend {
				err = fakeSend(email)
			} else {
				err = pipe.Send(email)
			}
			if err != nil {
				break
			}
		}
	}

Done:
	wg.Wait()
}

const softBounceWhere = "LOWER(`status`) REGEXP '^((4[0-9]{2})|(dial tcp)|(read tcp)|(proxy)|(eof)).+'"

func (c *campaign) resend(pipe *smtpSender.Pipe) {
	query, err := models.Db.Query("SELECT `id` FROM recipient WHERE campaign_id=? AND removed=0 AND "+softBounceWhere, c.ID)
	checkErr(err)
	defer query.Close()

	updateRecipientStatus, err := c.prepareQueryUpdateRecipientStatus()
	checkErr(err)

	wg := &sync.WaitGroup{}
	for query.Next() {
		select {
		case <-c.Stop:
			camplog.Printf("Stop signal for campaign %s recieved", c.ID)
			return
		default:
			var rID string
			err = query.Scan(&rID)
			checkErr(err)
			r, err := GetRecipient(rID)
			checkErr(err)

			if r.unsubscribed() && !c.SendUnsubscribe {
				_, err := updateRecipientStatus.Exec(models.StatusUnsubscribe, r.ID)
				checkErr(err)
				camplog.Printf("Recipient id %s email %s is unsubscribed", r.ID, r.Email)
				continue
			}

			wg.Add(1)
			_, err = updateRecipientStatus.Exec(models.StatusSending, r.ID)
			checkErr(err)

			bldr := new(smtpSender.Builder)
			bldr.SetFrom(c.FromName, c.FromEmail)
			bldr.SetTo(r.Name, r.Email)
			bldr.AddHeader(
				"List-Unsubscribe: "+r.unsubscribeEmailHeaderURL()+"\nPrecedence: bulk",
				"Message-ID: <"+strconv.FormatInt(time.Now().Unix(), 10)+c.ID+"."+r.ID+"@"+"gonder"+">", // ToDo hostname
				"X-Postmaster-Msgtype: campaign"+c.ID,
			)
			bldr.AddSubjectFunc(c.subjectTemplFunc(r))
			bldr.AddHTMLFunc(c.htmlTemplFunc(r, false, false))
			bldr.AddAttachment(c.Attachments...)
			email := bldr.Email(r.ID, func(result smtpSender.Result) {
				var res string
				if result.Err == nil {
					res = "Ok"
				} else {
					res = result.Err.Error()
				}
				_, err := updateRecipientStatus.Exec(res, result.ID)
				checkErr(err)
				camplog.Printf("Campaign %s for recipient id %s email %s is %s send time %s", c.ID, r.ID, r.Email, res, result.Duration.String())
				wg.Done()
			})
			if !models.Config.RealSend {
				err = fakeSend(email)
			} else {
				err = pipe.Send(email)
			}
			if err != nil {
				camplog.Println(err)
				break
			}
		}
		wg.Wait()
	}
}

var fakeStream int64

func fakeSend(email smtpSender.Email) error {
	for atomic.LoadInt64(&fakeStream) > 5 {
		runtime.Gosched()
	}
	atomic.AddInt64(&fakeStream, 1)
	go func() {
		var res error
		wait := time.Duration(rand.Int()/10000000000) * time.Nanosecond
		time.Sleep(wait)
		if rand.Intn(4) == 0 {
			res = errors.New("421 Test send")
		} else {
			res = errors.New("Ok Test send")
		}
		atomic.AddInt64(&fakeStream, -1)
		email.ResultFunc(smtpSender.Result{
			ID:       email.ID,
			Duration: wait,
			Err:      res,
		})
	}()
	return nil
}

func (c *campaign) HasResend() int {
	var count int
	err := models.Db.QueryRow("SELECT COUNT(`id`) FROM `recipient` WHERE `campaign_id`=? AND removed=0 AND "+softBounceWhere, c.ID).Scan(&count)
	checkErr(err)
	return count
}

func (c *campaign) subjectTemplFunc(r Recipient) func(io.Writer) error {
	return func(w io.Writer) error {
		return c.subjectTmplFunc.Execute(w, r.Params)
	}
}

func (c *campaign) htmlTemplFunc(r Recipient, web bool, preview bool) func(io.Writer) error {
	return func(w io.Writer) error {
		if preview {
			if web {
				r.Params["WebUrl"] = nil
			} else {
				r.Params["WebUrl"] = "/preview?id=" + r.ID + "&type=web"
			}
			r.Params["StatPng"] = "data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7"
			r.Params["UnsubscribeUrl"] = "/unsubscribe?campaignId=" + r.CampaignID
			c.htmlTmplFunc.Funcs(template.FuncMap{
				"RedirectUrl": func(p map[string]interface{}, u string) string {
					url := regexp.MustCompile(`\s*?(\[.*?\])\s*?`).Split(u, 2)
					return strings.TrimSpace(url[len(url)-1])
				},
				// ToDo more functions (example QRcode generator)
			}).Parse(c.htmlTmpl)
		} else {
			if web {
				r.Params["WebUrl"] = nil
			} else {
				r.Params["WebUrl"] = models.EncodeUTM("web", "", r.Params)
			}
			c.htmlTmplFunc.Funcs(template.FuncMap{
				"RedirectUrl": func(p map[string]interface{}, u string) string { return models.EncodeUTM("redirect", u, p) },
				// ToDo more functions (example QRcode generator)
			}).Parse(c.htmlTmpl)
		}

		return c.htmlTmplFunc.Execute(w, r.Params)
	}
}

var (
	reReplaceLink         = regexp.MustCompile(`(\s[hH][rR][eE][fF]\s*?=\s*?)["']\s*?(\[.*?\])?\s*?(\b[hH][tT]{2}[pP][sS]?\b:\/\/\b)(.*?)["']`)
	reReplaceRelativeSrc  = regexp.MustCompile(`(\s[sS][rR][cC]\s*?=\s*?)(["'])\s*?(.?\/?[files\/].*?)(["'])`)
	reReplaceRelativeHref = regexp.MustCompile(`(\s[hH][rR][eE][fF]\s*?=\s*?)(["'])\s*?(.?\/?[files\/].*?)(["'])`)
)

func prepareHTMLTemplate(html *string) {
	var tmp = *html

	part := make([]string, 5)

	// add StatPng if not exist
	if strings.Index(tmp, "{{.StatPng}}") == -1 {
		if strings.Index(tmp, "</body>") == -1 {
			tmp = tmp + "<img src='{{.StatPng}}' border='0px' width='10px' height='10px'/>"
		} else {
			tmp = strings.Replace(tmp, "</body>", "<img src='{{.StatPng}}' border='0px' width='10px' height='10px'/></body>", -1)
		}
	}

	// make absolute URL
	tmp = reReplaceRelativeSrc.ReplaceAllStringFunc(tmp, func(str string) string {
		part = reReplaceRelativeSrc.FindStringSubmatch(str)
		return part[1] + part[2] + filepath.Join(models.Config.URL, part[3]) + part[4]
	})
	tmp = reReplaceRelativeHref.ReplaceAllStringFunc(tmp, func(str string) string {
		part = reReplaceRelativeHref.FindStringSubmatch(str)
		return part[1] + part[2] + filepath.Join(models.Config.URL, part[3]) + part[4]
	})

	// replace http and https href link to utm redirect
	tmp = reReplaceLink.ReplaceAllStringFunc(tmp, func(str string) string {
		part = reReplaceLink.FindStringSubmatch(str)
		u, err := url.Parse(part[4])
		checkErr(err)
		part[4] = u.RequestURI()
		return part[1] + `"{{RedirectUrl . "` + part[2] + " " + part[3] + part[4] + `"}}"`
	})

	*html = tmp
}
