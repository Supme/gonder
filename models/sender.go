package models

import (
	"errors"
	"github.com/supme/smtpSender"
	"strings"
	"sync"
)

type emailPool struct {
	storage map[string]emailPoolStorage
	sync.RWMutex
}

type emailPoolStorage struct {
	name string
	config      []smtpSender.Config
	resendDelay int
	resendCount int
	pipe        *smtpSender.Pipe
}

var EmailPool emailPool

func InitEmailPool() error {
	EmailPool.Lock()
	defer EmailPool.Unlock()

	EmailPool.storage = map[string]emailPoolStorage{}

	q, err := Db.Query("SELECT `id`,`name`,`iface`,`host`,`stream`,`resend_delay`,`resend_count` FROM `profile` WHERE `host`<>'group'")
	if err != nil {
		return err
	}
	for q.Next() {
		var (
			id          string
			name string
			cnf         smtpSender.Config
			resendDelay int
			resendCount int
		)
		err = q.Scan(
			&id,
			&name,
			&cnf.Iface,
			&cnf.Hostname,
			&cnf.Stream,
			&resendDelay,
			&resendCount,
		)
		if err != nil {
			return err
		}
		pipe := smtpSender.NewPipe(cnf)
		pipe.Start()

		EmailPool.storage[id] = emailPoolStorage{
			name: name,
			config:      []smtpSender.Config{cnf},
			resendDelay: resendDelay,
			resendCount: resendCount,
			pipe:        &pipe,
		}
	}

	q, err = Db.Query("SELECT `id`,`name`,`iface`,`resend_delay`,`resend_count` FROM `profile` WHERE `host`='group'")
	if err != nil {
		return err
	}
	for q.Next() {
		var (
			id          string
			name string
			iface	string
			cnf	[]smtpSender.Config
			resendDelay int
			resendCount int
		)
		err = q.Scan(
			&id,
			&name,
			&iface,
			&resendDelay,
			&resendCount,
		)
		if err != nil {
			return err
		}
		for _, i := range strings.Split(iface, ",") {
			if _, ok := EmailPool.storage[strings.TrimSpace(i)]; ok {
				c := EmailPool.storage[strings.TrimSpace(i)].config
				if len(c) > 0 {
					cnf = append(cnf, c[0])
				}
			}
		}
		pipe := smtpSender.NewPipe(cnf...)
		pipe.Start()

		EmailPool.storage[id] = emailPoolStorage{
			name: name,
		 	config:      cnf,
			resendDelay: resendDelay,
			resendCount: resendCount,
			pipe:        &pipe,
		}
	}

	return nil
}

func (ep emailPool) Get(id string) (*smtpSender.Pipe, error) {
	ep.RLock()
	defer ep.RUnlock()

	if _, ok := ep.storage[id]; ok {
		return ep.storage[id].pipe, nil
	}
	return nil, errors.New("don't have this pipe")
}

func (ep emailPool) Stop(id string){
	ep.Lock()
	if _, ok := ep.storage[id]; ok {
		ep.storage[id].pipe.Stop()
		delete(ep.storage, id)
	}
	ep.Unlock()
}

func (ep emailPool) StopAll(){
	p := ep.List()
	for id := range p {
		ep.Stop(id)
	}
}


func (ep emailPool) List() map[string]string {
	pools := map[string]string{}

	ep.RLock()
	for k, v := range ep.storage {
		pools[k] = v.name
	}
	ep.RUnlock()

	return pools
}
