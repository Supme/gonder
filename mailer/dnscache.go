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
//
package mailer

import (
	"net"
	"time"
	"log"
)

type domainMx struct {
		time time.Time
		mxs []*net.MX
		connection int
}

var servers = map[string]domainMx{}

func getMX(domain string) ([]*net.MX, error) {

	if _, ok := servers[domain]; ok || servers[domain].time.Unix() > time.Now().Add(-5 * time.Minute).Unix() {
		log.Printf("Lookup %s from memory\n", domain)
		return servers[domain].mxs, nil
	}
	log.Printf("Lookup %s from net\n", domain)
	m, err := net.LookupMX(domain)
	if err != nil {
		return []*net.MX{}, err
	}
	servers[domain] = domainMx{
		mxs: m,
		time: time.Now(),
	}

	return servers[domain].mxs, nil
}

func upMXconnect(domain string)  {
	s := servers[domain]
	s.connection++
	servers[domain] = s
	log.Printf("%d connections to %s\n", getMXconnect(domain), domain)
}

func downMXconnect(domain string)  {
	s := servers[domain]
	s.connection--
	servers[domain] = s
	log.Printf("%d connections to %s\n", getMXconnect(domain), domain)
}

func getMXconnect(domain string) int {
	if _, ok := servers[domain]; !ok {
		return int(0)
	}
	return servers[domain].connection
}
