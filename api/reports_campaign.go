package api

import (
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
	case "recipients":
		if rFormat == "csv" {
			err = rc.recipientsCSV(&campaign)
			break
		}
		err = rc.recipientsJSON(&campaign)
	case "clicks":
		if rFormat == "csv" {
			err = rc.clicksCSV(&campaign)
			break
		}
		err = rc.clicksJSON(&campaign)
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
	res := make([]models.CampaignReportQuestion, 0, 64)
	q, err := c.ReportQuestion()
	if err != nil {
		return err
	}
	for q.Next() {
		var r models.CampaignReportQuestion
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
	q, err := c.ReportQuestion()
	if err != nil {
		return err
	}
	rc.w.Header().Set("Content-Disposition", "attachment; filename=campaign_"+c.StringID()+"_question.csv")
	rc.w.Header().Set("Content-Type", "text/csv")
	csvWriter := models.NewCSVWriter(rc.w)
	columns, err := q.Columns()
	if err != nil {
		return err
	}
	err = csvWriter.Write(columns)
	if err != nil {
		return err
	}
	for q.Next() {
		var r models.CampaignReportQuestion
		err = q.StructScan(&r)
		if err != nil {
			return err
		}
		r.Validate()
		err = csvWriter.Write([]string{
			strconv.Itoa(r.ID),
			r.Email,
			r.At,
			string(r.DataValid),
		})
	}
	csvWriter.Flush()
	return nil
}

func (rc reportsCampaign) unsubscribedJSON(c *models.Campaign) error {
	res := make([]models.CampaignReportUnsubscribed, 0, 64)
	q, err := c.ReportUnsubscribed()
	if err != nil {
		return err
	}
	for q.Next() {
		var r models.CampaignReportUnsubscribed
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
	q, err := c.ReportUnsubscribed()
	if err != nil {
		return err
	}
	rc.w.Header().Set("Content-Disposition", "attachment; filename=campaign_"+c.StringID()+"_unsubscribed.csv")
	rc.w.Header().Set("Content-Type", "text/csv")
	csvWriter := models.NewCSVWriter(rc.w)
	columns, err := q.Columns()
	if err != nil {
		return err
	}
	err = csvWriter.Write(columns)
	if err != nil {
		return err
	}
	for q.Next() {
		var r models.CampaignReportUnsubscribed
		err = q.StructScan(&r)
		if err != nil {
			return err
		}
		r.Validate()
		err = csvWriter.Write([]string{
			r.Email,
			r.At,
			string(r.DataValid),
		})
	}
	csvWriter.Flush()
	return nil
}

func (rc reportsCampaign) recipientsJSON(c *models.Campaign) error {
	res := make([]models.CampaignReportRecipients, 0, 64)
	q, err := c.ReportRecipients()
	if err != nil {
		return err
	}
	for q.Next() {
		var r models.CampaignReportRecipients
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

func (rc reportsCampaign) recipientsCSV(c *models.Campaign) error {
	q, err := c.ReportRecipients()
	if err != nil {
		return err
	}
	rc.w.Header().Set("Content-Disposition", "attachment; filename=campaign_"+c.StringID()+"_recipients.csv")
	rc.w.Header().Set("Content-Type", "text/csv")
	csvWriter := models.NewCSVWriter(rc.w)
	columns, err := q.Columns()
	if err != nil {
		return err
	}
	err = csvWriter.Write(columns)
	if err != nil {
		return err
	}
	for q.Next() {
		var r models.CampaignReportRecipients
		err = q.StructScan(&r)
		if err != nil {
			return err
		}
		r.Validate()
		err = csvWriter.Write([]string{
			strconv.Itoa(r.ID),
			r.Email,
			r.Name,
			r.At,
			r.StatusValid,
			strconv.FormatBool(r.Open),
			string(r.DataValid),
		})
	}
	csvWriter.Flush()
	return nil
}

func (rc reportsCampaign) clicksJSON(c *models.Campaign) error {
	res := make([]models.CampaignReportClicks, 0, 64)
	q, err := c.ReportClicks()
	if err != nil {
		return err
	}
	for q.Next() {
		var r models.CampaignReportClicks
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

func (rc reportsCampaign) clicksCSV(c *models.Campaign) error {
	q, err := c.ReportClicks()
	if err != nil {
		return err
	}
	rc.w.Header().Set("Content-Disposition", "attachment; filename=campaign_"+c.StringID()+"_clicks.csv")
	rc.w.Header().Set("Content-Type", "text/csv")
	csvWriter := models.NewCSVWriter(rc.w)
	columns, err := q.Columns()
	if err != nil {
		return err
	}
	err = csvWriter.Write(columns)
	if err != nil {
		return err
	}
	for q.Next() {
		var clx models.CampaignReportClicks
		err = q.StructScan(&clx)
		if err != nil {
			return err
		}
		clx.Validate()
		err = csvWriter.Write([]string{
			strconv.Itoa(clx.ID),
			clx.Email,
			clx.At,
			clx.URL,
		})
	}
	csvWriter.Flush()
	return nil
}
