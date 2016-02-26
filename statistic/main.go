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
package statistic

import (
	"github.com/supme/gonder/models"
	"encoding/base64"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"log"
	"strconv"
	"runtime"
	"os"
)

var (
	Port string
)

func Run() {

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		mem := new(runtime.MemStats)
		runtime.ReadMemStats(mem)
		c.String(http.StatusOK, "Welcome to San Tropez! (Conn: "  + strconv.Itoa(models.Db.Stats().OpenConnections) + " Allocate: " + strconv.FormatUint(mem.Alloc, 10) + ")")
	})

	router.Static("/files/", "files")

	router.GET("/data/:params", func(c *gin.Context) {

		var param *models.Json
		data, err := base64.URLEncoding.DecodeString(strings.Split(c.Param("params"), ".")[0])
		checkErr(err)

		err = json.Unmarshal([]byte(data), &param)
		checkErr(err)

		userAgent := c.ClientIP() + " " + c.Request.UserAgent()

		if param.Opened != "" {
			go statOpened(param.Campaign, param.Recipient, userAgent)
			// blank 16x16 png
			c.Header("Content-Type", "image/png")
			output, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAMAAAAoLQ9TAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAADUExURUxpcU3H2DoAAAABdFJOUwBA5thmAAAADUlEQVQY02NgGAXIAAABEAAB7JfjegAAAABJRU5ErkJggg==iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAMAAABEpIrGAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAADUExURUxpcU3H2DoAAAABdFJOUwBA5thmAAAAEklEQVQ4y2NgGAWjYBSMAuwAAAQgAAFWu83mAAAAAElFTkSuQmCC")
			c.String(http.StatusOK, string(output))
		} else if param.Url != "" {
			go statJump(param.Campaign, param.Recipient, param.Url, userAgent)
			// jump to url
			c.Redirect(http.StatusMovedPermanently, param.Url)
		} else if param.Webver != "" {
			go statWebVersion(param.Campaign, param.Recipient, userAgent)
			// web version
			message := models.WebMessage(param.Campaign, param.Recipient)
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(message))
		} else if param.Unsubscribe != "" {
			// unsubscribe ToDo CSRF
			data := gin.H{
				"campaignId":  param.Campaign,
				"recipientId": param.Recipient,
			}
			router.LoadHTMLFiles("statistic/templates/" + getTemplateName(param.Campaign) + "/accept.html")
			c.HTML(http.StatusOK, "accept.html", data)
		} else {
			c.String(http.StatusNotFound, "Not found")
		}
	})

	router.POST("/unsubscribe", func(c *gin.Context) {
		postUnsubscribe(c.PostForm("campaignId"), c.PostForm("recipientId"))
		router.LoadHTMLFiles("statistic/templates/" + getTemplateName(c.PostForm("campaignId")) + "/success.html")
		c.HTML(http.StatusOK, "success.html", gin.H{})
	})

	// Listen and server on 0.0.0.0:8080
	router.Run(":" + Port)
}

func getTemplateName(campaignId string) (name string) {
	models.Db.QueryRow("SELECT `group`.`template` FROM `campaign` INNER JOIN `group` ON `campaign`.`group_id`=`group`.`id` WHERE `group`.`template` IS NOT NULL AND `campaign`.`id`=?", campaignId).Scan(&name)
	if name == "" {
		name = "default"
	} else {
		if _, err := os.Stat("statistic/templates/" + name + "/accept.html"); err != nil {
			name = "default"
		}
		if _, err := os.Stat("statistic/templates/" + name + "/success.html"); err != nil {
			name = "default"
		}
	}
	return
}

func statOpened(campaignId string, recipientId string, userAgent string) {
	models.Db.Exec("INSERT INTO jumping (campaign_id, recipient_id, url) VALUES (?, ?, ?)", campaignId, recipientId, "open_trace")
	models.Db.Exec("UPDATE `recipient` SET `client_agent`= ? WHERE `id`=? AND `client_agent` IS NULL", userAgent, recipientId)
}

func statJump(campaignId string, recipientId string, url string, userAgent string) {
	models.Db.Exec("INSERT INTO jumping (campaign_id, recipient_id, url) VALUES (?, ?, ?)", campaignId, recipientId, url)
	models.Db.Exec("UPDATE `recipient` SET `web_agent`= ? WHERE `id`=? AND `web_agent` IS NULL", userAgent, recipientId)
}

func statWebVersion(campaignId string, recipientId string, userAgent string)  {
	models.Db.Exec("INSERT INTO jumping (campaign_id, recipient_id, url) VALUES (?, ?, ?)", campaignId, recipientId, "web_version")
	models.Db.Exec("UPDATE `recipient` SET `web_agent`= ? WHERE `id`=? AND `web_agent` IS NULL", userAgent, recipientId)
}

func postUnsubscribe(campaignId string, recipientId string) {
	models.Db.Exec("INSERT INTO unsubscribe (`group_id`, campaign_id, `email`) VALUE ((SELECT group_id FROM campaign WHERE id=?), ?, (SELECT email FROM recipient WHERE id=?))", campaignId, campaignId, recipientId)
}

func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}