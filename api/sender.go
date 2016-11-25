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
	"encoding/json"
	"github.com/supme/gonder/models"
	"net/http"
	"strconv"
	"log"
)

type Sender struct {
	Id    int64  `json:"recid"`
	Email string `json:"email"`
	Name  string `json:"name"`
}
type Senders struct {
	Total   int64    `json:"total"`
	Records []Sender `json:"records"`
}

func sender(w http.ResponseWriter, r *http.Request) {

	var err error
	var js []byte
	if r.FormValue("request") == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	req, err := parseRequest(r.FormValue("request"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch req.Cmd {

	case "get":
		if auth.Right("get-groups") && auth.GroupRight(req.Id) {
			var f Sender
			var fs Senders
			fs.Records = []Sender{}
			query, err := models.Db.Query("SELECT `id`, `email`, `name` FROM `sender` WHERE `group_id`=? LIMIT ? OFFSET ?", req.Id, req.Limit, req.Offset)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			defer query.Close()

			for query.Next() {
				err = query.Scan(&f.Id, &f.Email, &f.Name)
				fs.Records = append(fs.Records, f)
			}
			err = models.Db.QueryRow("SELECT COUNT(*) FROM `sender` WHERE `group_id`=?", req.Group).Scan(&fs.Total)
			js, err = json.Marshal(fs)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			js = []byte(`{"status": "error", "message": "Forbidden get groups"}`)
		}

	case "save":
		if auth.Right("save") {
			log.Print(req)
			var group int64
			err = models.Db.QueryRow("SELECT `group_id` FROM `sender` WHERE `id`=?", req.Id).Scan(&group)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			if auth.GroupRight(group) {
				_, err = models.Db.Exec("UPDATE `sender` SET `email`=?, `name`=? WHERE `id`=?", req.Email, req.Name, req.Id)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				js = []byte(`{"status": "success", "message": ""}`)
			} else {
				js = []byte(`{"status": "error", "message": "Forbidden right to this group"}`)
			}
		} else {
			js = []byte(`{"status": "error", "message": "Forbidden save groups"}`)
		}

	case "add":
		if auth.Right("save-groups") && auth.GroupRight(req.Id) {
			res, err := models.Db.Exec("INSERT INTO `sender` (`group_id`, `email`, `name`) VALUES (?, ?, ?);", req.Id, req.Email, req.Name)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			recid, err := res.LastInsertId()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			js = []byte(`{"status": "success", "message": "", "recid": ` + strconv.FormatInt(recid, 10) + `}`)
		} else {
			js = []byte(`{"status": "error", "message": "Forbidden save groups"}`)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
