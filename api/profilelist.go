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

type ProfileList struct {
	Id   int    `json:"id"`
	Name string `json:"text"`
}

func profilesList(w http.ResponseWriter, r *http.Request) {
	var js []byte
	if auth.Right("get-campaign") {
		psl, err := getProfilesList()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		js, err = json.Marshal(psl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		js = []byte(`{"status": "error", "message": "Forbidden get campaign"}`)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
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
