package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/alyu/configparser"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"text/template"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "c", "./dist_config.toml", "Path to config file")
	flag.Parse()

	config, err := configparser.Read(configFile)
	if err != nil {
		fmt.Printf("Error open config file: %s\r\n", err)
		os.Exit(1)
	}

	databaseSection, err := config.Section("database")
	if err != nil {
		fmt.Printf("Error read config file setion database: %s\r\n", err)
		os.Exit(1)
	}

	db, err := sql.Open("mysql", databaseSection.ValueOf("string"))
	if err != nil {
		fmt.Printf("Connect to database: %s\r\n", err)
		os.Exit(1)
	}

	type profile struct {
		ID          int
		Name        string
		Iface       string
		Hostname    string
		Stream      int
		ResendDelay int
		ResendCount int
	}

	var profiles []profile

	rows, err := db.Query("SELECT `id`,`name`,`iface`,`host`,`stream`,`resend_delay`,`resend_count` FROM `profile`")
	if err != nil {
		fmt.Printf("Get rows from database: %s\r\n", err)
		os.Exit(1)
	}

	for rows.Next() {
		var p profile
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Iface,
			&p.Hostname,
			&p.Stream,
			&p.ResendDelay,
			&p.ResendCount,
		)
		if err != nil {
			fmt.Printf("Scan row from database: %s\r\n", err)
			os.Exit(1)
		}
		profiles = append(profiles, p)
	}

	tmpl := `
# profile config imported from db
{{range .}}
[[profile]]
id = {{.ID}}
name = {{.Name}}
hostname = {{.Hostname}}
interface = {{.Iface}}{{if ne .Hostname "group"}}
stream = {{.Stream}}{{else}}
#stream not need for group{{end}}
resend_count = {{.ResendCount}}
resend_delay = {{.ResendDelay}}
{{end}}
# end profile config imported from db
`
	t, err := template.New("profiles").Parse(tmpl)
	if err != nil {
		fmt.Printf("Parse template error: %s\r\n", err)
		os.Exit(1)
	}
	err = t.Execute(os.Stdout, profiles)
	if err != nil {
		fmt.Printf("Template error: %s\r\n", err)
		os.Exit(1)
	}
}
