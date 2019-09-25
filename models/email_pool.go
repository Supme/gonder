package models

import (
	"errors"
	"fmt"
	"github.com/Supme/smtpSender"
	"strconv"
	"strings"
	"sync"
)

type emailPool struct {
	storage map[int]emailPoolStorage
	sync.RWMutex
}

type emailPoolStorage struct {
	name        string
	config      []smtpSender.Config
	resendDelay int
	resendCount int
	pipe        *smtpSender.Pipe
}

var EmailPool emailPool

type profile struct {
	id          int
	name        string
	hostname    string
	iface       string
	stream      int
	resendCount int
	resendDelay int
}

func initPool(profiles []profile) error {
	EmailPool.Lock()
	EmailPool.storage = map[int]emailPoolStorage{}

	// create interfaces pool
	for i := range profiles {
		if profiles[i].hostname == "group" {
			continue
		}

		var cnf smtpSender.Config

		cnf.Iface = profiles[i].iface
		cnf.Hostname = profiles[i].hostname
		cnf.Stream = profiles[i].stream

		pipe := smtpSender.NewPipe(cnf)
		pipe.Start()

		EmailPool.storage[profiles[i].id] = emailPoolStorage{
			name:        profiles[i].name,
			config:      []smtpSender.Config{cnf},
			resendDelay: profiles[i].resendDelay,
			resendCount: profiles[i].resendCount,
			pipe:        pipe,
		}
	}

	// create group pool
	for i := range profiles {
		if profiles[i].hostname != "group" {
			continue
		}

		var cnf []smtpSender.Config

		for _, i := range strings.Split(profiles[i].iface, ",") {
			n, err := strconv.Atoi(i)
			if err != nil {
				return fmt.Errorf("convert group interfaces id error: %s", err)
			}
			if _, ok := EmailPool.storage[n]; ok {
				c := EmailPool.storage[n].config
				if len(c) > 0 {
					cnf = append(cnf, c[0])
				}
			}
		}
		pipe := smtpSender.NewPipe(cnf...)
		pipe.Start()

		EmailPool.storage[profiles[i].id] = emailPoolStorage{
			name:        profiles[i].name,
			config:      cnf,
			resendDelay: profiles[i].resendDelay,
			resendCount: profiles[i].resendCount,
			pipe:        pipe,
		}
	}

	EmailPool.Unlock()

	return nil
}

func (ep *emailPool) Get(id int) (*smtpSender.Pipe, error) {
	ep.RLock()
	defer ep.RUnlock()

	if _, ok := ep.storage[id]; ok {
		return ep.storage[id].pipe, nil
	}
	return nil, errors.New("don't have this pipe")
}

func (ep *emailPool) Stop(id int) {
	ep.Lock()
	if _, ok := ep.storage[id]; ok {
		ep.storage[id].pipe.Stop()
		delete(ep.storage, id)
	}
	ep.Unlock()
}

func (ep *emailPool) StopAll() {
	p := ep.List()
	for id := range p {
		ep.Stop(id)
	}
}

// GetResendParams return delay, count, error for poll id
func (ep *emailPool) GetResendParams(id int) (int, int, error) {
	ep.RLock()
	defer ep.RUnlock()

	if _, ok := ep.storage[id]; ok {
		return ep.storage[id].resendDelay, ep.storage[id].resendCount, nil
	}
	return 0, 0, errors.New("don't have this pipe")
}

func (ep *emailPool) List() map[int]string {
	pools := map[int]string{}
	ep.RLock()
	for k, v := range ep.storage {
		pools[k] = v.name
	}
	ep.RUnlock()
	return pools
}
