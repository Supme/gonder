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
package api

import (
	"log"
	"regexp"
	"net/url"
)

func parseArrayForm(r url.Values) map[string]map[string]map[string][]string {
	dblarr := make(map[string]map[string]map[string][]string)
	for key,val := range r {
		res, err := regexp.MatchString(`([\w\d])+(\[([\w\d])+\])(\[([\w\d])+\])`, key)
		if  err != nil {
			log.Println(err)
		}
		if res {
			r := regexp.MustCompile(`([\w\d])+`)
			a := r.FindAllString(key,3)
			if dblarr[a[0]] == nil {
				dblarr[a[0]] = make(map[string]map[string][]string)
			}
			if dblarr[a[0]][a[1]] == nil {
				dblarr[a[0]][a[1]] = make(map[string][]string)
			}
			dblarr[a[0]][a[1]][a[2]] = val
		}
	}
	return dblarr
}
