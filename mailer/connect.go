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
)

type (
	Connect struct {
		ServerMx string
		domain string
		conn net.Conn
	}
)

func (c *Connect) Up(iface, domain string) (net.Conn, error) {
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

	record, err := getMX(c.domain)
	for i := range record {
		smx := net.JoinHostPort(record[i].Host, "25")
		if s != nil {
			c.conn, err = s.Dial("tcp", smx)
		} else {
			c.conn, err = n.Dial("tcp", smx)
		}
		if err == nil {
			c.ServerMx = record[i].Host
			break
		}
	}

	if err == nil {
		upMXconnect(c.domain)
	}

	return c.conn, err
}

func (c *Connect) Close(){
	c.conn.Close()
	downMXconnect(c.domain)
}
