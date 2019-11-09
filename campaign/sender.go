package campaign

import (
	"bytes"
	"database/sql"
	"github.com/Supme/smtpSender"
	"gonder/models"
	"log"
	"sync"
)

type SendType string

const (
	SendTypeStream SendType = "stream"
	SendTypeResend SendType = "resend"
)

func (s SendType) String() string {
	return string(s)
}

func GetResultFunc(wg *sync.WaitGroup, sendType SendType, campaignID, recipientID, recipientEmail string) func(result smtpSender.Result){
	return func(result smtpSender.Result) {
		var res string
		if result.Err == nil {
			res = "Ok"
		} else {
			res = result.Err.Error()
		}
		err := models.RecipientGetByStringID(result.ID).UpdateRecipientStatus(res)
		if err != nil {
			log.Print(err)
		}
		campLog.Printf("Campaign %s for recipient id %s email %s is %s send time %s", campaignID, recipientID, recipientEmail, res, result.Duration.String())
		models.Prometheus.Campaign.SendResult.WithLabelValues(campaignID, models.GetStatusCodeFromSendResult(result.Err), sendType.String()).Inc()
		wg.Done()
	}
}


type sending struct {
	campaigns map[string]campaign
	mu        sync.RWMutex
}

var Sending sending

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
		defer func() {
			if err := query.Close(); err != nil {
				log.Print(err)
			}
		}()
		for query.Next() {
			var id string
			if err := query.Scan(&id); err != nil {
				log.Print(err)
			}
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
}

func checkErr(err error) {
	if err != nil {
		campLog.Println(err)
	}
}
