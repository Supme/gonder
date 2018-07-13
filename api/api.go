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
	"github.com/NYTimes/gziphandler"
	"github.com/supme/gonder/models"
	"github.com/tdewolff/minify"
	minifyCSS "github.com/tdewolff/minify/css"
	minifyHTML "github.com/tdewolff/minify/html"
	minifyJS "github.com/tdewolff/minify/js"
	minifyJSON "github.com/tdewolff/minify/json"
	"html/template"
	"io"
		"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
)

var (
	user   auth
	apilog *log.Logger
	min    *minify.M
)

const (
	useGZIP = true

	useMinify     = false
	useMinifyCSS  = true
	useMinifyHTML = true
	useMinifyJS   = true
	useMinifyJSON = true
)

func init() {
	min = minify.New()
	if useMinifyCSS {
		min.AddFunc("text/css", minifyCSS.Minify)
	}
	if useMinifyHTML {
		min.AddFunc("text/html", minifyHTML.Minify)
	}
	if useMinifyJS {
		min.AddFunc("application/javascript", minifyJS.Minify)
	}
	if useMinifyJSON {
		min.AddFunc("application/json", minifyJSON.Minify)
	}
}

func apiHandler(fn http.HandlerFunc, checkAuth bool) http.Handler {
	var handler http.Handler
	if checkAuth {
		handler = user.Check(fn)
	} else {
		handler = fn
	}
	if useMinify {
		handler = min.Middleware(handler)
	}
	if useGZIP {
		handler = gziphandler.GzipHandler(handler)
	}
	return handler
}

// Run start api server
func Run() {
	l, err := os.OpenFile(models.FromRootDir("log/api.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("error opening api log file: %v", err)
	}
	defer l.Close()
	apilog = log.New(io.MultiWriter(l, os.Stdout), "", log.Ldate|log.Ltime)

	api := http.NewServeMux()

	api.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mem := new(runtime.MemStats)
		runtime.ReadMemStats(mem)
		w.Write([]byte("Welcome to San Tropez! (Conn: " + strconv.Itoa(models.Db.Stats().OpenConnections) + " Allocate: " + strconv.FormatUint(mem.Alloc, 10) + ")"))
	})

	// API
	//api.HandleFunc("/api/", user.Check(apiRequest))
	api.Handle("/api/", apiHandler(apiRequest, true))

	// Reports
	api.Handle("/report", apiHandler(report, true))
	api.Handle("/report/status", apiHandler(reportCampaignStatus, true))
	api.Handle("/report/jump", apiHandler(reportJumpDetailedCount, true))
	api.Handle("/report/unsubscribed", apiHandler(reportUnsubscribed, true))

	api.Handle("/preview", apiHandler(getMailPreview, true))
	api.Handle("/unsubscribe", apiHandler(getUnsubscribePreview, true))

	// Assets static dirs
	api.Handle("/assets/", apiHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", 3600 * 24 * 7))
			http.ServeFile(w, r, path.Join(models.FromRootDir("panel/"),
				r.URL.Path))
		}), false))

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

	api.Handle("/panel", user.Check(func(w http.ResponseWriter, r *http.Request) {
		if pusher, ok := w.(http.Pusher); ok {
			// Push is supported.
			for _, p := range []string{
				"/assets/jquery/jquery-3.1.1.min.js?" + models.Config.Version,
				"/assets/w2ui/w2ui.min.js?" + models.Config.Version,
				"/assets/w2ui/w2ui.min.css?" + models.Config.Version,
				"/assets/w2ui/locale/ru-ru.json?" + models.Config.Version,
				"/assets/panel/locale/ru-ru.json?" + models.Config.Version,
				"/assets/panel/layout.js?" + models.Config.Version,
				"/assets/panel/group.js?" + models.Config.Version,
				"/assets/panel/sender.js?" + models.Config.Version,
				"/assets/panel/campaign.js?" + models.Config.Version,
				"/assets/panel/recipient.js?" + models.Config.Version,
				"/assets/panel/profile.js?" + models.Config.Version,
				"/assets/panel/users.js?" + models.Config.Version,
				"/assets/panel/template.js?" + models.Config.Version,
			} {
				if err := pusher.Push(p, nil); err != nil {
					apilog.Printf("Failed to push: %v", err)
				}
			}
		} else {
			apilog.Print("Push not supported")
		}

		tmpl := template.Must(template.ParseFiles(models.FromRootDir("panel/index.html")))
		if err != nil {
			apilog.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = tmpl.Execute(w, map[string]string{
			"version": models.Config.Version,
		})
		if err != nil {
			apilog.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")

	req, err := parseRequest(r.FormValue("request"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"status": "error", "message": %s}`, strconv.Quote(err.Error()))))
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
		js = []byte(fmt.Sprintf(`{"status": "error", "message": %s}`, strconv.Quote(err.Error())))
	} else if js == nil || string(js) == "" {
		js = []byte(`{"status": "success", "message": ""}`)
	}
	w.Write(js)
}
