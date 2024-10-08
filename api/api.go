// Package api contains api service and web panel
// Project Gonder.
// Author Supme
// Copyright Supme 2016
// License http://opensource.org/licenses/MIT MIT License
//
//	THE SOFTWARE AND DOCUMENTATION ARE PROVIDED "AS IS" WITHOUT WARRANTY OF
//	ANY KIND, EITHER EXPRESSED OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
//	IMPLIED WARRANTIES OF MERCHANTABILITY AND/OR FITNESS FOR A PARTICULAR
//	PURPOSE.
//
// Please see the LICENSE file for more information.
package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/NYTimes/gziphandler"
	"github.com/Supme/httpreloader"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tdewolff/minify"
	minifyCSS "github.com/tdewolff/minify/css"
	minifyHTML "github.com/tdewolff/minify/html"
	minifyJS "github.com/tdewolff/minify/js"
	minifyJSON "github.com/tdewolff/minify/json"
	"gonder/models"
	"gonder/panel"
	"html/template"
	"log"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

var (
	apiLog *log.Logger
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
		handler = AuthHandler(fn)
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
func Run(logger *log.Logger) {
	var err error

	apiLog = logger

	lang, err = newLang()
	if err != nil {
		apiLog.Printf("error loading languages: %s", err)
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

	indexTmpl, err := template.New("index.html").Funcs(template.FuncMap{
		"tr": func(s string) template.HTML {
			return template.HTML(lang.tr(models.Config.APIPanelLocale, s))
		},
	}).Parse(panel.Index)
	if err != nil {
		log.Panic(err)
		return
	}

	api.Handle(models.Config.APIPanelPath, apiHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Cache-Control", "no-store")
			err = indexTmpl.Execute(w, map[string]string{
				"version": models.AppVersion,
				"locale":  models.Config.APIPanelLocale,
				"root":    models.Config.APIPanelPath,
			})
			if err != nil {
				log.Print(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}), false))

	// Assets static dirs
	api.Handle(models.Config.APIPanelPath+"/assets/", http.StripPrefix(models.Config.APIPanelPath, apiHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/") {
				http.NotFound(w, r)
				return
			}
			w.Header().Add("Cache-Control", "no-store")

			//filePath := path.Join("panel", r.URL.Path)
			//
			//if _, err := os.Stat(filePath); err == nil {
			//	http.ServeFile(w, r, path.Join(filePath))
			//	return
			//}

			filePath := strings.TrimPrefix(r.URL.Path, "/")
			file, err := panel.Assets.Open(filePath)
			if err != nil {
				apiLog.Print(err)
				http.NotFound(w, r)
				return
			}

			fileInfo, err := file.Stat()
			if err != nil {
				apiLog.Print(err)
				http.NotFound(w, r)
				return
			}
			fileContent, err := panel.Assets.ReadFile(filePath)
			if err != nil {
				apiLog.Print(err)
				http.NotFound(w, r)
				return
			}
			w.Header().Add("Content-Length", strconv.Itoa(int(fileInfo.Size())))
			//w.Header().Add("Date", fileInfo.ModTime().Format(time.RFC1123))
			w.Header().Add("Content-Type", mime.TypeByExtension(path.Ext(fileInfo.Name())))

			_, err = w.Write(fileContent)
			if err != nil {
				apiLog.Print(err)
				return
			}

		}), false)))

	// API
	api.Handle("/api/", apiHandler(apiRequest, true))
	// Recipient
	api.Handle("/recipient/upload", apiHandler(RecipientUploadHandlerFunc, true))
	// Reports
	api.Handle("/report/group", apiHandler(reportsGroupHandlerFunc, true))
	api.Handle("/report/campaign", apiHandler(reportsCampaignHandlerFunc, true))
	api.Handle("/report/started", apiHandler(reportStartedCampaign, true))
	api.Handle("/report/summary", apiHandler(reportSummary, true))
	api.Handle("/report/clickcount", apiHandler(reportClickCount, true))
	api.Handle("/report/recipients", apiHandler(reportRecipientsList, true))
	api.Handle("/report/clicks", apiHandler(reportRecipientClicks, true))
	api.Handle("/report/unsubscribed", apiHandler(reportUnsubscribed, true))
	api.Handle("/report/file/recipients", apiHandler(reportRecipientsCsv, true))
	api.Handle("/report/question", apiHandler(reportQuestionSummary, true))

	api.Handle("/preview", apiHandler(getMailPreview, true))
	api.Handle("/unsubscribe", apiHandler(getUnsubscribePreview, true))

	api.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(models.Config.UTMFilesDir))))

	api.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/x-icon")
		ico, _ := base64.StdEncoding.DecodeString("AAABAAEAEBAAAAEAIABoBAAAFgAAACgAAAAQAAAAIAAAAAEAIAAAAAAAAAQAABILAAASCwAAAAAAAAAAAAByGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/8q2uP9yGSL/yra4/3IZIv/Ktrj/yra4/3IZIv9yGSL/yra4/8q2uP9yGSL/yra4/3IZIv/Ktrj/chki/3IZIv/Ktrj/chki/+je3/9yGSL/yra4/3IZIv/Ktrj/chki/8q2uP9yGSL/chki/8q2uP9yGSL/yra4/3IZIv9yGSL/yra4/+je3//Ktrj/chki/8q2uP9yGSL/yra4/3IZIv/Ktrj/yra4/3IZIv/Ktrj/yra4/3IZIv9yGSL/chki/+je3/9yGSL/yra4/3IZIv/Ktrj/chki/8q2uP9yGSL/yra4/3IZIv9yGSL/yra4/3IZIv/Ktrj/chki/3IZIv/Ktrj/chki/8q2uP9yGSL/yra4/8q2uP9yGSL/chki/8q2uP/Ktrj/chki/8q2uP/Ktrj/yra4/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/+je3//o3t//6N7f/+je3//o3t//6N7f/+je3/9yGSL/6N7f/+je3//o3t//6N7f/+je3//o3t//chki/3IZIv/o3t//yra4/8q2uP/Ktrj/yra4/8q2uP/Ktrj/chki/8q2uP/Ktrj/yra4/8q2uP/Ktrj/6N7f/3IZIv9yGSL/6N7f/8q2uP9yGSL/chki/3IZIv/Ktrj/6N7f/3IZIv/Ktrj/yra4/3IZIv9yGSL/yra4/+je3/9yGSL/chki/+je3//Ktrj/chki/8q2uP/o3t//6N7f/8q2uP9yGSL/6N7f/+je3/9yGSL/chki/8q2uP/o3t//chki/3IZIv/o3t//yra4/3IZIv/o3t//yra4/8q2uP/Ktrj/chki/8q2uP/Ktrj/chki/3IZIv/Ktrj/6N7f/3IZIv9yGSL/6N7f/+je3/9yGSL/chki/3IZIv9yGSL/chki/3IZIv/Ktrj/yra4/8q2uP/Ktrj/6N7f/+je3/9yGSL/chki/+je3//o3t//6N7f/+je3//o3t//6N7f/+je3/9yGSL/6N7f/+je3//o3t//6N7f/+je3//o3t//chki/3IZIv/Ktrj/yra4/8q2uP/Ktrj/yra4/8q2uP/Ktrj/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==")
		_, err = w.Write(ico)
		if err != nil {
			log.Println(err)
		}
	})

	api.HandleFunc("/%7B%7B.StatUrl%7D%7D", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/gif")
		blank, _ := base64.StdEncoding.DecodeString("R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7")
		_, err = w.Write(blank)
		if err != nil {
			log.Println(err)
		}
	})

	api.HandleFunc("/logout", Logout)

	api.HandleFunc("/status/ws/campaign.log", AuthHandler(campaignStatus))
	api.HandleFunc("/status/ws/api.log", AuthHandler(apiStatus))
	api.HandleFunc("/status/ws/utm.log", AuthHandler(utmStatus))
	api.HandleFunc("/status/ws/main.log", AuthHandler(mainStatus))

	api.Handle("/metrics", promhttp.Handler())

	server, err := httpreloader.NewServer(":"+models.Config.APIPort, models.ServerPem, models.ServerKey, api)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGHUP)
		for range c {
			apiLog.Print("reload certificates")
			err := server.Reloader.UpdateCertificate(models.ServerPem, models.ServerKey)
			if err != nil {
				log.Print(err)
			}
		}
	}()

	apiLog.Println("API listening on port " + models.Config.APIPort + "...")
	log.Fatal(server.ListenAndServeTLS())
}

func apiRequest(w http.ResponseWriter, r *http.Request) {
	var (
		js  []byte
		req request
		err error
	)
	if r.FormValue("request") != "" {
		req, err = parseRequest([]byte(r.FormValue("request")))
	} else {
		err = json.NewDecoder(r.Body).Decode(&req)
		defer r.Body.Close()
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write([]byte(fmt.Sprintf(`{"status": "error", "message": %s}`, strconv.Quote(err.Error()))))
		if err != nil {
			log.Println(err)
		}
		return
	}
	req.auth = r.Context().Value(ContextAuth).(*Auth)

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
		js, err = recipientsReq(req)
	case "/api/sender":
		js, err = sender(req)
	case "/api/senderlist":
		js, err = senderList(req)
	case "/api/units":
		js, err = units(req)
	case "/api/account":
		js, err = account(req)
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
