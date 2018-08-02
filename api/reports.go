package api

import (
	"encoding/json"
	"fmt"
	"github.com/go-sql-driver/mysql"
	campSender "github.com/supme/gonder/campaign"
	"github.com/supme/gonder/models"
	"log"
	"net/http"
)

func reportCampaignStatus(w http.ResponseWriter, r *http.Request) {
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
	w.Write(js)
}

func reportJumpDetailedCount(w http.ResponseWriter, r *http.Request) {
	var err error
	var js []byte
	if err = r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if r.Form["campaign"] != nil && user.CampaignRight(r.Form["campaign"][0]) {
		var url string
		var count int
		res := make(map[string]int)
		query, err := models.Db.Query(fmt.Sprintf("SELECT `jumping`.`url` as link, ( SELECT COUNT(`jumping`.`id`) FROM `jumping` INNER JOIN `recipient` ON `jumping`.`recipient_id`=`recipient`.`id` WHERE `jumping`.`url`=link AND `jumping`.`campaign_id`=? AND `recipient`.`removed`=0 ) as cnt FROM `jumping` INNER JOIN `recipient` ON `jumping`.`recipient_id`=`recipient`.`id` WHERE `url` NOT IN ('%s', '%s', '%s') AND `jumping`.`campaign_id`=? AND `recipient`.`removed`=0 GROUP BY `jumping`.`url`", models.WebVersion, models.OpenTrace, models.Unsubscribe), r.Form["campaign"][0], r.Form["campaign"][0])
		if err != nil {
			apilog.Print(err)
		}
		defer query.Close()
		for query.Next() {
			err = query.Scan(&url, &count)
			if err != nil {
				apilog.Print(err)
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
	w.Write(js)
}

func reportUnsubscribed(w http.ResponseWriter, r *http.Request) {
	var err error
	var js []byte
	if err = r.ParseForm(); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if (r.Form["group"] != nil && user.GroupRight(r.Form["group"][0])) || (r.Form["campaign"] != nil && user.CampaignRight(r.Form["campaign"][0])) {
		var (
			id                 int64
			queryString, param string
			timestamp          mysql.NullTime
		)
		type U struct {
			Email string              `json:"email"`
			Date  int64               `json:"date"`
			Extra []map[string]string `json:"extra,omitempty"`
		}
		res := []U{}

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
		defer query.Close()
		for query.Next() {
			rs := U{}
			err = query.Scan(&id, &rs.Email, &timestamp)
			if err != nil {
				apilog.Print(err)
			}
			rs.Date = timestamp.Time.Unix()
			// extra data
			extra := map[string]string{}
			var (
				name  string
				value string
			)
			q, err := models.Db.Query("SELECT `name`, `value` FROM `unsubscribe_extra` WHERE `unsubscribe_id`=?", id)
			if err != nil {
				apilog.Print(err)
			}
			for q.Next() {
				err = q.Scan(&name, &value)
				if err != nil {
					log.Println(err)
					http.Error(w, "Param error", http.StatusInternalServerError)
					return
				}
				extra[name] = value
				rs.Extra = append(rs.Extra, extra)
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

func report(w http.ResponseWriter, r *http.Request) {
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
	if user.CampaignRight(campaignID) {
		reports := make(map[string]interface{})
		reports["Campaign"], err = reportCampaignInfo(campaignID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		reports["SendCount"] = reportSendCount(campaignID)
		reports["SuccessSendCount"] = reportSuccessSendCount(campaignID)
		reports["OpenMailCount"] = reportOpenMailCount(campaignID)
		reports["OpenWebVersionCount"] = reportOpenWebVersionCount(campaignID)
		reports["UnsubscribeCount"] = reportUnsubscribeCount(campaignID)
		reports["RecipientJumpCount"] = reportRecipientJumpCount(campaignID)

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

// Campaign info
func reportCampaignInfo(campaignID string) (map[string]interface{}, error) {
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
func reportSendCount(campaignID string) int {
	var count int
	err := models.Db.QueryRow("SELECT COUNT(*) FROM `recipient` WHERE `campaign_id`=? AND `status` IS NOT NULL AND `removed`=0", campaignID).Scan(&count)
	if err != nil {
		log.Println(err)
	}
	return count
}

// Success send
func reportSuccessSendCount(campaignID string) int {
	var count int
	err := models.Db.QueryRow("SELECT COUNT(*) FROM `recipient` WHERE `campaign_id`=? AND `status`='Ok' AND `removed`=0", campaignID).Scan(&count)
	if err != nil {
		apilog.Print(err)
	}
	return count
}

// Unsubscribe count
func reportUnsubscribeCount(campaignID string) int {
	var count int
	err := models.Db.QueryRow("SELECT COUNT(DISTINCT `email`) FROM `unsubscribe` WHERE `campaign_id`=?", campaignID).Scan(&count)
	if err != nil {
		log.Println(err)
	}
	return count
}

// Open mail count
func reportOpenMailCount(campaignID string) int {
	var count int
	err := models.Db.QueryRow("SELECT COUNT(*) FROM `recipient` WHERE `campaign_id`=? AND (`client_agent` IS NOT NULL OR `web_agent` IS NOT NULL) AND `removed`=0", campaignID).Scan(&count)
	if err != nil {
		log.Println(err)
	}
	return count
}

// Web version count
func reportOpenWebVersionCount(campaignID string) int {
	var count int
	err := models.Db.QueryRow(fmt.Sprintf("SELECT COUNT(DISTINCT `recipient`.`id`) FROM `jumping` INNER JOIN `recipient` ON `jumping`.`recipient_id`=`recipient`.`id` WHERE `jumping`.`campaign_id`=? AND `jumping`.`url`='%s' AND `recipient`.`removed`=0", models.WebVersion), campaignID).Scan(&count)
	if err != nil {
		log.Println(err)
	}
	return count
}

func reportRecipientJumpCount(campaignID string) int {
	var count int
	err := models.Db.QueryRow(fmt.Sprintf("SELECT COUNT(DISTINCT `recipient`.`id`) FROM `jumping` INNER JOIN `recipient` ON `jumping`.`recipient_id`=`recipient`.`id` WHERE `recipient`.`removed`=0 AND `jumping`.`url` NOT IN ('%s', '%s', '%s') AND `jumping`.`campaign_id`=?", models.OpenTrace, models.WebVersion, models.Unsubscribe), campaignID).Scan(&count)
	if err != nil {
		log.Println(err)
	}
	return count
}
