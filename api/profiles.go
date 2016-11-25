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
	"fmt"
	"github.com/supme/gonder/models"
	"net/http"
	"strconv"
)

type Profile struct {
	Id          int64  `json:"recid"`
	Name        string `json:"name"`
	Iface       string `json:"iface"`
	Host        string `json:"host"`
	Stream      int    `json:"stream"`
	ResendDelay int    `json:"resend_delay"`
	ResendCount int    `json:"resend_count"`
}

type Profiles struct {
	Status  string    `json:"status"`
	Message string    `json:"message"`
	Total   int64     `json:"total"`
	Records []Profile `json:"records"`
}

func profiles(w http.ResponseWriter, r *http.Request) {
	var err error
	var js []byte
	var ps Profiles
	var p Profile

	ps.Status = "success"

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
		if auth.Right("get-profiles") {
			ps, err = getProfiles()
			if err != nil {
				ps.Status = "error"
				ps.Message = err.Error()
			}
			js, err = json.Marshal(ps)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			js = []byte(`{"status": "error", "message": "Forbidden get profiles"}`)
		}

	case "add":
		if auth.Right("add-profiles") {
			p, err = addProfile()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			js, err = json.Marshal(p)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			js = []byte(`{"status": "error", "message": "Forbidden get profiles"}`)
		}

	case "delete":
		if auth.Right("delete-profiles") {
			fmt.Print(req.Selected)
			deleteProfiles(req.Selected)
			js, err = json.Marshal(ps)
		} else {
			js = []byte(`{"status": "error", "message": "Forbidden get profiles"}`)
		}

	case "save":
		if auth.Right("save-profiles") {
			err = saveProfiles(req.Changes)
			if err != nil {
				ps.Status = "error"
				ps.Message = err.Error()
			}
			js, err = json.Marshal(ps)
		} else {
			js = []byte(`{"status": "error", "message": "Forbidden get profiles"}`)
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func saveProfiles(changes []map[string]interface{}) (err error) {
	var e error
	err = nil
	var p Profile
	for c := range changes {
		p.Id, e = strconv.ParseInt(fmt.Sprint(changes[c]["recid"]), 10, 64)
		if e != nil {
			err = e
		}
		e = models.Db.QueryRow("SELECT `name`,`iface`,`host`,`stream`,`resend_delay`,`resend_count` FROM `profile` WHERE `id`=?", p.Id).Scan(&p.Name, &p.Iface, &p.Host, &p.Stream, &p.ResendDelay, &p.ResendCount)
		if e != nil {
			err = e
		}
		for i := range changes[c] {
			switch i {
			case "name":
				p.Name = fmt.Sprint(changes[c][i])
			case "iface":
				p.Iface = fmt.Sprint(changes[c][i])
			case "host":
				p.Host = fmt.Sprint(changes[c][i])
			case "stream":
				p.Stream, _ = strconv.Atoi(fmt.Sprint(changes[c][i]))
			case "resend_delay":
				p.ResendDelay, _ = strconv.Atoi(fmt.Sprint(changes[c][i]))
			case "resend_count":
				p.ResendCount, _ = strconv.Atoi(fmt.Sprint(changes[c][i]))
			}
		}
		_, e = models.Db.Exec("UPDATE `profile` SET `name`=?, `iface`=?, `host`=?, `stream`=?, `resend_delay`=?, `resend_count`=? WHERE id=?", p.Name, p.Iface, p.Host, p.Stream, p.ResendDelay, p.ResendCount, p.Id)
		if e != nil {
			err = e
		}
	}
	return
}

func deleteProfiles(selected []interface{}) {
	for _, s := range selected {
		models.Db.Exec("DELETE FROM `profile` WHERE `id`=?", s)
	}
}

func addProfile() (Profile, error) {
	var p Profile
	row, err := models.Db.Exec("INSERT INTO `profile` (`name`) VALUES ('')")
	if err != nil {
		return p, err
	}
	p.Id, err = row.LastInsertId()
	if err != nil {
		return p, err
	}

	return p, nil
}

func getProfiles() (Profiles, error) {
	var p Profile
	var ps Profiles
	ps.Records = []Profile{}
	query, err := models.Db.Query("SELECT `id`,`name`,`iface`,`host`,`stream`,`resend_delay`,`resend_count` FROM `profile`")
	if err != nil {
		return ps, err
	}
	defer query.Close()

	for query.Next() {
		err = query.Scan(&p.Id, &p.Name, &p.Iface, &p.Host, &p.Stream, &p.ResendDelay, &p.ResendCount)
		ps.Records = append(ps.Records, p)
	}
	err = models.Db.QueryRow("SELECT COUNT(*) FROM `profile`").Scan(&ps.Total)
	return ps, err
}
