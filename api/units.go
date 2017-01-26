package api

import (
	"github.com/supme/gonder/models"
	"encoding/json"
)

func units(req request) (js []byte, err error) {
	type UnitList struct {
		Id   int64  `json:"recid"`
		Name string `json:"name"`
	}

	if auth.IsAdmin() {
		var (
			sl = []UnitList{}
			id int64
			name string
		)
		switch req.Cmd {

		case "get":
			query, err := models.Db.Query("SELECT `id`, `name` FROM `auth_unit`")
			if err != nil {
				return js, err
			}
			defer query.Close()
			for query.Next() {
				err = query.Scan(&id, &name)

				sl = append(sl, UnitList{
					Id:   id,
					Name: name,
				})
			}
			js, err = json.Marshal(sl)
			if err != nil {
				return js, err
			}


		case "add":

		}

	}
	return js, err
}