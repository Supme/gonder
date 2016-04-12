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
	Id   int64 `json:"recid"`
	Name string `json:"name"`
}
type Campaigns struct {
	Total	    int64 `json:"total"`
	Records		[]Campaign `json:"records"`
}

func campaigns(w http.ResponseWriter, r *http.Request)  {

	var campaigns Campaigns
	var err error
	var js []byte

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
		if auth.Right("get-campaigns") {
			campaigns, err = getCampaigns(group, r.Form["offset"][0], r.Form["limit"][0])
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			js, err = json.Marshal(campaigns)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			js = []byte(`{"status": "error", "message": "Forbidden get campaigns"}`)
		}
		break

	case "save-records":
		if auth.Right("save-campaigns") {
			arrForm := parseArrayForm(r.Form)
			err := saveCampaigns(arrForm["changes"])
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			campaigns, err = getCampaigns(group, r.Form["offset"][0], r.Form["limit"][0])
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			js, err = json.Marshal(campaigns)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			js = []byte(`{"status": "error", "message": "Forbidden save campaigns"}`)
		}
		break

	case "add-record":
		if auth.Right("add-campaigns") {
			campaign, err := addCampaign(r.Form["group"][0])
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			js, err = json.Marshal(campaign)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			js = []byte(`{"status": "error", "message": "Forbidden add campaigns"}`)
		}
		break
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func addCampaign(groupId string) (Campaign, error) {
	c := Campaign{}
	c.Name = "New campaign"
	row, err := models.Db.Exec("INSERT INTO `campaign`(`group_id`, `name`) VALUES (?, ?)", groupId, c.Name)
	if err != nil {
		return c, err
	}
	c.Id, err = row.LastInsertId()
	if err != nil {
		return c, err
	}

	return c, nil
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

func getCampaigns(group, offset, limit string) (Campaigns, error) {
	var c Campaign
	var cs Campaigns
	var where string
	cs.Records = []Campaign{}

	if auth.IsAdmin() {
		where = "?"
	} else {
		where = "group_id IN (SELECT `group_id` FROM `auth_user_group` WHERE `auth_user_id`=?)"
	}

	query, err := models.Db.Query("SELECT `id`, `name` FROM `campaign` WHERE `group_id`=? AND " + where + " LIMIT ? OFFSET ?", group, auth.userId , limit, offset)
	if err != nil {
		return cs, err
	}
	defer query.Close()
	for query.Next() {
		err = query.Scan(&c.Id, &c.Name)
		cs.Records = append(cs.Records, c)
	}
	err = models.Db.QueryRow("SELECT COUNT(*) FROM `campaign` WHERE `group_id`=? AND " + where, group, auth.userId).Scan(&cs.Total)
	return cs, err
}
