package panel

import (
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"net/http"
	"github.com/supme/gonder/roxyFileman"
)

func Run() {

	//gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.LoadHTMLGlob("panel/templates/*")
	router.Static("/static/", "static")
	router.Static("/assets/", "assets")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "main.html", gin.H{})
	})



	router.GET("campaign/add/:id", func(c *gin.Context) {

	})

	router.GET("/campaign/recipient/:id", func(c *gin.Context) {
		data := gin.H{
			"campaignId": c.Param("id"),
			"recipients": getRecipients(c.Param("id"), "0", "10"),
		}
		c.HTML(http.StatusOK, "recipient.html", data)
	})

	router.GET("/campaign/recipient/param/:id", func(c *gin.Context) {
		data := gin.H{
			"recipient": getRecipient(c.Param("id")),
			"params":    getRecipientParam(c.Param("id")),
		}
		c.HTML(http.StatusOK, "recipientParam.html", data)
	})

	roxyFileman.FileMan(router.Group("fileman"))

	mailer := router.Group("mailer")
	{
		mailer.GET("group", func(c *gin.Context) {
			data := gin.H{
				"groups": getGroups(),
			}
			c.HTML(http.StatusOK, "group.html", data)
		})

		mailer.GET("campaign/:id", func(c *gin.Context) {
			data := gin.H{
				"campaigns": getCampaigns(c.Param("id")),
			}
			c.HTML(http.StatusOK, "campaign.html", data)
		})

		mailer.GET("campaign/edit/:id", func(c *gin.Context) {
			camp, err := getCampaignInfo(c.Param("id"))
			if err != nil {
				// blank 16x16 png
				c.Header("Content-Type", "image/png")
				output, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAMAAAAoLQ9TAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAADUExURUxpcU3H2DoAAAABdFJOUwBA5thmAAAADUlEQVQY02NgGAXIAAABEAAB7JfjegAAAABJRU5ErkJggg==iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAMAAABEpIrGAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAADUExURUxpcU3H2DoAAAABdFJOUwBA5thmAAAAEklEQVQ4y2NgGAWjYBSMAuwAAAQgAAFWu83mAAAAAElFTkSuQmCC")
				c.String(http.StatusOK, string(output))
			} else {
				data := gin.H{
					"ifaces":   getIfaces(),
					"campaign": camp,
				}
				c.HTML(http.StatusOK, "campaignEdit.html", data)
			}
		})

		mailer.POST("campaign/edit/:id", func(c *gin.Context) {
			data := gin.H{
				"ifaces": getIfaces(),
				"campaign": postCampaignInfo(
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

		mailer.GET("recipients/:id", hRecipients)
		mailer.POST("recipients/:id", uploadRecipients)

	}

	api := router.Group("api")
	{
		table := api.Group("mailer")
		{
			table.GET("recipients/:id", aRecipients)
		}
	}

	router.Run(":7777")
}
