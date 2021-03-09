package api

import (
	"encoding/json"
	"errors"
	"gonder/models"
	"log"
)

func users(req request) (js []byte, err error) {
	type users struct {
		Total   int64         `json:"total"`
		Records []models.User `json:"records"`
	}

	if req.auth.IsAdmin() {
		switch req.Cmd {
		case "get":
			var (
				userList           users
				partWhere, where   string
				partParams, params []interface{}
			)
			userList.Records = []models.User{}

			where = " WHERE 1=1 "
			partWhere, partParams, err = createSQLPart(req, where, params, map[string]string{"recid": "id", "name": "name"}, true)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			query, err := models.Db.Query("SELECT `id` FROM `auth_user`"+partWhere, partParams...)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			defer query.Close()

			var id int64
			for query.Next() {
				err = query.Scan(&id)
				if err != nil {
					log.Println(err)
					return nil, err
				}
				user, err := models.UserGetByID(id)
				if err != nil {
					log.Println(err)
					return nil, err
				}
				userList.Records = append(userList.Records, *user)
			}

			partWhere, partParams, err = createSQLPart(req, where, params, map[string]string{"recid": "id", "name": "name"}, false)
			if err != nil {
				log.Println(err)
			}
			err = models.Db.QueryRow("SELECT COUNT(*) FROM `auth_user` "+partWhere, partParams...).Scan(&userList.Total)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			js, err = json.Marshal(userList)
			if err != nil {
				log.Println(err)
			}
			return js, err

		case "save":
			// Change user
			user := models.User{
				ID:       req.Record.ID,
				Name:     req.Record.Name,
				Password: req.Record.Password,
				UnitID:   req.Record.Unit.ID,
				Blocked:  req.Record.Blocked != 0,
			}
			for i := range req.Record.Group {
				user.GroupsID = append(user.GroupsID, req.Record.Group[i].ID)
			}
			err = user.Update()
			return nil, err

		case "add":
			// Add user
			user := models.User{
				ID:       req.Record.ID,
				Name:     req.Record.Name,
				Password: req.Record.Password,
				UnitID:   req.Record.Unit.ID,
				Blocked:  req.Record.Blocked != 0,
			}
			for i := range req.Record.Group {
				user.GroupsID = append(user.GroupsID, req.Record.Group[i].ID)
			}
			err = user.Add()
			return nil, err

		default:
			err = errors.New("command not found")

		}
	} else {
		return nil, errors.New("access denied")
	}

	return js, err

}
