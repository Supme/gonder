package api

import (
	"encoding/json"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gonder/models"
	"log"
	"time"
)

type camp struct {
	ID   int64  `json:"recid"`
	Name string `json:"name"`
}
type camps struct {
	Total   int64  `json:"total"`
	Records []camp `json:"records"`
}

func campaigns(req request) (js []byte, err error) {

	var cs camps

	switch req.Cmd {

	case "get":
		if req.auth.Right("get-campaigns") {
			cs, err = getCampaigns(req)
			if err != nil {
				return js, err
			}
			js, err = json.Marshal(cs)
			if err != nil {
				return js, err
			}
		} else {
			return js, errors.New("Forbidden get campaigns")
		}

	case "save":
		if req.auth.Right("save-campaigns") {
			err = saveCampaigns(req.Changes, req.auth)
			if err != nil {
				return js, err
			}
			cs, err = getCampaigns(req)
			if err != nil {
				return js, err
			}
			js, err = json.Marshal(cs)
			if err != nil {
				return js, err
			}
		} else {
			return js, errors.New("Forbidden save campaigns")
		}

	case "add":
		if req.auth.Right("add-campaigns") {
			var c camp
			c, err = addCampaign(req.ID)
			if err != nil {
				return js, err
			}
			js, err = json.Marshal(c)
			if err != nil {
				return js, err
			}
		} else {
			return js, errors.New("Forbidden add campaigns")
		}

	case "clone":
		if req.auth.Right("add-campaigns") {
			var c camp
			c, err := cloneCampaign(req.ID)
			if err != nil {
				return js, err
			}
			js, err = json.Marshal(c)
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

func cloneCampaign(campaignID int64) (camp, error) {
	c := camp{}
	cData := campaignData{
		Accepted: false,
	}
	var (
		groupID    int64
		start, end mysql.NullTime
	)
	query := models.Db.QueryRow("SELECT `group_id`,`profile_id`,`sender_id`,`name`,`subject`,`template_html`,`template_text`,`start_time`,`end_time`,`compress_html`,`send_unsubscribe` FROM `campaign` WHERE `id`=?", campaignID)

	err := query.Scan(
		&groupID,
		&cData.ProfileID,
		&cData.SenderID,
		&cData.Name,
		&cData.Subject,
		&cData.TemplateHTML,
		&cData.TemplateText,
		&start,
		&end,
		&cData.CompressHTML,
		&cData.SendUnsubscribe)

	if err != nil {
		log.Println(err)
		return c, err
	}
	cData.Name = "[Clone] " + cData.Name
	row, err := models.Db.Exec("INSERT INTO `campaign` (`group_id`,`profile_id`,`sender_id`,`name`,`subject`,`template_html`,`template_text`,`start_time`,`end_time`,`compress_html`,`send_unsubscribe`,`accepted`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)",
		groupID,
		cData.ProfileID,
		cData.SenderID,
		cData.Name,
		cData.Subject,
		cData.TemplateHTML,
		cData.TemplateText,
		start,
		end,
		cData.CompressHTML,
		cData.SendUnsubscribe,
		false,
	)
	if err != nil {
		log.Println(err)
		return c, err
	}
	c.ID, err = row.LastInsertId()
	if err != nil {
		log.Println(err)
		return c, err
	}
	c.Name = cData.Name

	return c, nil
}

func addCampaign(groupID int64) (camp, error) {
	c := camp{}
	c.Name = "New campaign"
	t := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC)
	row, err := models.Db.Exec("INSERT INTO `campaign`(`group_id`, `name`, `start_time`, `end_time`) VALUES (?, ?, ?, ?)", groupID, c.Name, t, t)
	if err != nil {
		log.Println(err)
		return c, err
	}
	c.ID, err = row.LastInsertId()
	if err != nil {
		log.Println(err)
		return c, err
	}

	return c, nil
}

func saveCampaigns(changes []map[string]interface{}, user *Auth) error {
	var where string

	if user.IsAdmin() {
		where = "?"
	} else {
		where = "group_id IN (SELECT `group_id` FROM `auth_user_group` WHERE `auth_user_id`=?)"
	}

	for _, change := range changes {
		_, err := models.Db.Exec("UPDATE `campaign` SET `name`=? WHERE id=? AND "+where, change["name"], change["recid"], user.userID)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

func getCampaigns(req request) (camps, error) {
	var (
		c                  camp
		cs                 camps
		partWhere, where   string
		partParams, params []interface{}
		err                error
	)
	cs.Records = []camp{}
	params = append(params, req.ID)
	if req.auth.IsAdmin() {
		where = "`group_id`=?"
	} else {
		where = "`group_id`=? AND `group_id` IN (SELECT `group_id` FROM `auth_user_group` WHERE `auth_user_id`=?)"
		params = append(params, req.auth.userID)
	}
	partWhere, partParams, err = createSQLPart(req, where, params, map[string]string{"recid": "id", "name": "name"}, true)
	if err != nil {
		log.Println(err)
	}
	query, err := models.Db.Query("SELECT `id`, `name` FROM `campaign` WHERE "+partWhere, partParams...)
	if err != nil {
		log.Println(err)
		return cs, err
	}
	defer query.Close()
	for query.Next() {
		err = query.Scan(&c.ID, &c.Name)
		if err != nil {
			log.Println(err)
			return camps{}, err
		}
		cs.Records = append(cs.Records, c)
	}
	partWhere, partParams, err = createSQLPart(req, where, params, map[string]string{"recid": "id", "name": "name"}, false)
	if err != nil {
		apilog.Print(err)
	}
	err = models.Db.QueryRow("SELECT COUNT(*) FROM `campaign` WHERE "+partWhere, partParams...).Scan(&cs.Total)
	if err != nil {
		log.Println(err)
		return cs, err
	}
	return cs, nil
}
