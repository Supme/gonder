package api

import (
	"encoding/json"
	"errors"
	"gonder/models"
	"log"
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
		if req.auth.Right("get-groups") {
			gs, err = getGroups(req)
			if err != nil {
				return js, err
			}
			js, err = json.Marshal(gs)
			if err != nil {
				log.Println(err)
			}
			return js, err
		}
		return js, errors.New("Forbidden get group")

	case "save":
		if req.auth.Right("save-groups") {
			err = saveGroups(req.Changes, req.auth)
			if err != nil {
				return js, err
			}
			gs, err = getGroups(req)
			if err != nil {
				return js, err
			}
			js, err = json.Marshal(gs)
			if err != nil {
				log.Println(err)
			}
			return js, err
		}
		return js, errors.New("Forbidden save groups")

	case "add":
		if req.auth.Right("add-groups") {
			g, err = addGroup(req.Name)
			if err != nil {
				return js, err
			}
			js, err = json.Marshal(g)
			if err != nil {
				log.Println(err)
			}
			return js, err
		}
		return js, errors.New("Forbidden add groups")

	}

	return js, errors.New("Command not found")
}

func addGroup(name string) (grp, error) {
	g := grp{}
	if name == "" {
		name = "New group"
	}
	g.Name = name
	row, err := models.Db.Exec("INSERT INTO `group`(`name`) VALUES (?)", g.Name)
	if err != nil {
		log.Println(err)
		return g, err
	}
	g.ID, err = row.LastInsertId()
	if err != nil {
		log.Println(err)
		return g, err
	}

	return g, nil
}

func saveGroups(changes []map[string]interface{}, auth *Auth) (err error) {
	var where string

	if auth.IsAdmin() {
		where = "?"
	} else {
		where = "id IN (SELECT `group_id` FROM `auth_user_group` WHERE `auth_user_id`=?)"
	}

	for _, change := range changes {
		_, err := models.Db.Exec("UPDATE `group` SET `name`=? WHERE id=? AND "+where, change["name"], change["recid"], auth.user.ID)
		if err != nil {
			log.Println(err)
			return err
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
	if !req.auth.IsAdmin() {
		where = "WHERE id IN (SELECT `group_id` FROM `auth_user_group` WHERE `auth_user_id`=?)"
		params = append(params, req.auth.user.ID)
	} else {
		where = "WHERE 1=1"
	}
	partWhere, partParams, err = createSQLPart(req, where, params, map[string]string{"recid": "id", "name": "name"}, true)
	if err != nil {
		log.Println(err)
		return gs, err
	}
	query, err := models.Db.Query("SELECT `id`, `name` FROM `group` "+partWhere, partParams...)
	if err != nil {
		log.Println(err)
		return gs, err
	}
	defer query.Close()
	for query.Next() {
		err = query.Scan(&g.ID, &g.Name)
		if err != nil {
			log.Println(err)
			return grps{}, err
		}
		gs.Records = append(gs.Records, g)
	}
	partWhere, partParams, err = createSQLPart(req, where, params, map[string]string{"recid": "id", "name": "name"}, false)
	if err != nil {
		log.Println(err)
		return gs, err
	}
	err = models.Db.QueryRow("SELECT COUNT(*) FROM `group` "+partWhere, partParams...).Scan(&gs.Total)
	if err != nil {
		log.Println(err)
		return gs, err
	}
	return gs, nil
}
