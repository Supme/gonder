package models

import (
	"fmt"
	"github.com/alyu/configparser"
	"github.com/jmoiron/sqlx"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

type config struct {
	dbString         string
	dbConnections    int
	secretString     string
	APIPort          string
	APIPanelPath     string
	APIPanelLocale   string
	AdminEmail       string
	GonderEmail      string
	UTMDefaultURL    string
	UTMTemplatesDir  string
	UTMFilesDir      string
	UTMPort          string
	MaxCampaingns    int
	RealSend         bool
	DontUseTLS       bool
	DNScache         bool
	DefaultProfileID int
}

var (
	Db        *sqlx.DB
	Config    config
	LogDir    string
	ServerPem string
	ServerKey string
)

// ReadConfig read config file redefine variables from environment, if existed, check and connect to database and create email pool
// use env variables for redefine variables from config:
//  GONDER_MAIN_DEFAULT_PROFILE_ID (int)
//  GONDER_MAIN_ADMIN_EMAIL (string)
//  GONDER_MAIN_GONDER_EMAIL (string)
//  GONDER_MAIN_SECRET_STRING (string)
//  GONDER_DATABASE_STRING (string)
//  GONDER_DATABASE_CONNECTIONS (int)
//  GONDER_MAILER_SEND (bool)
//  GONDER_MAILER_DONT_USE_TLS (bool)
//  GONDER_MAILER_DNS_CACHE (bool)
//  GONDER_MAILER_MAX_CAMPAIGNS (int)
//  GONDER_UTM_DEFAULT_URL (string)
//  GONDER_UTM_TEMPLATES_DIR (string)
//  GONDER_UTM_FILES_DIR (string)
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
	c.secretString = getEnvString("GONDER_MAIN_SECRET_STRING", mainSection.ValueOf("secret_string"))

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
	c.DontUseTLS, err = getEnvBool("GONDER_MAILER_DONT_USE_TLS", mailerSection.ValueOf("dont_use_tls"))
	if err != nil {
		return fmt.Errorf("error parse config file value dont_use_tls: %s", err)
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
	c.UTMTemplatesDir = getEnvString("GONDER_UTM_TEMPLATES_DIR", utmSection.ValueOf("templates_dir"))
	c.UTMFilesDir = getEnvString("GONDER_UTM_FILES_DIR", utmSection.ValueOf("files_dir"))
	c.UTMPort = getEnvString("GONDER_UTM_PORT", utmSection.ValueOf("port"))

	apiSection, err := config.Section("api")
	if err != nil {
		return fmt.Errorf("error parse config file has not section API: %s", err)
	}
	c.APIPort = getEnvString("GONDER_API_PORT", apiSection.ValueOf("port"))
	c.APIPanelPath = getEnvString("GONDER_API_PANEL_PATH", apiSection.ValueOf("panel_path"))
	c.APIPanelLocale = getEnvString("GONDER_API_PANEL_LOCALE", apiSection.ValueOf("panel_locale"))

	profileSections, _ := config.Sections("profile")

	profiles := []profile{
		{
			id:          0,
			name:        "Default",
			hostname:    "",
			iface:       "",
			stream:      10,
			resendCount: 2,
			resendDelay: 1200,
		},
	}
	for i := range profileSections {
		var id, stream, resendCount, resendDelay int
		id, err = strconv.Atoi(profileSections[i].ValueOf("id"))
		if err != nil {
			return fmt.Errorf("error parse config file profile value id: %s", err)
		}
		if profileSections[i].ValueOf("stream") != "" {
			stream, err = strconv.Atoi(profileSections[i].ValueOf("stream"))
			if err != nil {
				return fmt.Errorf("error parse config file profile value stream: %s", err)
			}
		}
		resendCount, err = strconv.Atoi(profileSections[i].ValueOf("resend_count"))
		if err != nil {
			return fmt.Errorf("error parse config file profile value resend_count: %s", err)
		}
		resendDelay, err = strconv.Atoi(profileSections[i].ValueOf("resend_delay"))
		if err != nil {
			return fmt.Errorf("error parse config file profile value resend_delay: %s", err)
		}
		profiles = append(profiles, profile{
			id:          id,
			name:        profileSections[i].ValueOf("name"),
			hostname:    profileSections[i].ValueOf("hostname"),
			iface:       profileSections[i].ValueOf("interface"),
			stream:      stream,
			resendCount: resendCount,
			resendDelay: resendDelay,
		})
	}

	err = initPool(profiles)
	if err != nil {
		return fmt.Errorf("error initialize email pool: %s", err)
	}

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
