package api

import (
	"encoding/json"
	"errors"
	"gonder/models"
	"log"
)

type pList struct {
	ID   int    `json:"id"`
	Name string `json:"text"`
}

func profilesList(req request) (js []byte, err error) {
	if req.auth.Right("get-campaign") {
		psl, err := getProfilesList()
		if err != nil {
			return js, err
		}
		js, err = json.Marshal(psl)
		if err != nil {
			log.Println(err)
		}
		return js, err
	}
	return js, errors.New("Forbidden get campaign")

}

func getProfilesList() ([]pList, error) {
	var (
		p  pList
		ps []pList
	)
	ps = []pList{}
	query, err := models.Db.Query("SELECT `id`, `name` FROM `profile`")
	if err != nil {
		log.Println(err)
		return ps, err
	}
	defer func() {
		if err := query.Close(); err != nil {
			log.Print(err)
		}
	}()
	for query.Next() {
		err = query.Scan(&p.ID, &p.Name)
		if err != nil {
			log.Println(err)
			return ps, err
		}
		ps = append(ps, p)
	}
	return ps, nil
}
