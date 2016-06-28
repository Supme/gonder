package campaign

import (
	"github.com/supme/gonder/models"
	"time"
)

type campaign struct {
	id, from_email, from_name, subject, body, iface, host string
	send_unsubscribe bool
	stream, resend_delay, resend_count int
	attachments []Attachment
}


func (c campaign) run(id string) {
	c.get(id)
	c.send()
	c.resend()
}


// Send campaign
func (c campaign) send() {
	var r recipient
	count := 0
	stream := 0
	next := make(chan bool)

	query, err := models.Db.Prepare("SELECT `id`, `email`, `name` FROM recipient WHERE campaign_id=? AND removed=0 AND status IS NULL")
	checkErr(err)
	defer query.Close()

	q, err := query.Query(c.id)
	checkErr(err)
	defer q.Close()

	for q.Next() {
		err = q.Scan(&r.id, &r.to_email, &r.to_name)
		checkErr(err)
		count += 1
		if r.unsubscribe(c.id) == false  || c.send_unsubscribe {
			models.Db.Exec("UPDATE recipient SET status='Sending', date=NOW() WHERE id=?", r.id)
			stream += 1
			if stream > c.stream {
				<-next
				stream -= 1
			}
			go func(d recipient) {
				rs := d.send(&c)
				models.Db.Exec("UPDATE recipient SET status=?, date=NOW() WHERE id=?", rs, d.id)
				next <- true
			}(r)
		} else {
			models.Db.Exec("UPDATE recipient SET status='Unsubscribe', date=NOW() WHERE id=?", r.id)
			camplog.Printf("Recipient id %s email %s is unsubscribed", r.id, r.to_email)
		}
	}

	for stream != 0 {
		<-next
		stream -= 1
	}
	close(next)

	camplog.Printf("Done campaign %s. The number of recipients %d", c.id, count)
}

func (c campaign) resend() {
	var r recipient
	if c.resend_count != 0 {
		camplog.Printf("Start resend by campaign id %s ", c.id)
	}

	for n := 0; n < c.resend_count; n++ {
		time.Sleep(time.Duration(c.resend_delay) * time.Second)
		query, err := models.Db.Prepare("SELECT DISTINCT r.`id`, r.`email`, r.`name` FROM `recipient` as r,`status` as s WHERE r.`campaign_id`=? AND r.`removed`=0 AND s.`bounce_id`=2 AND UPPER(`r`.`status`) LIKE CONCAT(\"%\",s.`pattern`,\"%\")")
		checkErr(err)
		defer query.Close()

		q, err := query.Query(c.id)
		checkErr(err)
		defer q.Close()

		for q.Next() {
			err = q.Scan(&r.id, &r.to_email, &r.to_name)
			checkErr(err)
			if r.unsubscribe(c.id) == false || c.send_unsubscribe {
				models.Db.Exec("UPDATE recipient SET status='Sending', date=NOW() WHERE id=?", r.id)
				rs := r.send(&c)
				models.Db.Exec("UPDATE recipient SET status=?, date=NOW() WHERE id=?", rs, r.id)
			} else {
				models.Db.Exec("UPDATE recipient SET status='Unsubscribe', date=NOW() WHERE id=?", r.id)
				camplog.Printf("Recipient id %s email %s is unsubscribed", r.id, r.to_email)
			}
		}
	}
}

// Get all info for campaign
func (c *campaign) get(id string) {
	err := models.Db.QueryRow("SELECT t1.`id`,t3.`email`,t3.`name`,t1.`subject`,t1.`body`,t2.`iface`,t2.`host`,t2.`stream`,t1.`send_unsubscribe`,t2.`resend_delay`,t2.`resend_count` FROM `campaign` t1 INNER JOIN `profile` t2 ON t2.`id`=t1.`profile_id` INNER JOIN `sender` t3 ON t3.`id`=t1.`sender_id` WHERE t1.`id`=?", id).Scan(
		&c.id,
		&c.from_email,
		&c.from_name,
		&c.subject,
		&c.body,
		&c.iface,
		&c.host,
		&c.stream,
		&c.send_unsubscribe,
		&c.resend_delay,
		&c.resend_count,
	)

	checkErr(err)

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
		c.attachments = append(c.attachments, Attachment{Location: location, Name: name})
	}
}


