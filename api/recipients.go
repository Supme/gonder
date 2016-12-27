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
	"net/http"
	"os"
	"path"
	"time"
	"fmt"
	"strconv"
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

func recipients(w http.ResponseWriter, r *http.Request) {
	var err error
	var js []byte
	js = []byte(`{"status": "success", "message": ""}`)

	if r.FormValue("request") == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	req, err := parseRequest(r.FormValue("request"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if req.Recipient == 0 {
		switch req.Cmd {
		case "get":
			if auth.Right("get-recipients") && auth.CampaignRight(req.Campaign) {
				rs, err := getRecipients(req.Campaign, req.Offset, req.Limit)
				js, err = json.Marshal(rs)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				js = []byte(`{"status": "error", "message": "Forbidden get recipients"}`)
			}

		case "upload":
			if auth.Right("upload-recipients") && auth.CampaignRight(req.Campaign) {
				content, err := base64.StdEncoding.DecodeString(req.FileContent)
				if err != nil {
					js = []byte(`{"status": "error", "message": "Base64 decode"}`)
				}
				filename := strconv.FormatInt(time.Now().UnixNano(), 10)
				file := models.FromRootDir("tmp/" + filename)
				err = ioutil.WriteFile(file, content, 0644)
				if err != nil {
					js = []byte(`{"status": "error", "message": "Write file"}`)
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
					js = []byte(fmt.Sprintf(`{"status": "ok", "message": "%s"}`, filename))

				} else if path.Ext(req.FileName) == ".xlsx" {
					go func() {
						progress[filename] = 0
						err = recipientXlsx(req.Campaign, filename)
						if err != nil {
							apilog.Println(err)
						}
						delete(progress, filename)
					}()
					js = []byte(fmt.Sprintf(`{"status": "ok", "message": "%s"}`, filename))

				} else {
					js = []byte(`{"status": "error", "message": "This not csv or xlsx file"}`)
				}
			} else {
				js = []byte(`{"status": "error", "message": "Forbidden upload recipients"}`)
			}

		case "progress":
			if auth.Right("upload-recipients") && auth.CampaignRight(req.Campaign) {
				if val, ok := progress[req.Name]; ok {
					js = []byte(fmt.Sprintf(`{"status": "loading", "message": %d}`, val))
				} else {
					js = []byte(`{"status": "not found", "message": ""}`)
				}
			}

		case "clear":
			if auth.Right("delete-recipients") && auth.CampaignRight(req.Campaign) {
				err = delRecipients(req.Campaign)
				if err != nil {
					js = []byte(`{"status": "error", "message": "Can't delete all recipients"}`)
				}
			} else {
				js = []byte(`{"status": "error", "message": "Forbidden delete recipients"}`)
			}

		case "resend4xx":
			if auth.Right("accept-campaign") && auth.CampaignRight(req.Campaign) {
				err = resendCampaign(req.Campaign)
				if err != nil {
					js = []byte(`{"status": "error", "message": "Can't resend"}`)
				}
			} else {
				js = []byte(`{"status": "error", "message": "Forbidden resend campaign"}`)
			}
		}
	} else {
		if req.Cmd == "get" {
			rId, err := getRecipientCampaign(req.Recipient)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if auth.Right("get-recipient-parameters") && auth.CampaignRight(rId) {
				ps, err := getRecipientParams(req.Recipient, req.Offset, req.Limit)
				js, err = json.Marshal(ps)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				js = []byte(`{"status": "error", "message": "Forbidden get recipient parameters"}`)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func resendCampaign(campaignId int64) error {
	res, err := models.Db.Exec("UPDATE `recipient` SET `status`=NULL WHERE `campaign_id`=? AND `removed`=0 AND LOWER(`status`) REGEXP '^((4[0-9]{2})|(dial tcp)|(proxy)|(eof)).+'", campaignId)
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
func getRecipients(campaign, offset, limit int64) (Recipients, error) {
	var err error
	var rs Recipients
	rs.Records = []Recipient{}
	query, err := models.Db.Query("SELECT `id`, `name`, `email`, `status` FROM `recipient` WHERE `removed`!=1 AND `campaign_id`=? LIMIT ? OFFSET ?", campaign, limit, offset)
	if err != nil {
		return rs, err
	}
	defer query.Close()
	for query.Next() {
		r := Recipient{}
		err = query.Scan(&r.Id, &r.Name, &r.Email, &r.Result)
		rs.Records = append(rs.Records, r)
	}
	err = models.Db.QueryRow("SELECT COUNT(*) FROM `recipient` WHERE `removed`!=1 AND `campaign_id`=?", campaign).Scan(&rs.Total)
	return rs, nil

}

//ToDo check right errors
func getRecipientParams(recipient, offset, limit int64) (RecipientParams, error) {
	var err error
	var p RecipientParam
	var ps RecipientParams
	ps.Records = []RecipientParam{}
	query, err := models.Db.Query("SELECT `key`, `value` FROM `parameter` WHERE `recipient_id`=? LIMIT ? OFFSET ?", recipient, limit, offset)
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

// ToDo optimize this
func recipientCsv(campaignId int64, file string) error {

	title := make(map[int]string)
	data := make(map[string]string)
	var email, name string
	f := models.FromRootDir("tmp/" + file)
	csvfile, err := os.Open(f)
	if err != nil {
		return err
	}

	reader := csv.NewReader(csvfile)
	reader.FieldsPerRecord = -1
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return err
	}

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

			res, err := models.Db.Exec("INSERT INTO recipient (`campaign_id`, `email`, `name`) VALUES (?, ?, ?)", campaignId, email, name)
			if err != nil {
				return err
			}
			id, err := res.LastInsertId()
			if err != nil {
				return err
			}
			for i, t := range data {
				_, err := models.Db.Exec("INSERT INTO parameter (`recipient_id`, `key`, `value`) VALUES (?, ?, ?)", id, i, t)
				if err != nil {
					return err
				}
			}
		}
		progress[file] = int(k) * 100 / total
	}

	csvfile.Close()

	os.Remove(f)

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
	if xlFile.Sheets[0] != nil {
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

				res, err := models.Db.Exec("INSERT INTO recipient (`campaign_id`, `email`, `name`) VALUES (?, ?, ?)", campaignId, email, name)
				if err != nil {
					return err
				}
				id, err := res.LastInsertId()
				if err != nil {
					return err
				}
				for i, t := range data {
					_, err := models.Db.Exec("INSERT INTO parameter (`recipient_id`, `key`, `value`) VALUES (?, ?, ?)", id, i, t)
					if err != nil {
						return err
					}
				}

			}
			progress[file] = int(k) * 100 / total
		}
	}

	os.Remove(f)
	return err
}
