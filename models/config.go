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
package models

import (
	"database/sql"
	"strconv"
	"github.com/alyu/configparser"
	"github.com/kardianos/osext"
	"log"
	"strings"
	"path/filepath"
)

type config struct {
	dbType, dbString string
	RootPath string
	Version string
	Url string
	ApiPort string
	StatPort string
	MaxCampaingns int
	RealSend bool
}

var (
	Db *sql.DB
	Config config
)

func (c *config) Get() {
	var err error
	c.Update()
	Db, err = sql.Open(c.dbType, c.dbString)
	checkErr(err)
	Db.SetMaxIdleConns(10)
	Db.SetMaxOpenConns(10)
}

func (c *config) Close() {
	Db.Close()
}

func (c *config) Update() {
	var err error

	c.RootPath, err = osext.ExecutableFolder()
	checkErr(err)
	if strings.Contains(c.RootPath, "go-build") {
		c.RootPath = "."
	}

	config, err := configparser.Read(FromRootDir("config.ini"))
	checkErr(err)

	mainConfig, err := config.Section("Main")
	checkErr(err)

	dbConfig, err := config.Section("Database")
	checkErr(err)
	c.dbType = dbConfig.ValueOf("type")
	c.dbString = dbConfig.ValueOf("string")

	mailerConfig, err := config.Section("Mailer")
	checkErr(err)
	if mailerConfig.ValueOf("send") == "yes" {
		c.RealSend = true
	} else {
		c.RealSend = false
	}
	c.MaxCampaingns, err = strconv.Atoi(mailerConfig.ValueOf("maxcampaign"))
	if err != nil {
		panic("Error parse config parametr 'maxcampaign'")
	}

	statisticConfig, err := config.Section("Statistic")
	checkErr(err)
	apiConfig, err := config.Section("API")
	checkErr(err)
	c.Url = "http://" + mainConfig.ValueOf("host")
	c.StatPort = statisticConfig.ValueOf("port")
	c.ApiPort = apiConfig.ValueOf("port")

	c.Version = "Gonder 0.6"
}


func FromRootDir(path string) string {
	return filepath.Join(Config.RootPath, path)
}

func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}