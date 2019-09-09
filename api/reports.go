package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-sql-driver/mysql"
	campSender "gonder/campaign"
	"gonder/models"
	"log"
	"net/http"
)

func reportQuestionSummary(w http.ResponseWriter, r *http.Request) {
	var err error
	var js []byte
	type resultUnit struct {
		RecipientID int64             `json:"recipient_id"`
		Email       string            `json:"email"`
		At          int64             `json:"at"`
		Data        map[string]string `json:"data"`
	}

	user := r.Context().Value("Auth").(*Auth)
	if user.CampaignRight(r.FormValue("campaign")) {
		var result []resultUnit

		query, err := models.Db.Query("SELECT `question`.`id`, `question`.`recipient_id`, `recipient`.`email`, `question`.`at` FROM `question` LEFT JOIN `recipient` ON `question`.`recipient_id`=`recipient`.`id` WHERE `recipient`.`campaign_id`=?", r.FormValue("campaign"))
		if err != nil {
			log.Print(err)
		}
		defer func() {
			if err := query.Close(); err != nil {
				log.Print(err)
			}
		}()

		stmtData, err := models.Db.Prepare("SELECT `name`, `value` FROM `question_data` WHERE `question_id`=?")
		if err != nil {
			log.Print(err)
		}

		result = []resultUnit{}
		for query.Next() {
			var (
				res        resultUnit
				questionID int64
				at         mysql.NullTime
			)
			err = query.Scan(&questionID, &res.RecipientID, &res.Email, &at)
			if err != nil {
				log.Print(err)
				break
			}
			res.At = at.Time.Unix()
			questionData := map[string]string{}
			rows, err := stmtData.Query(questionID)
			if err != nil {
				log.Print(err)
			}
			for rows.Next() {
				var name, value string
				err = rows.Scan(&name, &value)
				if err != nil {
					log.Print(err)
					break
				}
				questionData[name] = value
			}
			res.Data = questionData
			result = append(result, res)
		}

		js, err = json.Marshal(result)
		if err != nil {
			js = []byte(`{"status": "error", "message": "JSON marshaling result"}`)
		}
	} else {
		js = []byte(`{"status": "error", "message": "Forbidden get reports for this campaign"}`)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)
	if err != nil {
		log.Println(err)
	}
}

func reportStartedCampaign(w http.ResponseWriter, r *http.Request) {
	var err error
	var js []byte
	type startedStruct struct {
		Started []string `json:"started"`
	}
	var campsVar startedStruct
	campsVar.Started = campSender.Sending.Started()
	js, err = json.Marshal(campsVar)
	if err != nil {
		js = []byte(`{"status": "error", "message": "Get reports for this campaign"}`)
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)
	if err != nil {
		log.Println(err)
	}
}

