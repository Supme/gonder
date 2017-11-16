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
	"errors"
	"github.com/supme/gonder/models"
	"strconv"
)

type Sender struct {
	Id           int64  `json:"recid"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	DkimSelector string `json:"dkimSelector"`
	DkimKey      string `json:"dkimKey"`
	DkimUse      bool   `json:"dkimUse"`
}
type Senders struct {
	Total   int64    `json:"total"`
	Records []Sender `json:"records"`
}

func sender(req request) (js []byte, err error) {

	switch req.Cmd {

	case "get":
		if auth.Right("get-groups") && auth.GroupRight(req.Id) {
			var f Sender
			var fs Senders
			fs.Records = []Sender{}
			query, err := models.Db.Query("SELECT `id`, `email`, `name`, `dkim_selector`, `dkim_key`, `dkim_use` FROM `sender` WHERE `group_id`=? LIMIT ? OFFSET ?", req.Id, req.Limit, req.Offset)
			if err != nil {
				return js, err
			}
			defer query.Close()

			for query.Next() {
				err = query.Scan(&f.Id, &f.Email, &f.Name, &f.DkimSelector, &f.DkimKey, &f.DkimUse)
				fs.Records = append(fs.Records, f)
			}
			err = models.Db.QueryRow("SELECT COUNT(*) FROM `sender` WHERE `group_id`=?", req.Group).Scan(&fs.Total)
			js, err = json.Marshal(fs)
			return js, err
		} else {
			return js, errors.New("Forbidden get groups")
		}

	case "save":
		if auth.Right("save-groups") && auth.GroupRight(req.Id) {
			var group int64
			err = models.Db.QueryRow("SELECT `group_id` FROM `sender` WHERE `id`=?", req.Id).Scan(&group)
			if err != nil {
				return js, err
			}
			if auth.GroupRight(group) {
				_, err = models.Db.Exec("UPDATE `sender` SET `email`=?, `name`=?, `dkim_selector`=?, `dkim_key`=?, `dkim_use`=? WHERE `id`=?", req.Email, req.Name, req.DkimSelector, req.DkimKey, req.DkimUse, req.Id)
				if err != nil {
					return js, err
				}
			} else {
				return js, errors.New("Forbidden right to this group")
			}
		} else {
			return js, errors.New("Forbidden save groups")
		}

	case "add":
		if auth.Right("save-groups") && auth.GroupRight(req.Id) {
			res, err := models.Db.Exec("INSERT INTO `sender` (`group_id`, `email`, `name`, `dkim_selector`, `dkim_key`, `dkim_use`) VALUES (?, ?, ?, ?, ?, ?);", req.Id, req.Email, req.Name, req.DkimSelector, req.DkimKey, req.DkimUse)
			if err != nil {
				return js, err
			}
			recid, err := res.LastInsertId()
			if err != nil {
				return js, err
			}
			js = []byte(`{"status": "success", "message": "", "recid": ` + strconv.FormatInt(recid, 10) + `}`)
		} else {
			return js, errors.New("Forbidden save groups")
		}
	default:
		err = errors.New("Command not found")
	}
	return js, err
}
