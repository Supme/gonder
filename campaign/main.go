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
package campaign

import (
	"sync"
	"time"
	"log"
)

var (
	MaxCampaingns int
)

func Run()  {
	MaxCampaingns = 2

	var w sync.WaitGroup
	// Работаем в цикле
	for {
		camps := get_active_campaigns(MaxCampaingns)
		for _, camp := range camps {
			w.Add(1)
			log.Println("Start campaign ", camp.id)
			go func(c campaign) {
				c.get_attachments()
				c.send()
				c.resend_soft_bounce()
				w.Done()
			}(camp)
		}
		w.Wait()
		time.Sleep(1 * time.Second)
	}
}
