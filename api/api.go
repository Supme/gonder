// Package api contains api service and web panel
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
package api

import (
	"encoding/base64"
	"errors"
	"fmt"
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
	user   auth
	apilog *log.Logger
)

// Run start api server
func Run() {
	l, err := os.OpenFile(models.FromRootDir("log/api.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("error opening api log file: %v", err)
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

	// API
	api.HandleFunc("/api/", user.Check(apiRequest))

	// Reports
	api.HandleFunc("/report", user.Check(report))
	api.HandleFunc("/report/status", user.Check(reportCampaignStatus))
	api.HandleFunc("/report/jump", user.Check(reportJumpDetailedCount))
	api.HandleFunc("/report/unsubscribed", user.Check(reportUnsubscribed))

	api.HandleFunc("/preview", user.Check(getMailPreview))
	api.HandleFunc("/unsubscribe", user.Check(getUnsubscribePreview))

	api.HandleFunc("/filemanager", user.Check(filemanager))

	// Static dirs
	api.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(models.FromRootDir("api/http/assets/")))))
	/*
		m := minify.New()
		m.AddFunc("text/css", css.Minify)
		m.AddFunc("text/html", html.Minify)
		m.AddFunc("application/javascript", js.Minify)
		api.Handle("/assets/",m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, path.Join(models.FromRootDir("api/http/"), r.URL.Path))
		})))
	*/
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

	api.HandleFunc("/panel", user.Check(func(w http.ResponseWriter, r *http.Request) {
		if pusher, ok := w.(http.Pusher); ok {
			// Push is supported.
			for _, p := range []string{
				"/assets/jquery/jquery-3.1.1.min.js",
				"/assets/w2ui/w2ui.min.js",
				"/assets/w2ui/w2ui.min.css",
				"/assets/locale/ru-ru.json",
				"/assets/ckeditor/ckeditor.js",
				"/assets/ckeditor/plugins/codemirror/js/codemirror.min.js",
				"/assets/ckeditor/plugins/codemirror/css/codemirror.min.css",
				"/assets/panel/layout.js",
				"/assets/panel/group.js",
				"/assets/panel/sender.js",
				"/assets/panel/campaign.js",
				"/assets/panel/recipient.js",
				"/assets/panel/profile.js",
				"/assets/panel/users.js",
				"/assets/panel/editor.js",
			} {
				if err := pusher.Push(p, nil); err != nil {
					apilog.Printf("Failed to push: %v", err)
				}
			}
		} else {
			apilog.Print("Push not supported")
		}
		if f, err := ioutil.ReadFile(models.FromRootDir("api/http/index.html")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.Write(f)
			//fmt.Fprintf(w, string(f))
		}
	}))

	api.HandleFunc("/logout", user.Logout)

	api.HandleFunc("/status/ws/campaign.log", user.Check(campaignLog))
	api.HandleFunc("/status/ws/api.log", user.Check(apiLog))
	api.HandleFunc("/status/ws/utm.log", user.Check(utmLog))
	api.HandleFunc("/status/ws/main.log", user.Check(mainLog))

	apilog.Println("API listening on port " + models.Config.APIPort + "...")
	apilog.Fatal(http.ListenAndServeTLS(":"+models.Config.APIPort, models.FromRootDir("/cert/server.pem"), models.FromRootDir("/cert/server.key"), api))
}

func apiRequest(w http.ResponseWriter, r *http.Request) {
	var js []byte
	if r.FormValue("request") == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	req, err := parseRequest(r.FormValue("request"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"status": "error", "message": "%s"}`, err)))
		return
	}

	switch r.URL.Path {
	case "/api/users":
		js, err = users(req)
	case "/api/groups":
		js, err = groups(req)
	case "/api/campaign":
		js, err = campaign(req)
	case "/api/campaigns":
		js, err = campaigns(req)
	case "/api/profilelist":
		js, err = profilesList(req)
	case "/api/recipients":
		js, err = recipients(req)
	case "/api/sender":
		js, err = sender(req)
	case "/api/senderlist":
		js, err = senderList(req)
	case "/api/units":
		js, err = units(req)
	case "/api/profiles":
		js, err = profiles(req)
	default:
		err = errors.New("Path not defined")
	}

	if err != nil {
		js = []byte(fmt.Sprintf(`{"status": "error", "message": "%s"}`, err))
	} else if js == nil {
		js = []byte(`{"status": "success", "message": ""}`)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
