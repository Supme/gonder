package api

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"github.com/supme/gonder/models"
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
		apilog.Println(err)
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
		apilog.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	var query *sql.Rows
	query, err = models.Db.Query("SELECT `id`, `name`, `email`, `status`, IF(COALESCE(`web_agent`,`client_agent`) IS NULL, 0, 1) as `open` FROM `recipient`"+partWhere, partParams...)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer query.Close()

	w.Header().Set("Content-Disposition", "attachment; filename=recipients_"+strconv.FormatInt(campaign, 10)+".csv")
	w.Header().Set("Content-Type", "text/csv")

	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = ';'
	csvWriter.UseCRLF = true

	err = csvWriter.Write([]string{lang.tr(panelLocale, "Id"), lang.tr(panelLocale, "Email"), lang.tr(panelLocale, "Name"), lang.tr(panelLocale, "Opened"), lang.tr(panelLocale, "Result")})
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
			apilog.Println(err)
			return
		}
		if opened {
			openedStr = lang.tr(panelLocale, "Yes")
		} else {
			openedStr = lang.tr(panelLocale, "No")
		}
		err = csvWriter.Write([]string{id, email, name, openedStr, result.String})
		if err != nil {
			log.Println(err)
			return
		}
	}

	csvWriter.Flush()

	return
}
