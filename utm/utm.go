// Package utm get status sended email
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
	"bytes"
	"encoding/base64"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"gonder/campaign"
	"gonder/models"
	"html/template"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

var (
	utmlog *log.Logger
)

// Run start utm server
func Run() {
	l, err := os.OpenFile(models.WorkDir("log/utm.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("error opening utm log file: %v", err)
	}
	defer l.Close()
	utmlog = log.New(io.MultiWriter(l, os.Stdout), "", log.Ldate|log.Ltime)

	utm := http.NewServeMux()

	utm.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mem := new(runtime.MemStats)
		runtime.ReadMemStats(mem)
		_, err := w.Write([]byte("Welcome to San Tropez! (Conn: " + strconv.Itoa(models.Db.Stats().OpenConnections) + " Allocate: " + strconv.FormatUint(mem.Alloc, 10) + ")"))
		if err != nil {
			log.Println(err)
			return
		}

	})

	// robots.txt
	// ToDo disallow all
	utm.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, err := w.Write([]byte("# " + models.Config.Version + "\nUser-agent: *\nDisallow: /data/\nDisallow: /files/\nDisallow: /unsubscribe/\nDisallow: /redirect/\nDisallow: /web/\nDisallow: /open/\n"))
		if err != nil {
			log.Println(err)
			return
		}

	})

	// favicon.ico
	utm.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/x-icon")
		ico, _ := base64.StdEncoding.DecodeString("AAABAAEAEBAAAAEAIABoBAAAFgAAACgAAAAQAAAAIAAAAAEAIAAAAAAAAAQAABILAAASCwAAAAAAAAAAAAByGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/8q2uP9yGSL/yra4/3IZIv/Ktrj/yra4/3IZIv9yGSL/yra4/8q2uP9yGSL/yra4/3IZIv/Ktrj/chki/3IZIv/Ktrj/chki/+je3/9yGSL/yra4/3IZIv/Ktrj/chki/8q2uP9yGSL/chki/8q2uP9yGSL/yra4/3IZIv9yGSL/yra4/+je3//Ktrj/chki/8q2uP9yGSL/yra4/3IZIv/Ktrj/yra4/3IZIv/Ktrj/yra4/3IZIv9yGSL/chki/+je3/9yGSL/yra4/3IZIv/Ktrj/chki/8q2uP9yGSL/yra4/3IZIv9yGSL/yra4/3IZIv/Ktrj/chki/3IZIv/Ktrj/chki/8q2uP9yGSL/yra4/8q2uP9yGSL/chki/8q2uP/Ktrj/chki/8q2uP/Ktrj/yra4/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/+je3//o3t//6N7f/+je3//o3t//6N7f/+je3/9yGSL/6N7f/+je3//o3t//6N7f/+je3//o3t//chki/3IZIv/o3t//yra4/8q2uP/Ktrj/yra4/8q2uP/Ktrj/chki/8q2uP/Ktrj/yra4/8q2uP/Ktrj/6N7f/3IZIv9yGSL/6N7f/8q2uP9yGSL/chki/3IZIv/Ktrj/6N7f/3IZIv/Ktrj/yra4/3IZIv9yGSL/yra4/+je3/9yGSL/chki/+je3//Ktrj/chki/8q2uP/o3t//6N7f/8q2uP9yGSL/6N7f/+je3/9yGSL/chki/8q2uP/o3t//chki/3IZIv/o3t//yra4/3IZIv/o3t//yra4/8q2uP/Ktrj/chki/8q2uP/Ktrj/chki/3IZIv/Ktrj/6N7f/3IZIv9yGSL/6N7f/+je3/9yGSL/chki/3IZIv9yGSL/chki/3IZIv/Ktrj/yra4/8q2uP/Ktrj/6N7f/+je3/9yGSL/chki/+je3//o3t//6N7f/+je3//o3t//6N7f/+je3/9yGSL/6N7f/+je3//o3t//6N7f/+je3//o3t//chki/3IZIv/Ktrj/yra4/8q2uP/Ktrj/yra4/8q2uP/Ktrj/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/chki/3IZIv9yGSL/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==")
		_, err := w.Write(ico)
		if err != nil {
			log.Println(err)
			return
		}
	})

	// folder files
	utm.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(models.WorkDir("files/")))))

	// unsubscribe
	utm.HandleFunc("/unsubscribe/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			splitURL := strings.Split(r.URL.Path, "/")
			if len(splitURL) != 3 {
				return
			}
			message, data, err := models.DecodeUTM(splitURL[2])
			if err != nil {
				log.Println(err)
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
			_, err = models.Db.Exec("UPDATE `recipient` SET `web_agent`= ? WHERE `id`=? AND `web_agent` IS NULL", models.GetIP(r)+" "+r.UserAgent(), message.RecipientID)
			if err != nil {
				log.Print(err)
			}
			if data == "mail" {
				if err := message.Unsubscribe(map[string]string{"Unsubscribed": "from header link"}); err != nil {
					log.Println(err)
					http.Error(w, "", http.StatusInternalServerError)
				} else {
					if t, err := template.ParseFiles(message.UnsubscribeTemplateDir() + "/success.html"); err != nil {
						log.Println(err)
						http.Error(w, "", http.StatusInternalServerError)
					} else {
						err = t.Execute(w, map[string]string{
							"CampaignId":     message.CampaignID,
							"RecipientId":    message.RecipientID,
							"RecipientEmail": message.RecipientEmail,
							"RecipientName":  message.RecipientName,
						})
						if err != nil {
							log.Println(err)
							return
						}
					}
				}
				return
			}
			if data == "web" {
				if t, err := template.ParseFiles(message.UnsubscribeTemplateDir() + "/accept.html"); err != nil {
					log.Println(err)
					http.Error(w, "", http.StatusInternalServerError)
				} else {
					err = t.Execute(w, map[string]string{
						"CampaignId":     message.CampaignID,
						"RecipientId":    message.RecipientID,
						"RecipientEmail": message.RecipientEmail,
						"RecipientName":  message.RecipientName,
					})
					if err != nil {
						log.Println(err)
						return
					}
				}
			}
		}
	})

	// unsubscribe with extra parameters from form
	utm.HandleFunc("/unsubscribe", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if r.Method == "POST" {
			var message models.Message
			err := message.New(r.PostFormValue("recipientId"))
			if err != nil {
				log.Println(err)
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
			_, err = models.Db.Exec("UPDATE `recipient` SET `web_agent`= ? WHERE `id`=? AND `web_agent` IS NULL", models.GetIP(r)+" "+r.UserAgent(), message.RecipientID)
			if err != nil {
				log.Print(err)
			}
			if message.CampaignID != r.PostFormValue("campaignId") {
				log.Println(err)
				http.Error(w, "Not valid request", http.StatusInternalServerError)
				return
			}
			if t, err := template.ParseFiles(message.UnsubscribeTemplateDir() + "/success.html"); err != nil {
				log.Println(err)
				http.Error(w, "", http.StatusInternalServerError)
			} else {
				var extra = map[string]string{}
				for name, value := range r.PostForm {
					if name != "campaignId" && name != "recipientId" && name != "unsubscribe" {
						extra[name] = strings.Join(value, "|")
					}
				}
				if err := message.Unsubscribe(extra); err != nil {
					log.Println(err)
					http.Error(w, "", http.StatusInternalServerError)
					return
				}
				err = t.Execute(w, map[string]string{
					"CampaignId":     message.CampaignID,
					"RecipientId":    message.RecipientID,
					"RecipientEmail": message.RecipientEmail,
					"RecipientName":  message.RecipientName,
				})
				if err != nil {
					log.Println(err)
					return
				}

			}
		}
	})

	// redirect link
	utm.HandleFunc("/redirect/", func(w http.ResponseWriter, r *http.Request) {
		splitURL := strings.Split(r.URL.Path, "/")
		if len(splitURL) != 3 {
			return
		}
		message, data, err := models.DecodeUTM(splitURL[2])
		if err != nil {
			log.Print(err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		_, err = models.Db.Exec("INSERT INTO jumping (campaign_id, recipient_id, url) VALUES (?, ?, ?)", message.CampaignID, message.RecipientID, data)
		if err != nil {
			log.Print(err)
		}
		_, err = models.Db.Exec("UPDATE `recipient` SET `web_agent`= ? WHERE `id`=? AND `web_agent` IS NULL", models.GetIP(r)+" "+r.UserAgent(), message.RecipientID)
		if err != nil {
			log.Print(err)
		}
		url := regexp.MustCompile(`\s*?(\[.*?\])\s*?`).Split(data, 2)
		http.Redirect(w, r, strings.TrimSpace(url[len(url)-1]), http.StatusFound)
	})

	utm.HandleFunc("/web/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		splitURL := strings.Split(r.URL.Path, "/")
		if len(splitURL) != 3 {
			return
		}
		message, _, err := models.DecodeUTM(splitURL[2])
		if err != nil {
			log.Println(err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		_, err = models.Db.Exec("INSERT INTO jumping (campaign_id, recipient_id, url) VALUES (?, ?, ?)", message.CampaignID, message.RecipientID, models.WebVersion)
		if err != nil {
			log.Print(err)
		}
		_, err = models.Db.Exec("UPDATE `recipient` SET `web_agent`= ? WHERE `id`=? AND `web_agent` IS NULL", models.GetIP(r)+" "+r.UserAgent(), message.RecipientID)
		if err != nil {
			log.Print(err)
		}

		recipient, err := campaign.GetRecipient(message.RecipientID)
		if err != nil {
			log.Println(err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		tmplFunc := recipient.WebHTML(true, false)
		err = tmplFunc(w)
		if err != nil {
			log.Println(err)
		}
	})

	// StatPng
	utm.HandleFunc("/open/", func(w http.ResponseWriter, r *http.Request) {
		splitURL := strings.Split(r.URL.Path, "/")
		if len(splitURL) != 3 {
			return
		}
		message, _, err := models.DecodeUTM(splitURL[2])
		if err != nil {
			log.Println(err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		_, err = models.Db.Exec("INSERT INTO jumping (campaign_id, recipient_id, url) VALUES (?, ?, ?)", message.CampaignID, message.RecipientID, models.OpenTrace)
		if err != nil {
			log.Print(err)
		}
		_, err = models.Db.Exec("UPDATE `recipient` SET `client_agent`= ? WHERE `id`=? AND `client_agent` IS NULL", models.GetIP(r)+" "+r.UserAgent(), message.RecipientID)
		if err != nil {
			log.Print(err)
		}
		w.Header().Set("Content-Type", "image/gif")
		//png, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAMAAAAoLQ9TAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAADUExURUxpcU3H2DoAAAABdFJOUwBA5thmAAAADUlEQVQY02NgGAXIAAABEAAB7JfjegAAAABJRU5ErkJggg==iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAMAAABEpIrGAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAADUExURUxpcU3H2DoAAAABdFJOUwBA5thmAAAAEklEQVQ4y2NgGAWjYBSMAuwAAAQgAAFWu83mAAAAAElFTkSuQmCC")
		gif, _ := base64.StdEncoding.DecodeString("R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7")
		_, err = w.Write(gif)
		if err != nil {
			log.Println(err)
			return
		}
	})

	// QRcode generator
	utm.HandleFunc("/code/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("qr") != "" {
			size := int(200)
			if r.URL.Query().Get("s") != "" {
				i, err := strconv.Atoi(r.FormValue("s"))
				if err != nil {
					log.Print(err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				size = i
			}
			qrCode, err := qr.Encode(r.URL.Query().Get("qr"), qr.M, qr.Auto)
			if err != nil {
				log.Print(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			qrCode, err = barcode.Scale(qrCode, size, size)
			if err != nil {
				log.Print(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			buffer := new(bytes.Buffer)
			if err := png.Encode(buffer, qrCode); err != nil {
				log.Print(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "image/png")
			w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
			if _, err := w.Write(buffer.Bytes()); err != nil {
				log.Print(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	})

	// Question
	utm.HandleFunc("/question/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			splitURL := strings.Split(r.URL.Path, "/")
			if len(splitURL) != 3 {
				return
			}
			message, _, err := models.DecodeUTM(splitURL[2])
			if err != nil {
				log.Println(err)
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
			if t, err := template.ParseFiles(message.QuestionTemplateDir() + "/question.html"); err != nil {
				log.Println(err)
				http.Error(w, "", http.StatusInternalServerError)
			} else {
				if err := r.ParseForm(); err != nil {
					log.Println(err)
					return
				}
				var data = map[string]string{}
				for name, value := range r.PostForm {
					data[name] = strings.Join(value, "|")
				}
				if err := message.Question(data); err != nil {
					log.Println(err)
					http.Error(w, "", http.StatusInternalServerError)
					return
				}
				err = t.Execute(w, map[string]string{
					"CampaignId":     message.CampaignID,
					"RecipientId":    message.RecipientID,
					"RecipientEmail": message.RecipientEmail,
					"RecipientName":  message.RecipientName,
				})
				if err != nil {
					log.Println(err)
					return
				}
			}
		}
	})

	utmlog.Println("UTM listening on port " + models.Config.StatPort + "...")
	log.Fatal(http.ListenAndServe(":"+models.Config.StatPort, muxLog(utm)))
}

func muxLog(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utmlog.Printf("host: %s %s %s", models.GetIP(r), r.Method, r.RequestURI)
		handler.ServeHTTP(w, r)
	})
}
