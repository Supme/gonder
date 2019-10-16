package api

import (
	"database/sql"
	"encoding/json"
	"gonder/models"
	"log"
	"net/http"
	"strconv"
)

func reportRecipientsCsv(w http.ResponseWriter, r *http.Request) {
	var (
		req                request
		partWhere, where   string
		partParams, params []interface{}
	)

	campaign, err := strconv.ParseInt(r.FormValue("campaign"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	user := r.Context().Value("Auth").(*Auth)
	if !user.Right("get-recipients") && !user.CampaignRight(campaign) {
		log.Println(err)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err = json.Unmarshal([]byte(r.FormValue("params")), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	params = append(params, campaign)

	where = " WHERE `removed`=0 AND `campaign_id`=?"
	partWhere, partParams, err = createSQLPart(req, where, params, map[string]string{
		"recid": "id", "name": "name", "email": "email", "result": "status", "open": "open",
	}, true)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	var query *sql.Rows
	query, err = models.Db.Query("SELECT `id`, `name`, `email`, `status`, IF(COALESCE(`web_agent`,`client_agent`) IS NULL, 0, 1) as `open` FROM `recipient`"+partWhere, partParams...)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer func() {
		if err := query.Close(); err != nil {
			log.Print(err)
		}
	}()

	w.Header().Set("Content-Disposition", "attachment; filename=recipients_"+strconv.FormatInt(campaign, 10)+".csv")
	w.Header().Set("Content-Type", "text/csv")

	csvWriter := models.NewCSVWriter(w)

	err = csvWriter.Write([]string{lang.tr(models.Config.APIPanelLocale, "Id"), lang.tr(models.Config.APIPanelLocale, "Email"), lang.tr(models.Config.APIPanelLocale, "Name"), lang.tr(models.Config.APIPanelLocale, "Opened"), lang.tr(models.Config.APIPanelLocale, "Result")})
	if err != nil {
		log.Println(err)
		return
	}

	for query.Next() {
		var (
			id, email, name, openedStr string
			result                     sql.NullString
			opened                     bool
		)
		err = query.Scan(&id, &name, &email, &result, &opened)
		if err != nil {
			log.Println(err)
			return
		}
		if opened {
			openedStr = lang.tr(models.Config.APIPanelLocale, "Yes")
		} else {
			openedStr = lang.tr(models.Config.APIPanelLocale, "No")
		}
		err = csvWriter.Write([]string{id, email, name, openedStr, result.String})
		if err != nil {
			log.Println(err)
			return
		}
	}

	csvWriter.Flush()
}
