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
// Please see the LICENSE file for more information.
package api

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/NYTimes/gziphandler"
	"github.com/tdewolff/minify"
	minifyCSS "github.com/tdewolff/minify/css"
	minifyHTML "github.com/tdewolff/minify/html"
	minifyJS "github.com/tdewolff/minify/js"
	minifyJSON "github.com/tdewolff/minify/json"
	"gonder/bindata"
	"gonder/models"
	"html/template"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	//user   Auth
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
		handler = CheckAuth(fn)
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

	lang, err = newLang(
		models.WorkDir("/panel/assets/w2ui/locale/*.json"),
		models.WorkDir("/panel/assets/gonder/locale/*.json"),
	)
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
	}).Parse(string(bindata.MustAsset("panel/index.html")))
	if err != nil {
		log.Panic(err)
		return
	}

	api.Handle(models.Config.APIPanelPath, apiHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			err = indexTmpl.Execute(w, map[string]string{
				"version": models.Version,
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
			w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", 3600*24*7))

			filePath := path.Join("panel", r.URL.Path)

			if _, err := os.Stat(filePath); err == nil {
				http.ServeFile(w, r, path.Join(filePath))
				return
			}

			fileInfo, err := bindata.AssetInfo(filePath)
			if err != nil {
				apiLog.Print(err)
				http.NotFound(w, r)
				return
			}
			fileContent, err := bindata.Asset(filePath)
			if err != nil {
				apiLog.Print(err)
				http.NotFound(w, r)
				return
			}
			w.Header().Add("Content-Length", strconv.Itoa(int(fileInfo.Size())))
			w.Header().Add("Date", fileInfo.ModTime().Format(time.RFC1123))
			w.Header().Add("Content-Type", mime.TypeByExtension(path.Ext(fileInfo.Name())))

			_, err = w.Write(fileContent)
			if err != nil {
				apiLog.Print(err)
				return
			}

		}), true)))

	// API
	api.Handle("/api/", apiHandler(apiRequest, true))

	// Reports
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

	api.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(models.WorkDir("files/")))))

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

	api.HandleFunc("/logout", Logout)

	api.HandleFunc("/status/ws/campaign.log", CheckAuth(campaignStatus))
	api.HandleFunc("/status/ws/api.log", CheckAuth(apiStatus))
	api.HandleFunc("/status/ws/utm.log", CheckAuth(utmStatus))
	api.HandleFunc("/status/ws/main.log", CheckAuth(mainStatus))

	apiLog.Println("API listening on port " + models.Config.APIPort + "...")
	log.Fatal(http.ListenAndServeTLS(":"+models.Config.APIPort, models.ServerPem, models.ServerKey, api))
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

	req.auth = r.Context().Value("Auth").(*Auth)

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
