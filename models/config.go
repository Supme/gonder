package models

import (
	"database/sql"
	"fmt"
	"github.com/alyu/configparser"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type config struct {
	dbType, dbString string
	dbConnections    int
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

func Init(configFile string) error {
	err := Config.Update(configFile)
	if err != nil {
		return err
	}
	Db, err = sql.Open(Config.dbType, Config.dbString)
	if err != nil {
		return fmt.Errorf("open database error: %s", err)
	}
	err = Db.Ping()
	if err != nil {
		return fmt.Errorf("ping database error: %s", err)
	}
	Db.SetMaxIdleConns(Config.dbConnections)
	Db.SetMaxOpenConns(Config.dbConnections)
	_, err = Db.Query("SELECT 1 FROM `auth_user`")
	if err != nil {
		return fmt.Errorf("database is empty, use -i key for create from a template")
	}
	err = InitEmailPool()
	if err != nil {
		return err
	}
	return nil
}

func InitDb() error {
	var confirm string
	fmt.Print("Initial database (y/N)? ")
	if _, err := fmt.Scanln(&confirm); err != nil {
		return err
	}
	if strings.ToLower(confirm) != "y" {
		return nil
	}

	sqlDump, err := ioutil.ReadFile("dump.sql")
	if err != nil {
		return err
	}
	query := strings.Split(string(sqlDump), ";")
	tx, err := Db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	for i := range query[:len(query)-1] {
		_, err = tx.Exec(query[i])
		if err != nil {
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (c *config) Update(configFile string) error {
	var err error

	config, err := configparser.Read(configFile)
	if err != nil {
		return fmt.Errorf("error parse config file: %s", err)
	}

	mainConfig, err := config.Section("Main")
	if err != nil {
		return fmt.Errorf("error parse config file: %s", err)
	}

	c.DefaultProfile, err = strconv.Atoi(mainConfig.ValueOf("defaultProfile"))
	if err != nil {
		return fmt.Errorf("error parse config file value defaultProfile: %s", err)
	}

	c.AdminMail = mainConfig.ValueOf("adminMail")
	c.GonderMail = mainConfig.ValueOf("gonderMail")

	dbConfig, err := config.Section("Database")
	if err != nil {
		return fmt.Errorf("error parse config file section Database: %s", err)
	}
	c.dbType = dbConfig.ValueOf("type")
	c.dbString = dbConfig.ValueOf("string")
	c.dbConnections, err = strconv.Atoi(dbConfig.ValueOf("connections"))
	if err != nil {
		c.dbConnections = 10
	}

	mailerConfig, err := config.Section("Mailer")
	if err != nil {
		return fmt.Errorf("error parse config file: %s", err)
	}
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
		return fmt.Errorf("error parse config file value maxcampaign: %s", err)
	}

	statisticConfig, err := config.Section("Statistic")
	if err != nil {
		return fmt.Errorf("error parse config file has not section Statistic: %s", err)
	}
	apiConfig, err := config.Section("API")
	if err != nil {
		return fmt.Errorf("error parse config file has not section API: %s", err)
	}
	c.URL = mainConfig.ValueOf("host")
	c.StatPort = statisticConfig.ValueOf("port")
	c.APIPort = apiConfig.ValueOf("port")
	return nil
}

func WorkDir(path string) string {
	wd, err := os.Getwd()
	checkErr(err)
	return filepath.Join(wd, path)
}

func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}
