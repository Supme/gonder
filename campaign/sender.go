package campaign

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/supme/gonder/models"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

type sending struct {
	campaigns map[string]campaign
	mu        sync.RWMutex
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
		time.Sleep(time.Second*10)
	}
}

func (s *sending) add(c campaign) {
	s.mu.Lock()
	s.campaigns[c.ID] = c
	s.mu.Unlock()
}

func (s *sending) Stop(id ...string) {
	for i := range id {
		s.stop(id[i])
	}
}

func (s *sending) stop(id string) {
	s.mu.Lock()
	if _, ok := s.campaigns[id]; ok {
		close(s.campaigns[id].Stop)
		<-s.campaigns[id].Finish
		delete(s.campaigns, id)
	}
	s.mu.Unlock()
}

func (s *sending) StopAll() {
	started := s.Started()
	for i := range started {
		s.stop(started[i])
	}
}

func (s *sending) Count() int {
	var count int
	s.mu.RLock()
	count = len(s.campaigns)
	s.mu.RUnlock()
	return count
}

func (s *sending) Started() []string {
	started := []string{}
	s.mu.RLock()
	for id := range s.campaigns {
		started = append(started, id)
	}
	s.mu.RUnlock()
	return started
}

func (s *sending) checkExpired() ([]string, error) {
	var (
		expired  []string
		launched bytes.Buffer
	)
	i := 0
	s.mu.RLock()
	for id := range s.campaigns {
		if i != 0 {
			launched.WriteString(",")
		}
		launched.WriteString("'" + id + "'")
		i++
	}
	s.mu.RUnlock()
	if launched.String() != "" {
		query, err := models.Db.Query("SELECT `id` FROM `campaign` WHERE `id` IN (" + launched.String() + ") AND NOW()>=`end_time`")
		if err != nil {
			return expired, err
		}
		defer query.Close()
		for query.Next() {
			var id string
			query.Scan(&id)
			expired = append(expired, id)
		}
	}
	return expired, nil
}

// ToDo return slice campaign id
func (s *sending) checkNext() (string, error) {
	var launched bytes.Buffer

	i := 0
	s.mu.RLock()
	for id := range s.campaigns {
		if i != 0 {
			launched.WriteString(",")
		}
		launched.WriteString("'" + id + "'")
		i++
	}
	s.mu.RUnlock()
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
	s.mu.Lock()
	if _, ok := s.campaigns[id]; ok {
		delete(s.campaigns, id)
	}
	s.mu.Unlock()
	return
}

func checkErr(err error) {
	if err != nil {
		camplog.Println(err)
	}
}
