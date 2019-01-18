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
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
)

var (
	user   auth
	apilog *log.Logger
	min    *minify.M
	lang   *languages
)

const (
	useGZIP = true

	useMinify     = true
	useMinifyCSS  = true
	useMinifyHTML = true
	useMinifyJS   = true
	useMinifyJSON = true

	panelRoot   = "/panel"
	panelLocale = "ru-ru"
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
		log.Printf("error opening api log file: %s", err)
	}
	defer l.Close()
	apilog = log.New(io.MultiWriter(l, os.Stdout), "", log.Ldate|log.Ltime)

	lang, err = newLang(
		models.FromRootDir("panel/assets/w2ui/locale/*.json"),
		models.FromRootDir("panel/assets/gonder/locale/*.json"),
	)
	if err != nil {
		apilog.Printf("error loading languages: %s", err)
	}

	api := http.NewServeMux()

	api.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mem := new(runtime.MemStats)
		runtime.ReadMemStats(mem)
		_, err = w.Write([]byte("Welcome to San Tropez! (Conn: " + strconv.Itoa(models.Db.Stats().OpenConnections) + " Allocate: " + strconv.FormatUint(mem.Alloc, 10) + ")"))
		if err != nil {
			log.Println(err)
		}
	})

	// API
	api.Handle("/api/", apiHandler(apiRequest, true))

	// Reports
	api.Handle("/report", apiHandler(report, true))
	api.Handle("/report/status", apiHandler(reportCampaignStatus, true))
	api.Handle("/report/jump", apiHandler(reportJumpDetailedCount, true))
	api.Handle("/report/unsubscribed", apiHandler(reportUnsubscribed, true))
	api.Handle("/report/recipients", apiHandler(reportRecipientsCsv, true))

	api.Handle("/preview", apiHandler(getMailPreview, true))
	api.Handle("/unsubscribe", apiHandler(getUnsubscribePreview, true))

	api.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(models.FromRootDir("files/")))))

	api.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/x-icon")
		ico, _ := base64.StdEncoding.DecodeString("AAABAAEAEBAAAAEAIABoBAAAFgAAACgAAAAQAAAAIAAAAAEAIAAAAAAAAAQAABILAAASCwAAAAAAAAAAAAByGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/8q2uP9yGSL/yra4/3IZIv/Ktrj/yra4/3IZIv9yGSL/yra4/8q2uP9yGSL/yra4/3IZIv/Ktrj/chki/3IZIv/Ktrj/chki/+je3/9yGSL/yra4/3IZIv/Ktrj/chki/8q2uP9yGSL/chki/8q2uP9yGSL/yra4/3IZIv9yGSL/yra4/+je3//Ktrj/chki/8q2uP9yGSL/yra4/3IZIv/Ktrj/yra4/3IZIv/Ktrj/yra4/3IZIv9yGSL/chki/+je3/9yGSL/yra4/3IZIv/Ktrj/chki/8q2uP9yGSL/yra4/3IZIv9yGSL/yra4/3IZIv/Ktrj/chki/3IZIv/Ktrj/chki/8q2uP9yGSL/yra4/8q2uP9yGSL/chki/8q2uP/Ktrj/chki/8q2uP/Ktrj/yra4/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/+je3//o3t//6N7f/+je3//o3t//6N7f/+je3/9yGSL/6N7f/+je3//o3t//6N7f/+je3//o3t//chki/3IZIv/o3t//yra4/8q2uP/Ktrj/yra4/8q2uP/Ktrj/chki/8q2uP/Ktrj/yra4/8q2uP/Ktrj/6N7f/3IZIv9yGSL/6N7f/8q2uP9yGSL/chki/3IZIv/Ktrj/6N7f/3IZIv/Ktrj/yra4/3IZIv9yGSL/yra4/+je3/9yGSL/chki/+je3//Ktrj/chki/8q2uP/o3t//6N7f/8q2uP9yGSL/6N7f/+je3/9yGSL/chki/8q2uP/o3t//chki/3IZIv/o3t//yra4/3IZIv/o3t//yra4/8q2uP/Ktrj/chki/8q2uP/Ktrj/chki/3IZIv/Ktrj/6N7f/3IZIv9yGSL/6N7f/+je3/9yGSL/chki/3IZIv9yGSL/chki/3IZIv/Ktrj/yra4/8q2uP/Ktrj/6N7f/+je3/9yGSL/chki/+je3//o3t//6N7f/+je3//o3t//6N7f/+je3/9yGSL/6N7f/+je3//o3t//6N7f/+je3//o3t//chki/3IZIv/Ktrj/yra4/8q2uP/Ktrj/yra4/8q2uP/Ktrj/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==")
		_, err = w.Write(ico)
		if err != nil {
			log.Println(err)
		}
	})

	api.HandleFunc("/{{.StatPng}}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/gif")
		blank, _ := base64.StdEncoding.DecodeString("R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7")
		_, err = w.Write(blank)
		if err != nil {
			log.Println(err)
		}
	})

	api.HandleFunc("/logout", user.Logout)

	api.HandleFunc("/status/ws/campaign.log", user.Check(campaignLog))
	api.HandleFunc("/status/ws/api.log", user.Check(apiLog))
	api.HandleFunc("/status/ws/utm.log", user.Check(utmLog))
	api.HandleFunc("/status/ws/main.log", user.Check(mainLog))

	api.Handle(panelRoot, apiHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if pusher, ok := w.(http.Pusher); ok {
				// Push is supported.
				for _, p := range []string{
					panelRoot + "/assets/jquery/jquery-3.1.1.min.js?" + models.Config.Version,
					panelRoot + "/assets/w2ui/w2ui.min.js?" + models.Config.Version,
					panelRoot + "/assets/w2ui/w2ui.min.css?" + models.Config.Version,
					panelRoot + "/assets/panel/layout.js?" + models.Config.Version,
					panelRoot + "/assets/panel/group.js?" + models.Config.Version,
					panelRoot + "/assets/panel/sender.js?" + models.Config.Version,
					panelRoot + "/assets/panel/campaign.js?" + models.Config.Version,
					panelRoot + "/assets/panel/recipient.js?" + models.Config.Version,
					panelRoot + "/assets/panel/profile.js?" + models.Config.Version,
					panelRoot + "/assets/panel/users.js?" + models.Config.Version,
					panelRoot + "/assets/panel/template.js?" + models.Config.Version,
				} {
					if err := pusher.Push(p, nil); err != nil {
						apilog.Printf("Failed to push: %v", err)
					}
				}
			} else {
				apilog.Println("Push not supported")
			}

			tmpl := template.Must(template.ParseFiles(models.FromRootDir("panel/index.html")))
			if err != nil {
				log.Print(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			err = tmpl.Execute(w, map[string]string{
				"version": models.Config.Version,
				"locale":  panelLocale,
				"root":    panelRoot,
			})
			if err != nil {
				log.Print(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}), false))

	// Assets static dirs
	api.Handle(panelRoot+"/assets/", http.StripPrefix(panelRoot, apiHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/") {
				http.NotFound(w, r)
				return
			}
			w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", 3600*24*7))
			http.ServeFile(w, r, path.Join(models.FromRootDir("panel/"), r.URL.Path))
		}), true)))

	apilog.Println("API listening on port " + models.Config.APIPort + "...")
	log.Fatal(http.ListenAndServeTLS(":"+models.Config.APIPort, models.FromRootDir("/cert/server.pem"), models.FromRootDir("/cert/server.key"), api))
}

func apiRequest(w http.ResponseWriter, r *http.Request) {
	var (
		js      []byte
		request []byte
		err     error
	)
	if r.FormValue("request") != "" {
		request = []byte(r.FormValue("request"))
	} else {
		request, err = ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if len(request) == 0 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	req, err := parseRequest(request)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write([]byte(fmt.Sprintf(`{"status": "error", "message": %s}`, strconv.Quote(err.Error()))))
		if err != nil {
			log.Println(err)
		}
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
	_, err = w.Write(js)
	if err != nil {
		log.Println(err)
	}

}
