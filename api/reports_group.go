package api

import (
	"gonder/models"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type reportsGroup struct {
	w http.ResponseWriter
	r *http.Request
}

func reportsGroupHandlerFunc(w http.ResponseWriter, r *http.Request) {
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
	rc := reportsGroup{w: w, r: r}
	rFormat := strings.ToLower(r.FormValue("format"))
	rType := strings.ToLower(r.FormValue("type"))
	group := models.Group(id)
	switch rType {
	case "campaigns":
		if rFormat == "csv" {
			err = rc.campaignsCSV(&group)
			break
		}
		err = rc.campaignsJSON(&group)
	case "unsubscribed":
		if rFormat == "csv" {
			err = rc.unsubscribedCSV(&group)
			break
		}
		err = rc.unsubscribedJSON(&group)
	default:
		rc.w.WriteHeader(http.StatusNotImplemented)
	}
	if err != nil {
		log.Print(err)
		rc.w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (rc reportsGroup) campaignsJSON(c *models.Group) error {
	res := make([]models.GroupCampaign, 0, 64)
	q, err := c.Campaigns()
	if err != nil {
		return err
	}
	for q.Next() {
		var r models.GroupCampaign
		err = q.StructScan(&r)
		if err != nil {
			return err
		}
		res = append(res, r)
	}
	rc.w.Header().Set("Content-Type", "application/json")
	err = models.JSONResponse{}.OkWriter(rc.w, res)
	if err != nil {
		return err
	}
	return nil
}

func (rc reportsGroup) campaignsCSV(c *models.Group) error {
	q, err := c.Campaigns()
	if err != nil {
		return err
	}
	rc.w.Header().Set("Content-Disposition", "attachment; filename=group_"+c.StringID()+"_campaigns.csv")
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
		var r models.GroupCampaign
		err = q.StructScan(&r)
		if err != nil {
			return err
		}
		err = csvWriter.Write([]string{
			strconv.Itoa(r.ID),
			r.Name,
			r.Subject,
			r.Start,
			r.End,
		})
		if err != nil {
			return err
		}
	}
	csvWriter.Flush()
	return nil
}

func (rc reportsGroup) unsubscribedJSON(c *models.Group) error {
	res := make([]models.GroupUnsubscribed, 0, 64)
	q, err := c.Unsubscribed()
	if err != nil {
		return err
	}
	for q.Next() {
		var r models.GroupUnsubscribed
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

func (rc reportsGroup) unsubscribedCSV(c *models.Group) error {
	q, err := c.Unsubscribed()
	if err != nil {
		return err
	}
	rc.w.Header().Set("Content-Disposition", "attachment; filename=group_"+c.StringID()+"_unsubscribed.csv")
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
		var r models.GroupUnsubscribed
		err = q.StructScan(&r)
		if err != nil {
			return err
		}
		r.Validate()
		err = csvWriter.Write([]string{
			strconv.Itoa(r.CampaignID),
			r.Email,
			r.At,
			string(r.DataValid),
		})
		if err != nil {
			return err
		}
	}
	csvWriter.Flush()
	return nil
}
