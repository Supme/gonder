package api

import (
	campSender "github.com/supme/gonder/campaign"
	"github.com/supme/gonder/models"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"text/template"
)

func getMailPreview(w http.ResponseWriter, r *http.Request) {
	if user.Right("get-recipients") {
		recipient, err := campSender.GetRecipient(r.FormValue("id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		if user.CampaignRight(recipient.CampaignID) {
			var tmplFunc func(io.Writer) error
			if r.FormValue("type") != "web" {
				tmplFunc = recipient.WebHTML(false, true)
			} else {
				tmplFunc = recipient.WebHTML(true, true)
			}
			w.Header().Set("Content-Type", "text/html")
			err = tmplFunc(w)
			if err != nil {
				apilog.Println(err)
			}

		} else {
			http.Error(w, "Forbidden get recipients from this campaign", http.StatusForbidden)
		}
	} else {
		http.Error(w, "Forbidden get recipients", http.StatusForbidden)
	}
}

// ToDo fix get right template then move unsubscribe template from models
func getUnsubscribePreview(w http.ResponseWriter, r *http.Request) {
	if user.Right("get-recipients") && user.CampaignRight(r.FormValue("campaignId")) {

		var tmpl string
		var content []byte
		var err error
		models.Db.QueryRow("SELECT `group`.`template` FROM `campaign` INNER JOIN `group` ON `campaign`.`group_id`=`group`.`id` WHERE `group`.`template` IS NOT NULL AND `campaign`.`id`=?", r.FormValue("id")).Scan(&tmpl)
		if tmpl == "" {
			tmpl = "default"
		} else {
			if _, err = os.Stat(models.FromRootDir("templates/" + tmpl + "/accept.html")); err != nil {
				tmpl = "default"
			}
			if _, err = os.Stat(models.FromRootDir("templates/" + tmpl + "/success.html")); err != nil {
				tmpl = "default"
			}
		}

		if r.Method == "GET" {
			content, _ = ioutil.ReadFile(models.FromRootDir("templates/" + tmpl + "/accept.html"))
		} else {
			content, _ = ioutil.ReadFile(models.FromRootDir("templates/" + tmpl + "/success.html"))
		}

		t := template.New("unsubscribe")
		t, err = t.Parse(string(content))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		err = t.Execute(w, map[string]string{"campaignId": r.FormValue("campaignId")})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	} else {
		http.Error(w, "Forbidden get from this campaign", http.StatusForbidden)
	}

}
