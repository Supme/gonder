package api

import (
	"encoding/json"
	"errors"
	"github.com/supme/gonder/models"
)

type sndrList struct {
	ID   int64  `json:"id"`
	Text string `json:"text"`
}

func senderList(req request) (js []byte, err error) {
	if user.Right("get-groups") && user.GroupRight(req.ID) {
		var id int64
		var email, name string
		var fs = []sndrList{}
		query, err := models.Db.Query("SELECT `id`, `name`, `email` FROM `sender` WHERE `group_id`=?", req.ID)
		if err != nil {
			return js, err
		}
		defer query.Close()
		for query.Next() {
			err = query.Scan(&id, &name, &email)
			if err != nil {
				return nil, err
			}
			fs = append(fs, sndrList{
				ID:   id,
				Text: name + " (" + email + ")",
			})
		}
		js, err = json.Marshal(fs)
		return js, err
	}
	return js, errors.New("Forbidden get from this group")
}
