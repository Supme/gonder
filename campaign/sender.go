package campaign

import (
	"github.com/supme/gonder/models"
	"sync"
	"time"
)

type campaign struct {
	id, fromEmail, fromName, subject, body string
	dkimSelector	string
	dkimPrivateKey  []byte
	dkimUse			bool
	sendUnsubscribe                        bool
	profileId, resendDelay, resendCount    int
	attachments                            []string
	wg                                     sync.WaitGroup
}

func (c campaign) run(id string) {
	c.get(id)
	c.send()
	c.resend()
}

// Send for recipient from campaign
func send(camp *campaign, id, email, name string, profileId int, iface, host string) {
	r := recipient{
		id:      id,
		toEmail: email,
		toName:  name,
	}

	if r.checkUnsubscribe(camp.id) == false || camp.sendUnsubscribe {
		_, err := models.Db.Exec("UPDATE recipient SET status='Sending', date=NOW() WHERE id=?", r.id)
		checkErr(err)

		result := r.send(camp, &iface, &host)
		models.ProfileFree(profileId)

		_, err = models.Db.Exec("UPDATE recipient SET status=?, date=NOW() WHERE id=?", result, r.id)
		checkErr(err)
	} else {
		_, err := models.Db.Exec("UPDATE recipient SET status='Unsubscribe', date=NOW() WHERE id=?", r.id)
		checkErr(err)
		camplog.Printf("Recipient id %s email %s is unsubscribed", r.id, r.toEmail)
	}
}

// Send campaign
func (c campaign) send() {
	query, err := models.Db.Query("SELECT `id`, `email`, `name` FROM recipient WHERE campaign_id=? AND removed=0 AND status IS NULL", c.id)
	checkErr(err)
	defer query.Close()

	for query.Next() {
		var id, email, name string
		err = query.Scan(&id, &email, &name)
		checkErr(err)

		profileId, iface, host := models.ProfileNext(c.profileId)
		go func(id, email, name string, profileId int, iface, host string) {
			c.wg.Add(1)
			defer c.wg.Done()
			send(&c, id, email, name, profileId, iface, host)
		}(id, email, name, profileId, iface, host)

	}
	c.wg.Wait()
	camplog.Printf("Done campaign id %s.", c.id)
}

func (c campaign) resend() {
	var id, email, name string

	count := c.countSoftBounce()
	if count == 0 {
		return
	}

	if c.resendCount != 0 {
		camplog.Printf("Start %d resend by campaign id %s ", count, c.id)
	}

	for n := 0; n < c.resendCount; n++ {
		if c.countSoftBounce() == 0 {
			return
		}
		time.Sleep(time.Duration(c.resendDelay) * time.Second)

		query, err := models.Db.Query("SELECT `id`, `email`, `name` FROM `recipient` WHERE `campaign_id`=? AND `removed`=0 AND LOWER(`status`) REGEXP '^((4[0-9]{2})|(dial tcp)|(read tcp)|(proxy)|(eof)).+'", c.id)
		checkErr(err)
		defer query.Close()

		for query.Next() {
			err = query.Scan(&id, &email, &name)
			checkErr(err)
			profileId, iface, host := models.ProfileNext(c.profileId)
			send(&c, id, email, name, profileId, iface, host)
		}
	}
}

func (c *campaign) countSoftBounce() int {
	var count int
	//err := models.Db.QueryRow("SELECT COUNT(DISTINCT r.`id`) FROM `recipient` as r,`status` as s WHERE r.`campaign_id`=? AND r.`removed`=0 AND s.`bounce_id`=2 AND UPPER(`r`.`status`) LIKE CONCAT('%',s.`pattern`,'%')", c.id).Scan(&count)
	err := models.Db.QueryRow("SELECT COUNT(`id`) FROM `recipient` WHERE `campaign_id`=? AND `removed`=0 AND LOWER(`status`) REGEXP '^((4[0-9]{2})|(dial tcp)|(proxy)|(eof)).+'", c.id).Scan(&count)
	checkErr(err)
	return count
}

// Get all info for campaign
func (c *campaign) get(id string) {
	err := models.Db.QueryRow("SELECT t1.`id`,t3.`email`,t3.`name`,t3.`dkim_selector`,t3.`dkim_key`,t3.`dkim_use`,t1.`subject`,t1.`body`,t2.`id`,t1.`send_unsubscribe`,t2.`resend_delay`,t2.`resend_count` FROM `campaign` t1 INNER JOIN `profile` t2 ON t2.`id`=t1.`profile_id` INNER JOIN `sender` t3 ON t3.`id`=t1.`sender_id` WHERE t1.`id`=?", id).Scan(
		&c.id,
		&c.fromEmail,
		&c.fromName,
		&c.dkimSelector,
		&c.dkimPrivateKey,
		&c.dkimUse,
		&c.subject,
		&c.body,
		&c.profileId,
		&c.sendUnsubscribe,
		&c.resendDelay,
		&c.resendCount,
	)
	checkErr(err)

	attachment, err := models.Db.Query("SELECT `path` FROM attachment WHERE campaign_id=?", c.id)
	checkErr(err)
	defer attachment.Close()

	c.attachments = nil
	var location string
	for attachment.Next() {
		err = attachment.Scan(&location)
		checkErr(err)
		c.attachments = append(c.attachments, location)
	}
}
