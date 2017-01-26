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

type ProfileList struct {
	Id   int    `json:"id"`
	Name string `json:"text"`
}

func profilesList(req request) (js []byte, err error) {
	if auth.Right("get-campaign") {
		psl, err := getProfilesList()
		if err != nil {
			return js, err
		}
		js, err = json.Marshal(psl)
		return js, err
	} else {
		return js, errors.New("Forbidden get campaign")
	}
	return js, err
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
