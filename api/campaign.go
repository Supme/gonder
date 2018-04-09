package api

import (
	"encoding/json"
	"errors"
	"github.com/go-sql-driver/mysql"
	campSender "github.com/supme/gonder/campaign"
	"github.com/supme/gonder/models"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var rexpRemoveQuotes = regexp.MustCompile(`(href|src)=["'](.*?)["']`)

func campaign(req request) (js []byte, err error) {
	switch req.Cmd {
	case "get":
		if user.Right("get-campaign") && user.CampaignRight(req.ID) {
			var start, end mysql.NullTime
			err = models.Db.QueryRow("SELECT `id`, `name`,`profile_id`,`subject`,`sender_id`,`start_time`,`end_time`,`send_unsubscribe`,`body`,`accepted` FROM campaign WHERE id=?", req.ID).Scan(
				&req.Content.ID,
				&req.Content.Name,
				&req.Content.ProfileID,
				&req.Content.Subject,
				&req.Content.SenderID,
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
		if user.Right("save-campaign") && user.CampaignRight(req.ID) {
			start := time.Unix(req.Content.StartDate, 0).Format(time.RFC3339)
			end := time.Unix(req.Content.EndDate, 0).Format(time.RFC3339)

			// fix visual editor replace &amp;
			req.Content.Template = rexpRemoveQuotes.ReplaceAllStringFunc(req.Content.Template, func(str string) string {
				return strings.Replace(str, "&amp;", "&", -1)
			})

			// ToDo check right working template

			_, err = models.Db.Exec("UPDATE campaign SET `name`=?,`profile_id`=?,`subject`=?,`sender_id`=?,`start_time`=?,`end_time`=?,`send_unsubscribe`=?,`body`=? WHERE id=?",
				req.Content.Name,
				req.Content.ProfileID,
				req.Content.Subject,
				req.Content.SenderID,
				start,
				end,
				req.Content.SendUnsubscribe,
				req.Content.Template,
				req.ID,
			)
			if err != nil {
				return js, err
			}

			js, err = json.Marshal(req.Content)
		} else {
			return js, errors.New("Forbidden save campaign")
		}

	case "accept":
		if user.Right("accept-campaign") && user.CampaignRight(req.ID) {
			var accepted int
			if req.Select {
				accepted = 1
			} else {
				accepted = 0
			}
			_, err = models.Db.Exec("UPDATE campaign SET `accepted`=? WHERE id=?", accepted, req.ID)
			if err != nil {
				return js, err
			}
			if accepted == 0 {
				go campSender.Sending.Stop(strconv.Itoa(int(req.ID)))
			}
		} else {
			return js, errors.New("Forbidden accept campaign")
		}

	default:
		err = errors.New("Command not found")
	}

	return js, err
}
