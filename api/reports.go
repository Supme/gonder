package api

import (
	"net/http"
	"encoding/json"
	"github.com/supme/gonder/models"
)

type Report struct {
	Name    string `json:"name"`
	Value	interface{} `json:"value"`
}

func report(w http.ResponseWriter, r *http.Request)  {

	var reports []Report
	var err error
	var js []byte
	if err = r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	campaign := r.Form["campaign"][0]
	if auth.CampaignRight(campaign) {
		reports = append(reports, Report{Name:"OpenWebVersionCount", Value:reportOpenWebVersionCount(campaign)})
		//ToDo
		reports = append(reports, Report{Name:"Other statistic", Value:"Comming soon"})
		js, err = json.Marshal(reports)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	} else {
		js = []byte(`{"status": "error", "message": "Forbidden add groups"}`)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}


func reportOpenWebVersionCount(id string) int {
	var count int
	models.Db.QueryRow("SELECT COUNT(*) FROM `jumping` INNER JOIN recipient ON `jumping`.`recipient_id`=`recipient`.`id` WHERE `jumping`.`campaign_id`=? AND `jumping`.`url`='web_version' AND `recipient`.`removed`!=1", id).Scan(&count)
	return count
}