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

type Recipient struct {
	Id   int64 `json:"recid"`
	Name string `json:"name"`
	Email string `json:"email"`
	Result string `json:"result"`
}

type Recipients struct {
	Total	    int `json:"total"`
	Records		[]Recipient `json:"records"`
}

type RecipientParam struct  {
	Key string `json:"key"`
	Value string `json:"value"`
}

type RecipientParams struct {
	Total	    int `json:"total"`
	Records		[]RecipientParam `json:"records"`
}


func recipients(w http.ResponseWriter, r *http.Request)  {
	var err error
	var js []byte

	if err = r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if r.Form["content"][0] == "recipients" {
		switch r.Form["cmd"][0] {
		case "get-records":
			if auth.Right("get-recipients") {
				rs, err := getRecipients( r.Form["campaign"][0], r.Form["offset"][0], r.Form["limit"][0])
				js, err = json.Marshal(rs)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}  else {
				js = []byte(`{"status": "error", "message": "Forbidden get recipients"}`)
			}

			break
		}
	}

	if r.Form["content"][0] == "parameters" {
		switch r.Form["cmd"][0] {
		case "get-records":
			if auth.Right("get-recipient-parameters") {
				ps, err := getRecipientParams( r.Form["recipient"][0], r.Form["offset"][0], r.Form["limit"][0])
				js, err = json.Marshal(ps)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}  else {
				js = []byte(`{"status": "error", "message": "Forbidden get recipient parameters"}`)
			}

			break
		}
	}


	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

//ToDo check right errors
func getRecipients(campaign, offset, limit string) (Recipients, error) {
	var err error
	var r Recipient
	var rs Recipients
	rs.Records = []Recipient{}
	query, err := models.Db.Query("SELECT `id`, `name`, `email`, `status` FROM `recipient` WHERE `campaign_id`=?  LIMIT ? OFFSET ?", campaign, limit, offset)
	if err != nil {
		return rs, err
	}
	defer query.Close()
	for query.Next() {
		err = query.Scan(&r.Id, &r.Name, &r.Email, &r.Result)
		rs.Records = append(rs.Records, r)
	}
	err = models.Db.QueryRow("SELECT COUNT(*) FROM `recipient` WHERE `campaign_id`=?", campaign).Scan(&rs.Total)
	return rs, nil

}

//ToDo check right errors
func getRecipientParams(recipient, offset, limit string) (RecipientParams, error) {
	var err error
	var p RecipientParam
	var ps RecipientParams
	ps.Records = []RecipientParam{}
	query, err := models.Db.Query("SELECT `key`, `value` FROM `parameter` WHERE `recipient_id`=?", recipient)
	if err != nil {
		return ps, err
	}
	defer query.Close()
	for query.Next() {
		err = query.Scan(&p.Key, &p.Value)
		ps.Records = append(ps.Records, p)
	}
	err = models.Db.QueryRow("SELECT COUNT(*) FROM `parameter` WHERE `recipient_id`=?", recipient).Scan(&ps.Total)
	return ps, err
}