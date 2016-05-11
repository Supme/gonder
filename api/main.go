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
	"os"
	"io"
	"github.com/supme/gonder/models"
)

var (
	auth Auth
	apilog *log.Logger
)

func Run()  {
	l, err := os.OpenFile(models.FromRootDir("log/api.log"), os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Println("error opening api log file: %v", err)
	}
	defer l.Close()

	multi := io.MultiWriter(l, os.Stdout)

	apilog = log.New(multi, "", log.Ldate|log.Ltime)

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

	http.HandleFunc("/api/sender", auth.Check(sender))
	http.HandleFunc("/api/senderlist", auth.Check(senderList))

	http.HandleFunc("/preview", auth.Check(getMailPreview))
	http.HandleFunc("/unsubscribe", auth.Check(getUnsubscribePreview))

	http.HandleFunc("/filemanager", auth.Check(filemanager))

	// Static dirs
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(models.FromRootDir("api/http/assets/")))))
	http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(models.FromRootDir("files/")))))
	http.HandleFunc("/{{.StatPng}}", func (w http.ResponseWriter, r *http.Request)  {
		blank, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAMAAAAoLQ9TAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAADUExURUxpcU3H2DoAAAABdFJOUwBA5thmAAAADUlEQVQY02NgGAXIAAABEAAB7JfjegAAAABJRU5ErkJggg==iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAMAAABEpIrGAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAADUExURUxpcU3H2DoAAAABdFJOUwBA5thmAAAAEklEQVQ4y2NgGAWjYBSMAuwAAAQgAAFWu83mAAAAAElFTkSuQmCC")
		w.Write(blank)
	})

	http.HandleFunc("/panel", auth.Check(func (w http.ResponseWriter, r *http.Request)  {
		if f, err := ioutil.ReadFile(models.FromRootDir("api/http/index.html")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else{
			w.Write(f)
		}
	}))

	http.HandleFunc("/status/ws/campaign.log", auth.Check(campaignLog))
	http.HandleFunc("/status/ws/api.log", auth.Check(apiLog))
	http.HandleFunc("/status/ws/statistic.log", auth.Check(statisticLog))
	http.HandleFunc("/status/ws/main.log", auth.Check(mainLog))

	apilog.Println("API listening on port " + models.Config.ApiPort + "...")
	apilog.Fatal(http.ListenAndServeTLS(":" + models.Config.ApiPort, "./cert/server.pem", "./cert/server.key", nil))
}


