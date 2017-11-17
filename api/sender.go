package api

import (
	"encoding/json"
	"errors"
	"github.com/supme/gonder/models"
	"strconv"
)

type sndr struct {
	ID           int64  `json:"recid"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	DkimSelector string `json:"dkimSelector"`
	DkimKey      string `json:"dkimKey"`
	DkimUse      bool   `json:"dkimUse"`
}
type sndrs struct {
	Total   int64  `json:"total"`
	Records []sndr `json:"records"`
}

func sender(req request) (js []byte, err error) {

	switch req.Cmd {

	case "get":
		if user.Right("get-groups") && user.GroupRight(req.ID) {
			var f sndr
			var fs sndrs
			fs.Records = []sndr{}
			query, err := models.Db.Query("SELECT `id`, `email`, `name`, `dkim_selector`, `dkim_key`, `dkim_use` FROM `sender` WHERE `group_id`=? LIMIT ? OFFSET ?", req.ID, req.Limit, req.Offset)
			if err != nil {
				return js, err
			}
			defer query.Close()

			for query.Next() {
				err = query.Scan(&f.ID, &f.Email, &f.Name, &f.DkimSelector, &f.DkimKey, &f.DkimUse)
				fs.Records = append(fs.Records, f)
			}
			err = models.Db.QueryRow("SELECT COUNT(*) FROM `sender` WHERE `group_id`=?", req.Group).Scan(&fs.Total)
			js, err = json.Marshal(fs)
			return js, err
		}
		return js, errors.New("Forbidden get groups")

	case "save":
		if user.Right("save-groups") && user.GroupRight(req.ID) {
			var group int64
			err = models.Db.QueryRow("SELECT `group_id` FROM `sender` WHERE `id`=?", req.ID).Scan(&group)
			if err != nil {
				return js, err
			}
			if user.GroupRight(group) {
				_, err = models.Db.Exec("UPDATE `sender` SET `email`=?, `name`=?, `dkim_selector`=?, `dkim_key`=?, `dkim_use`=? WHERE `id`=?", req.Email, req.Name, req.DkimSelector, req.DkimKey, req.DkimUse, req.ID)
				if err != nil {
					return js, err
				}
			} else {
				return js, errors.New("Forbidden right to this group")
			}
		}
		return js, errors.New("Forbidden save groups")

	case "add":
		if user.Right("save-groups") && user.GroupRight(req.ID) {
			res, err := models.Db.Exec("INSERT INTO `sender` (`group_id`, `email`, `name`, `dkim_selector`, `dkim_key`, `dkim_use`) VALUES (?, ?, ?, ?, ?, ?);", req.ID, req.Email, req.Name, req.DkimSelector, req.DkimKey, req.DkimUse)
			if err != nil {
				return js, err
			}
			recid, err := res.LastInsertId()
			if err != nil {
				return js, err
			}
			js = []byte(`{"status": "success", "message": "", "recid": ` + strconv.FormatInt(recid, 10) + `}`)
		}
		return js, errors.New("Forbidden save groups")

	}
	return js, errors.New("Command not found")
}
