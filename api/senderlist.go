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
	"errors"
)

type SenderList struct {
	Id   int64  `json:"id"`
	Text string `json:"text"`
}

func senderList(req request) (js []byte, err error) {
	if auth.Right("get-groups") && auth.GroupRight(req.Id) {
		var id int64
		var email, name string
		var fs = []SenderList{}
		query, err := models.Db.Query("SELECT `id`, `name`, `email` FROM `sender` WHERE `group_id`=?", req.Id)
		if err != nil {
			return js, err
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
		return js, err
	} else {
		return js, errors.New("Forbidden get from this group")
	}
	return js, err
}
