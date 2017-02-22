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

type Group struct {
	Id   int64  `json:"recid"`
	Name string `json:"name"`
}
type Groups struct {
	Total   int64   `json:"total"`
	Records []Group `json:"records"`
}

func groups(req request) (js []byte, err error) {

	var groups Groups

	switch req.Cmd {

	case "get":
		if auth.Right("get-groups") {
			groups, err = getGroups(req)
			if err != nil {
				return js, err
			}
			js, err = json.Marshal(groups)
			return js, err
		} else {
			return js, errors.New("Forbidden get group")
		}

	case "save":
		if auth.Right("save-groups") {
			err := saveGroups(req.Changes)
			if err != nil {
				return js, err
			}
			groups, err = getGroups(req)
			if err != nil {
				return js, err
			}
			js, err = json.Marshal(groups)
			return js, err
		} else {
			return js, errors.New("Forbidden save groups")
		}

	case "add":
		if auth.Right("add-groups") {
			group, err := addGroup()
			if err != nil {
				return js, err
			}
			js, err = json.Marshal(group)
			if err != nil {
				return js, err
			}
		} else {
			return js, errors.New("Forbidden add groups")
		}

	default:
		err = errors.New("Command not found")
	}

	return js, err
}

func addGroup() (Group, error) {
	g := Group{}
	g.Name = "New group"
	row, err := models.Db.Exec("INSERT INTO `group`(`name`) VALUES (?)", g.Name)
	if err != nil {
		return g, err
	}
	g.Id, err = row.LastInsertId()
	if err != nil {
		return g, err
	}

	return g, nil
}

func saveGroups(changes []map[string]interface{}) (err error) {
	var e error
	var where string
	err = nil

	if auth.IsAdmin() {
		where = "?"
	} else {
		where = "id IN (SELECT `group_id` FROM `auth_user_group` WHERE `auth_user_id`=?)"
	}

	for _, change := range changes {
		_, e = models.Db.Exec("UPDATE `group` SET `name`=? WHERE id=? AND "+where, change["name"], change["recid"], auth.userId)
		if e != nil {
			err = e
		}
	}
	return
}

func getGroups(req request) (Groups, error) {
	var (
		g Group
		gs Groups
		partWhere, where string
		partParams, params []interface{}
		err error
	)
	gs.Records = []Group{}
	if !auth.IsAdmin() {
		where = "WHERE id IN (SELECT `group_id` FROM `auth_user_group` WHERE `auth_user_id`=?)"
		params = append(params, auth.userId)
	} else {
		where = "WHERE 1=1"
	}
	partWhere, partParams, err = createSqlPart(req, where, params, map[string]string{"recid":"id", "name":"name"}, true)
	if err != nil {
		apilog.Print(err)
	}
	query, err := models.Db.Query("SELECT `id`, `name` FROM `group` " + partWhere , partParams...)
	if err != nil {
		return gs, err
	}
	defer query.Close()
	for query.Next() {
		err = query.Scan(&g.Id, &g.Name)
		gs.Records = append(gs.Records, g)
	}
	partWhere, partParams, err = createSqlPart(req, where, params, map[string]string{"recid":"id", "name":"name"}, false)
	if err != nil {
		apilog.Print(err)
	}
	err = models.Db.QueryRow("SELECT COUNT(*) FROM `group` " + partWhere, partParams...).Scan(&gs.Total)
	return gs, err
}
