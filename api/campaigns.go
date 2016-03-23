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


type Campaign struct {
	Id   int `json:"recid"`
	Name string `json:"name"`
}
type Campaigns struct {
	Total	    int `json:"total"`
	Records		[]Campaign `json:"records"`
}

func campaigns(w http.ResponseWriter, r *http.Request)  {

	var campaigns Campaigns
	var err error
	if err = r.ParseForm(); err != nil {
		//handle error http.Error() for example
		return
	}

	group := "0";
	if  r.Form["group"] != nil {
		group = r.Form["group"][0]
	}

	switch r.Form["cmd"][0] {

	case "get-records":
		campaigns, err = getCampaigns(group, r.Form["offset"][0], r.Form["limit"][0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		break

	case "save-records":
		arrForm := parseArrayForm(r.Form)
		err := saveCampaigns(arrForm["changes"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		campaigns, err = getCampaigns(group, r.Form["offset"][0], r.Form["limit"][0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		break
	}

	arrForm := parseArrayForm(r.PostForm)

	_ = arrForm

	js, err := json.Marshal(campaigns)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func saveCampaigns(changes map[string]map[string][]string) (err error) {
	var e error
	err = nil
	for _, change := range changes {
		_, e = models.Db.Exec("UPDATE `campaign` SET `name`=? WHERE id=?", change["name"][0], change["recid"][0])
		if e != nil {
			err = e
		}
	}
	return
}

//ToDo check right errors
func getCampaigns(group, offset, limit string) (Campaigns, error) {
	var c Campaign
	var cs Campaigns
	cs.Records = []Campaign{}
	query, err := models.Db.Query("SELECT `id`, `name` FROM `campaign` WHERE `group_id`=? LIMIT ? OFFSET ?", group, limit, offset)
	if err != nil {
		return cs, err
	}
	defer query.Close()
	for query.Next() {
		err = query.Scan(&c.Id, &c.Name)
		cs.Records = append(cs.Records, c)
	}
	err = models.Db.QueryRow("SELECT COUNT(*) FROM `campaign` WHERE `group_id`=?", campaign).Scan(&cs.Total)
	return cs, nil
}
