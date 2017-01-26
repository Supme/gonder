package api

import (
	"github.com/supme/gonder/models"
	"encoding/json"
	"fmt"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"errors"
)

type UserList struct {
	Id   int64  `json:"recid"`
	UnitId int64 `json:"unitid"`
	GroupsId []int64 `json:"groupsid"`
	Name string `json:"name"`
	Password string `json:"password"`
}

func users(req request) (js []byte, err error) {
	if auth.IsAdmin() {
		var (
			sl = []UserList{}
			id, unitid, grpid int64
			name string
		)
		switch req.Cmd {

		case "get":
			query, err := models.Db.Query("SELECT `id`, `auth_unit_id`, `name` FROM `auth_user`")
			if err != nil {
				return nil, err
			}
			defer query.Close()
			for query.Next() {
				err = query.Scan(&id, &unitid, &name)
				groupq, err := models.Db.Query("SELECT `group_id` FROM `auth_user_group` WHERE `auth_user_id`=?", id)
				if err != nil {
					return nil, err
				}
				defer groupq.Close()
				groupsid := []int64{}
				for groupq.Next() {
					groupq.Scan(&grpid)
					groupsid = append(groupsid, grpid)
				}
				sl = append(sl, UserList{
					Id:   id,
					UnitId: unitid,
					GroupsId: groupsid,
					Name: name,
				})
			}
			js, err = json.Marshal(sl)
			return js, err

		case "save":
			fmt.Printf("Id: '%v'\n",req.Record.Id)
			fmt.Printf("Name: '%v'\n",req.Record.Name)
			fmt.Printf("Password: '%v'\n",req.Record.Password)
			fmt.Printf("UnitId: '%v'\n",req.Record.Unit.Id)
			for _, g := range req.Record.Group {
				fmt.Printf("GroupId: '%v'\n", g)
			}

			// Change user
			if req.Record.Password != "" {
				hash := sha256.New()
				hash.Write([]byte(fmt.Sprint(req.Record.Password)))
				md := hash.Sum(nil)
				shaPassword := hex.EncodeToString(md)
				fmt.Println("Password change, SHA256=",shaPassword)
				if req.Record.Name != "" && req.Record.Id == 0 {
					//New user
					return nil, errors.New("New user function not complete")
				} else {
					_, err = models.Db.Exec("UPDATE `auth_user` SET `password`=?, auth_unit_id=? WHERE `id`=?", shaPassword, req.Record.Unit.Id, req.Record.Id)
					if err != nil {
						return nil, err
					}
				}
			} else {
				if req.Record.Name == "" {
					_, err = models.Db.Exec("UPDATE `auth_user` SET auth_unit_id=? WHERE `id`=?", req.Record.Unit.Id, req.Record.Id)
					if err != nil {
						return nil, err
					}
				} else {
					return nil, errors.New("New user must have password")
				}
			}

			_, err = models.Db.Exec("DELETE FROM `auth_user_group` WHERE `auth_user_id`=?", req.Record.Id)
			if err != nil {
				return nil, err
			}
			val := []string{}
			for _, g := range req.Record.Group {
				val = append(val, fmt.Sprintf(`("%d", "%d")`, req.Record.Id, g.Id))
			}
			if len(val) != 0 {
				fmt.Println(fmt.Sprintf("INSERT INTO `auth_user_group` (`auth_user_id`, `group_id`) VALUES %s", strings.Join(val, ", ")))
				_, err = models.Db.Exec(fmt.Sprintf("INSERT INTO `auth_user_group` (`auth_user_id`, `group_id`) VALUES %s", strings.Join(val, ", ")))
				if err != nil {
					return nil, err
				}
			}

		default:
			err = errors.New("Command not found")

		}
	} else {
		return nil, errors.New("Access denied")
	}

	return js, err

}