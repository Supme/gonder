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
)

type Group struct {
	Id   int64  `json:"recid"`
	Name string `json:"name"`
}
type Groups struct {
	Total   int64   `json:"total"`
	Records []Group `json:"records"`
}

func groups(w http.ResponseWriter, r *http.Request) {

	var groups Groups
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
		if auth.Right("get-groups") {
			groups, err = getGroups(req.Offset, req.Limit)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			js, err = json.Marshal(groups)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			js = []byte(`{"status": "error", "message": "Forbidden get group"}`)
		}

	case "save":
		if auth.Right("save-groups") {
			err := saveGroups(req.Changes)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			groups, err = getGroups(req.Offset, req.Limit)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			js, err = json.Marshal(groups)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			js = []byte(`{"status": "error", "message": "Forbidden save groups"}`)
		}

	case "add":
		if auth.Right("add-groups") {
			group, err := addGroup()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			js, err = json.Marshal(group)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			js = []byte(`{"status": "error", "message": "Forbidden add groups"}`)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func addGroup() (Group, error) {
	g := Group{}
	g.Name = "New group"
	row, err := models.Db.Exec("INSERT INTO `group`(`name`) VALUES (?)", g.Name)
	if err != nil {
		return g, err
	}
	g.Id, err = row.LastInsertId()
	if err != nil {
		return g, err
	}

	return g, nil
}

func saveGroups(changes []map[string]interface{}) (err error) {
	var e error
	var where string
	err = nil

	if auth.IsAdmin() {
		where = "?"
	} else {
		where = "id IN (SELECT `group_id` FROM `auth_user_group` WHERE `auth_user_id`=?)"
	}

	for _, change := range changes {
		_, e = models.Db.Exec("UPDATE `group` SET `name`=? WHERE id=? AND "+where, change["name"], change["recid"], auth.userId)
		if e != nil {
			err = e
		}
	}
	return
}

// ToDo order by reqest
func getGroups(offset, limit int64) (Groups, error) {
	var g Group
	var gs Groups
	var where string
	gs.Records = []Group{}

	if auth.IsAdmin() {
		where = "?"
	} else {
		where = "id IN (SELECT `group_id` FROM `auth_user_group` WHERE `auth_user_id`=?)"
	}

	query, err := models.Db.Query("SELECT `id`, `name` FROM `group` WHERE "+where+" ORDER BY `id` DESC LIMIT ? OFFSET ?", auth.userId, limit, offset)
	if err != nil {
		return gs, err
	}
	defer query.Close()
	for query.Next() {
		err = query.Scan(&g.Id, &g.Name)
		gs.Records = append(gs.Records, g)
	}
	err = models.Db.QueryRow("SELECT COUNT(*) FROM `group` WHERE "+where, auth.userId).Scan(&gs.Total)
	return gs, err
}
