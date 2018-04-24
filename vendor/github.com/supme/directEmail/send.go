// Package directEmail support direct send email from selected interface include SOCKS5 proxy server
package directEmail

import (
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/net/proxy"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// A Email contains options for send email.
type Email struct {
	// Ip contains local IP address wich use for send email
	// if blank use default interface
	Ip string
	// Host is host name
	// if blank use DNS resolv for field fill
	Host string
	// MapIp use for translate local IP to global if NAT
	// if use Socks server translate IP SOCKS server to real IP
	MapIp map[string]string
	// FromEmail sender email (Required)
	FromEmail string
	// FromName sender name
	FromName string
	// ToEmail recipient email (Required)
	ToEmail string
	// ToName recipient name
	ToName string
	// Subject email subject
	Subject string

	headers         []string
	textPlain       []byte
	textHtml        []byte
	attachments     [][]byte
	raw             bytes.Buffer
	bodyLenght 		int
}

const debugIs = false

// New returns a new Email instance for create and send email
func New() Email {
	return Email{}
}

// Send sending email message
func (self *Email) Send() error {
	var err error

	domain, err := self.domainFromEmail(self.ToEmail)
	if err != nil {
		return errors.New("550 Bad ToEmail")
	}

	client, err := self.dial(domain)
	if err != nil {
		return errors.New(fmt.Sprintf("550 %v", err))
	}

	if err := client.Hello(strings.TrimRight(self.Host, ".")); err != nil {
		return err
	}

	if err := client.Mail(self.FromEmail); err != nil {
		return err
	}

	if err := client.Rcpt(self.ToEmail); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(w, self.raw.String())
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()

}

type dialer func(network, address string) (net.Conn, error)

func (self *Email) dial(domain string) (client *smtp.Client, err error) {
	var (
		conn     net.Conn
		dialFunc dialer
	)

	if self.Ip == "" {
		iface := net.Dialer{}
		dialFunc = iface.Dial
		debug("Dial function is default network interface\n")
	} else {
		if strings.ToLower(self.Ip[0:8]) == "socks://" {
			iface, err := proxy.SOCKS5("tcp", self.Ip[8:], nil, proxy.FromEnvironment())
			if err != nil {
				return nil, err
			}
			dialFunc = iface.Dial
			debug("Dial function is socks proxy from ", self.Ip[8:], "\n")
		} else {
			addr := &net.TCPAddr{
				IP: net.ParseIP(self.Ip),
			}
			iface := net.Dialer{LocalAddr: addr}
			dialFunc = iface.Dial
			debug("Dial function is ", addr.String(), " network interface\n")
		}
	}

	records, err := net.LookupMX(domain)
	if err != nil {
		return
	}
	debug("MX for domain:\n")
	for i := range records {
		debug(" - ", records[i].Pref, " ", records[i].Host, "\n")
	}

	for i := range records {
		server := strings.TrimRight(strings.TrimSpace(records[i].Host), ".")
		debug("Connect to server ", server, "\n")
		conn, err = dialFunc("tcp", net.JoinHostPort(server, "25"))
		if err != nil {
			debug("Not connected\n")
			continue
		}
		debug("Connected\n")
		client, err = smtp.NewClient(conn, server)
		if err == nil {
			break
		}
	}
	if err != nil {
		return
	}

	conn.SetDeadline(time.Now().Add(5 * time.Minute)) // SMTP RFC

	if self.Ip == "" {
		self.Ip = conn.LocalAddr().String()
	}

	if self.Host == "" {
		var myGlobalIP string
		myIp, _, err := net.SplitHostPort(strings.TrimLeft(self.Ip, "socks://"))
		myGlobalIP, ok := self.MapIp[myIp]
		if !ok {
			myGlobalIP = myIp
		}
		names, err := net.LookupAddr(myGlobalIP)
		if err != nil && len(names) < 1 {
			return nil, err
		}
		debug("LookUp ", myGlobalIP, " this result ", names[0], "\n")
		self.Host = names[0]
	}

	return
}
