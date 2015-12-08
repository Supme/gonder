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

	log.Println("Start database mailer")
	go mailer.Sender()

	log.Println("Start statistics http server")
	go mailer.Stat(hostConfig.ValueOf("port"))

	log.Println("Start panel http server")
	go panel.Run()

	//log.Println("Press Enter for stop")
	//var input string
	//fmt.Scanln(&input)
	for {}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
