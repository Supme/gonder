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

type SenderList struct {
	Id   int64  `json:"id"`
	Text string `json:"text"`
}

func senderList(w http.ResponseWriter, r *http.Request) {
	var js []byte
	if auth.Right("get-group") && auth.GroupRight(r.FormValue("groupId")) {
		var id int64
		var email, name string
		var fs []SenderList
		fs = []SenderList{}
		query, err := models.Db.Query("SELECT `id`, `name`, `email` FROM `sender` WHERE `group_id`=?", r.FormValue("groupId"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer query.Close()
		for query.Next() {
			err = query.Scan(&id, &name, &email)

			fs = append(fs, SenderList{
				Id:   id,
				Text: name + " (" + email + ")",
			})
		}
		js, err = json.Marshal(fs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		js = []byte(`{"status": "error", "message": "Forbidden get from this group"}`)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
