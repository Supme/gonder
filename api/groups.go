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

type Group struct {
	Id   int `json:"recid"`
	Name string `json:"name"`
}
type Groups struct {
	Total	    int `json:"total"`
	Records		[]Group `json:"records"`
}

func groups(w http.ResponseWriter, r *http.Request)  {

	var groups Groups
	var err error

	if err = r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Form["cmd"][0] {

	case "get-records":
		groups, err = getGroups(r.Form["offset"][0], r.Form["limit"][0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		break

	case "save-records":
		arrForm := parseArrayForm(r.Form)
		err := saveGroups(arrForm["changes"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		groups, err = getGroups(r.Form["offset"][0], r.Form["limit"][0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		break
	}

	js, err := json.Marshal(groups)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func saveGroups(changes map[string]map[string][]string) (err error) {
	var e error
	err = nil
	for _, change := range changes {
		_, e = models.Db.Exec("UPDATE `group` SET `name`=? WHERE id=?", change["name"][0], change["recid"][0])
		if e != nil {
			err = e
		}
	}
	return
}

func getGroups(offset, limit string) (Groups, error) {
	var group Group
	var groups Groups
	groups.Records = []Group{}
	query, err := models.Db.Query("SELECT `id`, `name` FROM `group`")
	if err != nil {
		return groups, err
	}
	defer query.Close()
	for query.Next() {
		err = query.Scan(&group.Id, &group.Name)
		groups.Records = append(groups.Records, group)
	}
	groups.Total = len(groups.Records)
	return groups, nil
}
