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
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/supme/gonder/models"
)

type Campaign struct {
	Id   int64  `json:"recid"`
	Name string `json:"name"`
}
type Campaigns struct {
	Total   int64      `json:"total"`
	Records []Campaign `json:"records"`
}

func campaigns(req request) (js []byte, err error) {

	var campaigns Campaigns

	switch req.Cmd {

	case "get":
		if auth.Right("get-campaigns") {
			campaigns, err = getCampaigns(req)
			if err != nil {
				return js, err
			}
			js, err = json.Marshal(campaigns)
			if err != nil {
				return js, err
			}
		} else {
			return js, errors.New("Forbidden get campaigns")
		}

	case "save":
		if auth.Right("save-campaigns") {
			err := saveCampaigns(req.Changes)
			if err != nil {
				return js, err
			}
			campaigns, err = getCampaigns(req)
			if err != nil {
				return js, err
			}
			js, err = json.Marshal(campaigns)
			if err != nil {
				return js, err
			}
		} else {
			return js, errors.New("Forbidden save campaigns")
		}

	case "add":
		if auth.Right("add-campaigns") {
			campaign, err := addCampaign(req.Id)
			if err != nil {
				return js, err
			}
			js, err = json.Marshal(campaign)
			if err != nil {
				return js, err
			}
		} else {
			return js, errors.New("Forbidden add campaigns")
		}

	case "clone":
		if auth.Right("add-campaigns") {
			campaign, err := cloneCampaign(req.Id)
			if err != nil {
				return js, err
			}
			js, err = json.Marshal(campaign)
			if err != nil {
				return js, err
			}
		} else {
			return js, errors.New("Forbidden add campaigns")
		}

	default:
		err = errors.New("Command not found")
	}

	return js, err
}

func cloneCampaign(campaignId int64) (Campaign, error) {
	c := Campaign{}
	cData := CampaignData{
		Accepted: false,
	}
	var (
		groupId    int64
		start, end mysql.NullTime
	)
	query := models.Db.QueryRow("SELECT `group_id`,`profile_id`,`sender_id`,`name`,`subject`,`body`,`start_time`,`end_time`,`send_unsubscribe` FROM `campaign` WHERE `id`=?", campaignId)

	err := query.Scan(
		&groupId,
		&cData.ProfileId,
		&cData.SenderId,
		&cData.Name,
		&cData.Subject,
		&cData.Template,
		&start,
		&end,
		&cData.SendUnsubscribe)

	if err != nil {
		return c, err
	}
	cData.Name = "[Clone] " + cData.Name
	row, err := models.Db.Exec("INSERT INTO `campaign` (`group_id`,`profile_id`,`sender_id`,`name`,`subject`,`body`,`start_time`,`end_time`,`send_unsubscribe`,`accepted`) VALUES (?,?,?,?,?,?,?,?,?,?)",
		groupId,
		cData.ProfileId,
		cData.SenderId,
		cData.Name,
		cData.Subject,
		cData.Template,
		start,
		end,
		cData.SendUnsubscribe,
		cData.Accepted,
	)
	if err != nil {
		return c, err
	}
	c.Id, err = row.LastInsertId()
	if err != nil {
		return c, err
	}
	c.Name = cData.Name

	return c, err
}

func addCampaign(groupId int64) (Campaign, error) {
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

func saveCampaigns(changes []map[string]interface{}) (err error) {
	var e error
	err = nil
	var where string

	if auth.IsAdmin() {
		where = "?"
	} else {
		where = "group_id IN (SELECT `group_id` FROM `auth_user_group` WHERE `auth_user_id`=?)"
	}

	for _, change := range changes {
		_, e = models.Db.Exec("UPDATE `campaign` SET `name`=? WHERE id=? AND "+where, change["name"], change["recid"], auth.userId)
		if e != nil {
			err = e
		}
	}
	return
}

func getCampaigns(req request) (Campaigns, error) {
	var (
		c                  Campaign
		cs                 Campaigns
		partWhere, where   string
		partParams, params []interface{}
		err                error
	)
	cs.Records = []Campaign{}
	params = append(params, req.Id)
	if auth.IsAdmin() {
		where = "`group_id`=?"
	} else {
		where = "`group_id`=? AND `group_id` IN (SELECT `group_id` FROM `auth_user_group` WHERE `auth_user_id`=?)"
		params = append(params, auth.userId)
	}
	partWhere, partParams, err = createSqlPart(req, where, params, map[string]string{"recid": "id", "name": "name"}, true)
	if err != nil {
		fmt.Println("Create SQL Part error:", err)
	}
	query, err := models.Db.Query("SELECT `id`, `name` FROM `campaign` WHERE "+partWhere, partParams...)
	if err != nil {
		return cs, err
	}
	defer query.Close()
	for query.Next() {
		err = query.Scan(&c.Id, &c.Name)
		cs.Records = append(cs.Records, c)
	}
	partWhere, partParams, err = createSqlPart(req, where, params, map[string]string{"recid": "id", "name": "name"}, false)
	if err != nil {
		apilog.Print(err)
	}
	err = models.Db.QueryRow("SELECT COUNT(*) FROM `campaign` WHERE "+partWhere, partParams...).Scan(&cs.Total)
	return cs, err
}
