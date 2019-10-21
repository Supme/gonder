package api

import (
	"encoding/json"
	"fmt"
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
	case "useragent":
		if rFormat == "csv" {
			err = rc.userAgentCSV(&campaign)
			break
		}
		err = rc.userAgentJSON(&campaign)
	default:
		rc.w.WriteHeader(http.StatusNotImplemented)
		rc.w.Header().Set("Content-Type", "application/json")
		models.JSONResponse{}.ErrorWriter(rc.w, fmt.Errorf("this type not implemented"))
	}
	if err != nil {
		log.Print(err)
		rc.w.WriteHeader(http.StatusInternalServerError)
		rc.w.Header().Set("Content-Type", "application/json")
		models.JSONResponse{}.ErrorWriter(rc.w, fmt.Errorf("server error"))
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
	err = models.JSONResponse{}.OkWriter(rc.w, res)
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
	err = models.JSONResponse{}.OkWriter(rc.w, res)
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
	err = models.JSONResponse{}.OkWriter(rc.w, res)
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

func (rc reportsCampaign) userAgentJSON(c *models.Campaign) error {
	res := make([]models.CampaignReportUserAgent, 0, 64)
	q, err := c.ReportUserAgent()
	if err != nil {
		return err
	}
	for q.Next() {
		var r models.CampaignReportUserAgent
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

func (rc reportsCampaign) userAgentCSV(c *models.Campaign) error {

	q, err := c.ReportUserAgent()
	if err != nil {
		return err
	}
	rc.w.Header().Set("Content-Disposition", "attachment; filename=campaign_"+c.StringID()+"_clicks.csv")
	rc.w.Header().Set("Content-Type", "text/csv")
	csvWriter := models.NewCSVWriter(rc.w)
	err = csvWriter.Write([]string{
		"id", "email", "name",
		"client ip", "client is mobile", "client is bot", "client platform", "client os", "client engine name", "client engine version", "client browser name", "client browser version",
		"browser ip", "browser is mobile", "browser is bot", "browser platform", "browser os", "browser engine name", "browser engine version", "browser name", "browser version",
	})
	if err != nil {
		return err
	}
	for q.Next() {
		var ua models.CampaignReportUserAgent
		err = q.StructScan(&ua)
		if err != nil {
			return err
		}
		ua.Validate()
		csvRow := make([]string, 21)
		csvRow[0] = strconv.Itoa(ua.ID)
		csvRow[1] = ua.Email
		csvRow[2] = ua.Name
		if ua.ClientParsed != nil {
			csvRow[3] = ua.ClientParsed.IP
			csvRow[4] = strconv.FormatBool(ua.ClientParsed.IsMobile)
			csvRow[5] = strconv.FormatBool(ua.ClientParsed.IsBot)
			csvRow[6] = ua.ClientParsed.Platform
			csvRow[7] = ua.ClientParsed.OS
			csvRow[8] = ua.ClientParsed.EngineName
			csvRow[9] = ua.ClientParsed.EngineVersion
			csvRow[10] = ua.ClientParsed.BrowserName
			csvRow[11] = ua.ClientParsed.BrowserVersion
		}
		if ua.BrowserParsed != nil {
			csvRow[12] = ua.BrowserParsed.IP
			csvRow[13] = strconv.FormatBool(ua.BrowserParsed.IsMobile)
			csvRow[14] = strconv.FormatBool(ua.BrowserParsed.IsBot)
			csvRow[15] = ua.BrowserParsed.Platform
			csvRow[16] = ua.BrowserParsed.OS
			csvRow[17] = ua.BrowserParsed.EngineName
			csvRow[18] = ua.BrowserParsed.EngineVersion
			csvRow[19] = ua.BrowserParsed.BrowserName
			csvRow[20] = ua.BrowserParsed.BrowserVersion
		}
		err = csvWriter.Write(csvRow)
	}
	csvWriter.Flush()
	return nil
}
