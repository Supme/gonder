package api

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"gonder/models"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strconv"
)

var progress models.Progress

func RecipientUploadHandlerFunc(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024*30) // ToDo config variable?
	if err := r.ParseForm(); err != nil {
		log.Println(err)
		e := models.JSONResponse{}.ErrorWriter(w, err)
		if e != nil {
			apiLog.Print(e)
		}
		return
	}

	id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	auth := r.Context().Value(ContextAuth).(*Auth)
	if !auth.Right("upload-recipients") || !auth.CampaignRight(id) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if isAccepted(id) {
		e := models.JSONResponse{}.OkWriter(w, "Cannot add recipients to an accepted campaign.")
		if e != nil {
			apiLog.Print(e)
		}
		return
	}
	var content []byte
	content, err = base64.StdEncoding.DecodeString(r.FormValue("content"))
	if err != nil {
		log.Println(err)
		return
	}
	file, err := ioutil.TempFile("", "gonder_recipient_upload_")
	if err != nil {
		log.Println(err)
		e := models.JSONResponse{}.ErrorWriter(w, err)
		if e != nil {
			apiLog.Print(e)
		}
		return
	}
	if _, err = file.Write(content); err != nil {
		e := models.JSONResponse{}.ErrorWriter(w, err)
		if e != nil {
			apiLog.Print(e)
		}
		return
	}
	filename := file.Name()
	if err = file.Close(); err != nil {
		log.Println(err)
		e := models.JSONResponse{}.ErrorWriter(w, err)
		if e != nil {
			apiLog.Print(e)
		}
		return
	}
	apiLog.Print(auth.user.Name, " upload file ", r.FormValue("name"))

	switch path.Ext(r.FormValue("name")) {
	case ".csv":
		go func() {
			if err = models.Campaign(id).LoadRecipientCsv(filename, &progress); err != nil {
				apiLog.Println(err)
			}
			progress.Delete(filename)

		}()
		e := models.JSONResponse{}.OkWriter(w, filename)
		if e != nil {
			apiLog.Print(e)
		}

	case ".xlsx":
		go func() {
			if err = models.Campaign(id).LoadRecipientXlsx(filename, &progress); err != nil {
				apiLog.Println(err)
			}
			progress.Delete(filename)

		}()
		e := models.JSONResponse{}.OkWriter(w, filename)
		if e != nil {
			apiLog.Print(e)
		}

	default:
		e := models.JSONResponse{}.ErrorWriter(w, fmt.Errorf("this not csv or xlsx file"))
		if e != nil {
			apiLog.Print(e)
		}
	}
}

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

type recips []models.RecipientData

type recipParams struct {
	Total   int               `json:"total"`
	Records map[string]string `json:"records"`
}

func recipientsReq(req request) (js []byte, err error) {
	if req.Recipient == 0 {
		switch req.Cmd {
		case "get":
			if !req.auth.Right("get-recipients") && !req.auth.CampaignRight(req.Campaign) {
				return js, errors.New("Forbidden get recipients")
			}
			var rs recipsTable
			rs, err = getRecipients(req)
			if err != nil {
				return js, err
			}
			js, err = json.Marshal(rs)

		case "add":
			if !req.auth.Right("upload-recipients") && !req.auth.CampaignRight(req.Campaign) {
				return js, errors.New("Forbidden add recipients")
			}
			if isAccepted(req.Campaign) {
				return js, errors.New("Cannot add recipients to an accepted campaign.")
			}
			err = models.Campaign(req.Campaign).AddRecipients(req.Recipients)
			if err != nil {
				return js, err
			}

		case "delete":
			if !req.auth.Right("delete-recipients") {
				return js, errors.New("Forbidden delete recipients")
			}
			gs, err := models.RecipientsGroups(req.IDs...)
			if err != nil {
				return js, err
			}
			fmt.Println("groups:", gs)
			for _, g := range gs {
				if !req.auth.CampaignRight(g.IntID()) {
					return js, errors.New("Forbidden delete recipients from this group")
				}
				if isAccepted(int64(g.IntID())) {
					return js, errors.New("Cannot delete recipients from accepted campaign.")
				}
			}
			err = models.RecipientsDelete(req.IDs...)
			if err != nil {
				return js, err
			}

		case "progress":
			if req.auth.Right("upload-recipients") {
				if p := progress.Get(req.Name); p != nil {
					js = []byte(fmt.Sprintf(`{"status": "success", "message": %d}`, *p))
				} else {
					js = []byte(`{"status": "error", "message": "not found"}`)
				}
			}

		case "clear":
			if !req.auth.Right("delete-recipients") || !req.auth.CampaignRight(req.Campaign) {
				return js, errors.New("Forbidden delete recipients")
			}
			err = models.Campaign(req.Campaign).DeleteRecipients()
			if err != nil {
				return js, errors.New("Can't delete all recipients")
			}

		case "resend4xx":
			if !req.auth.Right("accept-campaign") && !req.auth.CampaignRight(req.Campaign) {
				return js, errors.New("Forbidden resend campaign")
			}
			err = resendCampaign(req.Campaign, req.auth)
			if err != nil {
				return js, errors.New("Can't resend")
			}

		case "deduplicate":
			if !req.auth.Right("delete-recipients") && !req.auth.CampaignRight(req.Campaign) {
				return js, errors.New("Forbidden delete recipients")
			}
			var cnt int64
			cnt, err = models.Campaign(req.Campaign).DeduplicateRecipient()
			if err != nil {
				apiLog.Println(err)
				return js, errors.New("Can't deduplicate recipients")
			}
			js = []byte(fmt.Sprintf(`{"status": "success", "message": %d}`, cnt))

		case "unavailable":
			if !req.auth.Right("delete-recipients") && !req.auth.CampaignRight(req.Campaign) {
				return js, errors.New("Forbidden mark unavailable recipients")
			}
			var cnt int64
			cnt, err = models.Campaign(req.Campaign).MarkUnavailableRecentTime(req.Interval)
			if err != nil {
				apiLog.Println(err)
				return js, errors.New("Can't mark unavailable recipients")
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
			if !req.auth.Right("get-recipient-parameters") || !req.auth.CampaignRight(rID) {
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

func resendCampaign(campaignID int64, auth *Auth) error {
	res, err := models.Db.Exec("UPDATE `recipient` SET `status`=NULL WHERE `campaign_id`=? AND `removed`=0 AND LOWER(`status`) REGEXP '^((4[0-9]{2})|(dial tcp)|(read tcp)|(proxy)|(eof)|(remote error)).+'", campaignID)
	if err != nil {
		log.Println(err)
		return err
	}
	c, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		return err
	}
	apiLog.Printf("User %s resend by 4xx code for campaign %d. Resend count %d", auth.user.Name, campaignID, c)
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
	defer func() {
		err := query.Close()
		if err != nil {
			log.Print(err)
		}
	}()
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
	var ps recipParams
	ps.Records = map[string]string{}
	query, err := models.Db.Query("SELECT `key`, `value` FROM `parameter` WHERE `recipient_id`=?", recipient)
	if err != nil {
		log.Println(err)
		return ps, err
	}
	defer func() {
		err := query.Close()
		if err != nil {
			log.Print(err)
		}
	}()
	for query.Next() {
		var k, v string
		err = query.Scan(&k, &v)
		if err != nil {
			log.Println(err)
			return recipParams{}, err
		}
		ps.Records[k] = v
	}
	err = models.Db.QueryRow("SELECT COUNT(*) FROM `parameter` WHERE `recipient_id`=?", recipient).Scan(&ps.Total)
	if err != nil {
		log.Println(err)
	}
	return ps, err
}
