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
	"net/http"
	"github.com/supme/gonder/models"
	"text/template"
	"os"
	"io/ioutil"
	"strconv"
)

func getMailPreview(w http.ResponseWriter, r *http.Request) {
	if auth.Right("get-recipients") {
		cId, err := getRecipientCampaign(r.FormValue("id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		if auth.CampaignRight(cId) {
			w.Header().Set("Content-Type", "text/html")
			var paramKey, paramValue string
			params := make(map[string]string)
			param, err := models.Db.Query("SELECT `key`, `value` FROM parameter WHERE recipient_id=?", r.FormValue("id"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			defer param.Close()
			for param.Next() {
				err = param.Scan(&paramKey, &paramValue)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				params[string(paramKey)] = string(paramValue)
			}

			var email, name string
			err = models.Db.QueryRow("SELECT `email`, `name` FROM `recipient` WHERE `id`=?", r.FormValue("id")).Scan(&email, &name)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			params["RecipientEmail"] = email
			params["RecipientName"] = name
			params["CampaignId"] = strconv.FormatInt(cId, 10)

			if r.FormValue("type") != "web" {
				params["WebUrl"] = "/preview?id=" + r.FormValue("id") + "&type=web"
			}
			params["UnsubscribeUrl"] = "/unsubscribe?campaignId=" + strconv.FormatInt(cId, 10)
			//data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAMAAAAoLQ9TAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAADUExURUxpcU3H2DoAAAABdFJOUwBA5thmAAAADUlEQVQY02NgGAXIAAABEAAB7JfjegAAAABJRU5ErkJggg==iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAMAAABEpIrGAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAADUExURUxpcU3H2DoAAAABdFJOUwBA5thmAAAAEklEQVQ4y2NgGAWjYBSMAuwAAAQgAAFWu83mAAAAAElFTkSuQmCC
			params["StatPng"] = `<img src="" border="0px" width="10px" height="10px"/>`

			tmpl := ""
			err = models.Db.QueryRow("SELECT `body` FROM campaign WHERE id=?", cId).Scan(&tmpl)
			t := template.New("preview")
			t, err = t.Parse(tmpl)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			err = t.Execute(w, params)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			http.Error(w, "Forbidden get recipients from this campaign", http.StatusForbidden)
		}
	} else {
		http.Error(w, "Forbidden get recipients", http.StatusForbidden)
	}
}

func getUnsubscribePreview(w http.ResponseWriter, r *http.Request)  {
	if auth.Right("get-recipients") && auth.CampaignRight(r.FormValue("campaignId")) {

		var tmpl string
		var content []byte
		var err error
		models.Db.QueryRow("SELECT `group`.`template` FROM `campaign` INNER JOIN `group` ON `campaign`.`group_id`=`group`.`id` WHERE `group`.`template` IS NOT NULL AND `campaign`.`id`=?", r.FormValue("id")).Scan(&tmpl)
		if tmpl == "" {
			tmpl = "default"
		} else {
			if _, err = os.Stat(models.FromRootDir("statistic/templates/" + tmpl + "/accept.html")); err != nil {
				tmpl = "default"
			}
			if _, err = os.Stat(models.FromRootDir("statistic/templates/" + tmpl + "/success.html")); err != nil {
				tmpl = "default"
			}
		}

		if r.Method == "GET" {
			content, _ = ioutil.ReadFile(models.FromRootDir("statistic/templates/" + tmpl + "/accept.html"))
		} else {
			content, _ = ioutil.ReadFile(models.FromRootDir("statistic/templates/" + tmpl + "/success.html"))
		}

		t := template.New("unsubscribe")
		t, err = t.Parse(string(content))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		err = t.Execute(w, map[string] string {"campaignId": r.FormValue("campaignId")})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	} else {
		http.Error(w, "Forbidden get from this campaign", http.StatusForbidden)
	}

}