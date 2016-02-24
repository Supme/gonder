// Project Gonder.
// Author Supme
// Copyright Supme 2016
// License http://opensource.org/licenses/MIT MIT License
//
//  THE SOFTWARE AND DOCUMENTATION ARE PROVIDED "AS IS" WITHOUT WARRANTY OF
//  ANY KIND, EITHER EXPRESSED OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
//  IMPLIED WARRANTIES OF MERCHANTABILITY AND/OR FITNESS FOR A PARTICULAR
//  PURPOSE.
//
// Please see the License.txt file for more information.
//
package campaign

import (
	"github.com/supme/gonder/mailer"
	"github.com/supme/gonder/models"
	"errors"
//	_ "github.com/eaigner/dkim"
	"log"
	"sync"
	"time"
	"strconv"
	"math/rand"
)

var Send bool

type (
	campaign struct {
		id, from, from_name, subject, body, iface, host, send_unsubscribe string
		stream      int
		delay       int
		attachments []mailer.Attachment
		recipients  []recipient
	}

	recipient struct {
		id, to, to_name	string
	}
)

func (c *campaign) get_null_recipients() {
	var d recipient
	c.recipients = nil

	query, err := models.Db.Prepare("SELECT `id`, `email`, `name` FROM recipient WHERE campaign_id=? AND status IS NULL LIMIT ?")
	checkErr(err)
	defer query.Close()

	q, err := query.Query(c.id, c.stream)
	checkErr(err)
	defer q.Close()

	for q.Next() {
		err = q.Scan(&d.id, &d.to, &d.to_name)
		checkErr(err)
		c.recipients = append(c.recipients, d)
	}
}

func (c *campaign) get_soft_bounce_recipients() {
	var	d recipient
	c.recipients = nil

	query, err := models.Db.Prepare("SELECT DISTINCT r.`id`, r.`email`, r.`name` FROM `recipient` as r,`status` as s WHERE r.`campaign_id`=? AND s.`bounce_id`=2 AND UPPER(`r`.`status`) LIKE CONCAT(\"%\",s.`pattern`,\"%\")")
	checkErr(err)
	defer query.Close()

	q, err := query.Query(c.id)
	checkErr(err)
	defer q.Close()

	for q.Next() {
		err = q.Scan(&d.id, &d.to, &d.to_name)
		checkErr(err)
		c.recipients = append(c.recipients, d)
	}

}

func (r *recipient) get(id string) {
	models.Db.QueryRow("SELECT `id`, `email`, `name` FROM recipient WHERE id=? AND status IS NULL LIMIT ?", id).Scan(r.id, r.to, r.to_name)
}

