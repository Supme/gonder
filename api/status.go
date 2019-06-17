package api

import (
	"github.com/gorilla/websocket"
	"github.com/gravitational/tail"
	"log"
	"net/http"
	"os"
)

func campaignLog(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("Auth").(*Auth)
	if user.Right("get-log-campaign") {
		logHandler(w, r, "./log/campaign.log")
	} else {
		http.Error(w, "Forbidden get log campaign", http.StatusForbidden)
		return
	}
}

func apiLog(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("Auth").(*Auth)
	if user.Right("get-log-api") {
		logHandler(w, r, "./log/api.log")
	} else {
		http.Error(w, "Forbidden get log api", http.StatusForbidden)
		return
	}
}

func utmLog(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("Auth").(*Auth)
	if user.Right("get-log-utm") {
		logHandler(w, r, "./log/utm.log")
	} else {
		http.Error(w, "Forbidden get log utm", http.StatusForbidden)
		return
	}
}

func mainLog(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("Auth").(*Auth)
	if user.Right("get-log-main") {
		logHandler(w, r, "./log/main.log")
	} else {
		http.Error(w, "Forbidden get log main", http.StatusForbidden)
		return
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  32,
	WriteBufferSize: 32,
}

func logHandler(w http.ResponseWriter, r *http.Request, file string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print(err)
		return
	}

	defer func() {
		_ = conn.Close()
	}()

	var more string
	var offset tail.SeekInfo
	offset.Whence = 2
	fi, err := os.Open(file)
	if err != nil {
		log.Print(err)
	}
	f, err := fi.Stat()
	if err != nil {
		log.Print(err)
	}
	if f.Size() < 50000 {
		offset.Offset = f.Size() * (-1)
	} else {
		offset.Offset = -50000
		more = "..."
	}
	if err := fi.Close(); err != nil {
		log.Print(err)
	}

	conf := tail.Config{
		Follow:   true,
		ReOpen:   true,
		Location: &offset,
		Logger:   tail.DiscardingLogger,
	}

	t, err := tail.TailFile(file, conf)
	if err != nil {
		apilog.Print(err)
	}

	for line := range t.Lines {
		if line.Err == nil {
			if err = conn.WriteMessage(websocket.TextMessage, []byte(more+line.Text)); err != nil {
				_ = conn.Close()
				return
			}
			more = ""
		} else {
			log.Print(err)
		}

	}
}
