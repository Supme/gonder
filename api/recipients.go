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
)

type Recipient struct {
	Id     int64  `json:"recid"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Result string `json:"result"`
}

type Recipients struct {
	Total   int         `json:"total"`
	Records []Recipient `json:"records"`
}

type RecipientParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type RecipientParams struct {
	Total   int              `json:"total"`
	Records []RecipientParam `json:"records"`
}

var progress = map[string]int{}

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

		case "upload":
			if auth.Right("upload-recipients") && auth.CampaignRight(req.Campaign) {
				content, err := base64.StdEncoding.DecodeString(req.FileContent)
				if err != nil {
					return js, err
				}
				filename := strconv.FormatInt(time.Now().UnixNano(), 10)
				file := models.FromRootDir("tmp/" + filename)
				err = ioutil.WriteFile(file, content, 0644)
				if err != nil {
					return js, err
				}
				apilog.Print(auth.Name," upload file ", req.FileName)

				if path.Ext(req.FileName) == ".csv" {
					go func() {
						progress[filename] = 0
						err = recipientCsv(req.Campaign, filename)
						if err != nil {
							apilog.Println(err)
						}
						delete(progress, filename)
					}()
					js = []byte(fmt.Sprintf(`{"status": "success", "message": "%s"}`, filename))

				} else if path.Ext(req.FileName) == ".xlsx" {
					go func() {
						progress[filename] = 0
						err = recipientXlsx(req.Campaign, filename)
						if err != nil {
							apilog.Println(err)
						}
						delete(progress, filename)
					}()
					js = []byte(fmt.Sprintf(`{"status": "success", "message": "%s"}`, filename))

				} else {
					return js, errors.New("This not csv or xlsx file")
				}
			} else {
				return js, errors.New("Forbidden upload recipients")
			}

		case "progress":
			if auth.Right("upload-recipients") && auth.CampaignRight(req.Campaign) {
				if val, ok := progress[req.Name]; ok {
					js = []byte(fmt.Sprintf(`{"status": "success", "message": %d}`, val))
				} else {
					js = []byte(`{"status": "error", "message": "not found"}`)
				}
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
func getRecipients(req request) (Recipients, error) {
	var (
		err error
		rs Recipients
		partWhere, where string
		partParams, params []interface{}
	)

	params = append(params, req.Campaign)

	rs.Records = []Recipient{}
	where = " WHERE `removed`!=1 AND `campaign_id`=?"
	partWhere, partParams, err = createSqlPart(req, where, params, map[string]string{
		"recid":"id", "name":"name", "email":"email", "result":"status",
	}, true)
	if err != nil {
		return rs, err
	}
	query, err := models.Db.Query("SELECT `id`, `name`, `email`, `status` FROM `recipient`" + partWhere, partParams...)
	if err != nil {
		return rs, err
	}
	defer query.Close()
	for query.Next() {
		r := Recipient{}
		err = query.Scan(&r.Id, &r.Name, &r.Email, &r.Result)
		rs.Records = append(rs.Records, r)
	}
	partWhere, partParams, err = createSqlPart(req, where, params, map[string]string{
		"recid":"id", "name":"name", "email":"email", "result":"status",
	}, false)
	if err != nil {
		return rs, err
	}
	err = models.Db.QueryRow("SELECT COUNT(*) FROM `recipient`" + partWhere, partParams...).Scan(&rs.Total)
	return rs, nil

}

//ToDo check right errors
func getRecipientParams(recipient int64) (RecipientParams, error) {
	var err error
	var p RecipientParam
	var ps RecipientParams
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
	_, err := models.Db.Exec("UPDATE `recipient` SET `removed`=1 WHERE `campaign_id`=?", campaignId)
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
					email = t
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
		progress[file] = int(k) * 100 / total
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
					t, err := cell.String()
					if err != nil {
						apilog.Println(err)
					}
					title[i] = t
				}
			} else {
				email = ""
				name = ""
				data = map[string]string{}
				for i, cell := range v.Cells {
					t, err := cell.String()
					if err != nil {
						apilog.Println(err)
					}
					if i == 0 {
						email = t
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
			progress[file] = int(k) * 100 / total
		}
		err = tx.Commit()
	}

	return err
}
