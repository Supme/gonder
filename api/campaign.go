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
	"github.com/go-sql-driver/mysql"
	"github.com/supme/gonder/models"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Data struct {
	Id              string `json:"recid"` //ToDo Id is int64
	Name            string `json:"name"`
	ProfileId       int    `json:"profileId"`
	Subject         string `json:"subject"`
	SenderId        int    `json:"senderId"`
	StartDate       int64  `json:"startDate"`
	EndDate         int64  `json:"endDate"`
	SendUnsubscribe bool   `json:"sendUnsubscribe"`
	Accepted        bool   `json:"accepted"`
	Template        string `json:"template"`
}

func campaign(w http.ResponseWriter, r *http.Request) {
	var err error
	var js []byte
	var data Data
	data = Data{}

	if err = r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	switch r.Form["cmd"][0] {
	case "get-data":
		data.Id = r.Form["recid"][0]
		dataId, err := strconv.ParseInt(data.Id, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if auth.Right("get-campaign") && auth.CampaignRight(dataId) {
			var start, end mysql.NullTime
			err = models.Db.QueryRow("SELECT `name`,`profile_id`,`subject`,`sender_id`,`start_time`,`end_time`,`send_unsubscribe`,`body`,`accepted` FROM campaign WHERE id=?", data.Id).Scan(
				&data.Name,
				&data.ProfileId,
				&data.Subject,
				&data.SenderId,
				&start,
				&end,
				&data.SendUnsubscribe,
				&data.Template,
				&data.Accepted,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			data.StartDate = start.Time.Unix()
			data.EndDate = end.Time.Unix()

			js, err = json.Marshal(data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			js = []byte(`{"status": "error", "message": "Forbidden get campaign"}`)
		}

	case "save-data":
		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		dataId, err := strconv.ParseInt(data.Id, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if auth.Right("save-campaign") && auth.CampaignRight(dataId) {
			start := time.Unix(data.StartDate, 0).Format(time.RFC3339)
			end := time.Unix(data.EndDate, 0).Format(time.RFC3339)

			// fix visual editor replace &amp;
			r := regexp.MustCompile(`(href|src)=["'](.*?)["']`)
			data.Template = r.ReplaceAllStringFunc(data.Template, func(str string) string {
				return strings.Replace(str, "&amp;", "&", -1)
			})

			_, err := models.Db.Exec("UPDATE campaign SET `name`=?,`profile_id`=?,`subject`=?,`sender_id`=?,`start_time`=?,`end_time`=?,`send_unsubscribe`=?,`body`=? WHERE id=?",
				data.Name,
				data.ProfileId,
				data.Subject,
				data.SenderId,
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

			js, err = json.Marshal(data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			js = []byte(`{"status": "error", "message": "Forbidden save campaign"}`)
		}

	case "accept":
		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		dataId, err := strconv.ParseInt(data.Id, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if auth.Right("accept-campaign") && auth.CampaignRight(dataId) {
			var accepted int
			if data.Accepted {
				accepted = 1
			} else {
				accepted = 0
			}

			_, err := models.Db.Exec("UPDATE campaign SET `accepted`=? WHERE id=?", accepted, data.Id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			js = []byte(`{"status": "ok", "message": ""}`)

		} else {
			js = []byte(`{"status": "error", "message": "Forbidden accept campaign"}`)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
