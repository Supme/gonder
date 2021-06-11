package models

import (
	"github.com/jmoiron/sqlx"
	"log"
	"strconv"
)

type Recipient int

func (id Recipient) IntID() int {
	return int(id)
}

func (id Recipient) StringID() string {
	return strconv.Itoa(id.IntID())
}

func RecipientGetByID(id int) Recipient {
	return Recipient(id)
}

func RecipientGetByStringID(id string) Recipient {
	intID, err := strconv.Atoi(id)
	if err != nil {
		log.Print(err)
	}
	return Recipient(intID)
}

func (id Recipient) UpdateRecipientStatus(status string) error {
	_, err := Db.Exec("UPDATE recipient SET status=?, date=NOW() WHERE id=?", status, id)
	return err
}

func (id Recipient) Delete() error {
	return RecipientsDelete(int(id))
}

func RecipientsDelete(id ...int) error {
	if len(id) == 0 {
		return nil
	}
	query, args, err := sqlx.In("UPDATE recipient SET removed=1 WHERE id IN (?);", id)
	if err != nil {
		return err
	}
	query = Db.Rebind(query)
	_, err = Db.Query(query, args...)
	return err
}

func RecipientsGroups(id ...int) ([]Group, error) {
	if len(id) == 0 {
		return []Group{}, nil
	}
	query, args, err := sqlx.In("SELECT DISTINCT campaign_id FROM recipient WHERE id IN (?);", id)
	if err != nil {
		return []Group{}, err
	}
	query = Db.Rebind(query)
	rows, err := Db.Query(query, args...)
	if err != nil {
		return []Group{}, err
	}
	var groups []Group
	for rows.Next() {
		var i int
		err = rows.Scan(&i)
		if err != nil {
			return groups, err
		}
		groups = append(groups, Group(i))
	}
	return groups, nil
}

type RecipientData struct {
	Name   string            `json:"name"`
	Email  string            `json:"email"`
	Params map[string]string `json:"params"`
}

type RecipientRemovedStatus int

const (
	RecipientRemovedStatusActive RecipientRemovedStatus = iota
	RecipientRemovedStatusDeleted
	RecipientRemovedStatusDuplicated
)

func (id RecipientRemovedStatus) IntID() int {
	return int(id)
}

func (id RecipientRemovedStatus) StringID() string {
	return strconv.Itoa(id.IntID())
}

func (id RecipientRemovedStatus) String() string {
	return [...]string{"active", "deleted", "duplicated"}[id]
}
