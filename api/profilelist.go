package api

import (
	"encoding/json"
	"errors"
	"gonder/models"
	"log"
	"sort"
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
	plist := models.EmailPool.List()
	ps := make([]pList, 0, len(plist))
	keys := make([]int, 0, len(plist))
	for k := range plist {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		ps = append(ps, pList{ID:k, Name:plist[k]})
	}
	return ps, nil
}
