// Project Gonder.
// Author Supme
// Copyright Supme 2016
// License http://opensource.org/licenses/MIT MIT License
//
//  THE SOFTWARE AND DOCUMENTATION ARE PROVIDED "AS IS" WITHOUT WARRANTY OF
//  ANY KIND, EITHER EXPRESSED OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
//  IMPLIED WARRANTIES OF MERCHANTABILITY AND/OR FITNESS FOR A PARTICULAR
//  PURPOSE.
//
// Please see the License.txt file for more information.
//
package mailer

import (
	"github.com/supme/gonder/models"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
//	_ "github.com/eaigner/dkim"
	"golang.org/x/net/proxy"
	"io/ioutil"
	"net"
	"net/smtp"
	"net/http"
	"strings"
	"golang.org/x/net/idna"
	"time"
)

type (
	MailData struct {
		Iface 	     string
		Host         string
		From_email   string
		From_name    string
		To_email     string
		To_name      string
		Extra_header string
		Subject      string
		Html         string
		Attachments  []Attachment
		conn 		 net.Conn
		socksConn	     	 proxy.Dialer
		netConn            net.Dialer
	}

	Attachment struct {
		Location string
		Name     string
	}
)

func (m *MailData) Send() error {

	// trim space
	m.To_email = strings.TrimSpace(m.To_email)
	// punycode convert
	splitEmail := strings.Split(m.To_email, "@")
	if len(splitEmail) != 2 {
		return errors.New(fmt.Sprintf("Bad email"))
	}
	domain, err := idna.ToASCII(splitEmail[1])
	if err != nil {
		return errors.New(fmt.Sprintf("Domain name failed: %v\r\n", err))
	}
	m.To_email = strings.Split(m.To_email, "@")[0] + "@" + domain

/*
	s := new(connect)
	conn, err := s.up(m.Iface, domain)
	if err != nil {
		return err
	}
*/

	if m.Iface == "" {
		// default interface
		m.netConn = net.Dialer{}
	} else {
		if m.Iface[0:8] == "socks://" {
			m.Iface = m.Iface[8:]
			var err error
			m.socksConn, err = proxy.SOCKS5("tcp", m.Iface, nil, proxy.FromEnvironment())
			if err != nil {
				return  err
			}
		} else {
			connectAddr := net.ParseIP(m.Iface)
			tcpAddr := &net.TCPAddr{
				IP: connectAddr,
			}
			m.netConn = net.Dialer{LocalAddr: tcpAddr}
		}
	}
	start := time.Now()
	//record, err := net.LookupMX(c.domain)
	record, err := models.DomainGetMX(domain)
	lookupTime := time.Since(start)
	start = time.Now()

	var serverMx string
	for i := range record {
		smx := net.JoinHostPort(record[i].Host, "25")
		if m.socksConn != nil {
			m.conn, err = m.socksConn.Dial("tcp", smx)
		} else {
			m.conn, err = m.netConn.Dial("tcp", smx)
		}
		if err == nil {
			serverMx = record[i].Host
			connTime := time.Since(start)
			fmt.Printf("Connect time to %s %s. Lookup time %s.\n", domain, connTime, lookupTime)
			break
		}
	}

	c, err := smtp.NewClient(m.conn, serverMx)
	if err != nil {
		return errors.New(fmt.Sprintf("%v (NewClient)\r\n", err))
	}

	if err := c.Hello(m.Host); err != nil {
		return errors.New(fmt.Sprintf("%v (Hello)\r\n", err))
	}

	// Set the sender and recipient first
	if err := c.Mail(m.From_email); err != nil {
		return errors.New(fmt.Sprintf("%v (Mail)\r\n", err))
	}

	if err := c.Rcpt(m.To_email); err != nil {
		return errors.New(fmt.Sprintf("%v (Rcpt)\r\n", err))
	}

	//dkim.New()

	w, err := c.Data()
	if err != nil {
		return errors.New(fmt.Sprintf("%v (Data)\r\n", err))
	}
	_, err = fmt.Fprint(w, m.makeMail())
	if err != nil {
		return errors.New(fmt.Sprintf("%v (wData)\r\n", err))
	}

	err = w.Close()
	if err != nil {
		return errors.New(fmt.Sprintf("%v (Close)\r\n", err))
	}

	return c.Quit()
}

