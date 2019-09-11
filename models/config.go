package models

import "C"
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
	dbString         string
	dbConnections    int
	UTMDefaultURL    string
	APIPort          string
	APIPanelPath     string
	APIPanelLocale   string
	UTMPort          string
	MaxCampaingns    int
	RealSend         bool
	DNScache         bool
	DefaultProfileID int
	AdminEmail       string
	GonderEmail      string
}

var (
	Db     *sql.DB
	Config config
	LogDir string
)

// ReadConfig read config file redefine variables from environment, if exist, check and connect to database and create email pool
// use env variables for redefine variables from config:
//  GONDER_MAIN_DEFAULT_PROFILE_ID (int)
//  GONDER_MAIN_ADMIN_EMAIL (string)
//  GONDER_MAIN_GONDER_EMAIL (string)
//  GONDER_DATABASE_STRING (string)
//  GONDER_DATABASE_CONNECTIONS (int)
//  GONDER_MAILER_SEND (bool)
//  GONDER_MAILER_DNS_CACHE (bool)
//  GONDER_MAILER_MAX_CAMPAIGNS (int)
//  GONDER_UTM_DEFAULT_URL (string)
//  GONDER_UTM_PORT (int)
//  GONDER_API_PORT (int)
//  GONDER_API_PANEL_PATH (string)
//  GONDER_API_PANEL_LOCALE (string)
func ReadConfig(configFile string) error {
	err := Config.read(configFile)
	if err != nil {
		return err
	}
	return nil
}

func ConnectDb() error {
	var err error
	Db, err = sql.Open("mysql", Config.dbString)
	if err != nil {
		return fmt.Errorf("open database error: %s", err)
	}
	err = Db.Ping()
	if err != nil {
		return fmt.Errorf("ping database error: %s", err)
	}
	Db.SetMaxIdleConns(Config.dbConnections)
	Db.SetMaxOpenConns(Config.dbConnections)
	return nil
}

func CheckDb() error {
	if _, err := Db.Query("SELECT 1 FROM `auth_user`"); err != nil {
		return fmt.Errorf("database is empty, use -i key for create from a template")
	}
	return nil
}

// InitDb initialize database
func InitDb(withoutConfirm bool) error {
	if !withoutConfirm {
		var confirm string
		fmt.Print("Initial database (y/N)? ")
		if _, err := fmt.Scanln(&confirm); err != nil {
			return err
		}
		if strings.ToLower(confirm) != "y" {
			return nil
		}
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

func (c *config) read(configFile string) error {
	config, err := configparser.Read(configFile)
	if err != nil {
		return fmt.Errorf("error parse config file: %s", err)
	}

	mainSection, err := config.Section("main")
	if err != nil {
		return fmt.Errorf("error parse config file: %s", err)
	}

	c.DefaultProfileID, err = getEnvInt("GONDER_MAIN_DEFAULT_PROFILE_ID", mainSection.ValueOf("default_profile_id"))
	if err != nil {
		return fmt.Errorf("error parse config file value defaultProfile: %s", err)
	}

	c.AdminEmail = getEnvString("GONDER_MAIN_ADMIN_EMAIL", mainSection.ValueOf("admin_email"))
	c.GonderEmail = getEnvString("GONDER_MAIN_GONDER_EMAIL", mainSection.ValueOf("gonder_email"))

	databaseSection, err := config.Section("database")
	if err != nil {
		return fmt.Errorf("error parse config file section Database: %s", err)
	}
	c.dbString = getEnvString("GONDER_DATABASE_STRING", databaseSection.ValueOf("string"))
	c.dbConnections, err = getEnvInt("GONDER_DATABASE_CONNECTIONS", databaseSection.ValueOf("connections"))
	if err != nil {
		c.dbConnections = 10
	}

	mailerSection, err := config.Section("mailer")
	if err != nil {
		return fmt.Errorf("error parse config file: %s", err)
	}
	c.RealSend, err = getEnvBool("GONDER_MAILER_SEND", mailerSection.ValueOf("send"))
	if err != nil {
		return fmt.Errorf("error parse config file value send: %s", err)
	}
	c.DNScache, err = getEnvBool("GONDER_MAILER_DNS_CACHE", mailerSection.ValueOf("dns_cache"))
	if err != nil {
		return fmt.Errorf("error parse config file value dns_cache: %s", err)
	}

	c.MaxCampaingns, err = getEnvInt("GONDER_MAILER_MAX_CAMPAIGNS", mailerSection.ValueOf("max_campaigns"))
	if err != nil {
		return fmt.Errorf("error parse config file value max_campaigns: %s", err)
	}

	utmSection, err := config.Section("utm")
	if err != nil {
		return fmt.Errorf("error parse config file has not section Statistic: %s", err)
	}
	c.UTMDefaultURL = getEnvString("GONDER_UTM_DEFAULT_URL", utmSection.ValueOf("default_url"))
	c.UTMPort = getEnvString("GONDER_UTM_PORT", utmSection.ValueOf("port"))

	apiSection, err := config.Section("api")
	if err != nil {
		return fmt.Errorf("error parse config file has not section API: %s", err)
	}
	c.APIPort = getEnvString("GONDER_API_PORT", apiSection.ValueOf("port"))
	c.APIPanelPath = getEnvString("GONDER_API_PANEL_PATH", apiSection.ValueOf("panel_path"))
	c.APIPanelLocale = getEnvString("GONDER_API_PANEL_LOCALE", apiSection.ValueOf("panel_locale"))

	return nil
}

func getEnvString(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue string) (int, error) {
	if value, exists := os.LookupEnv(key); exists {
		v, err := strconv.Atoi(value)
		return v, err
	}
	v, err := strconv.Atoi(defaultValue)
	return v, err
}

func getEnvBool(key string, defaultValue string) (bool, error) {
	if value, exists := os.LookupEnv(key); exists {
		v, err := strconv.ParseBool(value)
		return v, err
	}
	v, err := strconv.ParseBool(defaultValue)
	return v, err
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
