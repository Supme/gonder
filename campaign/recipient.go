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
	utmURL     string
}

func GetRecipient(id string) (Recipient, error) {
	recipient := Recipient{ID: id}
	err := models.Db.
		QueryRow("SELECT t3.`utm_url`, t1.`campaign_id`, t1.`email`, t1.`name` FROM `recipient` t1 INNER JOIN `campaign` t2 ON t1.`campaign_id`=t2.`id` INNER JOIN `sender` t3 ON t3.`id`=t2.`sender_id` WHERE t1.`id`=?", &recipient.ID).
		Scan(&recipient.utmURL, &recipient.CampaignID, &recipient.Email, &recipient.Name)
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
	recipient.Params["WebUrl"] = models.EncodeUTM("web", recipient.utmURL, "", recipient.Params)
	recipient.Params["StatPng"] = models.EncodeUTM("open", recipient.utmURL, "", recipient.Params)
	recipient.Params["QuestionUrl"] = models.EncodeUTM("question", recipient.utmURL, "", recipient.Params)
	recipient.Params["UnsubscribeUrl"] = models.EncodeUTM("unsubscribe", recipient.utmURL, "web", recipient.Params)

	return recipient, nil
}

func (r *Recipient) unsubscribed() bool {
	var unsubscribeCount int
	err := models.Db.QueryRow("SELECT COUNT(*) FROM `unsubscribe` t1 INNER JOIN `campaign` t2 ON t1.group_id = t2.group_id WHERE t2.id = ? AND t1.email = ?", r.CampaignID, r.Email).Scan(&unsubscribeCount)
	checkErr(err)
	return unsubscribeCount != 0
}

func (r Recipient) WebHTML(web bool, preview bool) func(io.Writer) error {
	camp, err := getCampaign(r.CampaignID)
	checkErr(err)
	return camp.htmlTemplFunc(r, web, preview)
}
