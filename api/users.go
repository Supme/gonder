package api

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"gonder/models"
	"log"
)

func users(req request) (js []byte, err error) {
	type userList struct {
		ID       int64   `json:"recid"`
		UnitID   int64   `json:"unitid"`
		GroupsID []int64 `json:"groupsid"`
		Name     string  `json:"name"`
		Password string  `json:"password"`
	}

	type users struct {
		Total   int64      `json:"total"`
		Records []userList `json:"records"`
	}

	if req.auth.IsAdmin() {
		switch req.Cmd {
		case "get":
			var (
				sl                 = users{}
				id, unitid, grpid  int64
				name               string
				partWhere, where   string
				partParams, params []interface{}
			)
			sl.Records = []userList{}

			where = " WHERE 1=1 "
			partWhere, partParams, err = createSQLPart(req, where, params, map[string]string{"recid": "id", "name": "name"}, true)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			query, err := models.Db.Query("SELECT `id`, `auth_unit_id`, `name` FROM `auth_user`"+partWhere, partParams...)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			defer query.Close()
			for query.Next() {
				err = query.Scan(&id, &unitid, &name)
				if err != nil {
					log.Println(err)
					return nil, err
				}
				groupq, err := models.Db.Query("SELECT `group_id` FROM `auth_user_group` WHERE `auth_user_id`=?", id)
				if err != nil {
					log.Println(err)
					return nil, err
				}
				defer groupq.Close()
				groupsid := []int64{}
				for groupq.Next() {
					err = groupq.Scan(&grpid)
					if err != nil {
						log.Println(err)
					}
					groupsid = append(groupsid, grpid)
				}
				sl.Records = append(sl.Records, userList{
					ID:       id,
					UnitID:   unitid,
					GroupsID: groupsid,
					Name:     name,
				})
			}

			partWhere, partParams, err = createSQLPart(req, where, params, map[string]string{"recid": "id", "name": "name"}, false)
			if err != nil {
				log.Println(err)
			}
			err = models.Db.QueryRow("SELECT COUNT(*) FROM `auth_user` "+partWhere, partParams...).Scan(&sl.Total)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			js, err = json.Marshal(sl)
			if err != nil {
				log.Println(err)
			}
			return js, err

		case "save":
			// Change user
			if req.Record.Password != "" {
				hash := sha256.New()
				if _, err := hash.Write([]byte(fmt.Sprint(req.Record.Password))); err != nil {
					log.Print(err)
					return nil, err
				}
				md := hash.Sum(nil)
				shaPassword := hex.EncodeToString(md)
				_, err = models.Db.Exec("UPDATE `auth_user` SET `password`=?, auth_unit_id=? WHERE `id`=?", shaPassword, req.Record.Unit.ID, req.Record.ID)
				if err != nil {
					log.Println(err)
					return nil, err
				}
			} else {
				_, err = models.Db.Exec("UPDATE `auth_user` SET auth_unit_id=? WHERE `id`=?", req.Record.Unit.ID, req.Record.ID)
				if err != nil {
					log.Println(err)
					return nil, err
				}
			}
			_, err = models.Db.Exec("DELETE FROM `auth_user_group` WHERE `auth_user_id`=?", req.Record.ID)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			for _, g := range req.Record.Group {
				_, err = models.Db.Exec("INSERT INTO `auth_user_group` (`auth_user_id`, `group_id`) VALUES (?, ?)", req.Record.ID, g.ID)
				if err != nil {
					log.Println(err)
					return nil, err
				}
			}

		case "add":
			// Add user
			row := models.Db.QueryRow("SELECT 1 FROM `auth_user` WHERE `name`=?", req.Record.Name).Scan()
			if row == sql.ErrNoRows {
				if req.Record.Password == "" {
					return nil, errors.New("Password can not be empty")
				}
				if req.Record.Name == "" {
					return nil, errors.New("name can not be empty")
				}
				hash := sha256.New()
				_, err = hash.Write([]byte(fmt.Sprint(req.Record.Password)))
				if err != nil {
					log.Println(err)
				}
				md := hash.Sum(nil)
				shaPassword := hex.EncodeToString(md)
				s, err := models.Db.Exec("INSERT INTO `auth_user`(`auth_unit_id`, `name`, `password`) VALUES (?, ?, ?)", req.Record.Unit.ID, req.Record.Name, shaPassword)
				if err != nil {
					log.Println(err)
					return nil, err
				}
				req.Record.ID, err = s.LastInsertId()
				if err != nil {
					log.Println(err)
					return nil, err
				}
				for _, g := range req.Record.Group {
					_, err = models.Db.Exec("INSERT INTO `auth_user_group` (`auth_user_id`, `group_id`) VALUES (?, ?)", req.Record.ID, g.ID)
					if err != nil {
						log.Println(err)
						return nil, err
					}
				}
			} else {
				return nil, errors.New("User name already exist")
			}

		default:
			err = errors.New("Command not found")

		}
	} else {
		return nil, errors.New("Access denied")
	}

	return js, err

}
