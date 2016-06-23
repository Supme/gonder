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
)

type (
	domainConn struct{
		now map[string]int
		max int
		update time.Time
	}
	connectStor struct {
		name map[string]int
		conn []domainConn
	}

	mxStor struct {
		records []*net.MX
		update time.Time
	}
)

var domains struct {
	connect connectStor
	sync.Mutex
}

var mx struct {
	stor map[string]mxStor
	sync.Mutex
}

func domainInit()  {
	domains.connect = connectStor{}
	domains.connect.conn = []domainConn{}
	domains.connect.name = map[string]int{}
	mx.stor = map[string]mxStor{}
}


func DomainGetMX(domain string) ([]*net.MX, error) {
	var record []*net.MX
	var err error
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

	return record, err
}

type (
	networkConnect struct {
		iface string
		host string
	}

	Network struct {
		connect []networkConnect
		i int
		sync.Mutex
	}
)

func (n *Network) Init(profileId int) error {
	var t networkConnect
	row, err := Db.Query("SELECT iface, host FROM network WHERE profile_id=?", profileId)
	if err != nil {
		return err
	}
	defer row.Close()

	n.Lock()
	defer n.Unlock()
	for row.Next() {
		row.Scan(t.iface, t.host)
		n.connect = append(n.connect, t)
	}
	n.i = 0
	return nil
}

func (n *Network) Next() (iface, host string) {
	n.Lock()
	defer n.Unlock()
	iface, host = n.connect[n.i].iface, n.connect[n.i].host
	if n.i++; n.i >= len(n.connect) { n.i = 0 }
	return
}

func domainGet(host, domain string) (connNow, connMax , i int) {
	domains.Lock()
	// check isset connect to domain
	if _, ok := domains.connect.name[domain]; !ok {
		domains.connect.conn = append(domains.connect.conn, domainConn{ max:0, update: time.Now() })
		domains.connect.name[domain] = len(domains.connect.conn) - 1
		// i is number record for domain
		i = domains.connect.name[domain]
		if domains.connect.conn[i].now == nil {
			domains.connect.conn[i].now = map[string]int{}
		}
		domains.connect.conn[i].now[host] = 0
	}
	connNow = domains.connect.conn[i].now[host]
	if time.Since(domains.connect.conn[i].update) > 15 * time.Minute {
		domains.connect.conn[i].max++
	}
	connMax = domains.connect.conn[i].max
	domains.Unlock()
	return connNow, connMax, i
}

// Get what have a free connection from host to domain
func DomainMaxConn(host, domain string) bool {
	connNow, connMax, _ := domainGet(host, domain)
	if connMax == 0 {
		return false
	}
	r := connNow <= connMax - 1
	return r
}

// Get count now connection from host to domain
func DomainGetConnNow(host, domain string) int {
	connNow, _, _ := domainGet(host, domain)
	r := connNow
	return r
}

// Up connection from host to domain
func DomainUpConn(host, domain string) {
	_, _, i := domainGet(host, domain)
	domains.Lock()
	domains.connect.conn[i].now[host]++
	domains.Unlock()
}

// Down connection from host to domain
func DomainDownConn(host, domain string) {
	_, _, i := domainGet(host, domain)
	domains.Lock()
	domains.connect.conn[i].now[host]--
	domains.Unlock()
}

func DomainDownMax(host, domain string) {
	_, _, i := domainGet(host, domain)
	domains.Lock()
	if domains.connect.conn[i].now[host] > 2 {
		domains.connect.conn[i].max = domains.connect.conn[i].now[host] - 1
	}
	domains.Unlock()
}
