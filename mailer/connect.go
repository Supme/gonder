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
	"golang.org/x/net/proxy"
	"time"
	"fmt"
)

type (
	connect struct {
		serverMx string
		domain string
		conn net.Conn
	}
)

func (c *connect) up(iface, domain string) (net.Conn, error) {
	var (
		s proxy.Dialer
		n net.Dialer
		err error
	)
	c.domain = domain

	if iface == "" {
		// default interface
		n = net.Dialer{}
	} else {
		if iface[0:8] == "socks://" {
			iface = iface[8:]
			var err error
			s, err = proxy.SOCKS5("tcp", iface, nil, proxy.FromEnvironment())
			if err != nil {
				return c.conn, err
			}
		} else {
			connectAddr := net.ParseIP(iface)
			tcpAddr := &net.TCPAddr{
				IP: connectAddr,
			}
			n = net.Dialer{LocalAddr: tcpAddr}
		}
	}
	start := time.Now()
	record, err := net.LookupMX(c.domain)
	lookupTime := time.Since(start)
	start = time.Now()
	for i := range record {
		smx := net.JoinHostPort(record[i].Host, "25")
		if s != nil {
			c.conn, err = s.Dial("tcp", smx)
		} else {
			c.conn, err = n.Dial("tcp", smx)
		}
		if err == nil {
			c.serverMx = record[i].Host
			connTime := time.Since(start)
			fmt.Printf("Connect time to %s %s. Lookup time %s.\n", c.domain, connTime, lookupTime)
			break
		}
	}

	return c.conn, err
}
