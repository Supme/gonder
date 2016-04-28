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
	"net/http"
	"net/smtp"
	"strings"
	"golang.org/x/net/idna"
)

type (
	MailData struct {
		Iface 	     string
		Host         string
		From         string
		From_name    string
		To           string
		To_name      string
		Extra_header string
		Subject      string
		Html         string
		Attachments  []Attachment
		s	     	 proxy.Dialer
		n            net.Dialer
	}

	Attachment struct {
		Location string
		Name     string
	}
)

func (m *MailData) Send() error {
	var smx string
	var mx []*net.MX
	var conn net.Conn

	if m.Iface == "" {
		// default interface
		m.n = net.Dialer{}
	} else {
		if m.Iface[0:8] == "socks://" {
			m.Iface = m.Iface[8:]
			var err error
			m.s, err = proxy.SOCKS5("tcp", m.Iface, nil, proxy.FromEnvironment())
			if err != nil {
				return err
			}
		} else {
			connectAddr := net.ParseIP(m.Iface)
			tcpAddr := &net.TCPAddr{
				IP: connectAddr,
			}
			m.n = net.Dialer{LocalAddr: tcpAddr}
		}
	}

	//ToDo cache MX servers
	// trim space
	m.To = strings.TrimSpace(m.To)
	// punycode convert
	splitEmail := strings.Split(m.To, "@")
	if len(splitEmail) != 2 {
		return errors.New(fmt.Sprintf("Bad email"))
	}
	domain, err := idna.ToASCII(splitEmail[1])
	if err != nil {
		return errors.New(fmt.Sprintf("Domain name failed: %v\r\n", err))
	}
	m.To = strings.Split(m.To, "@")[0] + "@" + domain

	mx, err = net.LookupMX(domain)
	if err != nil {
		return errors.New(fmt.Sprintf("LookupMX failed: %v\r\n", err))
	} else {
		for i := range mx {
			smx := net.JoinHostPort(mx[i].Host, "25")
			// Set ip (from MX records) and port mail server
			if m.s != nil {
				conn, err = m.s.Dial("tcp", smx)
			} else {
				conn, err = m.n.Dial("tcp", smx)
			}
			if err == nil {
				break
			}
		}
	}
	if err != nil {
		return err
	}
	defer conn.Close()

	host, _, _ := net.SplitHostPort(smx)
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	if err := c.Hello(m.Host); err != nil {
		return err
	}

	// Set the sender and recipient first
	if err := c.Mail(m.From); err != nil {
		return err
	}

	if err := c.Rcpt(m.To); err != nil {
		return err
	}

	msg := m.makeMail()

	//dkim.New()

	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, msg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	// Send the QUIT command and close the connection.
	return c.Quit()
}

func (m *MailData) makeMail() (msg string) {
	marker := makeMarker()

	msg = ""

	if m.From_name == "" {
		msg += `From: ` + m.From + "\r\n"
	} else {
		msg += `From: "` + encodeRFC2047(m.From_name) + `" <` + m.From + `>` +"\r\n"
	}

	if m.To_name == "" {
		msg += `To: ` + m.To + "\r\n"
	} else {
		msg += `To: "` + encodeRFC2047(m.To_name) + `" <` + m.To + `>` + "\r\n"
	}

	// -------------- head ----------------------------------------------------------

	msg += "Subject: " + encodeRFC2047(m.Subject) + "\r\n"
	msg += "MIME-Version: 1.0\r\n"
	msg += "Content-Type: multipart/mixed;\r\n	boundary=\"" + marker + "\"\r\n"
	msg += "X-Mailer: " + models.Config.Version + "\r\n"
	msg += m.Extra_header + "\r\n"
	// ------------- /head ---------------------------------------------------------

	// ------------- body ----------------------------------------------------------
	msg += "\r\n--" + marker + "\r\nContent-Type: text/html; charset=\"utf-8\"\r\nContent-Transfer-Encoding: 8bit\r\n\r\n"
	msg += m.Html
	// ------------ /body ---------------------------------------------------------

	// ----------- attachments ----------------------------------------------------
	for _, file := range m.Attachments {

		msg += "\r\n--" + marker
		//read and encode attachment
		content, _ := ioutil.ReadFile(file.Location + file.Name)
		encoded := base64.StdEncoding.EncodeToString(content)

		//split the encoded file in lines (doesn't matter, but low enough not to hit a max limit)
		lineMaxLength := 500
		nbrLines := len(encoded) / lineMaxLength

		//append lines to buffer
		var buf bytes.Buffer
		for i := 0; i < nbrLines; i++ {
			buf.WriteString(encoded[i*lineMaxLength:(i+1)*lineMaxLength] + "\n")
		} //for

		//append last line in buffer
		buf.WriteString(encoded[nbrLines*lineMaxLength:])

		//part 3 will be the attachment
		msg += fmt.Sprintf("\r\nContent-Type: %s;\r\n	name=\"%s\"\r\nContent-Transfer-Encoding: base64\r\nContent-Disposition: attachment;\r\n	filename=\"%s\"\r\n\r\n%s\r\n", http.DetectContentType(content), file.Name, file.Name, buf.String())
	}
	// ----------- /attachments ---------------------------------------------------

	return
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