func reportSummary(w http.ResponseWriter, r *http.Request) {
	var err error
	var js []byte
	if err = r.ParseForm(); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var campaignID string
	if len(r.Form["campaign"]) > 0 {
		campaignID = r.Form["campaign"][0]
	}
	user := r.Context().Value("Auth").(*Auth)
	if user.CampaignRight(campaignID) {
		reports := make(map[string]interface{})
		reports["Campaign"], err = _reportCampaignInfo(campaignID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		reports["SendCount"] = _reportSendCount(campaignID)
		reports["SuccessSendCount"] = _reportSuccessSendCount(campaignID)
		reports["OpenMailCount"] = _reportOpenMailCount(campaignID)
		reports["OpenWebVersionCount"] = _reportOpenWebVersionCount(campaignID)
		reports["UnsubscribeCount"] = _reportUnsubscribeCount(campaignID)
		reports["RecipientJumpCount"] = _reportRecipientJumpCount(campaignID)

		js, err = json.Marshal(reports)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	} else {
		js = []byte(`{"status": "error", "message": "Forbidden get reports for this campaign"}`)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)
	if err != nil {
		log.Println(err)
	}
}

func reportClickCount(w http.ResponseWriter, r *http.Request) {
	var err error
	var js []byte
	if err = r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user := r.Context().Value("Auth").(*Auth)
	if r.Form["campaign"] != nil && user.CampaignRight(r.Form["campaign"][0]) {
		var url string
		var count int
		res := make(map[string]int)
		query, err := models.Db.Query(fmt.Sprintf("SELECT `jumping`.`url` as link, ( SELECT COUNT(`jumping`.`id`) FROM `jumping` INNER JOIN `recipient` ON `jumping`.`recipient_id`=`recipient`.`id` WHERE `jumping`.`url`=link AND `jumping`.`campaign_id`=? AND `recipient`.`removed`=0 ) as cnt FROM `jumping` INNER JOIN `recipient` ON `jumping`.`recipient_id`=`recipient`.`id` WHERE `url` NOT IN ('%s', '%s', '%s') AND `jumping`.`campaign_id`=? AND `recipient`.`removed`=0 GROUP BY `jumping`.`url`", models.WebVersion, models.OpenTrace, models.Unsubscribe), r.Form["campaign"][0], r.Form["campaign"][0])
		if err != nil {
			log.Print(err)
		}
		defer func() {
			if err := query.Close(); err != nil {
				log.Print(err)
			}
		}()
		for query.Next() {
			err = query.Scan(&url, &count)
			if err != nil {
				log.Print(err)
			}
			res[url] = count
		}
		js, err = json.Marshal(res)
		if err != nil {
			log.Println(err)
			js = []byte(`{"status": "error", "message": "Get reports for this campaign"}`)
		}
	} else {
		js = []byte(`{"status": "error", "message": "Forbidden get reports for this campaign"}`)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)
	if err != nil {
		log.Print(err)
	}
}

func reportRecipientsList(w http.ResponseWriter, r *http.Request) {
	var err error
	if err = r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user := r.Context().Value("Auth").(*Auth)
	if r.Form["campaign"] != nil && user.CampaignRight(r.Form["campaign"][0]) && user.Right("get-recipient-parameters") {
		type rcptLineType struct {
			Id     int64  `json:"id"`
			Email  string `json:"email"`
			Name   string `json:"name"`
			Date   int64  `json:"date"`
			Open   bool   `json:"open"`
			Status string `json:"status"`
		}

		query, err := models.Db.Query("SELECT `id`, `email`, `name`, `date`, IF(COALESCE(`web_agent`,`client_agent`) IS NULL, 0, 1) as `open`, `status` FROM `recipient` WHERE `campaign_id`=? AND `removed`=0", r.Form["campaign"][0])
		if err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer query.Close()

		var rcpts []rcptLineType
		for query.Next() {
			var rcpt rcptLineType
			var date mysql.NullTime
			err = query.Scan(&rcpt.Id, &rcpt.Email, &rcpt.Name, &date, &rcpt.Open, &rcpt.Status)
			if err != nil {
				log.Print(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			rcpt.Date = date.Time.Unix()
			rcpts = append(rcpts, rcpt)
		}

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		err = enc.Encode(rcpts)
		if err != nil {
			log.Print(err)
			return
		}

		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write([]byte(`{"status": "error", "message": "Forbidden get reports for this campaign"}`)); err != nil {
		log.Print(err)
	}
}

//ToDo add to readme
func reportRecipientClicks(w http.ResponseWriter, r *http.Request) {
	var err error
	if err = r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var campaign string
	if len(r.Form["recipient"]) == 1 {
		err = models.Db.QueryRow("SELECT `campaign_id` FROM `recipient` WHERE `id`=?", r.Form["recipient"][0]).Scan(&campaign)
		if err == sql.ErrNoRows {
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"status": "error", "message": "unknown recipient"}`)); err != nil {
				log.Print(err)
			}
			return
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"status": "error", "message": "required parameter missing"}`)); err != nil {
			log.Print(err)
		}
		return
	}

	user := r.Context().Value("Auth").(*Auth)
	if err == nil && user.CampaignRight(campaign) && user.Right("get-recipient-parameters") {
		type clickType struct {
			URL  string `json:"url"`
			Date int64  `json:"date"`
		}

		query, err := models.Db.Query("SELECT `url`, `date` FROM `jumping` WHERE `recipient_id`=?", r.Form["recipient"][0])
		if err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := query.Close(); err != nil {
				log.Print(err)
			}
		}()

		var clicks []clickType
		for query.Next() {
			var click clickType
			var date mysql.NullTime
			err = query.Scan(&click.URL, &date)
			if err != nil {
				log.Print(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			click.Date = date.Time.Unix()
			clicks = append(clicks, click)
		}

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		err = enc.Encode(clicks)
		if err != nil {
			log.Print(err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write([]byte(`{"status": "error", "message": "Forbidden get reports for this recipient"}`)); err != nil {
		log.Print(err)
	}
}

func reportUnsubscribed(w http.ResponseWriter, r *http.Request) {
	var err error
	var js []byte
	if err = r.ParseForm(); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user := r.Context().Value("Auth").(*Auth)
	if (r.Form["group"] != nil && user.GroupRight(r.Form["group"][0])) || (r.Form["campaign"] != nil && user.CampaignRight(r.Form["campaign"][0])) {
		type U struct {
			Email string            `json:"email"`
			Date  int64             `json:"date"`
			Extra map[string]string `json:"extra,omitempty"`
		}
		var (
			queryString, param string
			timestamp          mysql.NullTime
			id                 int64
			res                []U
		)

		if r.Form["group"] != nil {
			queryString = "SELECT `id`, `email`, `date` FROM `unsubscribe` WHERE `group_id`=?"
			param = r.Form["group"][0]
		} else if r.Form["campaign"] != nil {
			queryString = "SELECT `id`, `email`, `date` FROM `unsubscribe` WHERE `campaign_id`=?"
			param = r.Form["campaign"][0]
		} else {
			http.Error(w, "Param error", http.StatusInternalServerError)
		}

		query, err := models.Db.Query(queryString, param)
		if err != nil {
			log.Println(err)
			http.Error(w, "Query to database error", http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := query.Close(); err != nil {
				log.Print(err)
			}
		}()
		for query.Next() {
			var rs U
			err = query.Scan(&id, &rs.Email, &timestamp)
			if err != nil {
				log.Print(err)
			}
			rs.Date = timestamp.Time.Unix()
			q, err := models.Db.Query("SELECT `name`, `value` FROM `unsubscribe_extra` WHERE `unsubscribe_id`=?", id)
			if err != nil {
				log.Print(err)
			}
			rs.Extra = map[string]string{}
			for q.Next() {
				var name, value string
				err = q.Scan(&name, &value)
				if err != nil {
					log.Println(err)
					http.Error(w, "Param error", http.StatusInternalServerError)
					return
				}
				rs.Extra[name] = value

			}
			err = q.Close()
			if err != nil {
				log.Println(err)
				http.Error(w, "Close request to database error", http.StatusInternalServerError)
				return
			}

			res = append(res, rs)
		}
		js, err = json.Marshal(res)
		if err != nil {
			http.Error(w, "Param error", http.StatusInternalServerError)
			js = []byte(`{"status": "error", "message": "Get reports"}`)
		}
	} else {
		js = []byte(`{"status": "error", "message": "Forbidden get reports"}`)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)
	if err != nil {
		log.Println(err)
	}
}

// Campaign info
func _reportCampaignInfo(campaignID string) (map[string]interface{}, error) {
	var name string
	var timestamp mysql.NullTime
	res := make(map[string]interface{})
	err := models.Db.QueryRow("SELECT `name`, `start_time` FROM `campaign` WHERE `id`=?", campaignID).Scan(&name, &timestamp)
	if err != nil {
		log.Println(err)
		return res, err
	}
	res["name"] = name
	res["Start"] = timestamp.Time.Unix()
	return res, nil
}

// Send
func _reportSendCount(campaignID string) int {
	var count int
	err := models.Db.QueryRow("SELECT COUNT(*) FROM `recipient` WHERE `campaign_id`=? AND `status` IS NOT NULL AND `removed`=0", campaignID).Scan(&count)
	if err != nil {
		log.Println(err)
	}
	return count
}

// Success send
func _reportSuccessSendCount(campaignID string) int {
	var count int
	err := models.Db.QueryRow("SELECT COUNT(*) FROM `recipient` WHERE `campaign_id`=? AND `status`='Ok' AND `removed`=0", campaignID).Scan(&count)
	if err != nil {
		log.Print(err)
	}
	return count
}

// Unsubscribe count
func _reportUnsubscribeCount(campaignID string) int {
	var count int
	err := models.Db.QueryRow("SELECT COUNT(DISTINCT `email`) FROM `unsubscribe` WHERE `campaign_id`=?", campaignID).Scan(&count)
	if err != nil {
		log.Println(err)
	}
	return count
}

// Open mail count
func _reportOpenMailCount(campaignID string) int {
	var count int
	err := models.Db.QueryRow("SELECT COUNT(*) FROM `recipient` WHERE `campaign_id`=? AND (`client_agent` IS NOT NULL OR `web_agent` IS NOT NULL) AND `removed`=0", campaignID).Scan(&count)
	if err != nil {
		log.Println(err)
	}
	return count
}

// Web version count
func _reportOpenWebVersionCount(campaignID string) int {
	var count int
	err := models.Db.QueryRow(fmt.Sprintf("SELECT COUNT(DISTINCT `recipient`.`id`) FROM `jumping` INNER JOIN `recipient` ON `jumping`.`recipient_id`=`recipient`.`id` WHERE `jumping`.`campaign_id`=? AND `jumping`.`url`='%s' AND `recipient`.`removed`=0", models.WebVersion), campaignID).Scan(&count)
	if err != nil {
		log.Println(err)
	}
	return count
}

func _reportRecipientJumpCount(campaignID string) int {
	var count int
	err := models.Db.QueryRow(fmt.Sprintf("SELECT COUNT(DISTINCT `recipient`.`id`) FROM `jumping` INNER JOIN `recipient` ON `jumping`.`recipient_id`=`recipient`.`id` WHERE `recipient`.`removed`=0 AND `jumping`.`url` NOT IN ('%s', '%s', '%s') AND `jumping`.`campaign_id`=?", models.OpenTrace, models.WebVersion, models.Unsubscribe), campaignID).Scan(&count)
	if err != nil {
		log.Println(err)
	}
	return count
}
