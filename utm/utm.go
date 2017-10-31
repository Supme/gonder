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
package utm

import (
	"encoding/base64"
	"github.com/supme/gonder/models"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"image/png"
	"bytes"
)

var (
	utmlog *log.Logger
)

func Run() {
	l, err := os.OpenFile(models.FromRootDir("log/utm.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("error opening utm log file: %v", err)
	}
	defer l.Close()

	multi := io.MultiWriter(l, os.Stdout)

	utmlog = log.New(multi, "", log.Ldate|log.Ltime)

	utm := http.NewServeMux()

	utm.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mem := new(runtime.MemStats)
		runtime.ReadMemStats(mem)
		w.Write([]byte("Welcome to San Tropez! (Conn: " + strconv.Itoa(models.Db.Stats().OpenConnections) + " Allocate: " + strconv.FormatUint(mem.Alloc, 10) + ")"))
	})

	// ToDo disallow all
	utm.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("# " + models.Config.Version + "\nUser-agent: *\nDisallow: /data/\nDisallow: /files/\nDisallow: /unsubscribe/\nDisallow: /redirect/\nDisallow: /web/\nDisallow: /open/\n"))
	})

	utm.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/x-icon")
		ico, _ := base64.StdEncoding.DecodeString("AAABAAEAEBAAAAEAIABoBAAAFgAAACgAAAAQAAAAIAAAAAEAIAAAAAAAAAQAABILAAASCwAAAAAAAAAAAAByGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/8q2uP9yGSL/yra4/3IZIv/Ktrj/yra4/3IZIv9yGSL/yra4/8q2uP9yGSL/yra4/3IZIv/Ktrj/chki/3IZIv/Ktrj/chki/+je3/9yGSL/yra4/3IZIv/Ktrj/chki/8q2uP9yGSL/chki/8q2uP9yGSL/yra4/3IZIv9yGSL/yra4/+je3//Ktrj/chki/8q2uP9yGSL/yra4/3IZIv/Ktrj/yra4/3IZIv/Ktrj/yra4/3IZIv9yGSL/chki/+je3/9yGSL/yra4/3IZIv/Ktrj/chki/8q2uP9yGSL/yra4/3IZIv9yGSL/yra4/3IZIv/Ktrj/chki/3IZIv/Ktrj/chki/8q2uP9yGSL/yra4/8q2uP9yGSL/chki/8q2uP/Ktrj/chki/8q2uP/Ktrj/yra4/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/+je3//o3t//6N7f/+je3//o3t//6N7f/+je3/9yGSL/6N7f/+je3//o3t//6N7f/+je3//o3t//chki/3IZIv/o3t//yra4/8q2uP/Ktrj/yra4/8q2uP/Ktrj/chki/8q2uP/Ktrj/yra4/8q2uP/Ktrj/6N7f/3IZIv9yGSL/6N7f/8q2uP9yGSL/chki/3IZIv/Ktrj/6N7f/3IZIv/Ktrj/yra4/3IZIv9yGSL/yra4/+je3/9yGSL/chki/+je3//Ktrj/chki/8q2uP/o3t//6N7f/8q2uP9yGSL/6N7f/+je3/9yGSL/chki/8q2uP/o3t//chki/3IZIv/o3t//yra4/3IZIv/o3t//yra4/8q2uP/Ktrj/chki/8q2uP/Ktrj/chki/3IZIv/Ktrj/6N7f/3IZIv9yGSL/6N7f/+je3/9yGSL/chki/3IZIv9yGSL/chki/3IZIv/Ktrj/yra4/8q2uP/Ktrj/6N7f/+je3/9yGSL/chki/+je3//o3t//6N7f/+je3//o3t//6N7f/+je3/9yGSL/6N7f/+je3//o3t//6N7f/+je3//o3t//chki/3IZIv/Ktrj/yra4/8q2uP/Ktrj/yra4/8q2uP/Ktrj/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==")
		w.Write(ico)
	})

	utm.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(models.FromRootDir("files/")))))

	utm.HandleFunc("/unsubscribe/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			message, data, err := models.DecodeData(strings.Split(r.URL.Path, "/")[2])
			if err != nil {
				utmlog.Println(err)
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
			_, err = models.Db.Exec("UPDATE `recipient` SET `web_agent`= ? WHERE `id`=? AND `web_agent` IS NULL", models.GetIP(r)+" "+r.UserAgent(), message.RecipientId)
			if err != nil {
				utmlog.Print(err)
			}
			if data == "mail" {
				if err := message.Unsubscribe(map[string]string{"Unsubscribed":"from header link"}); err != nil {
					utmlog.Println(err)
					http.Error(w, "", http.StatusInternalServerError)
					return
				} else {
					if t, err := template.ParseFiles(message.UnsubscribeTemplateDir() + "/success.html"); err != nil {
						utmlog.Println(err)
						http.Error(w, "", http.StatusInternalServerError)
						return
					} else {
						t.Execute(w, map[string]string{
							// ToDo remove
							"campaignId":  message.CampaignId,
							"recipientId": message.RecipientId,

							"CampaignId":  message.CampaignId,
							"RecipientId": message.RecipientId,
							"RecipientEmail": message.RecipientEmail,
							"RecipientName": message.RecipientName,
						})
					}
				}
			}
			if data == "web" {
				if t, err := template.ParseFiles(message.UnsubscribeTemplateDir() + "/accept.html"); err != nil {
					utmlog.Println(err)
					http.Error(w, "", http.StatusInternalServerError)
					return
				} else {
					t.Execute(w, map[string]string{
						// ToDo remove
						"campaignId":  message.CampaignId,
						"recipientId": message.RecipientId,

						"CampaignId":  message.CampaignId,
						"RecipientId": message.RecipientId,
						"RecipientEmail": message.RecipientEmail,
						"RecipientName": message.RecipientName,
					})
				}
			}
		}
	})

	utm.HandleFunc("/unsubscribe", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if r.Method == "POST" {
			var message models.Message
			err := message.New(r.PostFormValue("recipientId"))
			if err != nil {
				utmlog.Println(err)
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
			_, err = models.Db.Exec("UPDATE `recipient` SET `web_agent`= ? WHERE `id`=? AND `web_agent` IS NULL", models.GetIP(r)+" "+r.UserAgent(), message.RecipientId)
			if err != nil {
				utmlog.Print(err)
			}
			if message.CampaignId != r.PostFormValue("campaignId") {
				utmlog.Println(err)
				http.Error(w, "Not valid request", http.StatusInternalServerError)
				return
			}
			if t, err := template.ParseFiles(message.UnsubscribeTemplateDir() + "/success.html"); err != nil {
				utmlog.Println(err)
				http.Error(w, "", http.StatusInternalServerError)
				return
			} else {
				var extra = map[string]string{}
				for name, value := range r.PostForm {
					if name != "campaignId" && name != "recipientId" && name != "unsubscribe" {
						extra[name] = strings.Join(value, "|")
					}
				}
				if err := message.Unsubscribe(extra); err != nil {
					utmlog.Println(err)
					http.Error(w, "", http.StatusInternalServerError)
					return
				}
				t.Execute(w, map[string]string{
					// ToDo remove
					"campaignId":  message.CampaignId,
					"recipientId": message.RecipientId,

					"CampaignId":  message.CampaignId,
					"RecipientId": message.RecipientId,
					"RecipientEmail": message.RecipientEmail,
					"RecipientName": message.RecipientName,
				})
			}
		}
	})

	utm.HandleFunc("/redirect/", func(w http.ResponseWriter, r *http.Request) {
		message, data, err := models.DecodeData(strings.Split(r.URL.Path, "/")[2])
		if err != nil {
			utmlog.Print(err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		_, err = models.Db.Exec("INSERT INTO jumping (campaign_id, recipient_id, url) VALUES (?, ?, ?)", message.CampaignId, message.RecipientId, data)
		if err != nil {
			utmlog.Print(err)
		}
		_, err = models.Db.Exec("UPDATE `recipient` SET `web_agent`= ? WHERE `id`=? AND `web_agent` IS NULL", models.GetIP(r)+" "+r.UserAgent(), message.RecipientId)
		if err != nil {
			utmlog.Print(err)
		}
		url := regexp.MustCompile(`\s*?(\[.*?\])\s*?`).Split(data, 2)
		http.Redirect(w, r, strings.TrimSpace(url[len(url)-1]), http.StatusFound)
	})

	utm.HandleFunc("/web/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		message, _, err := models.DecodeData(strings.Split(r.URL.Path, "/")[2])
		if err != nil {
			utmlog.Println(err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		_, err = models.Db.Exec("INSERT INTO jumping (campaign_id, recipient_id, url) VALUES (?, ?, ?)", message.CampaignId, message.RecipientId, models.WebVersion)
		if err != nil {
			utmlog.Print(err)
		}
		_, err = models.Db.Exec("UPDATE `recipient` SET `web_agent`= ? WHERE `id`=? AND `web_agent` IS NULL", models.GetIP(r)+" "+r.UserAgent(), message.RecipientId)
		if err != nil {
			utmlog.Print(err)
		}
		data, err := message.RenderMessage()
		if err != nil {
			utmlog.Println(err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		w.Write([]byte(data))
	})

	utm.HandleFunc("/open/", func(w http.ResponseWriter, r *http.Request) {
		message, _, err := models.DecodeData(strings.Split(r.URL.Path, "/")[2])
		if err != nil {
			utmlog.Println(err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		_, err = models.Db.Exec("INSERT INTO jumping (campaign_id, recipient_id, url) VALUES (?, ?, ?)", message.CampaignId, message.RecipientId, models.OpenTrace)
		if err != nil {
			utmlog.Print(err)
		}
		_, err = models.Db.Exec("UPDATE `recipient` SET `client_agent`= ? WHERE `id`=? AND `client_agent` IS NULL", models.GetIP(r)+" "+r.UserAgent(), message.RecipientId)
		if err != nil {
			utmlog.Print(err)
		}
		w.Header().Set("Content-Type", "image/gif")
		//png, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAMAAAAoLQ9TAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAADUExURUxpcU3H2DoAAAABdFJOUwBA5thmAAAADUlEQVQY02NgGAXIAAABEAAB7JfjegAAAABJRU5ErkJggg==iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAMAAABEpIrGAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAADUExURUxpcU3H2DoAAAABdFJOUwBA5thmAAAAEklEQVQ4y2NgGAWjYBSMAuwAAAQgAAFWu83mAAAAAElFTkSuQmCC")
		gif, _ := base64.StdEncoding.DecodeString("R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7")
		w.Write(gif)
	})

	utm.HandleFunc("/code/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("qr") != "" {
			size := int(200)
			if r.URL.Query().Get("s") != "" {
				i, err := strconv.Atoi(r.FormValue("s"))
				if err != nil{
					utmlog.Print(err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				size = i
			}
			qrCode, err := qr.Encode(r.URL.Query().Get("qr"), qr.M, qr.Auto)
			if err != nil{
				utmlog.Print(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			qrCode, err = barcode.Scale(qrCode, size, size)
			if err != nil{
				utmlog.Print(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			buffer := new(bytes.Buffer)
			if err := png.Encode(buffer, qrCode); err != nil {
				utmlog.Print(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "image/png")
			w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
			if _, err := w.Write(buffer.Bytes()); err != nil {
				utmlog.Print(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	})

	utmlog.Println("UTM listening on port " + models.Config.StatPort + "...")
	utmlog.Fatal(http.ListenAndServe(":"+models.Config.StatPort, muxLog(utm)))
}

func muxLog(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utmlog.Printf("host: %s %s %s", models.GetIP(r), r.Method, r.RequestURI)
		handler.ServeHTTP(w, r)
	})
}
