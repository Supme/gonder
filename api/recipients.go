// Project Gonder.
// Author Supme
// Copyright Supme 2016
// License http://opensource.org/licenses/MIT MIT License
//
//  THE SOFTWARE AND DOCUMENTATION ARE PROVIDED "AS IS" WITHOUT WARRANTY OF
//  ANY KIND, EITHER EXPRESSED OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
//  IMPLIED WARRANTIES OF MERCHANTABILITY AND/OR FITNESS FOR A PARTICULAR
//  PURPOSE.
//
// Please see the License.txt file for more information.
//
package api

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"github.com/supme/gonder/models"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"os"
	"path"
	"time"
	"fmt"
	"strconv"
	"errors"
	"sync"
	"strings"
)

type RecipientTableLine struct {
	Id     int64  `json:"recid"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Result string `json:"result"`
	Open   bool `json:"open"`
}

type RecipientsTable struct {
	Total   int                  `json:"total"`
	Records []RecipientTableLine `json:"records"`
}

type Recipients []Recipient

type Recipient struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Params []RecipientParam `json:"params"`
}

type RecipientParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type RecipientTableParams struct {
	Total   int              `json:"total"`
	Records []RecipientParam `json:"records"`
}

type safeProgress struct {
	cnt map[string]int
	sync.RWMutex
}

var progress = safeProgress{cnt: map[string]int{}}

func recipients(req request) (js []byte, err error) {

	if req.Recipient == 0 {
		switch req.Cmd {
		case "get":
			if auth.Right("get-recipients") && auth.CampaignRight(req.Campaign) {
				rs, err := getRecipients(req)
				if err != nil {
					return js, err
				}
				js, err = json.Marshal(rs)
				return js, err
			} else {
				return js, errors.New("Forbidden get recipients")
			}
		case "add":
			if auth.Right("upload-recipients") && auth.CampaignRight(req.Campaign) {
				err := addRecipients(req.Campaign, req.Recipients)
				if err != nil {
					return js, err
				}
			}  else {
				return js, errors.New("Forbidden add recipients")
			}

		case "upload":
			if auth.Right("upload-recipients") && auth.CampaignRight(req.Campaign) {
				content, err := base64.StdEncoding.DecodeString(req.FileContent)
				if err != nil {
					return js, err
				}
				filename := strconv.FormatInt(time.Now().UnixNano(), 16)
				file := models.FromRootDir("tmp/" + filename)
				err = ioutil.WriteFile(file, content, 0644)
				if err != nil {
					return js, err
				}
				apilog.Print(auth.Name," upload file ", req.FileName)

				if path.Ext(req.FileName) == ".csv" {
					go func() {
						progress.Lock()
						progress.cnt[filename] = 0
						progress.Unlock()
						err = recipientCsv(req.Campaign, filename)
						if err != nil {
							apilog.Println(err)
						}
						progress.Lock()
						delete(progress.cnt, filename)
						progress.Unlock()
					}()
					js = []byte(fmt.Sprintf(`{"status": "success", "message": "%s"}`, filename))

				} else if path.Ext(req.FileName) == ".xlsx" {
					go func() {
						progress.Lock()
						progress.cnt[filename] = 0
						progress.Unlock()
						err = recipientXlsx(req.Campaign, filename)
						if err != nil {
							apilog.Println(err)
						}
						progress.Lock()
						delete(progress.cnt, filename)
						progress.Unlock()
					}()
					js = []byte(fmt.Sprintf(`{"status": "success", "message": "%s"}`, filename))

				} else {
					return js, errors.New("This not csv or xlsx file")
				}
			} else {
				return js, errors.New("Forbidden upload recipients")
			}

		case "progress":
			if auth.Right("upload-recipients") {
				progress.RLock()
				if val, ok := progress.cnt[req.Name]; ok {
					js = []byte(fmt.Sprintf(`{"status": "success", "message": %d}`, val))
				} else {
					js = []byte(`{"status": "error", "message": "not found"}`)
				}
				progress.RUnlock()
			}

		case "clear":
			if auth.Right("delete-recipients") && auth.CampaignRight(req.Campaign) {
				err = delRecipients(req.Campaign)
				if err != nil {
					return js, errors.New("Can't delete all recipients")
				}
			} else {
				return js, errors.New("Forbidden delete recipients")
			}

		case "resend4xx":
			if auth.Right("accept-campaign") && auth.CampaignRight(req.Campaign) {
				err = resendCampaign(req.Campaign)
				if err != nil {
					return js, errors.New("Can't resend")
				}
			} else {
				return js, errors.New("Forbidden resend campaign")
			}

		case "deduplicate":
			if auth.Right("delete-recipients") && auth.CampaignRight(req.Campaign) {
				cnt, err := deduplicateRecipient(req.Campaign)
				if err != nil {
					return js, errors.New("Can't deduplicate recipients")
				}
				js = []byte(fmt.Sprintf(`{"status": "success", "message": %d}`, cnt))
			} else {
				return js, errors.New("Forbidden delete recipients")
			}

		default:
			err = errors.New("Command not found")
		}
	} else {
		if req.Cmd == "get" {
			rId, err := getRecipientCampaign(req.Recipient)
			if err != nil {
				return js, err
			}
			if auth.Right("get-recipient-parameters") && auth.CampaignRight(rId) {
				ps, err := getRecipientParams(req.Recipient)
				js, err = json.Marshal(ps)
				if err != nil {
					return js, err
				}
			} else {
				return js, errors.New("Forbidden get recipient parameters")
			}
		}
	}
	return js, err
}

func deduplicateRecipient(campaignId int64) (cnt int64, err error) {
	q, err := models.Db.Query(`
	SELECT r1.id FROM recipient as r1
			JOIN (
				SELECT MIN(id) AS id, email FROM recipient WHERE
             	campaign_id=? AND removed=0
             	GROUP BY email HAVING COUNT(*)>1) as r2 ON (r1.email=r2.email AND r1.id!=r2.id
			)
     	WHERE r1.campaign_id=? AND removed=0
	`, campaignId, campaignId)
	if err != nil {
		return
	}

	tx, err := models.Db.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	dupl, err := tx.Prepare("UPDATE `recipient` SET `removed`=2 WHERE id=?")
	if err != nil {
		return
	}
	defer dupl.Close()

	cnt = 0
	for q.Next() {
		var id int64
		q.Scan(&id)
		_, err = dupl.Exec(id)
		if err != nil {
			return
		}
		cnt = cnt + 1
	}

	err = tx.Commit()
	return
}

func addRecipients(campaignId int64, recipients Recipients) error {
	tx, err := models.Db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stRecipient, err := tx.Prepare("INSERT INTO recipient (`campaign_id`, `email`, `name`) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stRecipient.Close()
	stParameter, err := tx.Prepare("INSERT INTO parameter (`recipient_id`, `key`, `value`) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stParameter.Close()

	for r := range recipients {
		res, err := stRecipient.Exec(campaignId, strings.TrimSpace(recipients[r].Email), recipients[r].Name)
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		for _, p := range recipients[r].Params {
			_, err := stParameter.Exec(id, p.Key, p.Value)
			if err != nil {
				return err
			}
		}
	}

	err = tx.Commit()
	return err
}

func resendCampaign(campaignId int64) error {
	res, err := models.Db.Exec("UPDATE `recipient` SET `status`=NULL WHERE `campaign_id`=? AND `removed`=0 AND LOWER(`status`) REGEXP '^((4[0-9]{2})|(dial tcp)|(read tcp)|(proxy)|(eof)).+'", campaignId)
	c, _ := res.RowsAffected()
	apilog.Printf("User %s resend by 4xx code for campaign %d. Resend count %d", auth.Name, campaignId, c)
	return err
}

func getRecipientCampaign(recipientId int64) (int64, error) {
	var id int64
	err := models.Db.QueryRow("SELECT `campaign_id` FROM `recipient` WHERE `id`=?", recipientId).Scan(&id)
	return id, err
}

//ToDo check right errors
//ToDo add unsubscribe column
func getRecipients(req request) (RecipientsTable, error) {
	var (
		err                error
		rs                 RecipientsTable
		partWhere, where   string
		partParams, params []interface{}
	)

	params = append(params, req.Campaign)

	rs.Records = []RecipientTableLine{}
	where = " WHERE `removed`=0 AND `campaign_id`=?"
	partWhere, partParams, err = createSqlPart(req, where, params, map[string]string{
		"recid":"id", "name":"name", "email":"email", "result":"status","open":"open",
	}, true)
	if err != nil {
		return rs, err
	}

	query, err := models.Db.Query("SELECT `id`, `name`, `email`, `status`, IF(COALESCE(`web_agent`,`client_agent`) IS NULL, 0, 1) FROM `recipient`" + partWhere, partParams...)
	if err != nil {
		return rs, err
	}
	defer query.Close()
	for query.Next() {
		r := RecipientTableLine{}
		err = query.Scan(&r.Id, &r.Name, &r.Email, &r.Result, &r.Open)
		rs.Records = append(rs.Records, r)
	}
	partWhere, partParams, err = createSqlPart(req, where, params, map[string]string{
		"recid":"id", "name":"name", "email":"email", "result":"status","open":"open",
	}, false)
	if err != nil {
		return rs, err
	}
	err = models.Db.QueryRow("SELECT COUNT(*) FROM `recipient`" + partWhere, partParams...).Scan(&rs.Total)
	return rs, nil

}

//ToDo check right errors
func getRecipientParams(recipient int64) (RecipientTableParams, error) {
	var err error
	var p RecipientParam
	var ps RecipientTableParams
	ps.Records = []RecipientParam{}
	query, err := models.Db.Query("SELECT `key`, `value` FROM `parameter` WHERE `recipient_id`=?", recipient)
	if err != nil {
		return ps, err
	}
	defer query.Close()
	for query.Next() {
		err = query.Scan(&p.Key, &p.Value)
		ps.Records = append(ps.Records, p)
	}
	err = models.Db.QueryRow("SELECT COUNT(*) FROM `parameter` WHERE `recipient_id`=?", recipient).Scan(&ps.Total)
	return ps, err
}

func delRecipients(campaignId int64) error {
	_, err := models.Db.Exec("UPDATE `recipient` SET `removed`=1 WHERE `campaign_id`=? AND `removed`=0", campaignId)
	return err
}

func recipientCsv(campaignId int64, file string) error {

	title := make(map[int]string)
	data := make(map[string]string)
	var email, name string
	f := models.FromRootDir("tmp/" + file)
	defer os.Remove(f)

	csvfile, err := os.Open(f)
	if err != nil {
		return err
	}
	defer csvfile.Close()

	reader := csv.NewReader(csvfile)
	reader.FieldsPerRecord = -1
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return err
	}

	tx, err := models.Db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stRecipient, err := tx.Prepare("INSERT INTO recipient (`campaign_id`, `email`, `name`) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stRecipient.Close()
	stParameter, err := tx.Prepare("INSERT INTO parameter (`recipient_id`, `key`, `value`) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stParameter.Close()

	total := len(rawCSVdata)
	for k, v := range rawCSVdata {
		if k == 0 {
			for i, t := range v {
				title[i] = t
			}
		} else {
			email = ""
			name = ""
			data = map[string]string{}
			for i, t := range v {
				if i == 0 {
					email = strings.TrimSpace(t)
				} else if i == 1 {
					name = t
				} else {
					data[title[i]] = t
				}
			}

			res, err := stRecipient.Exec(campaignId, email, name)
			if err != nil {
				return err
			}
			id, err := res.LastInsertId()
			if err != nil {
				return err
			}
			for i, t := range data {
				_, err := stParameter.Exec(id, i, t)
				if err != nil {
					return err
				}
			}
		}
		progress.Lock()
		progress.cnt[file] = int(k) * 100 / total
		progress.Unlock()
	}

	err = tx.Commit()

	return err
}

func recipientXlsx(campaignId int64, file string) error {

	title := make(map[int]string)
	data := make(map[string]string)
	var email, name string

	f := models.FromRootDir("tmp/" + file)
	xlFile, err := xlsx.OpenFile(f)
	if err != nil {
		return err
	}
	defer os.Remove(f)

	if xlFile.Sheets[0] != nil {

		tx, err := models.Db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		stRecipient, err := tx.Prepare("INSERT INTO recipient (`campaign_id`, `email`, `name`) VALUES (?, ?, ?)")
		if err != nil {
			return err
		}
		defer stRecipient.Close()
		stParameter, err := tx.Prepare("INSERT INTO parameter (`recipient_id`, `key`, `value`) VALUES (?, ?, ?)")
		if err != nil {
			return err
		}
		defer stParameter.Close()

		total := len(xlFile.Sheets[0].Rows)
		for k, v := range xlFile.Sheets[0].Rows {
			if k == 0 {
				for i, cell := range v.Cells {
					t := cell.String()
					title[i] = t
				}
			} else {
				email = ""
				name = ""
				data = map[string]string{}
				for i, cell := range v.Cells {
					t := cell.String()
					if i == 0 {
						email = strings.TrimSpace(t)
					} else if i == 1 {
						name = t
					} else {
						data[title[i]] = t
					}
				}

				res, err := stRecipient.Exec(campaignId, email, name)
				if err != nil {
					return err
				}
				id, err := res.LastInsertId()
				if err != nil {
					return err
				}
				for i, t := range data {
					_, err := stParameter.Exec(id, i, t)
					if err != nil {
						return err
					}
				}

			}
			progress.Lock()
			progress.cnt[file] = int(k) * 100 / total
			progress.Unlock()
		}
		err = tx.Commit()
	}

	return err
}
