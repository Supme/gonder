package models

import (
	"database/sql"
	"fmt"
	"github.com/alyu/configparser"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kardianos/osext"
	"io/ioutil"
	"log"
	"path/filepath"
	"strconv"
	"strings"
)

type config struct {
	dbType, dbString string
	dbConnections    int
	RootPath         string
	Version          string
	URL              string
	APIPort          string
	StatPort         string
	MaxCampaingns    int
	RealSend         bool
	DNScache         bool
	DefaultProfile   int
	AdminMail        string
	GonderMail       string
}

var (
	Db     *sql.DB
	Config config
)

func init() {
	var err error
	Config.Update()
	Db, err = sql.Open(Config.dbType, Config.dbString)
	checkErr(err)
	Db.SetMaxIdleConns(Config.dbConnections)
	Db.SetMaxOpenConns(Config.dbConnections)
	_, err = Db.Query("SELECT 1 FROM `auth_user`")
	if err != nil {
		checkErr(createDb())
	}
	InitEmailPool()
}

func createDb() error {
	var input string
	fmt.Print("Install database (y/N)? ")
	fmt.Scanln(&input)
	if input == "y" || input == "Y" {
		sql, err := ioutil.ReadFile("dump-my.sql")
		if err != nil {
			return err
		}
		query := strings.Split(string(sql), ";")
		for i := range query {
			_, err = Db.Exec(query[i])
			if err != nil {
				return err
			}
		}
	}
	return nil
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

	c.DefaultProfile, err = strconv.Atoi(mainConfig.ValueOf("defaultProfile"))
	if err != nil {
		panic("Error parse config parametr 'defaultProfile'")
	}
	c.AdminMail = mainConfig.ValueOf("adminMail")
	c.GonderMail = mainConfig.ValueOf("gonderMail")

	dbConfig, err := config.Section("Database")
	checkErr(err)
	c.dbType = dbConfig.ValueOf("type")
	c.dbString = dbConfig.ValueOf("string")
	c.dbConnections, err = strconv.Atoi(dbConfig.ValueOf("connections"))
	if err != nil {
		c.dbConnections = 10
	}

	mailerConfig, err := config.Section("Mailer")
	checkErr(err)
	if mailerConfig.ValueOf("send") == "yes" {
		c.RealSend = true
	} else {
		c.RealSend = false
	}
	if mailerConfig.ValueOf("dnscache") == "yes" {
		c.DNScache = true
	} else {
		c.DNScache = false
	}
	c.MaxCampaingns, err = strconv.Atoi(mailerConfig.ValueOf("maxcampaign"))
	if err != nil {
		panic("Error parse config parametr 'maxcampaign'")
	}

	statisticConfig, err := config.Section("Statistic")
	checkErr(err)
	apiConfig, err := config.Section("API")
	checkErr(err)
	c.URL = "http://" + mainConfig.ValueOf("host")
	c.StatPort = statisticConfig.ValueOf("port")
	c.APIPort = apiConfig.ValueOf("port")

	c.Version = "0.9"
}

func FromRootDir(path string) string {
	return filepath.Join(Config.RootPath, path)
}

func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}
