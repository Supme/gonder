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
	"net/http"
)

func Run()  {
	http.HandleFunc("/api/groups", groups)
	http.HandleFunc("/api/campaigns", campaigns)
	http.HandleFunc("/api/campaign", campaign)
	http.HandleFunc("/api/profiles", profiles)
	http.Handle("/files/", http.FileServer(http.Dir("./files/")))
	http.Handle("/", http.FileServer(http.Dir("./api/http/")))

	log.Println("API listening...")
	http.ListenAndServe(":3000", nil)

}


