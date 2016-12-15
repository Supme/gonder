package api

import (
"net/http"
"github.com/supme/gonder/models"
"encoding/json"
)

type UnitList struct {
	Id   int64  `json:"recid"`
	Name string `json:"name"`
}

func units(w http.ResponseWriter, r *http.Request) {

	var err error
	var js []byte
	if r.FormValue("request") == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	req, err := parseRequest(r.FormValue("request"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
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
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}


		case "add":

		}

	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}