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

type sending struct {
	campaigns map[string]Campaign
	sync.RWMutex
}

var (
	Sending sending
	camplog *log.Logger
)

// Run start look database for ready campaign for send
func Run() {
	l, err := os.OpenFile("log/campaign.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("error opening campaign log file: %v", err)
	}
	defer l.Close()
	camplog = log.New(io.MultiWriter(l, os.Stdout), "", log.Ldate|log.Ltime)

	Sending.campaigns = map[string]Campaign{}

	for {
		for Sending.Count() >= models.Config.MaxCampaingns {
			time.Sleep(1 * time.Second)
		}

		Sending.Lock()
		if id, err := Sending.checkNext(); err == nil {
			camp, err := GetCampaign(id)
			checkErr(err)
			Sending.campaigns[id] = camp
			go func() {
				camp.Send()
				Sending.removeStarted(id)
			}()
		}
		Sending.Unlock()
		time.Sleep(10 * time.Second)
	}
}

func (s *sending) Count() int {
	s.Lock()
	defer s.Unlock()
	return len(s.campaigns)
}

func (s *sending) Started() []string {
	started := []string{}

	s.Lock()
	for id := range s.campaigns {
		started = append(started, id)
	}
 	s.Unlock()

	return started
}

func (s *sending) checkNext() (string, error) {
	var launched bytes.Buffer

	i := 0
	for id := range s.campaigns {
		if i != 0 {
			launched.WriteString(",")
		}
		launched.WriteString("'" + id + "'")
		i++
	}

	var query bytes.Buffer
	query.WriteString("SELECT t1.`id` FROM `campaign` t1 WHERE t1.`accepted`=1 AND (NOW() BETWEEN t1.`start_time` AND t1.`end_time`) AND (SELECT COUNT(*) FROM `recipient` WHERE campaign_id=t1.`id` AND removed=0 AND status IS NULL) > 0")
	if launched.String() != "" {
		query.WriteString(" AND t1.`id` NOT IN (" + launched.String() + ")")
	}
	query.WriteString(" LIMIT 1")

	var id string
	err := models.Db.QueryRow(query.String()).Scan(&id)
	if err == sql.ErrNoRows {
		return "", err
	}
	checkErr(err)
	return id, err
}

func (s *sending) removeStarted(id string) {
	s.Lock()
	delete(s.campaigns, id)
	s.Unlock()
	return
}

func checkErr(err error) {
	if err != nil {
		camplog.Println(err)
	}
}
