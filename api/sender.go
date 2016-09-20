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
	if err = r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Form["cmd"][0] {

	case "get-records":
		if auth.Right("get-groups") && auth.GroupRight(r.Form["groupId"][0]) {
			var f Sender
			var fs Senders
			fs.Records = []Sender{}
			query, err := models.Db.Query("SELECT `id`, `email`, `name` FROM `sender` WHERE `group_id`=? LIMIT ? OFFSET ?", r.Form["groupId"][0], r.Form["limit"][0], r.Form["offset"][0])
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			defer query.Close()

			for query.Next() {
				err = query.Scan(&f.Id, &f.Email, &f.Name)
				fs.Records = append(fs.Records, f)
			}
			err = models.Db.QueryRow("SELECT COUNT(*) FROM `sender` WHERE `group_id`=?", r.Form["groupId"][0]).Scan(&fs.Total)
			js, err = json.Marshal(fs)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			js = []byte(`{"status": "error", "message": "Forbidden get groups"}`)
		}

	case "save-record":
		if auth.Right("save-groups") {
			var group int64
			err = models.Db.QueryRow("SELECT `group_id` FROM `sender` WHERE `id`=?", r.FormValue("recid")).Scan(&group)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			if auth.GroupRight(group) {
				_, err = models.Db.Exec("UPDATE `sender` SET `email`=?, `name`=? WHERE `id`=?", r.Form["email"][0], r.Form["name"][0], r.Form["recid"][0])
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

	case "add-record":
		if auth.Right("save-groups") && auth.GroupRight(r.Form["groupId"][0]) {
			res, err := models.Db.Exec("INSERT INTO `sender` (`group_id`, `email`, `name`) VALUES (?, ?, ?);", r.Form["groupId"][0], r.Form["email"][0], r.Form["name"][0])
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