func (m *MailData) makeMail() string {
	marker := makeMarker()

	var msg bytes.Buffer

	if m.From_name == "" {
		msg.WriteString(`From: ` + m.From_email + "\r\n")
	} else {
		msg.WriteString(`From: "` + encodeRFC2047(m.From_name) + `" <` + m.From_email + `>` +"\r\n")
	}

	if m.To_name == "" {
		msg.WriteString(`To: ` + m.To_email + "\r\n")
	} else {
		msg.WriteString(`To: "` + encodeRFC2047(m.To_name) + `" <` + m.To_email + `>` + "\r\n")
	}

	// -------------- head ----------------------------------------------------------

	msg.WriteString("Subject: " + encodeRFC2047(m.Subject) + "\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: multipart/mixed;\r\n	boundary=\"" + marker + "\"\r\n")
	msg.WriteString("X-Mailer: Gonder v" + models.Config.Version + "\r\n")
	msg.WriteString(m.Extra_header + "\r\n")
	// ------------- /head ---------------------------------------------------------

	// ------------- body ----------------------------------------------------------
	msg.WriteString("\r\n--" + marker + "\r\nContent-Type: text/html; charset=\"utf-8\"\r\nContent-Transfer-Encoding: 8bit\r\n\r\n")
	msg.WriteString(m.Html)
	// ------------ /body ---------------------------------------------------------

	// ----------- attachments ----------------------------------------------------
	for _, file := range m.Attachments {

		msg.WriteString("\r\n--" + marker)
		//read and encode attachment
		content, err := ioutil.ReadFile(file.Location + file.Name)
		if err != nil {
			fmt.Println(err)
		}
		encoded := base64.StdEncoding.EncodeToString(content)

		//part 3 will be the attachment
		msg.WriteString(fmt.Sprintf("\r\nContent-Type: %s;\r\n	name=\"%s\"\r\nContent-Transfer-Encoding: base64\r\nContent-Disposition: attachment;\r\n	filename=\"%s\"\r\n\r\n", http.DetectContentType(content), file.Name, file.Name))
		//split the encoded file in lines (doesn't matter, but low enough not to hit a max limit)
		lineMaxLength := 500
		nbrLines := len(encoded) / lineMaxLength
		for i := 0; i < nbrLines; i++ {
			msg.WriteString(encoded[i*lineMaxLength:(i+1)*lineMaxLength] + "\n")
		}

		//append last line in buffer
		msg.WriteString(encoded[nbrLines*lineMaxLength:])
		msg.WriteString("\r\n")

	}
	// ----------- /attachments ---------------------------------------------------

	return msg.String()
}

func makeMarker() string {
	b := make([]byte, 30)
	rand.Read(b)
	en := base64.StdEncoding // or URLEncoding
	d := make([]byte, en.EncodedLen(len(b)))
	en.Encode(d, b)
	return string(d)
}

func encodeRFC2047(s string) string {
	// use code from net/mail for rfc2047 encode any string
	// UTF-8 "Q" encoding
	b := bytes.NewBufferString("=?utf-8?q?")
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == ' ':
			b.WriteByte('_')
		case isVchar(c) && c != '=' && c != '?' && c != '_':
			b.WriteByte(c)
		default:
			fmt.Fprintf(b, "=%02X", c)
		}
	}
	b.WriteString("?= ")
	return b.String()
}

// isVchar returns true if c is an RFC 5322 VCHAR character.
func isVchar(c byte) bool {
	// Visible (printing) characters.
	return '!' <= c && c <= '~'
}
