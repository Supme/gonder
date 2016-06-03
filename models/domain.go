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
)

type (
	domainConn struct{
		now int
		max int
	}

	domainStor struct {
		name map[string]int
		conn []domainConn
		mu sync.Mutex
	}
)

var domains domainStor

func domainGet(domain string) (domainConn, int) {
	domains.mu.Lock()
	if _, ok := domains.name[domain]; !ok {
		domains.conn = append(domains.conn, domainConn{now:0, max:50})
		domains.name[domain] = len(domains.conn) - 1
	}
	i := domains.name[domain]
	r := domains.conn[i]
	domains.mu.Unlock()
	return r, i
}

func DomainMaxConn(domain string) bool {
	conn, _ := domainGet(domain)
	r := conn.now >= conn.max
	return r
}

func DomainGetConn(domain string) int {
	conn, _ := domainGet(domain)
	r := conn.now
	return r
}

func DomainUpConn(domain string) {
	_, i := domainGet(domain)
	domains.mu.Lock()
	domains.conn[i].now++
	domains.mu.Unlock()
}

func DomainDownConn(domain string) {
	_, i := domainGet(domain)
	domains.mu.Lock()
	domains.conn[i].now--
	domains.mu.Unlock()
}


