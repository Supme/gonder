package api

import (
	"database/sql"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/supme/gonder/models"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"sync"
)

type recipTable struct {
	ID     int64  `json:"recid"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Result string `json:"result"`
	Open   bool   `json:"open"`
}

type recipsTable struct {
	Total   int          `json:"total"`
	Records []recipTable `json:"records"`
}

type recips []recip

type recip struct {
	Name   string       `json:"name"`
	Email  string       `json:"email"`
	Params []recipParam `json:"params"`
}

type recipParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type recipParams struct {
	Total   int          `json:"total"`
	Records []recipParam `json:"records"`
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
			if !user.Right("get-recipients") && !user.CampaignRight(req.Campaign) {
				return js, errors.New("Forbidden get recipients")
			}
			var rs recipsTable
			rs, err = getRecipients(req)
			if err != nil {
				return js, err
			}
			js, err = json.Marshal(rs)

		case "add":
			if !user.Right("upload-recipients") && !user.CampaignRight(req.Campaign) {
				return js, errors.New("Forbidden add recipients")
			}
			err = addRecipients(req.Campaign, req.Recipients)
			if err != nil {
				return js, err
			}

		case "upload":
			if user.Right("upload-recipients") && user.CampaignRight(req.Campaign) {
				var content []byte
				content, err = base64.StdEncoding.DecodeString(req.FileContent)
				if err != nil {
					log.Println(err)
					return js, err
				}
				file, err := ioutil.TempFile("", "gonder_recipient_load_")
				if err != nil {
					log.Println(err)
					return js, err
				}
				_, err = file.Write(content)
				if err != nil {
					log.Println(err)
					return js, err
				}
				filename := file.Name()
				err = file.Close()
				if err != nil {
					log.Println(err)
					return js, err
				}
				apilog.Print(user.name, " upload file ", req.FileName)

				switch path.Ext(req.FileName) {
				case ".csv":
					go func() {
						progress.Lock()
						progress.cnt[file.Name()] = 0
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

				case ".xlsx":
					go func() {
						progress.Lock()
						progress.cnt[file.Name()] = 0
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

				default:
					return js, errors.New("This not csv or xlsx file")
				}

			} else {
				return js, errors.New("Forbidden upload recipients")
			}
			if err != nil {
				log.Println(err)
				return js, err
			}

		case "progress":
			if user.Right("upload-recipients") {
				progress.RLock()
				if val, ok := progress.cnt[req.Name]; ok {
					js = []byte(fmt.Sprintf(`{"status": "success", "message": %d}`, val))
				} else {
					js = []byte(`{"status": "error", "message": "not found"}`)
				}
				progress.RUnlock()
			}

		case "clear":
			if !user.Right("delete-recipients") || !user.CampaignRight(req.Campaign) {
				return js, errors.New("Forbidden delete recipients")
			}
			err = delRecipients(req.Campaign)
			if err != nil {
				return js, errors.New("Can't delete all recipients")
			}

		case "resend4xx":
			if !user.Right("accept-campaign") && !user.CampaignRight(req.Campaign) {
				return js, errors.New("Forbidden resend campaign")
			}
			err = resendCampaign(req.Campaign)
			if err != nil {
				return js, errors.New("Can't resend")
			}

		case "deduplicate":
			if !user.Right("delete-recipients") && !user.CampaignRight(req.Campaign) {
				return js, errors.New("Forbidden delete recipients")
			}
			var cnt int64
			cnt, err = deduplicateRecipient(req.Campaign)
			if err != nil {
				apilog.Println(err)
				return js, errors.New("Can't deduplicate recipients")
			}
			js = []byte(fmt.Sprintf(`{"status": "success", "message": %d}`, cnt))

		case "unavaible":
			if !user.Right("delete-recipients") && !user.CampaignRight(req.Campaign) {
				return js, errors.New("Forbidden mark unavaible recipients")
			}
			var cnt int64
			cnt, err = markUnavaibleRecentTime(req.Campaign)
			if err != nil {
				apilog.Println(err)
				return js, errors.New("Can't mark unavaible recipients")
			}
			js = []byte(fmt.Sprintf(`{"status": "success", "message": %d}`, cnt))

		default:
			err = errors.New("Command not found")
		}

	} else {
		if req.Cmd == "get" {
			var rID int64
			rID, err = getRecipientCampaign(req.Recipient)
			if err != nil {
				return js, err
			}
			if !user.Right("get-recipient-parameters") || !user.CampaignRight(rID) {
				return js, errors.New("Forbidden get recipient parameters")
			}
			var ps recipParams
			ps, err = getRecipientParams(req.Recipient)
			if err != nil {
				return js, err
			}
			js, err = json.Marshal(ps)
			if err != nil {
				log.Println(err)
			}

		} else {
			err = errors.New("Command not found")
		}
	}
	return js, err
}

// Remove later unavaible status like:
//  invalid mailbox
//  no such user
//  does not exist
//  unknown user
//  user unknown
//  user not found
//  bad destination mailbox
//  mailbox unavailable
// ToDo ALTER TABLE `recipient` ADD FULLTEXT(`status`); ??? why this slowly ???
// ToDo optimize this
func markUnavaibleRecentTime(campaignID int64) (cnt int64, err error) {
	p, err := models.Db.Prepare(fmt.Sprintf(`UPDATE recipient SET status="%s" WHERE id=?`, models.StatusUnavaibleRecentTime))
	if err != nil {
		log.Println(err)
		return
	}
	defer p.Close()

	q, err := models.Db.Query(`
SELECT id FROM recipient WHERE email IN
 (SELECT rs.email FROM recipient as rs WHERE
    date>(NOW() - INTERVAL 30 DAY)
   AND
   (rs.status LIKE "%invalid mailbox%" OR
    rs.status LIKE "%no such user%" OR
    rs.status LIKE "%does not exist%" OR
    rs.status LIKE "%unknown user%" OR
    rs.status LIKE "%user unknown%" OR
    rs.status LIKE "%user not found%" OR
    rs.status LIKE "%bad destination mailbox%" OR
    rs.status LIKE "%mailbox unavailable%" OR
    rs.status="Ok")
  GROUP BY rs.email
  HAVING SUM(rs.status!="Ok")>0 AND SUM(rs.status="Ok")=0)
AND removed=0
AND status IS NULL
AND campaign_id=?`, campaignID)
	if err != nil {
		log.Println(err)
		return
	}
	defer q.Close()

	cnt = 0
	for q.Next() {
		var id int64
		err = q.Scan(&id)
		if err != nil {
			log.Println(err)
			return
		}
		// ToDo check q.NextResultSet() and batch update
		_, err = p.Exec(id)
		if err != nil {
			log.Println(err)
			return
		}
		cnt++
	}

	return
}

func deduplicateRecipient(campaignID int64) (cnt int64, err error) {
	q, err := models.Db.Query(`
	SELECT r1.id FROM recipient as r1
		JOIN (
			SELECT MIN(id) AS id, email FROM recipient WHERE
             	campaign_id=? AND removed=0 AND status IS NULL
             	GROUP BY email HAVING COUNT(*)>1) as r2 ON (r1.email=r2.email AND r1.id!=r2.id
		)
	WHERE r1.campaign_id=? AND removed=0 AND status IS NULL;
	`, campaignID, campaignID)
	if err != nil {
		log.Println(err)
		return
	}

	tx, err := models.Db.Begin()
	if err != nil {
		log.Println(err)
		return
	}
	defer tx.Rollback()

	dupl, err := tx.Prepare("UPDATE `recipient` SET `removed`=2 WHERE id=?")
	if err != nil {
		log.Println(err)
		return
	}
	defer dupl.Close()

	cnt = 0
	for q.Next() {
		var id int64
		err = q.Scan(&id)
		if err != nil {
			log.Println(err)
			return
		}
		_, err = dupl.Exec(id)
		if err != nil {
			log.Println(err)
			return
		}
		cnt = cnt + 1
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
	}
	return
}

func addRecipients(campaignID int64, recipients recips) error {
	tx, err := models.Db.Begin()
	if err != nil {
		log.Println(err)
		return err
	}
	defer tx.Rollback()

	stRecipient, err := tx.Prepare("INSERT INTO recipient (`campaign_id`, `email`, `name`) VALUES (?, ?, ?)")
	if err != nil {
		log.Println(err)
		return err
	}
	defer stRecipient.Close()
	stParameter, err := tx.Prepare("INSERT INTO parameter (`recipient_id`, `key`, `value`) VALUES (?, ?, ?)")
	if err != nil {
		log.Println(err)
		return err
	}
	defer stParameter.Close()

	for r := range recipients {
		res, err := stRecipient.Exec(campaignID, strings.TrimSpace(recipients[r].Email), recipients[r].Name)
		if err != nil {
			log.Println(err)
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			log.Println(err)
			return err
		}
		for _, p := range recipients[r].Params {
			_, err := stParameter.Exec(id, p.Key, p.Value)
			if err != nil {
				log.Println(err)
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
	}
	return err
}

func resendCampaign(campaignID int64) error {
	res, err := models.Db.Exec("UPDATE `recipient` SET `status`=NULL WHERE `campaign_id`=? AND `removed`=0 AND LOWER(`status`) REGEXP '^((4[0-9]{2})|(dial tcp)|(read tcp)|(proxy)|(eof)).+'", campaignID)
	if err != err {
		log.Println(err)
		return err
	}
	c, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		return err
	}
	apilog.Printf("User %s resend by 4xx code for campaign %d. Resend count %d", user.name, campaignID, c)
	return nil
}

func getRecipientCampaign(recipientID int64) (int64, error) {
	var id int64
	err := models.Db.QueryRow("SELECT `campaign_id` FROM `recipient` WHERE `id`=?", recipientID).Scan(&id)
	if err != nil {
		log.Println(err)
	}
	return id, err
}

//ToDo check right errors
//ToDo add unsubscribe column
func getRecipients(req request) (recipsTable, error) {
	var (
		err                error
		rs                 recipsTable
		partWhere, where   string
		partParams, params []interface{}
	)

	params = append(params, req.Campaign)

	rs.Records = []recipTable{}
	where = " WHERE `removed`=0 AND `campaign_id`=?"
	partWhere, partParams, err = createSQLPart(req, where, params, map[string]string{
		"recid": "id", "name": "name", "email": "email", "result": "status", "open": "open",
	}, true)
	if err != nil {
		log.Println(err)
		return rs, err
	}

	var query *sql.Rows
	query, err = models.Db.Query("SELECT `id`, `name`, `email`, `status`, IF(COALESCE(`web_agent`,`client_agent`) IS NULL, 0, 1) as `open` FROM `recipient`"+partWhere, partParams...)
	if err != nil {
		log.Println(err)
		return rs, err
	}
	defer query.Close()
	for query.Next() {
		r := recipTable{}
		var result sql.NullString
		err = query.Scan(&r.ID, &r.Name, &r.Email, &result, &r.Open)
		if err != nil {
			log.Println(err)
			return rs, err
		}
		r.Result = result.String
		rs.Records = append(rs.Records, r)
	}
	partWhere, partParams, err = createSQLPart(req, where, params, map[string]string{
		"recid": "id", "name": "name", "email": "email", "result": "status", "open": "open",
	}, false)
	if err != nil {
		log.Println(err)
		return rs, err
	}
	err = models.Db.QueryRow("SELECT COUNT(*) FROM `recipient`"+partWhere, partParams...).Scan(&rs.Total)
	if err != nil {
		log.Println(err)
	}
	return rs, err

}

//ToDo check right errors
func getRecipientParams(recipient int64) (recipParams, error) {
	var p recipParam
	var ps recipParams
	ps.Records = []recipParam{}
	query, err := models.Db.Query("SELECT `key`, `value` FROM `parameter` WHERE `recipient_id`=?", recipient)
	if err != nil {
		log.Println(err)
		return ps, err
	}
	defer query.Close()
	for query.Next() {
		err = query.Scan(&p.Key, &p.Value)
		if err != nil {
			log.Println(err)
			return recipParams{}, err
		}
		ps.Records = append(ps.Records, p)
	}
	err = models.Db.QueryRow("SELECT COUNT(*) FROM `parameter` WHERE `recipient_id`=?", recipient).Scan(&ps.Total)
	if err != nil {
		log.Println(err)
	}
	return ps, err
}

func delRecipients(campaignID int64) error {
	_, err := models.Db.Exec("UPDATE `recipient` SET `removed`=1 WHERE `campaign_id`=? AND `removed`=0", campaignID)
	if err != nil {
		log.Println(err)
	}
	return err
}

func recipientCsv(campaignID int64, file string) error {
	title := make(map[int]string)
	var email, name string

	csvfile, err := os.Open(file)
	if err != nil {
		log.Println(err)
		return err
	}
	defer csvfile.Close()
	defer os.Remove(file)

	reader := csv.NewReader(csvfile)
	reader.FieldsPerRecord = -1
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		log.Println(err)
		return err
	}

	tx, err := models.Db.Begin()
	if err != nil {
		log.Println(err)
		return err
	}
	defer tx.Rollback()

	stRecipient, err := tx.Prepare("INSERT INTO recipient (`campaign_id`, `email`, `name`) VALUES (?, ?, ?)")
	if err != nil {
		log.Println(err)
		return err
	}
	defer stRecipient.Close()
	stParameter, err := tx.Prepare("INSERT INTO parameter (`recipient_id`, `key`, `value`) VALUES (?, ?, ?)")
	if err != nil {
		log.Println(err)
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
			data := map[string]string{}
			for i, t := range v {
				if i == 0 {
					email = strings.TrimSpace(t)
				} else if i == 1 {
					name = t
				} else {
					data[title[i]] = t
				}
			}

			res, err := stRecipient.Exec(campaignID, email, name)
			if err != nil {
				log.Println(err)
				return err
			}
			id, err := res.LastInsertId()
			if err != nil {
				log.Println(err)
				return err
			}
			for i, t := range data {
				_, err := stParameter.Exec(id, i, t)
				if err != nil {
					log.Println(err)
					return err
				}
			}
		}
		progress.Lock()
		progress.cnt[file] = int(k) * 100 / total
		progress.Unlock()
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
	}
	return err
}

func recipientXlsx(campaignID int64, file string) error {
	title := make(map[int]string)

	var email, name string

	xlFile, err := xlsx.OpenFile(file)
	if err != nil {
		log.Println(err)
		return err
	}
	defer os.Remove(file)

	if xlFile.Sheets[0] != nil {
		var tx *sql.Tx
		tx, err = models.Db.Begin()
		if err != nil {
			log.Println(err)
			return err
		}
		defer tx.Rollback()

		var stRecipient *sql.Stmt
		stRecipient, err = tx.Prepare("INSERT INTO recipient (`campaign_id`, `email`, `name`) VALUES (?, ?, ?)")
		if err != nil {
			log.Println(err)
			return err
		}
		defer stRecipient.Close()

		var stParameter *sql.Stmt
		stParameter, err = tx.Prepare("INSERT INTO parameter (`recipient_id`, `key`, `value`) VALUES (?, ?, ?)")
		if err != nil {
			log.Println(err)
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
				data := make(map[string]string)
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

				var res sql.Result
				res, err = stRecipient.Exec(campaignID, email, name)
				if err != nil {
					log.Println(err)
					return err
				}

				var id int64
				id, err = res.LastInsertId()
				if err != nil {
					log.Println(err)
					return err
				}
				for i, t := range data {
					_, err = stParameter.Exec(id, i, t)
					if err != nil {
						log.Println(err)
						return err
					}
				}

			}
			progress.Lock()
			progress.cnt[file] = int(k) * 100 / total
			progress.Unlock()
		}
		err = tx.Commit()
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}
