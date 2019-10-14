package models

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"strconv"
)

type Campaign int

func (id Campaign) IntID() int {
	return int(id)
}

func (id Campaign) StringID() string {
	return strconv.Itoa(id.IntID())
}

type CampaignUnsubscribed struct {
	Email     string         `db:"email" json:"email"`
	At        string         `db:"at" json:"at"`
	Data      sql.NullString `db:"data" json:"-"`
	DataValid string         `json:"data"`
}

func (cq *CampaignUnsubscribed) Validate() {
	if cq.Data.Valid {
		cq.DataValid = cq.Data.String
	} else {
		cq.DataValid = "NULL"
	}
}

func (id Campaign) Unsubscribed() (*sqlx.Rows, error) {
	return Db.Queryx("SELECT  u.`email` AS 'email', u.`date` AS 'at', (SELECT GROUP_CONCAT(e.`name`, ':`', e.`value`, '`' ORDER BY e.`name` SEPARATOR ' | ') FROM `unsubscribe_extra` e WHERE e.`unsubscribe_id`=u.`id`) AS 'data' FROM `unsubscribe` u WHERE u.`campaign_id`=?", id)
}

type CampaignQuestion struct {
	ID        int            `db:"id" json:"id"`
	Email     string         `db:"email" json:"email"`
	At        string         `db:"at" json:"at"`
	Data      sql.NullString `db:"data" json:"-"`
	DataValid string         `json:"data"`
}

func (cq *CampaignQuestion) Validate() {
	if cq.Data.Valid {
		cq.DataValid = cq.Data.String
	} else {
		cq.DataValid = "NULL"
	}
}

func (id Campaign) Question() (*sqlx.Rows, error) {
	return Db.Queryx("SELECT q.`recipient_id` AS 'id', r.`email` AS 'email', q.`at` AS 'at', (SELECT GROUP_CONCAT(d.`name`, ': `', d.`value`,'`' ORDER BY d.`name` SEPARATOR ' | ') FROM `question_data` d WHERE d.`question_id`=q.`id`) AS 'data' FROM `question` q LEFT JOIN `recipient` r ON q.`recipient_id`=r.`id` WHERE r.`campaign_id`=?", id)
}
