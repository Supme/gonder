package statistic

import (
	"github.com/supme/gonder/models"
	"encoding/base64"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"net"
	"log"
	"strconv"
)

var (
	Port string
)
func Run() {

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.LoadHTMLGlob("statistic/templates/*")

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to San Tropez! ("  + strconv.Itoa(models.Db.Stats().OpenConnections) + ")")
	})

	router.Static("/static/", "static")

	router.GET("/data/:params", func(c *gin.Context) {

		var param *models.Json
		data, err := base64.URLEncoding.DecodeString(strings.Split(c.Param("params"), ".")[0])
		checkErr(err)

		err = json.Unmarshal([]byte(data), &param)
		checkErr(err)

		rIP, _, err := net.SplitHostPort(c.ClientIP())
		if err != nil {
			rIP = "bad IP address"
		}
		userAgent := rIP + " " + c.Request.UserAgent()

		if param.Opened != "" {
			//ToDo Записывать параметры клиента (клиент, браузер, ip и т.д.)
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
			c.HTML(http.StatusOK, "unsubscribe.html", data)
		} else {
			c.String(http.StatusNotFound, "Not found")
		}
	})

	router.POST("/unsubscribe", func(c *gin.Context) {
		postUnsubscribe(c.PostForm("campaignId"), c.PostForm("recipientId"))
		c.HTML(http.StatusOK, "unsubscribeOk.html", gin.H{})
	})

	// Listen and server on 0.0.0.0:8080
	router.Run(":" + Port)
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