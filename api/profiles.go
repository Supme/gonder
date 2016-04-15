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
	"net/http"
	"encoding/json"
	"github.com/supme/gonder/models"
)

type ProfileList struct {
	Id   int `json:"id"`
	Name string `json:"text"`
}

type Profile struct {
	Id   int `json:"recid"`
	Name string `json:"name"`
	Iface string `json:"iface"`
	Host string `json:"host"`
	Stream int `json:"stream"`
	ResendDelay int `json:"resend_delay"`
	ResendCount int `json:"resend_count"`
}

type Profiles struct {
	Total	    int64 `json:"total"`
	Records		[]Profile `json:"records"`
}
//ToDo check rights
func profiles(w http.ResponseWriter, r *http.Request)  {
	var err error
	var js []byte

	if err = r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Form["cmd"][0] {

	case "get-list":
		ps, err := getProfilesList()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		js, err = json.Marshal(ps)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		break

	case "get-records":
		ps, err := getProfiles()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		js, err = json.Marshal(ps)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		break
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func getProfiles() (Profiles, error) {
	var p Profile
	var ps Profiles
	ps.Records = []Profile{}
	query, err := models.Db.Query("SELECT `id`,`name`,`iface`,`host`,`stream`,`resend_delay`,`resend_count` FROM `profile`")
//	query, err := models.Db.Query("SELECT `id`,`name` FROM `profile`")
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

func getProfilesList() ([]ProfileList, error) {
	var p ProfileList
	var ps []ProfileList
	ps = []ProfileList{}
	query, err := models.Db.Query("SELECT `id`, `name` FROM `profile`")
	if err != nil {
		return ps, err
	}
	defer query.Close()
	for query.Next() {
		err = query.Scan(&p.Id, &p.Name)
		ps = append(ps, p)
	}
	return ps, nil
}