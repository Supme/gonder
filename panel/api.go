package panel

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"log"
)

type profile struct  {
	Id string `form:"id" json:"id"`
	Name string `form:"name" json:"name"`
	Iface string `form:"iface" json:"iface"`
	Server string `form:"server" json:"server"`
	Host string `form:"host" json:"host"`
	Stream string `form:"stream" json:"stream"`
	Delay string `form:"delay" json:"delay"`
}

func apiGetProfiles(c *gin.Context)  {
	var p profile
	var ps []profile
	rows, err := Db.Query("SELECT `id`, `name`, `iface`, `host`, `stream`, `delay` FROM `profile`")
	checkErr(err)
	for rows.Next() {
		rows.Scan(&p.Id, &p.Name, &p.Iface, &p.Host, &p.Stream, &p.Delay)
		ps = append(ps, p)
	}
	log.Print(ps)
	c.JSON(http.StatusOK, ps)
}

func apiGetProfile(c *gin.Context) {
	var p profile
	err := Db.QueryRow("SELECT `id`, `name`, `iface`, `host`, `stream`, `delay` FROM `profile` WHERE `id`=?", c.Param("id")).Scan(
		&p.Id, &p.Name, &p.Iface, &p.Host, &p.Stream, &p.Delay,
	)
	if err != nil {
		p.Id = ""
		p.Name = ""
		p.Iface = ""
		p.Host = ""
		p.Server = ""
		p.Stream = "100"
		p.Delay = "0"
	} else {
		if p.Iface[:8] == "socks://" {
			p.Server = p.Iface[8:]
			p.Iface = p.Iface[:8]
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"id": p.Id,
		"name": p.Name,
		"iface": p.Iface,
		"server": p.Server,
		"host": p.Host,
		"stream": p.Stream,
		"delay": p.Delay,
	})
}

func apiPostProfile(c *gin.Context)  {
	var p profile
	if c.BindJSON(&p) == nil {
		if p.Iface[:8] == "socks://" {
			p.Iface = p.Iface + p.Server
		}
		res, err := Db.Exec("INSERT INTO `profile`(`name`, `iface`, `host`, `stream`, `delay`) VALUES (?,?,?,?,?)", p.Name, p.Iface, p.Host, p.Stream, p.Delay)
		checkErr(err)
		id, err := res.LastInsertId()
		p.Id = strconv.FormatInt(id, 10)
		c.JSON(http.StatusOK, gin.H{
			"status": "Ok",
			"id": p.Id,
		})
	} else {
		c.JSON(http.StatusNotAcceptable, gin.H{"status": "NotAcceptable"})
	}
}

func apiPutProfile(c *gin.Context)  {
	var p profile
	if c.BindJSON(&p) == nil {
		if p.Iface[:8] == "socks://" {
			p.Iface = p.Iface + p.Server
		}
		res, err := Db.Exec("UPDATE `profile` SET `name`=?,`iface`=?,`host`=?,`stream`=?,`delay`=? WHERE `id`=?", p.Name, p.Iface, p.Host, p.Stream, p.Delay, p.Id)
		checkErr(err)
		n, err := res.RowsAffected()
		checkErr(err)
		if n == 1 {
			c.JSON(http.StatusOK, gin.H{"status": "Ok"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"status": "Bad request"})
		}
	} else {
		c.JSON(http.StatusNotAcceptable, gin.H{"status": "NotAcceptable"})
	}
}