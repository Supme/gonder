package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/supme/gonder/models"
	"log"
	"strings"
)

func units(req request) (js []byte, err error) {
	type unitList struct {
		ID   int64  `json:"recid"`
		Name string `json:"name"`
	}
	type units struct {
		Total   int64      `json:"total"`
		Records []unitList `json:"records"`
	}
	type unitRight map[string]bool

	if user.IsAdmin() {
		switch req.Cmd {

		case "get":
			var (
				sl                 = units{}
				id                 int64
				name               string
				partWhere, where   string
				partParams, params []interface{}
			)
			sl.Records = []unitList{}
			where = " WHERE 1=1 "
			partWhere, partParams, err = createSQLPart(req, where, params, map[string]string{"recid": "id", "name": "name"}, true)
			if err != nil {
				return nil, err
			}
			query, err := models.Db.Query("SELECT `id`, `name` FROM `auth_unit`"+partWhere, partParams...)
			if err != nil {
				log.Println(err)
				return js, err
			}
			defer query.Close()
			for query.Next() {
				err = query.Scan(&id, &name)
				if err != nil {
					log.Println(err)
					return nil, err
				}
				sl.Records = append(sl.Records, unitList{
					ID:   id,
					Name: name,
				})
			}
			partWhere, partParams, err = createSQLPart(req, where, params, map[string]string{"recid": "id", "name": "name"}, false)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			err = models.Db.QueryRow("SELECT COUNT(*) FROM `auth_unit` "+partWhere, partParams...).Scan(&sl.Total)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			js, err = json.Marshal(sl)
			if err != nil {
				log.Println(err)
				return js, err
			}

		case "add":
			res, err := models.Db.Exec("INSERT INTO `auth_unit`(`name`) VALUES (?)", req.Record.Name)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			req.Record.ID, err = res.LastInsertId()
			if err != nil {
				log.Println(err)
				return nil, err
			}
			err = updateUnitRights(req)
			if err != nil {
				log.Println(err)
				return nil, err
			}

		case "save":
			_, err = models.Db.Exec("UPDATE `auth_unit` SET `name`=? WHERE `id`=?", req.Record.Name, req.Record.ID)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			err = updateUnitRights(req)
			if err != nil {
				log.Println(err)
				return nil, err
			}

		case "rights":
			var (
				name  string
				right bool
			)
			sl := unitRight{}
			query, err := models.Db.Query("SELECT auth_right.name AS name, IF (auth_unit_right.auth_unit_id IS NOT NULL, true, false) AS `right` FROM auth_right LEFT OUTER JOIN auth_unit_right ON auth_unit_right.auth_right_id=auth_right.id AND auth_unit_right.auth_unit_id=?", req.ID)
			if err != nil {
				log.Println(err)
				return js, err
			}
			defer query.Close()
			for query.Next() {
				err = query.Scan(&name, &right)
				if err != nil {
					log.Println(err)
					return nil, err
				}
				sl[name] = right
			}
			js, err = json.Marshal(sl)
			if err != nil {
				log.Println(err)
				return js, err
			}

		default:
			err = errors.New("Command not found")
		}

	} else {
		return nil, errors.New("Access denied")
	}

	return js, err
}

func updateUnitRights(req request) error {
	_, err := models.Db.Exec("DELETE FROM `auth_unit_right` WHERE `auth_unit_id`=?", req.Record.ID)
	if err != nil {
		return err
	}

	var (
		id   int64
		name string
	)
	rights := map[string]int64{}
	query, err := models.Db.Query("SELECT `id`, `name` FROM `auth_right`")
	if err != nil {
		log.Println(err)
		return err
	}
	defer query.Close()
	for query.Next() {
		err = query.Scan(&id, &name)
		if err != nil {
			log.Println(err)
			return err
		}
		rights[name] = id
	}

	queryData := []string{}
	if req.Record.GetGroups == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.ID, rights["get-groups"]))
	}
	if req.Record.SaveGroups == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.ID, rights["save-groups"]))
	}
	if req.Record.AddGroups == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.ID, rights["add-groups"]))
	}
	if req.Record.GetCampaigns == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.ID, rights["get-campaigns"]))
	}
	if req.Record.SaveCampaigns == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.ID, rights["save-campaigns"]))
	}
	if req.Record.AddCampaigns == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.ID, rights["add-campaigns"]))
	}
	if req.Record.GetCampaign == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.ID, rights["get-campaign"]))
	}
	if req.Record.SaveCampaign == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.ID, rights["save-campaign"]))
	}
	if req.Record.GetRecipients == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.ID, rights["get-recipients"]))
	}
	if req.Record.GetRecipientParameters == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.ID, rights["get-recipient-parameters"]))
	}
	if req.Record.UploadRecipients == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.ID, rights["upload-recipients"]))
	}
	if req.Record.DeleteRecipients == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.ID, rights["delete-recipients"]))
	}
	if req.Record.GetProfiles == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.ID, rights["get-profiles"]))
	}
	if req.Record.AddProfiles == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.ID, rights["add-profiles"]))
	}
	if req.Record.DeleteProfiles == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.ID, rights["delete-profiles"]))
	}
	if req.Record.SaveProfiles == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.ID, rights["save-profiles"]))
	}
	if req.Record.AcceptCampaign == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.ID, rights["accept-campaign"]))
	}
	if req.Record.GetLogMain == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.ID, rights["get-log-main"]))
	}
	if req.Record.GetLogAPI == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.ID, rights["get-log-api"]))
	}
	if req.Record.GetLogCampaign == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.ID, rights["get-log-campaign"]))
	}
	if req.Record.GetLogUtm == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.ID, rights["get-log-utm"]))
	}
	data := strings.Join(queryData, ", ")
	fmt.Println(data)

	_, err = models.Db.Exec("INSERT INTO `auth_unit_right`(`auth_unit_id`, `auth_right_id`) VALUES " + data)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
