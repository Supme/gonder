package api

import (
	"net/http"
	"encoding/json"
	"github.com/supme/gonder/models"
	"github.com/go-sql-driver/mysql"
)

/*
Название рассылки
Дата рассылки (Fact)
Кол-во отправленных писем
Количество недоставленных писем
Кол-во открытых писем
Кол-во переходов из письма
 */

func report(w http.ResponseWriter, r *http.Request)  {

	var err error
	var js []byte
	if err = r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	campaign := r.Form["campaign"][0]
	if auth.CampaignRight(campaign) {
		reports := make(map[string]interface{})
		reports["Campaign"], err = reportCampaignInfo(campaign)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		reports["SendCount"] = reportSendCount(campaign)
		reports["SuccessSendCount"] = reportSuccessSendCount(campaign)
		reports["OpenMailCount"] = reportOpenMailCount(campaign)
		reports["OpenWebVersionCount"] = reportOpenWebVersionCount(campaign)
		reports["UnsubscribeCount"] = reportUnsubscribeCount(campaign)
		reports["JumpDetailedCount"] = reportJumpDetailedCount(campaign)
		reports["RecipientJumpCount"] = reportRecipientJumpCount(campaign)

		js, err = json.Marshal(reports)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	} else {
		js = []byte(`{"status": "error", "message": "Forbidden get reports for this campaign"}`)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// Campaign info
func reportCampaignInfo(campaignId string) (map[string]interface{}, error) {
	var name string
	var timestamp mysql.NullTime
	res := make(map[string]interface{})
	err := models.Db.QueryRow("SELECT `name`, `start_time` FROM `campaign` WHERE `id`=?", campaignId).Scan(&name, &timestamp)
	if err != nil {
		apilog.Print(err)
		return res, err
	}
	res["Name"] = name
	res["Start"] = timestamp.Time.Unix()
	return res, nil
}

// Send
func reportSendCount(campaignId string) int {
	var count int
	err := models.Db.QueryRow("SELECT COUNT(*) FROM `recipient` WHERE `campaign_id`=? AND `status` IS NOT NULL AND `removed`=0", campaignId).Scan(&count)
	if err != nil {
		apilog.Print(err)
	}
	return count
}

// Success send
func reportSuccessSendCount(campaignId string) int {
	var count int
	err := models.Db.QueryRow("SELECT COUNT(*) FROM `recipient` WHERE `campaign_id`=? AND `status`='Ok' AND `removed`=0", campaignId).Scan(&count)
	if err != nil {
		apilog.Print(err)
	}
	return count
}

// Unsubscribe count
func reportUnsubscribeCount(campaignId string) int {
	var count int
	err := models.Db.QueryRow("SELECT COUNT(DISTINCT `email`) FROM `unsubscribe` WHERE `campaign_id`=?", campaignId).Scan(&count)
	if err != nil {
		apilog.Print(err)
	}
	return count
}

// Open mail count
func reportOpenMailCount(campaignId string) int {
	var count int
	err := models.Db.QueryRow("SELECT COUNT(*) FROM `recipient` WHERE `campaign_id`=? AND (`client_agent` IS NOT NULL OR `web_agent` IS NOT NULL) AND `removed`=0", campaignId).Scan(&count)
	if err != nil {
		apilog.Print(err)
	}
	return count
}

// Web version count
func reportOpenWebVersionCount(campaignId string) int {
	var count int
	err := models.Db.QueryRow("SELECT COUNT(DISTINCT `recipient`.`id`) FROM `jumping` INNER JOIN `recipient` ON `jumping`.`recipient_id`=`recipient`.`id` WHERE `jumping`.`campaign_id`=? AND `jumping`.`url`='web_version' AND `recipient`.`removed`=0", campaignId).Scan(&count)
	if err != nil {
		apilog.Print(err)
	}
	return count
}

func reportRecipientJumpCount(campaignId string) int {
	var count int
	err := models.Db.QueryRow("SELECT COUNT(DISTINCT `recipient`.`id`) FROM `jumping` INNER JOIN `recipient` ON `jumping`.`recipient_id`=`recipient`.`id` WHERE `recipient`.`removed`=0 AND `jumping`.`url` NOT IN ('open_trace', 'web_version', 'unsubscribe') AND `jumping`.`campaign_id`=?", campaignId).Scan(&count)
	if err != nil {
		apilog.Print(err)
	}
	return count
}

func reportJumpDetailedCount(campaignId string) map[string]int {
	var url string
	var count int
	res := make(map[string]int)
	query, err := models.Db.Query("SELECT `jumping`.`url` as link, ( SELECT COUNT(`jumping`.`id`) FROM `jumping` INNER JOIN `recipient` ON `jumping`.`recipient_id`=`recipient`.`id` WHERE `jumping`.`url`=link AND `jumping`.`campaign_id`=? AND `recipient`.`removed`=0 ) as cnt FROM `jumping` INNER JOIN `recipient` ON `jumping`.`recipient_id`=`recipient`.`id` WHERE `url` NOT IN ('open_trace', 'web_version', 'unsubscribe') AND `jumping`.`campaign_id`=? AND `recipient`.`removed`=0 GROUP BY `jumping`.`url`", campaignId, campaignId)
	if err != nil {
		apilog.Print(err)
		return res
	}
	defer query.Close()
	for query.Next() {
		err = query.Scan(&url, &count)
		if err != nil {
			apilog.Print(err)
		}
		res[url] = count
	}
	return res
}
