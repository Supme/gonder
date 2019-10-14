package api

import (
	"encoding/csv"
	"encoding/json"
	"gonder/models"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type reportsCampaign struct {
	w http.ResponseWriter
	r *http.Request
}

func reportsCampaignHandlerFunc(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user := r.Context().Value("Auth").(*Auth)
	if !user.CampaignRight(id) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	rc := reportsCampaign{w: w, r: r}
	rFormat := strings.ToLower(r.FormValue("format"))
	rType := strings.ToLower(r.FormValue("type"))
	campaign := models.Campaign(id)
	switch rType {
	case "question":
		if rFormat == "csv" {
			err = rc.questionCSV(&campaign)
			break
		}
		err = rc.questionJSON(&campaign)
	case "unsubscribed":
		if rFormat == "csv" {
			err = rc.unsubscribedCSV(&campaign)
			break
		}
		err = rc.unsubscribedJSON(&campaign)
	default:
		rc.w.WriteHeader(http.StatusNotImplemented)
	}
	if err != nil {
		log.Print(err)
		rc.w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (rc reportsCampaign) questionJSON(c *models.Campaign) error {
	res := make([]models.CampaignQuestion, 0, 64)
	q, err := c.Question()
	if err != nil {
		return err
	}
	for q.Next() {
		var r models.CampaignQuestion
		err = q.StructScan(&r)
		if err != nil {
			return err
		}
		r.Validate()
		res = append(res, r)
	}
	rc.w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(rc.w).Encode(res)
	if err != nil {
		return err
	}
	return nil
}

func (rc reportsCampaign) questionCSV(c *models.Campaign) error {
	q, err := c.Question()
	if err != nil {
		return err
	}
	rc.w.Header().Set("Content-Disposition", "attachment; filename=campaign_"+c.StringID()+"_question.csv")
	rc.w.Header().Set("Content-Type", "text/csv")
	csvWriter := csv.NewWriter(rc.w)
	csvWriter.Comma = ';'
	csvWriter.UseCRLF = true
	columns, err := q.Columns()
	if err != nil {
		return err
	}
	err = csvWriter.Write(columns)
	if err != nil {
		return err
	}
	for q.Next() {
		var r models.CampaignQuestion
		err = q.StructScan(&r)
		if err != nil {
			return err
		}
		r.Validate()
		err = csvWriter.Write([]string{
			strconv.Itoa(r.ID),
			r.Email,
			r.At,
			r.DataValid,
		})
	}
	csvWriter.Flush()
	return nil
}

func (rc reportsCampaign) unsubscribedJSON(c *models.Campaign) error {
	res := make([]models.CampaignUnsubscribed, 0, 64)
	q, err := c.Unsubscribed()
	if err != nil {
		return err
	}
	for q.Next() {
		var r models.CampaignUnsubscribed
		err = q.StructScan(&r)
		if err != nil {
			return err
		}
		r.Validate()
		res = append(res, r)
	}
	rc.w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(rc.w).Encode(res)
	if err != nil {
		return err
	}
	return nil
}

func (rc reportsCampaign) unsubscribedCSV(c *models.Campaign) error {
	q, err := c.Unsubscribed()
	if err != nil {
		return err
	}
	rc.w.Header().Set("Content-Disposition", "attachment; filename=campaign_"+c.StringID()+"_unsubscribed.csv")
	rc.w.Header().Set("Content-Type", "text/csv")
	csvWriter := csv.NewWriter(rc.w)
	csvWriter.Comma = ';'
	csvWriter.UseCRLF = true
	columns, err := q.Columns()
	if err != nil {
		return err
	}
	err = csvWriter.Write(columns)
	if err != nil {
		return err
	}
	for q.Next() {
		var r models.CampaignUnsubscribed
		err = q.StructScan(&r)
		if err != nil {
			return err
		}
		r.Validate()
		err = csvWriter.Write([]string{
			r.Email,
			r.At,
			r.DataValid,
		})
	}
	csvWriter.Flush()
	return nil
}
