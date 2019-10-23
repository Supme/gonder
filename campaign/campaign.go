package campaign

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/Supme/smtpSender"
	"github.com/tdewolff/minify"
	"gonder/campaign/minifyEmail"
	"gonder/models"
	"io"
	"log"
	"math/rand"
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
	FromDomain      string
	DkimSelector    string
	DkimPrivateKey  []byte
	DkimUse         bool
	SendUnsubscribe bool
	ProfileID       int
	ResendDelay     int
	ResendCount     int
	Attachments     []string

	subjectTmplFunc  *template.Template
	templateHTML     string
	templateHTMLFunc *template.Template
	compressHTML     bool
	templateText     string
	templateTextFunc *template.Template

	utmURL string

	Stop   chan struct{}
	Finish chan struct{}
}

var campLog *log.Logger

// Run start look database for ready campaign for send
func Run(logger *log.Logger) {
	campLog = logger

	Sending.campaigns = map[string]campaign{}

	for {
		for Sending.Count() >= models.Config.MaxCampaingns {
			time.Sleep(1 * time.Second)
		}
		if Sending.Count() > 0 {
			camp, err := Sending.checkExpired()
			checkErr(err)
			Sending.Stop(camp...)
		}
		if id, err := Sending.checkNext(); err == nil {
			fmt.Println("check next campaign for send id:", id)
			camp, err := getCampaign(id)
			checkErr(err)
			Sending.add(camp)
			go func(id string) {
				camp.send()
				Sending.removeStarted(id)
			}(id)
			continue
		}
		models.Prometheus.Campaign.Started.Set(float64(Sending.Count()))
		time.Sleep(time.Second * 10)
	}
}

func getCampaign(id string) (campaign, error) {
	var (
		subject string
		utmURL  string
	)
	c := campaign{ID: id}
	c.Stop = make(chan struct{})
	c.Finish = make(chan struct{})

	err := models.Db.QueryRow("SELECT t2.`email`,t2.`name`, t2.`utm_url`, t2.`dkim_selector`,t2.`dkim_key`,t2.`dkim_use`,t1.`subject`,t1.`template_html`,`template_text`,t1.`compress_html`, t1.`profile_id`,t1.`send_unsubscribe` FROM `campaign` t1  INNER JOIN `sender` t2 ON t2.`id`=t1.`sender_id` WHERE t1.`id`=?", id).Scan(
		&c.FromEmail,
		&c.FromName,
		&utmURL,
		&c.DkimSelector,
		&c.DkimPrivateKey,
		&c.DkimUse,
		&subject,
		&c.templateHTML,
		&c.templateText,
		&c.compressHTML,
		&c.ProfileID,
		&c.SendUnsubscribe,
	)
	if err != nil {
		return c, err
	}

	c.ResendDelay, c.ResendCount, err = models.EmailPool.GetResendParams(c.ProfileID)
	if err != nil {
		return c, fmt.Errorf("error get params for pool id %d: %s", c.ProfileID, err)
	}

	splitEmail := strings.Split(c.FromEmail, "@")
	if len(splitEmail) == 2 {
		c.FromDomain = strings.ToLower(strings.TrimSpace(splitEmail[1]))
	}

	c.subjectTmplFunc, err = template.New("Subject").Parse(subject)
	if err != nil {
		return c, fmt.Errorf("error parse campaign '%s' subject template: %s", c.ID, err)
	}

	prepareHTMLTemplate(&c.templateHTML, c.compressHTML)
	//fmt.Println(c.templateHTML)
	c.templateHTMLFunc = template.New("HTML")
	c.templateTextFunc = template.New("Text")

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

	campLog.Printf("Start campaign id %s.", c.ID)
	c.streamSend(pipe)

	select {
	case <-c.Stop:
		campLog.Printf("Campaign %s stoped", c.ID)
		goto End
	default:
		resend := c.HasResend()
		if resend > 0 && c.ResendCount > 0 {
			campLog.Printf("Done stream send campaign id %s but need %d resend.", c.ID, resend)
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

					campLog.Printf("Start %s resend by %d email from campaign id %s ", models.Conv1st2nd(i), resend, c.ID)
					c.resend(pipe)
				}

				select {
				case <-c.Stop:
					goto Done
				default:
					campLog.Printf("Done %s resend campaign id %s", models.Conv1st2nd(i), c.ID)
				}
			}
		}
	Finish:
		campLog.Printf("Finish campaign id %s", c.ID)
	}
