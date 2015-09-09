package mailer

import (
	"database/sql"
	"golang.org/x/net/proxy"
	"net"
)

var HostName string

var Db *sql.DB

type message struct {
	Subject string
	Body    string
}

type pJson struct {
	Campaign    string `json:"c"`
	Recipient   string `json:"r"`
	Url         string `json:"u"`
	Webver      string `json:"w"`
	Opened      string `json:"o"`
	Unsubscribe string `json:"s"`
}


type attachmentData struct {
	Location string
	Name     string
}

type mailData struct {
	Host         string
	From         string
	From_name    string
	To           string
	To_name      string
	Extra_header string
	Subject      string
	Html         string
	Attachments  []attachmentData
}

var s proxy.Dialer
var n net.Dialer
var (
	recipientId string
	campaignId  string
	iface       string
	stream      int
	delay       int
)
