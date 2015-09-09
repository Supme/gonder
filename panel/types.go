package panel

import (
	"database/sql"
	"log"
)

var Db *sql.DB

type iFace struct {
	Id     string
	Name   string
	Iface  string
	Host   string
	Stream string
	Delay  string
}

type campaign struct {
	Id        string
	IfaceId   string
	Name      string
	Subject   string
	From      string
	FromName  string
	Message   string
	StartTime string
	EndTime   string
}

type recipient struct {
	Id         string
	CampaignId string
	Email      string
	Name       string
}

func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}
