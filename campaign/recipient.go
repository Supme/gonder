package campaign

import (
	"database/sql"
	"errors"
	"github.com/supme/gonder/models"
)

type Recipient struct {
	ID         string
	CampaignID string
	Email      string
	Name       string
	Params     map[string]interface{}
}

func GetRecipient(id string) (Recipient, error) {
	recipient := Recipient{ID: id}
	err := models.Db.
		QueryRow("SELECT `campaign_id`,`email`,`name` FROM `recipient` WHERE `id`=?", &recipient.ID).
		Scan(&recipient.CampaignID, &recipient.Email, &recipient.Name)
	if err == sql.ErrNoRows {
		return recipient, errors.New("recipient does not exist")
	} else if err != nil {
		return recipient, err
	}

	recipient.Params = map[string]interface{}{}

	paramQuery, err := models.Db.Query("SELECT `key`, `value` FROM `parameter` WHERE `recipient_id`=?", id)
	if err != nil {
		return recipient, err
	}

	for paramQuery.Next() {
		var k, v string
		err = paramQuery.Scan(&k, &v)
		recipient.Params[k] = v
	}

	return recipient, nil
}

func (recipient *Recipient) Unsubscribed() bool {
	var unsubscribeCount int
	models.Db.QueryRow("SELECT COUNT(*) FROM `unsubscribe` t1 INNER JOIN `campaign` t2 ON t1.group_id = t2.group_id WHERE t2.id = ? AND t1.email = ?", recipient.CampaignID, recipient.Email).Scan(&unsubscribeCount)
	if unsubscribeCount == 0 {
		return false
	}
	return true
}

func (recipient *Recipient) UnsubscribeEmailHeaderURL() string {
	return models.EncodeUTM("unsubscribe", "mail", map[string]interface{}{"RecipientId": recipient.ID, "RecipientEmail": recipient.Email})
}
