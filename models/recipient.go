package models

import (
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

type RecipientData struct {
	Name   string           `json:"name"`
	Email  string           `json:"email"`
	Params []RecipientParam `json:"params"`
}

type RecipientParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
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
