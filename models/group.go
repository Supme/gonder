package models

import (
	"database/sql"
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"strconv"
)

type Group int

func (id Group) IntID() int {
	return int(id)
}

func (id Group) StringID() string {
	return strconv.Itoa(id.IntID())
}

type GroupCampaign struct {
	ID      int    `db:"id" json:"id"`
	Name    string `db:"name" json:"name"`
	Subject string `db:"subject" json:"subject"`
	Start   string `db:"start" json:"start"`
	End     string `db:"end" json:"end"`
}

func (id Group) Campaigns() (*sqlx.Rows, error) {
	return Db.Queryx("SELECT `id`, `name`, `subject`, `start_time` AS 'start', `end_time` AS 'end' FROM `campaign` WHERE `group_id`=?", id)
}

type GroupUnsubscribed struct {
	CampaignID int             `db:"campaign_id" json:"campaign_id"`
	Email      string          `db:"email" json:"email"`
	At         string          `db:"at" json:"at"`
	Data       sql.NullString  `db:"data" json:"-"`
	DataValid  json.RawMessage `json:"data"`
}

func (cq *GroupUnsubscribed) Validate() {
	if cq.Data.Valid {
		cq.DataValid = []byte(cq.Data.String)
	} else {
		cq.DataValid = []byte("null")
	}
}

func (id Group) Unsubscribed() (*sqlx.Rows, error) {
	return Db.Queryx(`
		SELECT
        	u.email,
        	u.campaign_id,
			u.date as "at",
			`+SQLKeyValueTableToJSON("d.name", "d.value", "unsubscribe_extra d", "d.unsubscribe_id=u.id")+` AS "data"
		FROM unsubscribe u
		WHERE u.group_id=?`, id)

}
