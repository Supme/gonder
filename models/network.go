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