func (r recipient) send(c *campaign) string {
	data := new(mailer.MailData)
	data.Iface = c.iface
	data.From = c.from
	data.From_name = c.from_name
	data.Host = c.host
	data.Attachments = c.attachments
	data.To = r.to
	data.To_name = r.to_name

	var rs string
	d, e := models.MailMessage(c.id, r.id, c.subject, c.body)
	if e == nil {
		data.Subject = d.Subject
		data.Html = d.Body
		data.Extra_header = "List-Unsubscribe: " + models.UnsubscribeUrl(c.id, r.id) + "\r\nPrecedence: bulk\r\n"
		data.Extra_header += "Message-ID: <" + strconv.FormatInt(time.Now().Unix(), 10) + c.id + "." + r.id +"@" + c.host + ">" + "\r\n"

		var res error
		if Send {
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

	log.Printf("Campaign %s for recipient id %s email %s is %s", c.id, r.id, data.To, rs)
	return rs
}

func (c *campaign) get(id string) {
	models.Db.QueryRow("SELECT t1.`id`,t1.`from`,t1.`from_name`,t1.`subject`,t1.`body`,t2.`iface`,t2.`host`,t2.`stream`,t2.`delay`, t1.`send_unsubscribe`  FROM campaign t1 INNER JOIN `profile` t2 ON t2.`id`=t1.`profile_id` WHERE id=?", id).Scan(
		c.id,
		c.from,
		c.from_name,
		c.subject,
		c.body,
		c.iface,
		c.host,
		c.stream,
		c.delay,
		c.send_unsubscribe,
	)
}

func (c *campaign) get_attachments() {
	// добавим приложенные к кампании файлы
	attachment, err := models.Db.Prepare("SELECT `path`, `file` FROM attachment WHERE campaign_id=?")
	checkErr(err)
	defer attachment.Close()

	attach, err := attachment.Query(c.id)
	checkErr(err)
	defer attach.Close()

	c.attachments = nil
	var location string
	var name string
	for attach.Next() {
		err = attach.Scan(&location, &name)
		checkErr(err)
		c.attachments = append(c.attachments, mailer.Attachment{Location: location, Name: name})
	}
}

func (c campaign) send() {
	var w sync.WaitGroup
	s := true
	for s {
		c.get_null_recipients()
		if len(c.recipients) == 0 {
			s = false
		}
		for _, r := range c.recipients {
			// если пользователь ни разу не отказался от подписки в этой группе
			var unsubscribeCount int
			models.Db.QueryRow("SELECT COUNT(*) FROM `unsubscribe` t1 INNER JOIN `campaign` t2 ON t1.group_id = t2.group_id WHERE t2.id = ? AND t1.email = ?", c.id, r.to).Scan(&unsubscribeCount)
			if unsubscribeCount == 0 || c.send_unsubscribe == "y" {
				w.Add(1)
				models.Db.Exec("UPDATE recipient SET status='Sending', date=NOW() WHERE id=?", r.id)
				go func(d recipient) {
					rs := d.send(&c)
					models.Db.Exec("UPDATE recipient SET status=?, date=NOW() WHERE id=?", rs, d.id)
					w.Done()
				}(r)
			} else {
				models.Db.Exec("UPDATE recipient SET status='Unsubscribe', date=NOW() WHERE id=?", r.id)
				log.Printf("Recipient id %s email %s is unsubscribed", r.id, r.to)
			}
		}
		w.Wait()
		time.Sleep(time.Second + time.Duration(c.delay) * time.Second)
	}
}

func (c campaign) resend_soft_bounce() {
	c.get_soft_bounce_recipients()
	if len(c.recipients) == 0 {
		return
	}

	for n := 0; n < resendCount; n++ {
		time.Sleep(time.Duration(resendPause) * time.Second)
		for _, r := range c.recipients {
			// если пользователь ни разу не отказался от подписки в этой группе
			var unsubscribeCount int
			models.Db.QueryRow("SELECT COUNT(*) FROM `unsubscribe` t1 INNER JOIN `campaign` t2 ON t1.group_id = t2.group_id WHERE t2.id = ? AND t1.email = ?", c.id, r.to).Scan(&unsubscribeCount)
			if unsubscribeCount == 0 || c.send_unsubscribe == "y" {
				models.Db.Exec("UPDATE recipient SET status='Sending', date=NOW() WHERE id=?", r.id)
				rs := r.send(&c)
				models.Db.Exec("UPDATE recipient SET status=?, date=NOW() WHERE id=?", rs, r.id)
			} else {
				models.Db.Exec("UPDATE recipient SET status='Unsubscribe', date=NOW() WHERE id=?", r.id)
				log.Printf("Recipient id %s email %s is unsubscribed", r.id, r.to)
			}
		}
	}
}

// Get active campaigns
func get_active_campaigns(limit int) []campaign{
	var c []campaign
	var d campaign

	campaign, err := models.Db.Prepare("SELECT t1.`id`,t1.`from`,t1.`from_name`,t1.`subject`,t1.`body`,t2.`iface`,t2.`host`,t2.`stream`,t2.`delay`, t1.`send_unsubscribe`  FROM `campaign` t1 INNER JOIN `profile` t2 ON t2.`id`=t1.`profile_id` WHERE NOW() BETWEEN t1.`start_time` AND t1.`end_time` AND (SELECT COUNT(*) FROM `recipient` WHERE campaign_id=t1.`id` AND status IS NULL) > 0 LIMIT ?")
	checkErr(err)
	defer campaign.Close()

	camp, err := campaign.Query(limit)
	checkErr(err)
	defer camp.Close()

	for camp.Next() {
		err = camp.Scan(
			&d.id,
			&d.from,
			&d.from_name,
			&d.subject,
			&d.body,
			&d.iface,
			&d.host,
			&d.stream,
			&d.delay,
			&d.send_unsubscribe,
		)
		checkErr(err)

		c = append(c, d)
	}
	return c
}

func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}
