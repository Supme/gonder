package models

import (
	"database/sql"
	"errors"
	"gonder/templates"
	"html/template"
	"log"
	"os"
	"path"
)

type Message struct {
	RecipientID    string // ToDo int?
	RecipientEmail string
	RecipientName  string
	RecipientParam map[string]string
	CampaignID     string // ToDo int?
}

func (m *Message) New(recipientID string) error {
	m.RecipientID = recipientID
	err := Db.QueryRow("SELECT `campaign_id`,`email`,`name` FROM `recipient` WHERE `id`=?", m.RecipientID).Scan(&m.CampaignID, &m.RecipientEmail, &m.RecipientName)
	if err != nil && err != sql.ErrNoRows {
		log.Print(err)
		return errors.New("find recipient with error")
	}
	if err == sql.ErrNoRows {
		return errors.New("the recipient does not exist")
	}
	return nil
}

// Unsubscribe recipient from group
// ToDo move to campaign/Recipient
func (m *Message) Unsubscribe(extra map[string]string) error {
	r, err := Db.Exec("INSERT INTO unsubscribe (`group_id`, `campaign_id`, `email`) VALUE ((SELECT group_id FROM campaign WHERE id=?), ?, ?)", m.CampaignID, m.CampaignID, m.RecipientEmail)
	if err != nil {
		log.Print(err)
	}
	id, e := r.LastInsertId()
	if e != nil {
		log.Print(e)
	}
	for name, value := range extra {
		_, err = Db.Exec("INSERT INTO unsubscribe_extra (`unsubscribe_id`, `name`, `value`) VALUE (?, ?, ?)", id, name, value)
		if err != nil {
			log.Print(err)
		}
	}
	return err
}

func (m *Message) Form(data map[string]string) error {
	r, err := Db.Exec("INSERT INTO question (`recipient_id`) VALUE (?)", m.RecipientID)
	if err != nil {
		log.Print(err)
	}
	id, e := r.LastInsertId()
	if e != nil {
		log.Print(e)
	}
	for name, value := range data {
		_, err = Db.Exec("INSERT INTO question_data (`question_id`, `name`, `value`) VALUE (?, ?, ?)", id, name, value)
		if err != nil {
			log.Print(err)
		}
	}
	return err
}

// GetTemplate return template from templates dir and if file not exist then find in bindata
func (m *Message) GetTemplate(fileName string) (*template.Template, error) {
	var dirName string
	err := Db.QueryRow("SELECT `group`.`template` FROM `campaign` INNER JOIN `group` ON `campaign`.`group_id`=`group`.`id` WHERE `group`.`template` IS NOT NULL AND `campaign`.`id`=?", m.CampaignID).Scan(&dirName)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if dirName == "" {
		dirName = "default"
	}
	filePath := path.Join(Config.UTMTemplatesDir, dirName, fileName)

	_, err = os.Stat(filePath)
	if err != nil {
		data, err := templates.Default.ReadFile(path.Join(dirName, fileName))
		if err != nil {
			return nil, err
		}
		return template.New(filePath).Parse(string(data))
	}
	return template.ParseFiles(filePath)
}
