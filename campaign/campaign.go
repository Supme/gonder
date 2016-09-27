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
	"bytes"
	"database/sql"
	"github.com/supme/gonder/models"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

var (
	startedCampaign struct {
		campaigns []string
		sync.Mutex
	}
	campSum int
	camplog *log.Logger
)

// Start look database for ready campaign for send
func Run() {
	l, err := os.OpenFile("log/campaign.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
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

		startedCampaign.Lock()
		if id, err := checkNextCampaign(); err == nil {
			startedCampaign.campaigns = append(startedCampaign.campaigns, id)
			c := campaign{}
			go func() {
				camplog.Printf("Start campaign id %s.", id)
				c.run(id)
				removeStartedCampaign(id)
				camplog.Printf("Finish campaign id %s", id)
			}()
		}
		startedCampaign.Unlock()
		time.Sleep(10 * time.Second)
	}
}

func checkNextCampaign() (string, error) {
	var launched bytes.Buffer
	for i, s := range startedCampaign.campaigns {
		if i != 0 {
			launched.WriteString(",")
		}
		launched.WriteString("'" + s + "'")
	}
	var query bytes.Buffer
	query.WriteString("SELECT t1.`id` FROM `campaign` t1 WHERE t1.`accepted`=1 AND (NOW() BETWEEN t1.`start_time` AND t1.`end_time`) AND (SELECT COUNT(*) FROM `recipient` WHERE campaign_id=t1.`id` AND removed=0 AND status IS NULL) > 0")
	if launched.String() != "" {
		query.WriteString(" AND t1.`id` NOT IN (" + launched.String() + ")")
	}
	var id string
	err := models.Db.QueryRow(query.String()).Scan(&id)
	if err == sql.ErrNoRows {
		return "", err
	}
	checkErr(err)
	return id, err
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

func checkErr(err error) {
	if err != nil {
		camplog.Println(err)
	}
}
