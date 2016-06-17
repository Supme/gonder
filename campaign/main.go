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

package campaign

import (
	"time"
	"github.com/supme/gonder/models"
	"log"
	"os"
	"io"
	"bytes"
	"sync"
)

var (
	startedCampaign struct{
		campaigns []string
		sync.Mutex
	}
	campSum int
	camplog *log.Logger
)

// Start look database for ready campaign for send
func Run()  {
	l, err := os.OpenFile("log/campaign.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Println("error opening campaign log file: %v", err)
	}
	defer l.Close()

	multi := io.MultiWriter(l, os.Stdout)

	camplog = log.New(multi, "", log.Ldate|log.Ltime)

	for {
		for {
			startedCampaign.Lock()
			campSum = len(startedCampaign.campaigns)
			startedCampaign.Unlock()
			if campSum <= models.Config.MaxCampaingns {
				break
			}
			time.Sleep(1 * time.Second)
		}

		c := nextCampaign()
		if c.id != "" {
			startedCampaign.Lock()
			startedCampaign.campaigns = append(startedCampaign.campaigns, c.id)
			startedCampaign.Unlock()
			go run_campaign(c)

		}
	}
}

func nextCampaign() campaign {
	var c campaign
	var launched bytes.Buffer
	startedCampaign.Lock()
	for i, s := range startedCampaign.campaigns {
		if i != 0 {
			launched.WriteString(",")
		}
		launched.WriteString("'" + s + "'")
	}
	startedCampaign.Unlock()
	var query bytes.Buffer
	query.WriteString("SELECT t1.`id`,t3.`email`,t3.`name`,t1.`subject`,t1.`body`,t2.`iface`,t2.`host`,t2.`stream`,t1.`send_unsubscribe`,t2.`resend_delay`,t2.`resend_count` FROM `campaign` t1 INNER JOIN `profile` t2 ON t2.`id`=t1.`profile_id` INNER JOIN `sender` t3 ON t3.`id`=t1.`sender_id` WHERE t1.`accepted`=1 AND (NOW() BETWEEN t1.`start_time` AND t1.`end_time`) AND (SELECT COUNT(*) FROM `recipient` WHERE campaign_id=t1.`id` AND removed=0 AND status IS NULL) > 0")
/* ToDo sqlite
	if models.Config.DbType == "sqlite3" {
		query = "SELECT t1.`id`,t3.`email`,t3.`name`,t1.`subject`,t1.`body`,t2.`iface`,t2.`host`,t2.`stream`,t1.`send_unsubscribe`,t2.`resend_delay`,t2.`resend_count` FROM `campaign` t1 INNER JOIN `profile` t2 ON t2.`id`=t1.`profile_id` INNER JOIN `sender` t3 ON t3.`id`=t1.`sender_id` WHERE t1.`accepted`=1 AND (strftime('%s','now') BETWEEN strftime('%s',t1.`start_time`) AND strftime('%s',t1.`end_time`)) AND (SELECT COUNT(*) FROM `recipient` WHERE campaign_id=t1.`id` AND removed=0 AND status IS NULL) > 0"
	}
	if (models.Config.DbType == "mysql") {
		query = "SELECT t1.`id`,t3.`email`,t3.`name`,t1.`subject`,t1.`body`,t2.`iface`,t2.`host`,t2.`stream`,t1.`send_unsubscribe`,t2.`resend_delay`,t2.`resend_count` FROM `campaign` t1 INNER JOIN `profile` t2 ON t2.`id`=t1.`profile_id` INNER JOIN `sender` t3 ON t3.`id`=t1.`sender_id` WHERE t1.`accepted`=1 AND (NOW() BETWEEN t1.`start_time` AND t1.`end_time`) AND (SELECT COUNT(*) FROM `recipient` WHERE campaign_id=t1.`id` AND removed=0 AND status IS NULL) > 0"
	}
*/
	if launched.String() != "" {
		query.WriteString(" AND t1.`id` NOT IN (" + launched.String() + ")")
	}

	models.Db.QueryRow(query.String()).Scan(
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
	return c
}

func removeStartedCampaign(id string) {
	startedCampaign.Lock()
	for i := range startedCampaign.campaigns {
		if startedCampaign.campaigns[i] == id {
			startedCampaign.campaigns = append(startedCampaign.campaigns[:i], startedCampaign.campaigns[i+1:]...)
			break
		}
	}
	startedCampaign.Unlock()
	return
}

func run_campaign(c campaign) {
	c.getAttachments()
	c.send()
	c.resendSoftBounce()
	removeStartedCampaign(c.id)
	camplog.Println("Finish campaign id", c.id)
}
