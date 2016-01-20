package main

import (
	"database/sql"
	"fmt"
	"github.com/alyu/configparser"
	_ "github.com/go-sql-driver/mysql"
	"github.com/supme/gonder/mailer"
	"github.com/supme/gonder/panel"
	"log"
	"runtime"
	"net"
	"os"
	"os/exec"
)

func main() {

	i, _ :=net.InterfaceAddrs()
	fmt.Println(i)

	runtime.GOMAXPROCS(runtime.NumCPU())

	log.Println("Read config file")
	config, err := configparser.Read("./config.ini")
	checkErr(err)

	dbConfig, err := config.Section("Database")
	checkErr(err)

	hostConfig, err := config.Section("Host")
	checkErr(err)

	log.Println("Connect to database")
	db, err := sql.Open(dbConfig.ValueOf("type"), dbConfig.ValueOf("string"))
	checkErr(err)
	defer db.Close()
	checkErr(db.Ping())

	mailer.Db = db
	panel.Db = db

	if hostConfig.ValueOf("port") == "80" {
		mailer.HostName = "http://" + hostConfig.ValueOf("name")
	} else {
		mailer.HostName = "http://" + hostConfig.ValueOf("name") + ":" + hostConfig.ValueOf("port")
	}

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
				mailer.Sender()

				for {}
			}

			if os.Args[2] == "stat" {

				log.Println("Start statistics http server")
				mailer.Stat(hostConfig.ValueOf("port"))

				for {}
			}

		}

	}

	if len(os.Args) == 1 {
		log.Println("Start panel http server")
		go panel.Run()

		log.Println("Start database mailer")
		go mailer.Sender()

		log.Println("Start statistics http server")
		go mailer.Stat(hostConfig.ValueOf("port"))

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
