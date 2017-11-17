package api

import (
	"encoding/json"
	"errors"
	"github.com/supme/gonder/models"
)

type pList struct {
	ID   int    `json:"id"`
	Name string `json:"text"`
}

func profilesList(req request) (js []byte, err error) {
	if user.Right("get-campaign") {
		psl, err := getProfilesList()
		if err != nil {
			return js, err
		}
		js, err = json.Marshal(psl)
		return js, err
	}
	return js, errors.New("Forbidden get campaign")

}

func getProfilesList() ([]pList, error) {
	var p pList
	var ps []pList
	ps = []pList{}
	query, err := models.Db.Query("SELECT `id`, `name` FROM `profile`")
	if err != nil {
		return ps, err
	}
	defer query.Close()
	for query.Next() {
		err = query.Scan(&p.ID, &p.Name)
		ps = append(ps, p)
	}
	return ps, nil
}
