package api

import (
	campSender "gonder/campaign"
	"gonder/models"
	"html/template"
	"io"
	"log"
	"net/http"
)

func getMailPreview(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ContextAuth).(*Auth)
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
				log.Println(err)
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
	user := r.Context().Value(ContextAuth).(*Auth)
	err := r.ParseForm()
	if err != nil {
		apiLog.Println(err)
		http.Error(w, "Not valid request", http.StatusInternalServerError)
		return
	}

	var message models.Message
	err = message.New(r.Form.Get("recipientId"))
	if err != nil {
		apiLog.Println(err)
		http.Error(w, "Not valid request", http.StatusInternalServerError)
		return
	}

	if user.Right("get-recipients") && user.CampaignRight(message.CampaignID) {
		var t *template.Template
		if r.Method == "GET" {
			t, err = message.GetTemplate("accept.html")
			if err != nil {
				apiLog.Println(err)
				http.Error(w, "Not valid request", http.StatusInternalServerError)
				return
			}
		} else {
			t, err = message.GetTemplate("success.html")
			if err != nil {
				apiLog.Println(err)
				http.Error(w, "Not valid request", http.StatusInternalServerError)
				return
			}
		}

		err = t.Execute(w, map[string]string{
			"CampaignId":     message.CampaignID,
			"RecipientId":    message.RecipientID,
			"RecipientEmail": message.RecipientEmail,
			"RecipientName":  message.RecipientName,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	} else {
		http.Error(w, "Forbidden get from this campaign", http.StatusForbidden)
	}

}
