package campaign

import (
	"database/sql"
	"errors"
	"gonder/models"
	"io"
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
		if err != nil {
			return Recipient{}, err
		}
		recipient.Params[k] = v
	}

	recipient.Params["RecipientId"] = recipient.ID
	recipient.Params["CampaignId"] = recipient.CampaignID
	recipient.Params["RecipientEmail"] = recipient.Email
	recipient.Params["RecipientName"] = recipient.Name
	recipient.Params["WebUrl"] = models.EncodeUTM("web", "", recipient.Params)
	recipient.Params["StatPng"] = models.EncodeUTM("open", "", recipient.Params)
	recipient.Params["QuestionUrl"] = models.EncodeUTM("question", "", recipient.Params)
	recipient.Params["UnsubscribeUrl"] = models.EncodeUTM("unsubscribe", "web", recipient.Params)

	return recipient, nil
}

func (r *Recipient) unsubscribed() bool {
	var unsubscribeCount int
	err := models.Db.QueryRow("SELECT COUNT(*) FROM `unsubscribe` t1 INNER JOIN `campaign` t2 ON t1.group_id = t2.group_id WHERE t2.id = ? AND t1.email = ?", r.CampaignID, r.Email).Scan(&unsubscribeCount)
	checkErr(err)
	if unsubscribeCount == 0 {
		return false
	}
	return true
}

func (r *Recipient) unsubscribeEmailHeaderURL() string {
	return models.EncodeUTM("unsubscribe", "mail", map[string]interface{}{"RecipientId": r.ID, "RecipientEmail": r.Email})
}

func (r Recipient) WebHTML(web bool, preview bool) func(io.Writer) error {
	camp, err := getCampaign(r.CampaignID)
	checkErr(err)
	return camp.htmlTemplFunc(r, web, preview)
}
