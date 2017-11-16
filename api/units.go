package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/supme/gonder/models"
	"strings"
)

func units(req request) (js []byte, err error) {
	type UnitList struct {
		Id   int64  `json:"recid"`
		Name string `json:"name"`
	}
	type Units struct {
		Total   int64      `json:"total"`
		Records []UnitList `json:"records"`
	}
	type UnitRight map[string]bool

	if auth.IsAdmin() {
		switch req.Cmd {

		case "get":
			var (
				sl                 = Units{}
				id                 int64
				name               string
				partWhere, where   string
				partParams, params []interface{}
			)
			sl.Records = []UnitList{}
			where = " WHERE 1=1 "
			partWhere, partParams, err = createSqlPart(req, where, params, map[string]string{"recid": "id", "name": "name"}, true)
			if err != nil {
				return nil, err
			}
			query, err := models.Db.Query("SELECT `id`, `name` FROM `auth_unit`"+partWhere, partParams...)
			if err != nil {
				return js, err
			}
			defer query.Close()
			for query.Next() {
				err = query.Scan(&id, &name)

				sl.Records = append(sl.Records, UnitList{
					Id:   id,
					Name: name,
				})
			}
			partWhere, partParams, err = createSqlPart(req, where, params, map[string]string{"recid": "id", "name": "name"}, false)
			if err != nil {
				apilog.Print(err)
			}
			err = models.Db.QueryRow("SELECT COUNT(*) FROM `auth_unit` "+partWhere, partParams...).Scan(&sl.Total)
			if err != nil {
				return nil, err
			}
			js, err = json.Marshal(sl)
			if err != nil {
				return js, err
			}

		case "add":
			res, err := models.Db.Exec("INSERT INTO `auth_unit`(`name`) VALUES (?)", req.Record.Name)
			if err != nil {
				return nil, err
			}
			req.Record.Id, err = res.LastInsertId()
			if err != nil {
				return nil, err
			}
			err = updateUnitRights(req)
			if err != nil {
				return nil, err
			}

		case "save":
			_, err = models.Db.Exec("UPDATE `auth_unit` SET `name`=? WHERE `id`=?", req.Record.Name, req.Record.Id)
			if err != nil {
				return nil, err
			}
			err = updateUnitRights(req)
			if err != nil {
				return nil, err
			}

		case "rights":
			var (
				name  string
				right bool
			)
			sl := UnitRight{}
			query, err := models.Db.Query("SELECT auth_right.name AS name, IF (auth_unit_right.auth_unit_id IS NOT NULL, true, false) AS `right` FROM auth_right LEFT OUTER JOIN auth_unit_right ON auth_unit_right.auth_right_id=auth_right.id AND auth_unit_right.auth_unit_id=?", req.Id)
			if err != nil {
				return js, err
			}
			defer query.Close()
			for query.Next() {
				err = query.Scan(&name, &right)

				sl[name] = right
			}
			js, err = json.Marshal(sl)
			if err != nil {
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
	_, err := models.Db.Exec("DELETE FROM `auth_unit_right` WHERE `auth_unit_id`=?", req.Record.Id)
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
		return err
	}
	defer query.Close()
	for query.Next() {
		err = query.Scan(&id, &name)
		rights[name] = id
	}

	queryData := []string{}
	if req.Record.GetGroups == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.Id, rights["get-groups"]))
	}
	if req.Record.SaveGroups == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.Id, rights["save-groups"]))
	}
	if req.Record.AddGroups == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.Id, rights["add-groups"]))
	}
	if req.Record.GetCampaigns == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.Id, rights["get-campaigns"]))
	}
	if req.Record.SaveCampaigns == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.Id, rights["save-campaigns"]))
	}
	if req.Record.AddCampaigns == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.Id, rights["add-campaigns"]))
	}
	if req.Record.GetCampaign == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.Id, rights["get-campaign"]))
	}
	if req.Record.SaveCampaign == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.Id, rights["save-campaign"]))
	}
	if req.Record.GetRecipients == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.Id, rights["get-recipients"]))
	}
	if req.Record.GetRecipientParameters == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.Id, rights["get-recipient-parameters"]))
	}
	if req.Record.UploadRecipients == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.Id, rights["upload-recipients"]))
	}
	if req.Record.DeleteRecipients == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.Id, rights["delete-recipients"]))
	}
	if req.Record.GetProfiles == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.Id, rights["get-profiles"]))
	}
	if req.Record.AddProfiles == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.Id, rights["add-profiles"]))
	}
	if req.Record.DeleteProfiles == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.Id, rights["delete-profiles"]))
	}
	if req.Record.SaveProfiles == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.Id, rights["save-profiles"]))
	}
	if req.Record.AcceptCampaign == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.Id, rights["accept-campaign"]))
	}
	if req.Record.GetLogMain == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.Id, rights["get-log-main"]))
	}
	if req.Record.GetLogApi == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.Id, rights["get-log-api"]))
	}
	if req.Record.GetLogCampaign == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.Id, rights["get-log-campaign"]))
	}
	if req.Record.GetLogUtm == 1 {
		queryData = append(queryData, fmt.Sprintf(`(%d, %d)`, req.Record.Id, rights["get-log-utm"]))
	}
	data := strings.Join(queryData, ", ")
	fmt.Println(data)

	_, err = models.Db.Exec("INSERT INTO `auth_unit_right`(`auth_unit_id`, `auth_right_id`) VALUES " + data)
	if err != nil {
		return err
	}

	return nil
}
