package mailer

import (
	"encoding/base64"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func Stat(hostPort string) {

	//gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.LoadHTMLGlob("mailer/templates/*")

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to San Tropez!")
	})

	router.Static("/static", "../static")

	router.GET("/data/:params", func(c *gin.Context) {

		var param *pJson
		data, err := base64.URLEncoding.DecodeString(strings.Split(c.Param("params"), ".")[0])
		checkErr(err)

		err = json.Unmarshal([]byte(data), &param)
		checkErr(err)

		if param.Opened != "" {
			// blank 16x16 png
			c.Header("Content-Type", "image/png")
			output, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAMAAAAoLQ9TAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAADUExURUxpcU3H2DoAAAABdFJOUwBA5thmAAAADUlEQVQY02NgGAXIAAABEAAB7JfjegAAAABJRU5ErkJggg==iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAMAAABEpIrGAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAADUExURUxpcU3H2DoAAAABdFJOUwBA5thmAAAAEklEQVQ4y2NgGAWjYBSMAuwAAAQgAAFWu83mAAAAAElFTkSuQmCC")
			c.String(http.StatusOK, string(output))
			go statOpened(param.Campaign, param.Recipient)
		} else if param.Url != "" {
			// jump to url
			c.Redirect(http.StatusMovedPermanently, param.Url)
			go statJump(param.Campaign, param.Recipient, param.Url)
		} else if param.Webver != "" {
			// web version
			message := getWebMessage(param.Campaign, param.Recipient)
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

	router.GET("/code", func(c *gin.Context) {
		d := pJson{
			c.Query("c"),
			c.Query("r"),
			c.Query("u"),
			c.Query("w"),
			c.Query("o"),
			c.Query("s"),
		}
		j, err := json.Marshal(d)
		checkErr(err)
		data := base64.URLEncoding.EncodeToString(j)
		c.String(http.StatusOK, "encoded: %s", data)
	})

	// Listen and server on 0.0.0.0:8080
	router.Run(":" + hostPort)
}
