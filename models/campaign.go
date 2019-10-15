package models

import (
	"database/sql"
	"encoding/json"
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

type CampaignRecipients struct {
	ID        int             `db:"id" json:"id"`
	Email     string          `db:"email" json:"email"`
	Name      string          `db:"name" json:"name"`
	At        string          `db:"at" json:"at"`
	Status    sql.NullString  `db:"status" json:"-"`
	StatusValid    string  `json:"status"`
	Open      bool            `db:"open" json:"open"`
	Data      sql.NullString  `db:"data" json:"-"`
	DataValid json.RawMessage `json:"data"`
}

func (cq *CampaignRecipients) Validate() {
	if cq.Data.Valid {
		cq.DataValid = []byte(cq.Data.String)
	} else {
		cq.DataValid = []byte("null")
	}
	if cq.Status.Valid {
		cq.StatusValid = cq.Status.String
	} else {
		cq.StatusValid = "null"
	}
}

func (id Campaign) Recipients() (*sqlx.Rows, error) {
	return Db.Queryx(`
		SELECT  
  			r.id,
  			r.email AS "email",
  			r.name,
			r.date AS "at",
  			r.status,
  			IF(COALESCE(r.web_agent,r.client_agent) IS NULL, false, true) as "open",
    		`+SQLKeyValueTableToJSON("d.key", "d.value", "parameter d", "d.recipient_id=r.id")+` AS "data"
 		FROM recipient r
 		WHERE r.removed=0 AND r.campaign_id=?`, id)
}

type CampaignUnsubscribed struct {
	Email     string          `db:"email" json:"email"`
	At        string          `db:"at" json:"at"`
	Data      sql.NullString  `db:"data" json:"-"`
	DataValid json.RawMessage `json:"data"`
}

func (cq *CampaignUnsubscribed) Validate() {
	if cq.Data.Valid {
		cq.DataValid = []byte(cq.Data.String)
	} else {
		cq.DataValid = []byte("null")
	}
}

func (id Campaign) Unsubscribed() (*sqlx.Rows, error) {
	return Db.Queryx(`
		SELECT
        	u.email,
			u.date as "at",
            `+SQLKeyValueTableToJSON("d.name", "d.value", "unsubscribe_extra d", "d.unsubscribe_id=u.id")+` AS "data"
		FROM unsubscribe u
		WHERE u.campaign_id=?`, id)
}

type CampaignQuestion struct {
	ID        int             `db:"id" json:"id"`
	Email     string          `db:"email" json:"email"`
	At        string          `db:"at" json:"at"`
	Data      sql.NullString  `db:"data" json:"-"`
	DataValid json.RawMessage `json:"data"`
}

func (cq *CampaignQuestion) Validate() {
	if cq.Data.Valid {
		cq.DataValid = []byte(cq.Data.String)
	} else {
		cq.DataValid = []byte("null")
	}
}

func (id Campaign) Question() (*sqlx.Rows, error) {
	return Db.Queryx(`
		SELECT
			q.recipient_id AS id,
			r.email,
			q.at,
			`+SQLKeyValueTableToJSON("d.name", "d.value", "question_data d", "d.question_id=q.id")+` AS "data"
		FROM question q
		LEFT JOIN recipient r ON q.recipient_id=r.id
		WHERE r.campaign_id=?`, id)
}
