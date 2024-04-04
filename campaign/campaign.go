package campaign

import (
	"errors"
	"fmt"
	"github.com/Supme/smtpSender"
	"gonder/models"
	"io"
	"log"
	"math/rand"
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
	Data        *models.CampaignData
	ResendDelay int
	ResendCount int
	Stop        chan struct{}
	Finish      chan struct{}
}

var campLog *log.Logger

// Run start look database for ready campaign for send
func Run(logger *log.Logger) {
	campLog = logger

	Sending.campaigns = map[string]*campaign{}

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

func getCampaign(id string) (*campaign, error) {
	var (
		c   campaign
		err error
	)

	c.Data, err = models.CampaignGetByStringID(id).GetData()
	if err != nil {
		return &c, err
	}
	c.Stop = make(chan struct{})
	c.Finish = make(chan struct{})

	c.ResendDelay, c.ResendCount, err = models.EmailPool.GetResendParams(c.Data.ProfileID)
	if err != nil {
		return &c, fmt.Errorf("error get params for pool id %d: %s", c.Data.ProfileID, err)
	}

	return &c, nil
}

func (c *campaign) send() {
	pipe, err := models.EmailPool.Get(c.Data.ProfileID)
	checkErr(err)

	campLog.Printf("Start campaign id %s.", c.Data.Campaign.StringID())

	for c.Data.Campaign.HasNotSent() {
		select {
		case <-c.Stop:
			goto Done
		default:
			c.streamSend(pipe)
		}
	}

	select {
	case <-c.Stop:
		campLog.Printf("Campaign %s stoped", c.Data.Campaign.StringID())
		goto End
	default:
		resend := c.Data.Campaign.CountResend()
		if resend > 0 && c.ResendCount > 0 {
			campLog.Printf("Done stream send campaign id %s but need %d resend.", c.Data.Campaign.StringID(), resend)
		}
		if resend > 0 {
			for i := 1; i <= c.ResendCount; i++ {
				select {
				case <-c.Stop:
					goto Done
				default:
					resend := c.Data.Campaign.CountResend()
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

					campLog.Printf("Start %s resend by %d email from campaign id %s ", models.Conv1st2nd(i), resend, c.Data.Campaign.StringID())
					c.resend(pipe)
				}

				select {
				case <-c.Stop:
					goto Done
				default:
					campLog.Printf("Done %s resend campaign id %s", models.Conv1st2nd(i), c.Data.Campaign.StringID())
				}
			}
		}
	Finish:
		campLog.Printf("Finish campaign id %s", c.Data.Campaign.StringID())
	}
Done:
	select {
	case <-c.Stop:
		campLog.Printf("Campaign %s stoped", c.Data.Campaign.StringID())
	default:
		close(c.Stop)
	}
End:
	close(c.Finish)
}

func (c *campaign) streamSend(pipe *smtpSender.Pipe) {
	query, err := models.Db.Query("SELECT `id` FROM recipient WHERE campaign_id=? AND removed=0 AND status IS NULL", c.Data.Campaign.StringID())
	checkErr(err)
	defer query.Close()

	wg := &sync.WaitGroup{}
	for query.Next() {
		select {
		case <-c.Stop:
			campLog.Printf("Stop signal for campaign %s received", c.Data.Campaign.StringID())
			goto Done
		default:
			var rID string
			err = query.Scan(&rID)
			checkErr(err)
			r, err := GetRecipient(rID)
			checkErr(err)

			if r.unsubscribed() && !c.Data.SendUnsubscribe {
				err = models.RecipientGetByStringID(r.ID).UpdateRecipientStatus(models.StatusUnsubscribe)
				checkErr(err)
				campLog.Printf("Campaign %s recipient id %s email %s is unsubscribed", c.Data.Campaign.StringID(), r.ID, r.Email)
				models.Prometheus.Campaign.SendResult.WithLabelValues(c.Data.Campaign.StringID(), "unsubscribed", "stream").Inc()
				continue
			}

			err = models.RecipientGetByStringID(r.ID).UpdateRecipientStatus(models.StatusSending)
			checkErr(err)

			email := c.getBuilder(r).Email(r.ID, GetResultFunc(wg, SendTypeStream, c.Data.Campaign.StringID(), r.ID, r.Email))
			email.DontUseTLS = models.Config.DontUseTLS

			wg.Add(1)
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
	query, err := models.Db.Query("SELECT `id` FROM recipient WHERE campaign_id=? AND removed=0 AND "+softBounceWhere, c.Data.Campaign.StringID())
	checkErr(err)
	defer func() {
		if err := query.Close(); err != nil {
			log.Print(err)
		}
	}()

	wg := &sync.WaitGroup{}
	for query.Next() {
		select {
		case <-c.Stop:
			campLog.Printf("Stop signal for campaign %s received", c.Data.Campaign.StringID())
			return
		default:
			var rID string
			err = query.Scan(&rID)
			checkErr(err)
			r, err := GetRecipient(rID)
			checkErr(err)

			if r.unsubscribed() && !c.Data.SendUnsubscribe {
				err = models.RecipientGetByStringID(r.ID).UpdateRecipientStatus(models.StatusUnsubscribe)
				checkErr(err)
				campLog.Printf("Recipient id %s email %s is unsubscribed", r.ID, r.Email)
				models.Prometheus.Campaign.SendResult.WithLabelValues(c.Data.Campaign.StringID(), "unsubscribed", "resend").Inc()
				continue
			}

			wg.Add(1)
			err = models.RecipientGetByStringID(r.ID).UpdateRecipientStatus(models.StatusSending)
			checkErr(err)

			email := c.getBuilder(r).Email(r.ID, GetResultFunc(wg, SendTypeResend, c.Data.Campaign.StringID(), r.ID, r.Email))
			email.DontUseTLS = models.Config.DontUseTLS

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
			res = errors.New("421 Fake send")
			models.Prometheus.Campaign.SendResult.WithLabelValues("", "421", "fake").Inc()
		} else {
			res = errors.New("Ok Fake send")
			models.Prometheus.Campaign.SendResult.WithLabelValues("", "250", "fake").Inc()
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

func (c campaign) getBuilder(r Recipient) *smtpSender.Builder {
	bldr := new(smtpSender.Builder)
	bldr.SetFrom(c.Data.Name, c.Data.Email)
	if c.Data.BimiSelector != "" {
		bldr.AddHeader("BIMI-Selector: v=BIMI1; s=" + c.Data.BimiSelector + ";")
	}
	bldr.SetTo(r.Name, r.Email)
	bldr.AddHeader(
		"List-Unsubscribe: "+models.EncodeUTM("unsubscribe", c.Data.UtmURL, "mail", r.Params)+"\nPrecedence: bulk",
		"Message-ID: <"+strconv.FormatInt(time.Now().Unix(), 10)+c.Data.Campaign.StringID()+"."+r.ID+"@"+"gonder"+">", // ToDo hostname
		"X-Postmaster-Msgtype: campaign"+c.Data.Campaign.StringID(),
	)
	bldr.AddSubjectFunc(c.subjectTemplFunc(r))
	if err := bldr.AddHTMLFunc(c.htmlTemplFunc(r, false, false)); err != nil {
		log.Print(err)
	}
	if c.Data.HasTextTemplate() {
		bldr.AddTextFunc(c.textTemplFunc(r, false))
	}
	if c.Data.HasAMPTemplate() {
		bldr.AddAMPFunc(c.ampTemplFunc(r, false))
	}
	if err := bldr.AddAttachment(c.Data.Attachments...); err != nil {
		log.Print(err)
	}
	if c.Data.DkimUse {
		bldr.SetDKIM(c.Data.DkimDomain, c.Data.DkimSelector, c.Data.DkimPrivateKey)
	}
	return bldr
}

func (c campaign) HasResend() int {
	var count int
	err := models.Db.QueryRow("SELECT COUNT(`id`) FROM `recipient` WHERE `campaign_id`=? AND removed=0 AND "+softBounceWhere, c.Data.Campaign.IntID()).Scan(&count)
	checkErr(err)
	return count
}

func (c campaign) subjectTemplFunc(r Recipient) func(io.Writer) error {
	return func(w io.Writer) error {
		tmpl, err := c.Data.GetSubjectTemplate()
		if err != nil {
			return err
		}
		return tmpl.Execute(w, r.Params)
	}
}

func (c campaign) htmlTemplFunc(r Recipient, web bool, preview bool) func(io.Writer) error {
	return func(w io.Writer) error {
		tmpl, err := c.Data.GetHTMLTemplate()
		if err != nil {
			return err
		}
		if preview {
			if web {
				r.Params["WebUrl"] = nil
			} else {
				r.Params["WebUrl"] = "/preview?id=" + r.ID + "&type=web"
			}
			r.Params["StatUrl"] = "data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7"
			r.Params["UnsubscribeUrl"] = "/unsubscribe?recipientId=" + r.ID
			r.Params["QuestionUrl"] = "/question?recipientId=" + r.ID
			tmpl = tmpl.Funcs(
				template.FuncMap{
					"RedirectUrl": func(p map[string]interface{}, u string) string {
						url := regexp.MustCompile(`\s*?(\[.*?\])\s*?`).Split(u, 2)
						return strings.TrimSpace(url[len(url)-1])
					},
					// ToDo more functions (example QRcode generator)
				})
		} else {
			if web {
				r.Params["WebUrl"] = nil
			}
		}

		return tmpl.Execute(w, r.Params)
	}
}

func (c campaign) textTemplFunc(r Recipient, web bool) func(io.Writer) error {
	return func(w io.Writer) error {
		if web {
			r.Params["WebUrl"] = nil
		} else {
			r.Params["WebUrl"] = models.EncodeUTM("web", c.Data.UtmURL, "", r.Params)
		}
		tmpl, err := c.Data.GetTextTemplate()
		if err != nil {
			return err
		}
		return tmpl.Execute(w, r.Params)
	}
}

func (c campaign) ampTemplFunc(r Recipient, web bool) func(io.Writer) error {
	return func(w io.Writer) error {
		if web {
			r.Params["WebUrl"] = nil
		}
		tmpl, err := c.Data.GetAMPTemplate()
		if err != nil {
			return err
		}
		return tmpl.Execute(w, r.Params)
	}
}
