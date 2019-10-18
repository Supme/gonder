package models

import "strconv"

type Recipient int

func (id Recipient) IntID() int {
	return int(id)
}

func (id Recipient) StringID() string {
	return strconv.Itoa(id.IntID())
}

type RecipientStatus int

const (
	RecipientStatusActive RecipientStatus = iota
	RecipientStatusDeleted
	RecipientStatusDuplicated
)

func (id RecipientStatus) IntID() int {
	return int(id)
}

func (id RecipientStatus) StringID() string {
	return strconv.Itoa(id.IntID())
}

func (id RecipientStatus) String() string {
	return [...]string{"active", "deleted", "duplicated"}[id]
}
