package panel

import (
//	"encoding/base64"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Run() {

	//gin.SetMode(gin.ReleaseMode)
	gin.SetMode(gin.DebugMode)

	router := gin.Default()
	router.LoadHTMLGlob("panel/templates/*")
	router.Static("/static/", "static")
	router.Static("/assets/", "assets")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "main.html", gin.H{})
	})
/*
	router.GET("campaign/add", func(c *gin.Context) {

	})
*/


	mailer := router.Group("mailer")
	{
		mailer.GET("group", func(c *gin.Context) {
			data := gin.H{
				"groups": getGroups(),
			}
			c.HTML(http.StatusOK, "group.html", data)
		})

		mailer.GET("group/:id", func(c *gin.Context) {
			data := gin.H{
				"campaigns": getCampaigns(c.Param("id")),
			}
			c.HTML(http.StatusOK, "campaign.html", data)
		})
/*
		mailer.GET("campaign/edit/:id/{{.StatPng}}", func(c *gin.Context) {
			// blank 16x16 png
			c.Header("Content-Type", "image/png")
			output, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAMAAAAoLQ9TAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAADUExURUxpcU3H2DoAAAABdFJOUwBA5thmAAAADUlEQVQY02NgGAXIAAABEAAB7JfjegAAAABJRU5ErkJggg==iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAMAAABEpIrGAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAADUExURUxpcU3H2DoAAAABdFJOUwBA5thmAAAAEklEQVQ4y2NgGAWjYBSMAuwAAAQgAAFWu83mAAAAAElFTkSuQmCC")
			c.String(http.StatusOK, string(output))
		})
*/
		mailer.GET("campaign/edit/:id", func(c *gin.Context) {
			camp, err := getCampaignInfo(c.Param("id"))

			if err == nil {
				data := gin.H{
					"ifaces":   getIfaces(),
					"campaign": camp,
				}
				c.HTML(http.StatusOK, "campaignEdit.html", data)
			} else {
				c.String(http.StatusNotFound, "Campaign not found")
			}
		})

		mailer.POST("campaign/edit/:id", func(c *gin.Context) {
			data := gin.H{
				"ifaces": getIfaces(),
				"campaign": updateCampaignInfo(
					campaign{
						Id:        c.Param("id"),
						IfaceId:   c.PostForm("ifaceId"),
						Name:      c.PostForm("name"),
						Subject:   c.PostForm("subject"),
						From:      c.PostForm("from"),
						FromName:  c.PostForm("fromName"),
						Message:   c.PostForm("message"),
						StartTime: c.PostForm("startTime"),
						EndTime:   c.PostForm("endTime"),
					},
				),
			}

			c.HTML(http.StatusOK, "campaignEdit.html", data)
		})

		mailer.GET("campaign/recipient/get/:id", func(c *gin.Context) {
			data := gin.H{
				"campaignId": c.Param("id"),
				"recipients": getRecipients(c.Param("id"), "0", "10"),
			}
			c.HTML(http.StatusOK, "recipient.html", data)
		})

		mailer.GET("campaign/recipient/param/:id", func(c *gin.Context) {
			data := gin.H{
				"recipient": getRecipient(c.Param("id")),
				"params":    getRecipientParam(c.Param("id")),
			}
			c.HTML(http.StatusOK, "recipientParam.html", data)
		})

		mailer.GET("recipients/:id", hRecipients)
		mailer.POST("recipients/:id", uploadRecipients)

	}

	router.GET("filemanager", func(c *gin.Context) {
		if c.Query("mode") == "download" {
			var n string
			var d []byte

			if c.Query("width") != "" && c.Query("height") != "" {
				// Download resized
				d = FilemanagerResize(c.Query("path"), c.Query("width"), c.Query("height"))
			} else {
				// Download file
				n, d = FilemanagerDownload(c.Query("path"))
				c.Header("Content-Disposition", "attachment; filename='" + n + "'")
			}
			c.Data(http.StatusOK,http.DetectContentType(d),d)

		} else {
			c.JSON(http.StatusOK, Filemanager(c.Query("mode"), c.Query("path"), c.Query("name"), c.Query("old"), c.Query("new")))
		}
	})

	router.POST("filemanager", func(c *gin.Context){
		if c.PostForm("mode") == "add" {
			file, head, err := c.Request.FormFile("newfile")
			checkErr(err)
			c.JSON(http.StatusOK, FilemanagerAdd(c.PostForm("currentpath"), head.Filename, file))
		}
	})

	api := router.Group("api")
	{
		table := api.Group("mailer")
		{
			table.GET("recipients/:id", aRecipients)
		}
	}

	router.Run(":7777")
}
