package api

import (
	"encoding/json"
	"errors"
	"github.com/supme/gonder/models"
)

type grp struct {
	ID   int64  `json:"recid"`
	Name string `json:"name"`
}
type grps struct {
	Total   int64 `json:"total"`
	Records []grp `json:"records"`
}

func groups(req request) (js []byte, err error) {

	var (
		g  grp
		gs grps
	)

	switch req.Cmd {

	case "get":
		if user.Right("get-groups") {
			gs, err = getGroups(req)
			if err != nil {
				return js, err
			}
			js, err = json.Marshal(gs)
			return js, err
		}
		return js, errors.New("Forbidden get group")

	case "save":
		if user.Right("save-groups") {
			err = saveGroups(req.Changes)
			if err != nil {
				return js, err
			}
			gs, err = getGroups(req)
			if err != nil {
				return js, err
			}
			js, err = json.Marshal(gs)
			return js, err
		}
		return js, errors.New("Forbidden save groups")

	case "add":
		if user.Right("add-groups") {
			g, err = addGroup()
			if err != nil {
				return js, err
			}
			js, err = json.Marshal(g)
			if err != nil {
				return js, err
			}
		}
		return js, errors.New("Forbidden add groups")

	}

	return js, errors.New("Command not found")
}

func addGroup() (grp, error) {
	g := grp{}
	g.Name = "New group"
	row, err := models.Db.Exec("INSERT INTO `group`(`name`) VALUES (?)", g.Name)
	if err != nil {
		return g, err
	}
	g.ID, err = row.LastInsertId()
	if err != nil {
		return g, err
	}

	return g, nil
}

func saveGroups(changes []map[string]interface{}) (err error) {
	var e error
	var where string
	err = nil

	if user.IsAdmin() {
		where = "?"
	} else {
		where = "id IN (SELECT `group_id` FROM `auth_user_group` WHERE `auth_user_id`=?)"
	}

	for _, change := range changes {
		_, e = models.Db.Exec("UPDATE `group` SET `name`=? WHERE id=? AND "+where, change["name"], change["recid"], user.userID)
		if e != nil {
			err = e
		}
	}
	return
}

func getGroups(req request) (grps, error) {
	var (
		g                  grp
		gs                 grps
		partWhere, where   string
		partParams, params []interface{}
		err                error
	)
	gs.Records = []grp{}
	if !user.IsAdmin() {
		where = "WHERE id IN (SELECT `group_id` FROM `auth_user_group` WHERE `auth_user_id`=?)"
		params = append(params, user.userID)
	} else {
		where = "WHERE 1=1"
	}
	partWhere, partParams, err = createSQLPart(req, where, params, map[string]string{"recid": "id", "name": "name"}, true)
	if err != nil {
		apilog.Print(err)
	}
	query, err := models.Db.Query("SELECT `id`, `name` FROM `group` "+partWhere, partParams...)
	if err != nil {
		return gs, err
	}
	defer query.Close()
	for query.Next() {
		err = query.Scan(&g.ID, &g.Name)
		gs.Records = append(gs.Records, g)
	}
	partWhere, partParams, err = createSQLPart(req, where, params, map[string]string{"recid": "id", "name": "name"}, false)
	if err != nil {
		apilog.Print(err)
	}
	err = models.Db.QueryRow("SELECT COUNT(*) FROM `group` "+partWhere, partParams...).Scan(&gs.Total)
	return gs, err
}
