package main

import (
	"github.com/supme/gonder/mailer"
	"github.com/supme/gonder/statistic"
	"github.com/supme/gonder/panel"
	"github.com/supme/gonder/models"
	"fmt"
	"github.com/alyu/configparser"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"os/exec"
	"runtime"
	"database/sql"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	log.Println("Read config file")
	config, err := configparser.Read("./config.ini")
	checkErr(err)

	mainConfig, err := config.Section("Main")
	checkErr(err)

	dbConfig, err := config.Section("Database")
	checkErr(err)

	mailerConfig, err := config.Section("Mailer")
	checkErr(err)

	statisticConfig, err := config.Section("Statistic")
	checkErr(err)

	panelConfig, err := config.Section("Panel")
	checkErr(err)

	// Init models
	log.Println("Connect to database")
	models.Db, err = sql.Open(dbConfig.ValueOf("type"), dbConfig.ValueOf("string"))
	checkErr(err)
	defer models.Db.Close()
	checkErr(models.Db.Ping())

	models.Db.SetMaxIdleConns(5)
	models.Db.SetMaxOpenConns(5)

	if statisticConfig.ValueOf("port") == "80" {
		models.StatUrl = "http://" + mainConfig.ValueOf("name")
	} else {
		models.StatUrl = "http://" + mainConfig.ValueOf("name") + ":" + statisticConfig.ValueOf("port")
	}

	// Init mailer
	if mailerConfig.ValueOf("send") == "yes" {
		mailer.Send = true
	} else {
		mailer.Send = false
	}

	// Init statistic
	statistic.Port = statisticConfig.ValueOf("port")

	// Init control panel
	panel.Port = panelConfig.ValueOf("port")

	// Start
	if len(os.Args) == 2 {

		if os.Args[1] == "start" {
			p := exec.Command(os.Args[0], "start", "panel")
			p.Start()
			fmt.Println("Panel [PID]", p.Process.Pid)

			sd := exec.Command(os.Args[0], "start", "sender")
			sd.Start()
			fmt.Println("Sender [PID]", sd.Process.Pid)

			st := exec.Command(os.Args[0], "start", "stat")
			st.Start()
			fmt.Println("Statistic [PID]", st.Process.Pid)

			os.Exit(0)
		}

		if os.Args[1] == "stop" {
			c := exec.Command("killall", os.Args[0])
			c.Start()
			fmt.Println("Stop all services")

			os.Exit(0)
		}
	}

	if len(os.Args) == 3 {

		if os.Args[1] == "start" {

			if os.Args[2] == "panel" {

				log.Println("Start panel http server")
				panel.Run()

				for {}
			}

			if os.Args[2] == "sender" {

				log.Println("Start database mailer")
				mailer.Run()

				for {}
			}

			if os.Args[2] == "stat" {

				log.Println("Start statistics http server")
				statistic.Run()

				for {}
			}

		}

	}

	if len(os.Args) == 1 {
		log.Println("Start panel http server")
		go panel.Run()

		log.Println("Start database mailer")
		go mailer.Run()

		log.Println("Start statistics http server")
		go statistic.Run()

		log.Println("Press Enter for stop")
		var input string
		fmt.Scanln(&input)
	}


}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
