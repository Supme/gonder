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
	"encoding/base64"
	"github.com/supme/gonder/models"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
)

var (
	auth   Auth
	apilog *log.Logger
)

func Run() {
	l, err := os.OpenFile(models.FromRootDir("log/api.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println("error opening api log file: %v", err)
	}
	defer l.Close()

	multi := io.MultiWriter(l, os.Stdout)

	apilog = log.New(multi, "", log.Ldate|log.Ltime)

	api := http.NewServeMux()

	api.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mem := new(runtime.MemStats)
		runtime.ReadMemStats(mem)
		w.Write([]byte("Welcome to San Tropez! (Conn: " + strconv.Itoa(models.Db.Stats().OpenConnections) + " Allocate: " + strconv.FormatUint(mem.Alloc, 10) + ")"))
	})

	// Groups
	// Example:
	// Get groups: http://host/api/groups?cmd=get-records&limit=100&offset=0
	// Rename groups: http://host/api/groups?cmd=save-records&selected[]=20&limit=100&offset=0&changes[0][recid]=1&changes[0][name]=Test+1&changes[1][recid]=2&changes[1][name]=Test+2
	// ...
	api.HandleFunc("/api/groups", auth.Check(groups))

	// Campaigns from group
	// Example:
	// Get campaigns: http://host/api/campaigns?group=2&cmd=get-records&limit=100&offset=0
	// Rename campaign: http://host/api/campaigns?cmd=save-records&selected[]=6&limit=100&offset=0&changes[0][recid]=6&changes[0][name]=Test+campaign
	// ...
	api.HandleFunc("/api/campaigns", auth.Check(campaigns))

	// Campaign data
	// Example:
	// Get data: http://host/api/campaign?cmd=get-data&recid=4
	// ...
	api.HandleFunc("/api/campaign", auth.Check(campaign))

	api.HandleFunc("/api/profilelist", auth.Check(profilesList))

	// Profiles
	// Example:
	// Get list http://host/api/profiles?cmd=get-list
	// ...
	api.HandleFunc("/api/profiles", auth.Check(profiles))

	// Get recipients from campaign
	// Example:
	// Get list recipients: http://host/api/recipients?content=recipients&campaign=4&cmd=get-records&limit=100&offset=0
	// Get recipient parameters: http://host/api/recipients?content=parameters&recipient=149358&cmd=get-records&limit=100&offset=0
	// ...
	api.HandleFunc("/api/recipients", auth.Check(recipients))

	api.HandleFunc("/api/sender", auth.Check(sender))
	api.HandleFunc("/api/senderlist", auth.Check(senderList))

	// Reports
	// Example:
	//   Get reports: http://host/api/report?campaign=4
	// /api/report?campaign=xxx
	// /api/report/jump?campaign=xxx
	// /api/report/unsubscribed?group=xxx
	// /api/report/unsubscribed?campaign=xxx
	api.HandleFunc("/api/report", auth.Check(report))
	api.HandleFunc("/api/report/jump", auth.Check(reportJumpDetailedCount))
	api.HandleFunc("/api/report/unsubscribed", auth.Check(reportUnsubscribed))

	api.HandleFunc("/preview", auth.Check(getMailPreview))
	api.HandleFunc("/unsubscribe", auth.Check(getUnsubscribePreview))

	api.HandleFunc("/filemanager", auth.Check(filemanager))

	// Static dirs
	api.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(models.FromRootDir("api/http/assets/")))))
	api.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(models.FromRootDir("files/")))))

	api.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/x-icon")
		ico, _ := base64.StdEncoding.DecodeString("AAABAAEAEBAAAAEAIABoBAAAFgAAACgAAAAQAAAAIAAAAAEAIAAAAAAAAAQAABILAAASCwAAAAAAAAAAAAByGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/8q2uP9yGSL/yra4/3IZIv/Ktrj/yra4/3IZIv9yGSL/yra4/8q2uP9yGSL/yra4/3IZIv/Ktrj/chki/3IZIv/Ktrj/chki/+je3/9yGSL/yra4/3IZIv/Ktrj/chki/8q2uP9yGSL/chki/8q2uP9yGSL/yra4/3IZIv9yGSL/yra4/+je3//Ktrj/chki/8q2uP9yGSL/yra4/3IZIv/Ktrj/yra4/3IZIv/Ktrj/yra4/3IZIv9yGSL/chki/+je3/9yGSL/yra4/3IZIv/Ktrj/chki/8q2uP9yGSL/yra4/3IZIv9yGSL/yra4/3IZIv/Ktrj/chki/3IZIv/Ktrj/chki/8q2uP9yGSL/yra4/8q2uP9yGSL/chki/8q2uP/Ktrj/chki/8q2uP/Ktrj/yra4/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/+je3//o3t//6N7f/+je3//o3t//6N7f/+je3/9yGSL/6N7f/+je3//o3t//6N7f/+je3//o3t//chki/3IZIv/o3t//yra4/8q2uP/Ktrj/yra4/8q2uP/Ktrj/chki/8q2uP/Ktrj/yra4/8q2uP/Ktrj/6N7f/3IZIv9yGSL/6N7f/8q2uP9yGSL/chki/3IZIv/Ktrj/6N7f/3IZIv/Ktrj/yra4/3IZIv9yGSL/yra4/+je3/9yGSL/chki/+je3//Ktrj/chki/8q2uP/o3t//6N7f/8q2uP9yGSL/6N7f/+je3/9yGSL/chki/8q2uP/o3t//chki/3IZIv/o3t//yra4/3IZIv/o3t//yra4/8q2uP/Ktrj/chki/8q2uP/Ktrj/chki/3IZIv/Ktrj/6N7f/3IZIv9yGSL/6N7f/+je3/9yGSL/chki/3IZIv9yGSL/chki/3IZIv/Ktrj/yra4/8q2uP/Ktrj/6N7f/+je3/9yGSL/chki/+je3//o3t//6N7f/+je3//o3t//6N7f/+je3/9yGSL/6N7f/+je3//o3t//6N7f/+je3//o3t//chki/3IZIv/Ktrj/yra4/8q2uP/Ktrj/yra4/8q2uP/Ktrj/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==")
		w.Write(ico)
	})

	api.HandleFunc("/{{.StatPng}}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/gif")
		blank, _ := base64.StdEncoding.DecodeString("R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7")
		w.Write(blank)
	})

	api.HandleFunc("/panel", auth.Check(func(w http.ResponseWriter, r *http.Request) {
		if f, err := ioutil.ReadFile(models.FromRootDir("api/http/index.html")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.Write(f)
		}
	}))

	api.HandleFunc("/logout", auth.Logout)

	api.HandleFunc("/status/ws/campaign.log", auth.Check(campaignLog))
	api.HandleFunc("/status/ws/api.log", auth.Check(apiLog))
	api.HandleFunc("/status/ws/utm.log", auth.Check(utmLog))
	api.HandleFunc("/status/ws/main.log", auth.Check(mainLog))

	apilog.Println("API listening on port " + models.Config.ApiPort + "...")
	apilog.Fatal(http.ListenAndServeTLS(":"+models.Config.ApiPort, "./cert/server.pem", "./cert/server.key", api))
}
