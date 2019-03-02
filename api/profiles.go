package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/supme/gonder/models"
	"log"
	"strconv"
)

type prof struct {
	ID          int64  `json:"recid"`
	Name        string `json:"name"`
	Iface       string `json:"iface"`
	Host        string `json:"host"`
	Stream      int    `json:"stream"`
	ResendDelay int    `json:"resend_delay"`
	ResendCount int    `json:"resend_count"`
}

type profs struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Total   int64  `json:"total"`
	Records []prof `json:"records"`
}

func profiles(req request) (js []byte, err error) {
	var ps profs
	var p prof

	switch req.Cmd {

	case "get":
		if req.auth.Right("get-profiles") {
			ps, err = getProfiles(req)
			if err != nil {
				return js, err
			}
			js, err = json.Marshal(ps)
			if err != nil {
				log.Println(err)
			}
			return js, err
		}
		return js, errors.New("Forbidden get profiles")

	case "add":
		if req.auth.Right("add-profiles") {
			p, err = addProfile()
			if err != nil {
				return js, err
			}
			js, err = json.Marshal(p)
			if err != nil {
				log.Println(err)
			}
			return js, err
		}
		return js, errors.New("Forbidden get profiles")

	case "delete":
		if req.auth.Right("delete-profiles") {
			deleteProfiles(req.Selected)
			js, err = json.Marshal(ps)
			if err != nil {
				log.Println(err)
			}
			return js, err
		}
		return js, errors.New("Forbidden get profiles")

	case "save":
		if req.auth.Right("save-profiles") {
			err = saveProfiles(req.Changes)
			if err != nil {
				return js, err
			}
			js, err = json.Marshal(ps)
			if err != err {
				log.Println(err)
			}
			return js, err
		}
		return js, errors.New("Forbidden get profiles")

	}

	return js, errors.New("Command not found")
}

func saveProfiles(changes []map[string]interface{}) error {
	var err error
	var p prof
	for c := range changes {
		p.ID, err = strconv.ParseInt(fmt.Sprint(changes[c]["recid"]), 10, 64)
		if err != nil {
			log.Println(err)
			return err
		}
		err = models.Db.QueryRow("SELECT `name`,`iface`,`host`,`stream`,`resend_delay`,`resend_count` FROM `profile` WHERE `id`=?", p.ID).Scan(&p.Name, &p.Iface, &p.Host, &p.Stream, &p.ResendDelay, &p.ResendCount)
		if err != nil {
			log.Println(err)
			return err
		}
		for i := range changes[c] {
			switch i {
			case "name":
				p.Name = fmt.Sprint(changes[c][i])
			case "iface":
				p.Iface = fmt.Sprint(changes[c][i])
			case "host":
				p.Host = fmt.Sprint(changes[c][i])
			case "stream":
				p.Stream, _ = strconv.Atoi(fmt.Sprint(changes[c][i]))
			case "resend_delay":
				p.ResendDelay, _ = strconv.Atoi(fmt.Sprint(changes[c][i]))
			case "resend_count":
				p.ResendCount, _ = strconv.Atoi(fmt.Sprint(changes[c][i]))
			}
		}
		_, err = models.Db.Exec("UPDATE `profile` SET `name`=?, `iface`=?, `host`=?, `stream`=?, `resend_delay`=?, `resend_count`=? WHERE id=?", p.Name, p.Iface, p.Host, p.Stream, p.ResendDelay, p.ResendCount, p.ID)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

func deleteProfiles(selected []interface{}) {
	for _, s := range selected {
		_, err := models.Db.Exec("DELETE FROM `profile` WHERE `id`=?", fmt.Sprintf("%d", s))
		if err != nil {
			log.Println(err)
		}
	}
}

func addProfile() (prof, error) {
	var p prof
	row, err := models.Db.Exec("INSERT INTO `profile` (`name`) VALUES ('')")
	if err != nil {
		log.Println(err)
		return p, err
	}
	p.ID, err = row.LastInsertId()
	if err != nil {
		log.Println(err)
		return p, err
	}

	return p, nil
}

func getProfiles(req request) (profs, error) {
	var (
		p                  prof
		ps                 profs
		partWhere          string
		partParams, params []interface{}
		err                error
	)
	ps.Records = []prof{}

	partWhere, partParams, err = createSQLPart(req, " WHERE 1=1", params, map[string]string{
		"recid": "id", "name": "name", "iface": "iface", "host": "host", "stream": "stream", "resend_delay": "resend_delay", "resend_count": "resend_count",
	}, true)
	if err != nil {
		log.Println(err)
		return ps, err
	}
	query, err := models.Db.Query("SELECT `id`,`name`,`iface`,`host`,`stream`,`resend_delay`,`resend_count` FROM `profile`"+partWhere, partParams...)
	if err != nil {
		log.Println(err)
		return ps, err
	}
	defer query.Close()

	for query.Next() {
		err = query.Scan(&p.ID, &p.Name, &p.Iface, &p.Host, &p.Stream, &p.ResendDelay, &p.ResendCount)
		if err != nil {
			log.Println(err)
			return profs{}, err
		}
		ps.Records = append(ps.Records, p)
	}
	err = models.Db.QueryRow("SELECT COUNT(*) FROM `profile`"+partWhere, partParams...).Scan(&ps.Total)
	if err != nil {
		log.Println(err)
	}
	return ps, err
}
