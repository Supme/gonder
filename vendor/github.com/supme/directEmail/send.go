// Package directEmail support direct send email from selected interface include SOCKS5 proxy server
package directEmail

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"golang.org/x/net/proxy"
	"net"
	"net/smtp"
	"net/url"
	"strings"
	"time"
)

// A Email contains options for send email.
type Email struct {
	// Ip contains local IP address which use for send email
	// if blank use default interface
	// if use socks5 proxy "socks://123.124.125.126:8080"
	// and with auth "socks://user:password@123.124.125.126:8080"
	Ip string
	// Host is host name
	// if blank use DNS resolv for field fill
	Host string
	// Port SMTP server port
	Port uint16
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

	headers     []string
	textPlain   []byte
	textHTML    []byte
	attachments [][]byte
	raw         bytes.Buffer
	bodyLenght  int
}

const (
	debugIs     = false
	connTimeout = 5 * time.Minute // SMTP RFC
)

// New returns a new Email instance for create and send email
func New() Email {
	return Email{Port: 25}
}

// SendThroughServer send email from SMTP server
func (slf *Email) SendThroughServer(host string, port uint16, username, password string) error {
	slf.Port = port

	dialFunc, err := slf.dialFunction()
	debug("Dialer selected, now dial to server\n")

	conn, err := dialFunc("tcp", net.JoinHostPort(host, fmt.Sprintf("%d", slf.Port)))
	if err != nil {
		debug("Not connected\n")
		return err
	}
	defer conn.Close()
	debug("Connected\n")

	err = conn.SetDeadline(time.Now().Add(connTimeout + time.Millisecond*10))
	if err != nil {
		return err
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	auth := smtp.PlainAuth(
		"",
		username,
		password,
		host,
	)
	slf.Host = "localhost"

	debug("Connected, now send email\n")

	slf.cleanEmail()
	return slf.sendWithTimeout(auth, host, client)
}

// Send email directly
func (slf *Email) Send() error {
	var err error

	server, err := slf.domainFromEmail(slf.ToEmail)
	if err != nil {
		return errors.New("553 Bad ToEmail")
	}

	dialFunc, err := slf.dialFunction()
	if err != nil {
		return fmt.Errorf("421 %v", err)
	}

	conn, err := slf.newClient(server, dialFunc)
	if err != nil {
		debug("Not connected\n")
		return fmt.Errorf("421 %v", err)
	}
	defer conn.Close()
	debug("Connected\n")

	slf.cleanEmail()
	return slf.sendWithTimeout(nil, "", conn)
}

func (slf *Email) cleanEmail() {
	slf.ToEmail = strings.TrimSpace(slf.ToEmail)
	slf.FromEmail = strings.TrimSpace(slf.FromEmail)
}

var testHookStartTLS func(*tls.Config)

// sendWithTimeout hack for bad server
func (slf *Email) sendWithTimeout(auth smtp.Auth, host string, client *smtp.Client) error {
	var res error
	err := make(chan error, 1)
	go func() {
		err <- slf.send(auth, host, client)
	}()

	select {
	case res = <-err:

	case <-time.After(connTimeout):
		res = fmt.Errorf("421 send timeout after %s", connTimeout.String())
	}

	return res
}

// send sending email message
func (slf *Email) send(auth smtp.Auth, host string, client *smtp.Client) error {
	var err error

	if err := client.Hello(strings.TrimRight(slf.Host, ".")); err != nil {
		return err
	}

	if auth != nil {
		if ok, _ := client.Extension("STARTTLS"); ok {
			config := &tls.Config{ServerName: host}
			if testHookStartTLS != nil {
				testHookStartTLS(config)
			}
			if err = client.StartTLS(config); err != nil {
				return err
			}
		}
		if auth != nil {
			if err = client.Auth(auth); err != nil {
				return err
			}
		}
	}

	if err := client.Mail(slf.FromEmail); err != nil {
		return err
	}

	if err := client.Rcpt(slf.ToEmail); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(w, slf.raw.String())
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()

}

type conn func(network, address string) (net.Conn, error)

func (slf *Email) dialFunction() (conn, error) {
	var dialFunc conn

	if slf.Ip == "" {
		iface := net.Dialer{}
		dialFunc = iface.Dial
		debug("Dial function is default network interface\n")
	} else {
		if strings.ToLower(slf.Ip[0:8]) == "socks://" {
			u, err := url.Parse(slf.Ip)
			if err != nil {
				return nil, fmt.Errorf("Error parse socks: %s", err.Error())
			}
			var iface proxy.Dialer
			if u.User != nil {
				auth := proxy.Auth{}
				auth.User = u.User.Username()
				auth.Password, _ = u.User.Password()
				iface, err = proxy.SOCKS5("tcp", u.Host, &auth, proxy.FromEnvironment())
				if err != nil {
					return dialFunc, err
				}
			} else {
				iface, err = proxy.SOCKS5("tcp", u.Host, nil, proxy.FromEnvironment())
				if err != nil {
					return dialFunc, err
				}
			}
			slf.Ip = u.Host
			dialFunc = iface.Dial
			debug("Dial function is socks proxy from ", slf.Ip[8:], "\n")
		} else {
			addr := &net.TCPAddr{
				IP: net.ParseIP(slf.Ip),
			}
			iface := net.Dialer{
				LocalAddr: addr,
			}
			dialFunc = iface.Dial
			debug("Dial function is ", addr.String(), " network interface\n")
		}
	}

	return dialFunc, nil
}

func (slf *Email) newClient(server string, dialFunc conn) (client *smtp.Client, err error) {
	var conn net.Conn

	records, err := net.LookupMX(server)
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
		conn, err = dialFunc("tcp", net.JoinHostPort(server, fmt.Sprintf("%d", slf.Port)))
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

	err = conn.SetDeadline(time.Now().Add(connTimeout + time.Millisecond*10))
	if err != nil {
		return
	}

	if slf.Ip == "" {
		slf.Ip = conn.LocalAddr().String()
	}

	if slf.Host == "" {
		var myGlobalIP string
		myIP, _, err := net.SplitHostPort(strings.TrimLeft(slf.Ip, "socks://"))
		myGlobalIP, ok := slf.MapIp[myIP]
		if !ok {
			myGlobalIP = myIP
		}
		names, err := net.LookupAddr(myGlobalIP)
		if err != nil && len(names) < 1 {
			return nil, err
		}
		debug("LookUp ", myGlobalIP, " this result ", names[0], "\n")
		slf.Host = names[0]
	}

	return
}
