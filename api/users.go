package api

import (
	"net/http"
	"github.com/supme/gonder/models"
	"encoding/json"
)

type UserList struct {
	Id   int64  `json:"recid"`
	UnitId int64 `json:"unitid"`
	GroupsId []int64 `json:"groupsid"`
	Name string `json:"name"`
	Password string `json:"password"`
}

func users(w http.ResponseWriter, r *http.Request) {

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
			sl = []UserList{}
			id, unitid, grpid int64
			groupsid []int64
			name string
		)
		switch req.Cmd {

		case "get":
			query, err := models.Db.Query("SELECT `id`, `auth_unit_id`, `name` FROM `auth_user`")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer query.Close()
			for query.Next() {
				err = query.Scan(&id, &unitid, &name)
				groupq, err := models.Db.Query("SELECT `group_id` FROM `auth_user_group` WHERE `auth_user_id`=?", id)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer groupq.Close()
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
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		case "save":

			apilog.Print(req.Record)
		/*
			for _, change := range req.Changes {
				if change["name"] != nil {
					_, err = models.Db.Exec("UPDATE `auth_user` SET `name`=? WHERE `id`=?", change["name"], change["recid"])
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}
				if change["password"] != nil {
					hash := sha256.New()
					hash.Write([]byte(fmt.Sprint(change["password"])))
					md := hash.Sum(nil)
					shaPassword := hex.EncodeToString(md)
					_, err = models.Db.Exec("UPDATE `auth_user` SET `password`=? WHERE `id`=?", shaPassword, change["recid"])
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}
				apilog.Println(change["recid"]," - ",change["name"], " - ",change["password"])
				js = []byte(`{"status": "success"}`)
			}*/
		}

	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}