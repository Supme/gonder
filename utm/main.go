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
	"os"
	"github.com/supme/gonder/models"
	"log"
	"io"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"html/template"
	"encoding/base64"
	"encoding/json"
	"regexp"
	"net"
)

var (
	utmlog *log.Logger
)

// todo remove later
type oldJson struct {
	Campaign    string `json:"c"`
	Recipient   string `json:"r"`
	Url         string `json:"u"`
	Webver      string `json:"w"`
	Opened      string `json:"o"`
	Unsubscribe string `json:"s"`
}


func Run() {
	l, err := os.OpenFile(models.FromRootDir("log/utm.log"), os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Println("error opening utm log file: %v", err)
	}
	defer l.Close()

	multi := io.MultiWriter(l, os.Stdout)

	utmlog = log.New(multi, "", log.Ldate | log.Ltime)

	utm := http.NewServeMux()

	utm.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mem := new(runtime.MemStats)
		runtime.ReadMemStats(mem)
		w.Write([]byte("Welcome to San Tropez! (Conn: "  + strconv.Itoa(models.Db.Stats().OpenConnections) + " Allocate: " + strconv.FormatUint(mem.Alloc, 10) + ")"))
	})

	utm.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("# Gonder " + models.Config.Version + "\nUser-agent: *\nDisallow: /data/\nDisallow: /files/\nDisallow: /unsubscribe/\nDisallow: /unsubscribe\nDisallow: /redirect/\nDisallow: /web/\nDisallow: /open/\n"))
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
			message, data, err := models.Decode_data(strings.Split(r.URL.Path, "/")[2])
			if err != nil {
				utmlog.Println(err)
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
			if data == "mail" {
				if err := message.Unsubscribe(); err != nil {
					utmlog.Println(err)
					http.Error(w, "", http.StatusInternalServerError)
					return
				} else {
					if t, err := template.ParseFiles(message.Unsubscribe_template_dir() + "/success.html"); err != nil {
						utmlog.Println(err)
						http.Error(w, "", http.StatusInternalServerError)
						return
					} else {
						t.Execute(w, map[string]string{
							"campaignId":  message.CampaignId,
							"recipientId": message.RecipientId,
						})
					}
				}
			}
			if data == "web" {
				if t, err := template.ParseFiles(message.Unsubscribe_template_dir() + "/accept.html"); err != nil {
					utmlog.Println(err)
					http.Error(w, "", http.StatusInternalServerError)
					return
				} else {
					t.Execute(w, map[string]string{
						"campaignId":  message.CampaignId,
						"recipientId": message.RecipientId,
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
			if message.CampaignId != r.PostFormValue("campaignId") {
				utmlog.Println(err)
				http.Error(w, "Not valid request", http.StatusInternalServerError)
				return
			}
			if t, err := template.ParseFiles(message.Unsubscribe_template_dir() + "/success.html"); err != nil {
				utmlog.Println(err)
				http.Error(w, "", http.StatusInternalServerError)
				return
			} else {
				if err := message.Unsubscribe(); err != nil {
					utmlog.Println(err)
					http.Error(w, "", http.StatusInternalServerError)
					return
				}
				t.Execute(w, map[string]string{
					"campaignId":  message.CampaignId,
					"recipientId": message.RecipientId,
				})
			}
		}
	})

	utm.HandleFunc("/redirect/", func(w http.ResponseWriter, r *http.Request) {
		message, data, err := models.Decode_data(strings.Split(r.URL.Path, "/")[2])
		if err != nil {
			utmlog.Println(err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		models.Db.Exec("INSERT INTO jumping (campaign_id, recipient_id, url) VALUES (?, ?, ?)", message.CampaignId, message.RecipientId, data)
		models.Db.Exec("UPDATE `recipient` SET `web_agent`= ? WHERE `id`=? AND `web_agent` IS NULL", getIP(r) + " " + r.UserAgent(), message.RecipientId)
		url := regexp.MustCompile(`\s*?(\[.*?\])\s*?`).Split(data, 2)
		http.Redirect(w, r, strings.TrimSpace(url[len(url)-1]), http.StatusFound)
	})

	utm.HandleFunc("/web/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		message, _, err := models.Decode_data(strings.Split(r.URL.Path, "/")[2])
		if err != nil {
			utmlog.Println(err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		models.Db.Exec("INSERT INTO jumping (campaign_id, recipient_id, url) VALUES (?, ?, 'web_version')", message.CampaignId, message.RecipientId)
		models.Db.Exec("UPDATE `recipient` SET `web_agent`= ? WHERE `id`=? AND `web_agent` IS NULL", getIP(r) + " " + r.UserAgent(), message.RecipientId)
		data, err := message.Render_message()
		if err != nil {
			utmlog.Println(err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		w.Write([]byte(data))
	})

	utm.HandleFunc("/open/", func(w http.ResponseWriter, r *http.Request) {
		message, _, err := models.Decode_data(strings.Split(r.URL.Path, "/")[2])
		if err != nil {
			utmlog.Println(err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		models.Db.Exec("INSERT INTO jumping (campaign_id, recipient_id, url) VALUES (?, ?, 'open_trace')", message.CampaignId, message.RecipientId)
		models.Db.Exec("UPDATE `recipient` SET `client_agent`= ? WHERE `id`=? AND `client_agent` IS NULL", getIP(r) + " " + r.UserAgent(), message.RecipientId)
		w.Header().Set("Content-Type", "image/png")
		png, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAMAAAAoLQ9TAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAADUExURUxpcU3H2DoAAAABdFJOUwBA5thmAAAADUlEQVQY02NgGAXIAAABEAAB7JfjegAAAABJRU5ErkJggg==iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAMAAABEpIrGAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAADUExURUxpcU3H2DoAAAABdFJOUwBA5thmAAAAEklEQVQ4y2NgGAWjYBSMAuwAAAQgAAFWu83mAAAAAElFTkSuQmCC")
		w.Write(png)
	})




	// Old statistic todo delete later
	utm.HandleFunc("/data/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			var param *oldJson

			data, err := base64.URLEncoding.DecodeString(strings.Split(strings.Split(r.URL.Path, "/")[2], ".")[0])
			if err != nil {
				utmlog.Println(err)
				http.Error(w, "", http.StatusInternalServerError)
				return
			}

			err = json.Unmarshal([]byte(data), &param)
			if err != nil {
				utmlog.Println(err)
				http.Error(w, "", http.StatusInternalServerError)
				return
			}

			userAgent := getIP(r) + " " + r.UserAgent()

			if param.Opened != "" {
				models.Db.Exec("INSERT INTO jumping (campaign_id, recipient_id, url) VALUES (?, ?, ?)", param.Campaign, param.Recipient, "open_trace")
				models.Db.Exec("UPDATE `recipient` SET `client_agent`= ? WHERE `id`=? AND `client_agent` IS NULL", userAgent, param.Recipient)
				w.Header().Set("Content-Type", "image/png")
				png, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAMAAAAoLQ9TAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAADUExURUxpcU3H2DoAAAABdFJOUwBA5thmAAAADUlEQVQY02NgGAXIAAABEAAB7JfjegAAAABJRU5ErkJggg==iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAMAAABEpIrGAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAADUExURUxpcU3H2DoAAAABdFJOUwBA5thmAAAAEklEQVQ4y2NgGAWjYBSMAuwAAAQgAAFWu83mAAAAAElFTkSuQmCC")
				w.Write(png)
			} else if param.Url != "" {
				models.Db.Exec("INSERT INTO jumping (campaign_id, recipient_id, url) VALUES (?, ?, ?)", param.Campaign, param.Recipient, param.Url)
				models.Db.Exec("UPDATE `recipient` SET `web_agent`= ? WHERE `id`=? AND `web_agent` IS NULL", userAgent, param.Recipient)
				http.Redirect(w, r, param.Url, http.StatusFound)
			} else if param.Webver != "" {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				models.Db.Exec("INSERT INTO jumping (campaign_id, recipient_id, url) VALUES (?, ?, ?)", param.Campaign, param.Recipient, "web_version")
				models.Db.Exec("UPDATE `recipient` SET `web_agent`= ? WHERE `id`=? AND `web_agent` IS NULL", userAgent, param.Recipient)
				var m models.Message
				m.New(param.Recipient)
				message, err :=  m.Render_message()
				if err != nil {
					utmlog.Println(err)
					http.Error(w, "", http.StatusInternalServerError)
					return
				}
				w.Write([]byte(message))
			} else if param.Unsubscribe != "" {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				var name string
				models.Db.QueryRow("SELECT `group`.`template` FROM `campaign` INNER JOIN `group` ON `campaign`.`group_id`=`group`.`id` WHERE `group`.`template` IS NOT NULL AND `campaign`.`id`=?", param.Campaign).Scan(&name)
				if name == "" {
					name = "default"
				} else {
					if _, err := os.Stat(models.FromRootDir("templates/" + name + "/accept.html")); err != nil {
						name = "default"
					}
					if _, err := os.Stat(models.FromRootDir("templates/" + name + "/success.html")); err != nil {
						name = "default"
					}
				}
				if t, err := template.ParseFiles(models.FromRootDir("templates/" + name + "/accept.html")); err != nil {
					utmlog.Println(err)
					http.Error(w, "", http.StatusInternalServerError)
					return
				} else {
					t.Execute(w, map[string]string{
						"campaignId":  param.Campaign,
						"recipientId": param.Recipient,
					})
				}
			}
		}
	})
	// /Old statistic todo delete later





	utmlog.Println("UTM listening on port " + models.Config.StatPort + "...")
	utmlog.Fatal(http.ListenAndServe(":" + models.Config.StatPort, mux_log(utm)))
}

func mux_log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utmlog.Printf("host: %s %s %s", getIP(r), r.Method, r.RequestURI)
		handler.ServeHTTP(w, r)
	})
}

func getIP(r *http.Request) string {
	if ipProxy := r.Header.Get("X-FORWARDED-FOR"); len(ipProxy) > 0 {
		return ipProxy
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}