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
	"github.com/go-sql-driver/mysql"
	"time"
)

type Data struct {
	Id   string `json:"recid"`
	Name string `json:"name"`
	ProfileId int `json:"profileId"`
	Subject string `json:"subject"`
	FromName string `json:"fromName"`
	FromEmail string `json:"fromEmail"`
	StartDate int64 `json:"startDate"`
	EndDate int64 `json:"endDate"`
	SendUnsubscribe string `json:"sendUnsubscribe"`
	Template string `json:"template"`
}

func campaign(w http.ResponseWriter, r *http.Request)  {
	var err error
	var data Data
	data = Data{}

	if err = r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Form["cmd"][0] {
	case "get-data":
		var start, end mysql.NullTime
		err = models.Db.QueryRow("SELECT `id`,`name`,`profile_id`,`subject`,`from_name`,`from`,`start_time`,`end_time`,`send_unsubscribe`,`body` FROM campaign WHERE id=?", r.Form["recid"][0]).Scan(
			&data.Id,
			&data.Name,
			&data.ProfileId,
			&data.Subject,
			&data.FromName,
			&data.FromEmail,
			&start,
			&end,
			&data.SendUnsubscribe,
			&data.Template,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data.StartDate = start.Time.Unix()
		data.EndDate = end.Time.Unix()
		data.EndDate = end.Time.Unix()
		break

	case "save-data":
		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		start := time.Unix(data.StartDate, 0).UTC().Format(time.RFC3339)
		end := time.Unix(data.EndDate, 0).UTC().Format(time.RFC3339)

		_, err := models.Db.Exec("UPDATE campaign SET `name`=?, `profile_id`=?, `subject`=?,`from_name`=?,`from`=?,`start_time`=?,`end_time`=?,`send_unsubscribe`=?,`body`=? WHERE id=?",
			data.Name,
			data.ProfileId,
			data.Subject,
			data.FromName,
			data.FromEmail,
			start,
			end,
			data.SendUnsubscribe,
			data.Template,
			data.Id,
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		break
	}


	js, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
