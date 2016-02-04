package panel

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"time"
	"io"
	"encoding/csv"
	"github.com/supme/gonder/models"
)

type tRecipients struct {
	Draw string
	Start string
	Length string
	CampaignId string
}

func hRecipients(c *gin.Context) {
	c.HTML(http.StatusOK, "recipients.html", gin.H{
		"campaignId": c.Param("id"),
	})
}

func aRecipients(c *gin.Context) {
	t := tRecipients{}
	t.Draw = c.Query("_")
	t.CampaignId = c.Param("id")
	t.Start = c.DefaultQuery("start", "0")
	t.Length = c.DefaultQuery("length", "10")

	total := t.RecordsTotal()
	c.JSON(http.StatusOK, gin.H{
		"draw": t.Draw,
		"recordsTotal": total,
		"recordsFiltered": total,
		"data": t.Recipients(),
	})
}

func (t *tRecipients) Recipients() [][]string {
	var id, email, name string
	var r [][]string

	query, err := models.Db.Query("SELECT `id`, `email`, `name` FROM `recipient` WHERE `campaign_id`=? ORDER BY `id` LIMIT ?,?", t.CampaignId, t.Start, t.Length)
	checkErr(err)
	defer query.Close()

	for query.Next() {
		err = query.Scan(&id, &email, &name)
		checkErr(err)
		r = append(r,[]string{ 0:id, 1:email, 2:name })
	}

	if r == nil {
		return [][]string{}
	}

	return r
}

func (t *tRecipients) RecordsTotal() string {
	var count string

	err := models.Db.QueryRow("SELECT COUNT(*) FROM `recipient` WHERE `campaign_id`=?", t.CampaignId).Scan(&count)
	checkErr(err)

	return count
}

func uploadRecipients(c *gin.Context) {
	if c.PostForm("delete") != "" {
		_, err := models.Db.Exec("DELETE FROM `recipient` WHERE `campaign_id`=?", c.Param("id"))
		checkErr(err)
	}

	if c.PostForm("submit") != "" {
		file, head, err := c.Request.FormFile("file")
		checkErr(err)
		if err == nil {
			fpath := "tmp/" + string(time.Now().UnixNano()) + head.Filename
			out, err := os.Create(fpath)
			checkErr(err)
			defer out.Close()
			_, err = io.Copy(out, file)
			checkErr(err)
			postRecipientCsv(c.Param("id"), fpath)
			os.Remove(fpath)
		}

	}

	hRecipients(c)
}

// ToDo optimize this
func postRecipientCsv(campaignId string, file string) error {
	title := make(map[int]string)
	data := make(map[string]string)

	csvfile, err := os.Open(file)
	checkErr(err)
	defer csvfile.Close()

	reader := csv.NewReader(csvfile)
	reader.FieldsPerRecord = -1
	rawCSVdata, err := reader.ReadAll()
	checkErr(err)

	for k, v := range rawCSVdata {
		if k == 0 {
			for i, t := range v {
				title[i] = t
			}
		} else {
			email := ""
			name := ""
			for i, t := range v {
				if i == 0 {
					email = t
				} else if i == 1 {
					name = t
				} else {
					data[title[i]] = t
				}
			}

			res, err := models.Db.Exec("INSERT INTO recipient (`campaign_id`, `email`, `name`) VALUES (?, ?, ?)", campaignId, email, name)
			checkErr(err)

			id, err := res.LastInsertId()
			checkErr(err)

			for i, t := range data {
				_, err := models.Db.Exec("INSERT INTO parameter (`recipient_id`, `key`, `value`) VALUES (?, ?, ?)", id, i, t)
				checkErr(err)
			}
		}
	}

	return err
}
