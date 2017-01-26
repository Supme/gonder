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
	"regexp"
	"strings"
	"time"
	"errors"
)

type Data struct {
	Id              int64 `json:"recid"`
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

func campaign(req request) (js []byte, err error) {
	switch req.Cmd {
	case "get":
		if auth.Right("get-campaign") && auth.CampaignRight(req.Id) {
			var start, end mysql.NullTime
			err = models.Db.QueryRow("SELECT `id`, `name`,`profile_id`,`subject`,`sender_id`,`start_time`,`end_time`,`send_unsubscribe`,`body`,`accepted` FROM campaign WHERE id=?", req.Id).Scan(
				&req.Content.Id,
				&req.Content.Name,
				&req.Content.ProfileId,
				&req.Content.Subject,
				&req.Content.SenderId,
				&start,
				&end,
				&req.Content.SendUnsubscribe,
				&req.Content.Template,
				&req.Content.Accepted,
			)
			if err != nil {
				return js, err
			}
			req.Content.StartDate = start.Time.Unix()
			req.Content.EndDate = end.Time.Unix()

			js, err = json.Marshal(req.Content)
			if err != nil {
				return js, err
			}
		} else {
			return js, errors.New("Forbidden get campaign")
		}

	case "save":
		if auth.Right("save-campaign") && auth.CampaignRight(req.Id) {
			start := time.Unix(req.Content.StartDate, 0).Format(time.RFC3339)
			end := time.Unix(req.Content.EndDate, 0).Format(time.RFC3339)

			// fix visual editor replace &amp;
			r := regexp.MustCompile(`(href|src)=["'](.*?)["']`)
			req.Content.Template = r.ReplaceAllStringFunc(req.Content.Template, func(str string) string {
				return strings.Replace(str, "&amp;", "&", -1)
			})

			_, err := models.Db.Exec("UPDATE campaign SET `name`=?,`profile_id`=?,`subject`=?,`sender_id`=?,`start_time`=?,`end_time`=?,`send_unsubscribe`=?,`body`=? WHERE id=?",
				req.Content.Name,
				req.Content.ProfileId,
				req.Content.Subject,
				req.Content.SenderId,
				start,
				end,
				req.Content.SendUnsubscribe,
				req.Content.Template,
				req.Id,
			)

			if err != nil {
				return js, err
			}

			js, err = json.Marshal(req.Content)
		} else {
			return js, errors.New("Forbidden save campaign")
		}

	case "accept":
		if auth.Right("accept-campaign") && auth.CampaignRight(req.Id) {
			var accepted int
			if req.Select {
				accepted = 1
			} else {
				accepted = 0
			}

			_, err := models.Db.Exec("UPDATE campaign SET `accepted`=? WHERE id=?", accepted, req.Id)
			if err != nil {
				return js, err
			}
		} else {
			return js, errors.New("Forbidden accept campaign")
		}

	default:
		err = errors.New("Command not found")
	}

	return js, err
}
