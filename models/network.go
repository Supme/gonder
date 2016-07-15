// Project gonder.
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
package models

import (
	"sync"
	"net"
	"time"
	"log"
	"strings"
)

type mxStor struct {
	records []*net.MX
	update time.Time
}

var mx = struct {
	stor map[string]mxStor
	sync.Mutex
} {
	stor: make(map[string]mxStor),
}

func DomainGetMX(domain string) ([]*net.MX, error) {
	var (
		record []*net.MX
		err error
	)

	if Config.DnsCache {
		mx.Lock()
		defer mx.Unlock()
		if _, ok := mx.stor[domain]; !ok || time.Since(mx.stor[domain].update) > 15 * time.Minute {
			record, err = net.LookupMX(domain)
			if err == nil {
				mx.stor[domain] = mxStor{
					records: record,
					update:time.Now(),
				}
			}
		} else {
			record = mx.stor[domain].records
		}
	} else {
		record, err = net.LookupMX(domain)
	}

	return record, err
}


type (
	NetStore struct {
		params map[int]profile
	}

	profile struct {
		iface, host string
		streamMax int
		streamNow int
		sync.Mutex
		free chan bool
	}
)

var NetProfile NetStore

func (n *NetStore) Update() {
	var (
		iface, host string
		id, stream  int
	)
	query, err := Db.Query("SELECT `id`,`iface`,`host`,`stream` FROM `profile`")
	if err != nil {
		log.Print(err)
	}
	defer query.Close()
	for query.Next() {
		iface, host = "", ""
		id, stream  = 0, 0
		err = query.Scan(&id, &iface, &host, &stream)
		if err != nil {
			log.Print(err)
		}
		n.params[id].Lock()
		n.params[id].iface, n.params[id].host, n.params[id].streamMax = iface, host, stream
		n.params[id].Unlock()
	}
}

func (n *NetStore) next(id int) (string, string) {
	n.params[id].Lock()
	if n.params[id].streamMax > n.params[id].streamNow {
		n.params[id].streamNow++
		n.params[id].Unlock()
		return n.params[id].iface, n.params[id].host
	}
	<-n.params[id].free
	n.params[id].Unlock()
	return n.params[id].iface, n.params[id].host
}

var netRotate struct{
	store map [int]int
	sync.Mutex
}

func (n *NetStore) Get(id int) (iface, host string) {
	iface, host = n.next(id)
	if iface[0:5] == "group" {
		netRotate.Lock()
		p := strings.Split(host, ",")
		netRotate.store[id]++
		if len(p)-1 > netRotate.store[id] {
			netRotate.store[id] = 0
		}
		iface, host = n.next(p[netRotate.store[id]])
		netRotate.Unlock()
	}
	return
}

