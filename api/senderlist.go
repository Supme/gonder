package api

import (
	"encoding/json"
	"errors"
	"github.com/supme/gonder/models"
	"log"
)

type sndrList struct {
	ID   int64  `json:"id"`
	Text string `json:"text"`
}

func senderList(req request) (js []byte, err error) {
	if req.auth.Right("get-groups") && req.auth.GroupRight(req.ID) {
		var id int64
		var email, name string
		var fs = []sndrList{}
		query, err := models.Db.Query("SELECT `id`, `name`, `email` FROM `sender` WHERE `group_id`=?", req.ID)
		if err != nil {
			log.Println(err)
			return js, err
		}
		defer query.Close()
		for query.Next() {
			err = query.Scan(&id, &name, &email)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			fs = append(fs, sndrList{
				ID:   id,
				Text: name + " (" + email + ")",
			})
		}
		js, err = json.Marshal(fs)
		if err != nil {
			log.Println(err)
		}
		return js, err
	}
	return js, errors.New("Forbidden get from this group")
}
