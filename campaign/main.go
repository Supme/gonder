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
	"time"
	"github.com/supme/gonder/models"
	"log"
)

var (
	MaxCampaingns int
	resendPause int
	resendCount int

	startedCampaign []string
)

func Run()  {
	MaxCampaingns = 2
	resendPause = 60
	resendCount = 3

	for {
		for len(startedCampaign) >= MaxCampaingns {
			time.Sleep(1 * time.Second)
		}

		c := next_campaign()
		if c.id != "" {
			startedCampaign = append(startedCampaign, c.id)
			go run_campaign(c)
		}
		time.Sleep(1 * time.Second)
	}
}

func next_campaign() campaign {
	var c campaign

	started := ""
	for i, s := range startedCampaign {
		if i != 0 {
			started += ","
		}
		started += "'" + s + "'"
	}

	query := "SELECT t1.`id`,t1.`from`,t1.`from_name`,t1.`subject`,t1.`body`,t2.`iface`,t2.`host`,t2.`stream`,t2.`delay`, t1.`send_unsubscribe`  FROM `campaign` t1 INNER JOIN `profile` t2 ON t2.`id`=t1.`profile_id` WHERE NOW() BETWEEN t1.`start_time` AND t1.`end_time` AND (SELECT COUNT(*) FROM `recipient` WHERE campaign_id=t1.`id` AND status IS NULL) > 0"
	if started != "" {
		query += " AND t1.`id` NOT IN (" + started + ")"
	}

	models.Db.QueryRow(query).Scan(
		&c.id,
		&c.from,
		&c.from_name,
		&c.subject,
		&c.body,
		&c.iface,
		&c.host,
		&c.stream,
		&c.delay,
		&c.send_unsubscribe,
	)
	return c
}

func remove_started_campaign(id string) {
	for i, d := range startedCampaign {
		if d == id {
			startedCampaign = append(startedCampaign[:i], startedCampaign[i+1:]...)
			return
		}
	}
	return
}

func run_campaign(c campaign) {
	log.Println("Start campaign ", c.id)
	c.get_attachments()
	c.send()
	log.Println("Resend bounce mail for campaign id", c.id)
	c.resend_soft_bounce()
	remove_started_campaign(c.id)
	log.Println("Finish campaign id", c.id)
}
