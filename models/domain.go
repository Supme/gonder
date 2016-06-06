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
	"fmt"
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
		return true
	}
	r := connNow >= connMax - 1
	fmt.Printf("Domain %s have %d connections\n",domain, DomainGetConnNow(host, domain))
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

func DomainGetMX(domain string) ([]*net.MX, error) {
	var record []*net.MX
	var err error
	mx.Lock()
	if _, ok := mx.stor[domain]; ok {
		if time.Since(mx.stor[domain].update) < 15 * time.Minute {
			record = mx.stor[domain].records
			mx.Unlock()
			return record, nil
		}
	}

	record, err = net.LookupMX(domain)
	mx.stor[domain] = mxStor{
		records: record,
		update:time.Now(),
	}
	//mx.stor[domain].records = record
	//mx.stor[domain].update = time.Now()
	mx.Unlock()
/*
	conn, i := domainGet(domain)
	if len(conn.mx) != 0 || conn.mxRefresh.Unix() > time.Now().Add(-5 * time.Minute).Unix() {
		record = conn.mx
	} else {
		record, err = net.LookupMX(domain)
		domains.mu.Lock()
		domains.conn[i].mx = record
		domains.conn[i].mxRefresh = time.Now()
		domains.mu.Unlock()
	}
	fmt.Print(record)
*/
	return record, err
}