Done:
	select {
	case <-c.Stop:
		campLog.Printf("Campaign %s stoped", c.ID)
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
			campLog.Printf("Stop signal for campaign %s received", c.ID)
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
				campLog.Printf("Campaign %s recipient id %s email %s is unsubscribed", c.ID, r.ID, r.Email)
				models.Prometheus.Campaign.SendResult.WithLabelValues(c.ID, models.GetDomainFromEmail(r.Email), "unsubscribed", "stream").Inc()
				continue
			}

			wg.Add(1)
			_, err = updateRecipientStatus.Exec(models.StatusSending, r.ID)
			checkErr(err)

			bldr := new(smtpSender.Builder)
			bldr.SetFrom(c.FromName, c.FromEmail)
			bldr.SetTo(r.Name, r.Email)
			bldr.AddHeader(
				"List-Unsubscribe: "+models.EncodeUTM("unsubscribe", c.utmURL, "mail", r.Params)+"\nPrecedence: bulk",
				"Message-ID: <"+strconv.FormatInt(time.Now().Unix(), 10)+c.ID+"."+r.ID+"@"+"gonder"+">", // ToDo hostname
				"X-Postmaster-Msgtype: campaign"+c.ID,
			)
			bldr.AddSubjectFunc(c.subjectTemplFunc(r))
			if err := bldr.AddHTMLFunc(c.htmlTemplFunc(r, false, false)); err != nil {
				log.Print(err)
			}
			if c.templateText != "" {
				bldr.AddTextFunc(c.textTemplFunc(r, false, false))
			}
			if err := bldr.AddAttachment(c.Attachments...); err != nil {
				log.Print(err)
			}
			if c.DkimUse {
				bldr.SetDKIM(c.FromDomain, c.DkimSelector, c.DkimPrivateKey)
			}
			email := bldr.Email(r.ID, func(result smtpSender.Result) {
				var res string
				if result.Err == nil {
					res = "Ok"
				} else {
					res = result.Err.Error()
				}
				_, err := updateRecipientStatus.Exec(res, result.ID)
				checkErr(err)
				campLog.Printf("Campaign %s for recipient id %s email %s is %s send time %s", c.ID, r.ID, r.Email, res, result.Duration.String())
				models.Prometheus.Campaign.SendResult.WithLabelValues(c.ID, models.GetDomainFromEmail(r.Email), models.GetStatusCodeFromSendResult(result.Err), "stream").Inc()
				wg.Done()
			})
			if !models.Config.RealSend {
				err = fakeSend(*email)
			} else {
				err = pipe.Send(*email)
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
	defer func() {
		if err := query.Close(); err != nil {
			log.Print(err)
		}
	}()

	updateRecipientStatus, err := c.prepareQueryUpdateRecipientStatus()
	checkErr(err)

	wg := &sync.WaitGroup{}
	for query.Next() {
		select {
		case <-c.Stop:
			campLog.Printf("Stop signal for campaign %s received", c.ID)
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
				campLog.Printf("Recipient id %s email %s is unsubscribed", r.ID, r.Email)
				models.Prometheus.Campaign.SendResult.WithLabelValues(c.ID, models.GetDomainFromEmail(r.Email), "unsubscribed", "resend").Inc()
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
			if err := bldr.AddHTMLFunc(c.htmlTemplFunc(r, false, false)); err != nil {
				log.Print(err)
			}
			if c.templateText != "" {
				bldr.AddTextFunc(c.textTemplFunc(r, false, false))
			}
			if err := bldr.AddAttachment(c.Attachments...); err != nil {
				log.Print(err)
			}
			if c.DkimUse {
				bldr.SetDKIM(c.FromDomain, c.DkimSelector, c.DkimPrivateKey)
			}
			email := bldr.Email(r.ID, func(result smtpSender.Result) {
				var res string
				if result.Err == nil {
					res = "Ok"
				} else {
					res = result.Err.Error()
				}
				_, err := updateRecipientStatus.Exec(res, result.ID)
				checkErr(err)
				campLog.Printf("Campaign %s for recipient id %s email %s is %s send time %s", c.ID, r.ID, r.Email, res, result.Duration.String())
				models.Prometheus.Campaign.SendResult.WithLabelValues(c.ID, models.GetDomainFromEmail(r.Email), models.GetStatusCodeFromSendResult(result.Err), "resend").Inc()
				wg.Done()
			})
			if !models.Config.RealSend {
				err = fakeSend(*email)
			} else {
				err = pipe.Send(*email)
			}
			if err != nil {
				campLog.Println(err)
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
			models.Prometheus.Campaign.SendResult.WithLabelValues("", models.GetDomainFromEmail(email.To), "421", "test").Inc()
		} else {
			res = errors.New("Ok Test send")
			models.Prometheus.Campaign.SendResult.WithLabelValues("", models.GetDomainFromEmail(email.To), "250", "test").Inc()
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
			r.Params["UnsubscribeUrl"] = "/unsubscribe?recipientId=" + r.ID
			r.Params["QuestionUrl"] = "/question?recipientId=" + r.ID
			_, err := c.templateHTMLFunc.Funcs(
				template.FuncMap{
					"RedirectUrl": func(p map[string]interface{}, u string) string {
						url := regexp.MustCompile(`\s*?(\[.*?\])\s*?`).Split(u, 2)
						return strings.TrimSpace(url[len(url)-1])
					},
					// ToDo more functions (example QRcode generator)
				}).Parse(c.templateHTML)
			if err != nil {
				log.Print(err)
			}
		} else {
			if web {
				r.Params["WebUrl"] = nil
			}
			_, err := c.templateHTMLFunc.Funcs(
				template.FuncMap{
					"RedirectUrl": func(p map[string]interface{}, u string) string {
						return models.EncodeUTM("redirect", c.utmURL, u, p)
					},
					// ToDo more functions (example QRcode generator)
				}).Parse(c.templateHTML)
			if err != nil {
				log.Print(err)
			}
		}

		return c.templateHTMLFunc.Execute(w, r.Params)
	}
}

func (c *campaign) textTemplFunc(r Recipient, web bool, preview bool) func(io.Writer) error {
	return func(w io.Writer) error {
		if preview {
			if web {
				r.Params["WebUrl"] = nil
			} else {
				r.Params["WebUrl"] = "/preview?id=" + r.ID + "&type=web"
			}
			r.Params["StatPng"] = "data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7"
			r.Params["UnsubscribeUrl"] = "/unsubscribe?recipientId=" + r.ID
			_, err := c.templateTextFunc.Funcs(
				template.FuncMap{
					"RedirectUrl": func(p map[string]interface{}, u string) string {
						url := regexp.MustCompile(`\s*?(\[.*?\])\s*?`).Split(u, 2)
						return strings.TrimSpace(url[len(url)-1])
					},
					// ToDo more functions (example QRcode generator)
				}).Parse(c.templateText)
			if err != nil {
				log.Print(err)
			}
		} else {
			if web {
				r.Params["WebUrl"] = nil
			} else {
				r.Params["WebUrl"] = models.EncodeUTM("web", c.utmURL, "", r.Params)
			}
			_, err := c.templateTextFunc.Funcs(
				template.FuncMap{
					"RedirectUrl": func(p map[string]interface{}, u string) string {
						return models.EncodeUTM("redirect", c.utmURL, u, p)
					},
					// ToDo more functions (example QRcode generator)
				}).Parse(c.templateText)
			if err != nil {
				log.Print(err)
			}
		}

		return c.templateTextFunc.Execute(w, r.Params)
	}
}

var (
	reReplaceLink         = regexp.MustCompile(`(\s[hH][rR][eE][fF]\s*?=\s*?)["']\s*?(\[.*?\])?\s*?(\b[hH][tT]{2}[pP][sS]?\b:\/\/\b)(.*?)["']`)
	reReplaceRelativeSrc  = regexp.MustCompile(`(\s[sS][rR][cC]\s*?=\s*?)(["'])(\.?\/?files\/.*?)(["'])`)
	reReplaceRelativeHref = regexp.MustCompile(`(\s[hH][rR][eE][fF]\s*?=\s*?)(["'])(\.?\/?files\/.*?)(["'])`)
)

func prepareHTMLTemplate(htmlTmpl *string, useCompress bool) {
	var (
		tmp string
		err error
	)
	if useCompress {
		m := minify.New()
		m.Add("email/html", &minifyEmail.Minifier{
			KeepComments:            true,
			KeepConditionalComments: true,
			KeepDefaultAttrVals:     false,
			KeepDocumentTags:        false,
			KeepEndTags:             false,
			KeepWhitespace:          false,
		})

		tmp, err = m.String("email/html", *htmlTmpl)
		if err != nil {
			campLog.Print(err)
		}
	} else {
		tmp = *htmlTmpl
	}

	part := make([]string, 5)

	// add StatPng if not exist
	if !strings.Contains(tmp, "{{.StatPng}}") {
		if !strings.Contains(tmp, "</body>") {
			tmp = tmp + "<img src=\"{{.StatPng}}\" border=\"0px\" width=\"10px\" height=\"10px\" alt=\"\"/>"
		} else {
			tmp = strings.Replace(tmp, "</body>", "<img src=\"{{.StatPng}}\" border=\"0px\" width=\"10px\" height=\"10px\" alt=\"\"/></body>", -1)
		}
	}

	// make absolute URL
	tmp = reReplaceRelativeSrc.ReplaceAllStringFunc(tmp, func(str string) string {
		part = reReplaceRelativeSrc.FindStringSubmatch(str)
		return part[1] + part[2] + filepath.Join(models.Config.UTMDefaultURL, part[3]) + part[4]
	})
	tmp = reReplaceRelativeHref.ReplaceAllStringFunc(tmp, func(str string) string {
		part = reReplaceRelativeHref.FindStringSubmatch(str)
		return part[1] + part[2] + filepath.Join(models.Config.UTMDefaultURL, part[3]) + part[4]
	})

	// replace http and https href link to utm redirect
	tmp = reReplaceLink.ReplaceAllStringFunc(tmp, func(str string) string {
		part = reReplaceLink.FindStringSubmatch(str)
		return part[1] + `"{{RedirectUrl . "` + part[2] + " " + part[3] + part[4] + `"}}"`
	})

	*htmlTmpl = tmp
}
