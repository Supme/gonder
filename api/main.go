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
	"io/ioutil"
	"encoding/base64"
)

var (
	auth Auth
	Port string
)

func Run()  {

	// Groups
	// Example:
	// Get groups: http://host/api/groups?cmd=get-records&limit=100&offset=0
	// Rename groups: http://host/api/groups?cmd=save-records&selected[]=20&limit=100&offset=0&changes[0][recid]=1&changes[0][name]=Test+1&changes[1][recid]=2&changes[1][name]=Test+2
	// ...
	http.HandleFunc("/api/groups", auth.Check(groups))

	// Campaigns from group
	// Example:
	// Get campaigns: http://host/api/campaigns?group=2&cmd=get-records&limit=100&offset=0
	// Rename campaign: http://host/api/campaigns?cmd=save-records&selected[]=6&limit=100&offset=0&changes[0][recid]=6&changes[0][name]=Test+campaign
	// ...
	http.HandleFunc("/api/campaigns", auth.Check(campaigns))

	// Campaign data
	// Example:
	// Get data: http://host/api/campaign?cmd=get-data&recid=4
	// ...
	http.HandleFunc("/api/campaign", auth.Check(campaign))

	http.HandleFunc("/api/profilelist", auth.Check(profilesList))

	// Profiles
	// Example:
	// Get list http://host/api/profiles?cmd=get-list
	// ...
	http.HandleFunc("/api/profiles", auth.Check(profiles))

	// Get recipients from campaign
	// Example:
	// Get list recipients: http://host/api/recipients?content=recipients&campaign=4&cmd=get-records&limit=100&offset=0
	// Get recipient parameters: http://host/api/recipients?content=parameters&recipient=149358&cmd=get-records&limit=100&offset=0
	// ...
	http.HandleFunc("/api/recipients", auth.Check(recipients))

	http.HandleFunc("/filemanager", auth.Check(filemanager))

	// Static dirs
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./api/http/assets/"))))
	http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir("./files/"))))
	http.HandleFunc("/{{.StatPng}}", func (w http.ResponseWriter, r *http.Request)  {
		blank, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAMAAAAoLQ9TAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAADUExURUxpcU3H2DoAAAABdFJOUwBA5thmAAAADUlEQVQY02NgGAXIAAABEAAB7JfjegAAAABJRU5ErkJggg==iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAMAAABEpIrGAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAADUExURUxpcU3H2DoAAAABdFJOUwBA5thmAAAAEklEQVQ4y2NgGAWjYBSMAuwAAAQgAAFWu83mAAAAAElFTkSuQmCC")
		w.Write(blank)
	})

	http.HandleFunc("/panel", auth.Check(func (w http.ResponseWriter, r *http.Request)  {
		if f, err := ioutil.ReadFile("./api/http/index.html"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else{
			w.Write(f)
		}
	}))

	log.Println("API listening on port " + Port + "...")
	log.Fatal(http.ListenAndServeTLS(":" + Port, "./cert/server.pem", "./cert/server.key", nil))

}
