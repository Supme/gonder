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
package panel

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"encoding/base64"
	"os"
	"bufio"
	"strings"
)


func routers() {

	router := gin.Default()
	router.LoadHTMLGlob("panel/templates/*")
	router.Static("/static/", "static")
	router.Static("/fonts/", "fonts")
	router.Static("/assets/", "assets")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "main.html", gin.H{})
	})

	// get users from file
	users := make(gin.Accounts)
	file, err := os.Open("users.txt")
	checkErr(err)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		user :=strings.Split(line,":")
		users[user[0]] = user[1]
	}
	checkErr(err)


	mailer := router.Group("mailer", gin.BasicAuth(users))
	{
		mailer.GET("logout", func(c *gin.Context) {
			c.HTML(http.StatusUnauthorized, "logout.html", nil)
		})

		mailer.GET("group", func(c *gin.Context) {
			data := gin.H{
				"groups": getGroups(),
			}
			c.HTML(http.StatusOK, "group.html", data)
		})

		mailer.POST("group", func(c *gin.Context) {
			addGroup(c.PostForm("name"))
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

		mailer.POST("group/:id", func(c *gin.Context) {
			addCampaigns(c.Param("id"), c.PostForm("name"))
			data := gin.H{
				"campaigns": getCampaigns(c.Param("id")),
			}
			c.HTML(http.StatusOK, "campaign.html", data)
		})

		mailer.GET("campaign/edit/:id", func(c *gin.Context) {
			camp, err := getCampaignInfo(c.Param("id"))

			if err == nil {
				data := gin.H{
					"ifaces":   getProfiles(),
					"campaign": camp,
				}
				c.HTML(http.StatusOK, "campaignEdit.html", data)
			} else {
				// blank 16x16 png
				c.Header("Content-Type", "image/png")
				output, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAMAAAAoLQ9TAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAADUExURUxpcU3H2DoAAAABdFJOUwBA5thmAAAADUlEQVQY02NgGAXIAAABEAAB7JfjegAAAABJRU5ErkJggg==iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAMAAABEpIrGAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAADUExURUxpcU3H2DoAAAABdFJOUwBA5thmAAAAEklEQVQ4y2NgGAWjYBSMAuwAAAQgAAFWu83mAAAAAElFTkSuQmCC")
				c.String(http.StatusOK, string(output))
			}
		})

		mailer.POST("campaign/edit/:id", func(c *gin.Context) {
			data := gin.H{
				"ifaces": getProfiles(),
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
						SendUnsubscribe: c.PostForm("sendUnsubscribe"),
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
		mailer := api.Group("mailer", gin.BasicAuth(users))
		{
			mailer.GET("recipients/:id", aRecipients)
			mailer.GET("profile/", apiGetProfiles)
			mailer.GET("profile/:id", apiGetProfile)
			mailer.POST("profile/", apiPostProfile)
			mailer.PUT("profile/", apiPutProfile)
		}
	}

	settings := router.Group("settings", gin.BasicAuth(users))
	{

		settings.GET("interfaces", func(c *gin.Context){
			data := gin.H{
				"ifaces":   getProfiles(),
				"ip": getIfaces(),
			}
			c.HTML(http.StatusOK, "settings.html", data)
		})

	}

	router.Run(":" + Port)


}
