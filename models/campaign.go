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

type CampaignReportRecipients struct {
	ID          int             `db:"id" json:"id"`
	Email       string          `db:"email" json:"email"`
	Name        string          `db:"name" json:"name"`
	At          string          `db:"at" json:"at"`
	Status      sql.NullString  `db:"status" json:"-"`
	StatusValid string          `json:"status"`
	Open        bool            `db:"open" json:"open"`
	Data        sql.NullString  `db:"data" json:"-"`
	DataValid   json.RawMessage `json:"data"`
}

func (cq *CampaignReportRecipients) Validate() {
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

func (id Campaign) ReportRecipients() (*sqlx.Rows, error) {
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

type CampaignReportUnsubscribed struct {
	Email     string          `db:"email" json:"email"`
	At        string          `db:"at" json:"at"`
	Data      sql.NullString  `db:"data" json:"-"`
	DataValid json.RawMessage `json:"data"`
}

func (cq *CampaignReportUnsubscribed) Validate() {
	if cq.Data.Valid {
		cq.DataValid = []byte(cq.Data.String)
	} else {
		cq.DataValid = []byte("null")
	}
}

func (id Campaign) ReportUnsubscribed() (*sqlx.Rows, error) {
	return Db.Queryx(`
		SELECT
        	u.email,
			u.date as "at",
            `+SQLKeyValueTableToJSON("d.name", "d.value", "unsubscribe_extra d", "d.unsubscribe_id=u.id")+` AS "data"
		FROM unsubscribe u
		WHERE u.campaign_id=?`, id)
}

type CampaignReportQuestion struct {
	ID        int             `db:"id" json:"id"`
	Email     string          `db:"email" json:"email"`
	At        string          `db:"at" json:"at"`
	Data      sql.NullString  `db:"data" json:"-"`
	DataValid json.RawMessage `json:"data"`
}

func (cq *CampaignReportQuestion) Validate() {
	if cq.Data.Valid {
		cq.DataValid = []byte(cq.Data.String)
	} else {
		cq.DataValid = []byte("null")
	}
}

func (id Campaign) ReportQuestion() (*sqlx.Rows, error) {
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

type CampaignReportClicks struct {
	ID    int    `db:"id" json:"id"`
	Email string `db:"email" json:"email"`
	At    string `db:"at" json:"at"`
	URL   string `db:"url" json:"url"`
}

func (cq *CampaignReportClicks) Validate() {}

func (id Campaign) ReportClicks() (*sqlx.Rows, error) {
	return Db.Queryx(`
		SELECT
			r.id,
			r.email,
			j.date as at,
			j.url
		FROM jumping j INNER JOIN recipient r ON j.recipient_id=r.id
		WHERE r.removed=0
			AND j.url NOT IN ('`+OpenTrace+`','`+WebVersion+`','`+Unsubscribe+`')
			AND j.campaign_id=?
		ORDER BY r.id, id`, id)
}
