package mailer

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	_ "github.com/eaigner/dkim"
	"golang.org/x/net/proxy"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"sync"
	"time"
	"golang.org/x/net/idna"
	"database/sql"
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

type MailData struct {
	Iface 	     string
	Host         string
	From         string
	From_name    string
	To           string
	To_name      string
	Extra_header string
	Subject      string
	Html         string
	Attachments  []attachmentData
	s	     proxy.Dialer
	n            net.Dialer
}

func (m *MailData) SendMail() error {
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
	// punycode convert
	domain, err := idna.ToASCII(strings.Split(m.To, "@")[1])
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
				defer conn.Close()
				break
			}
		}
	}
	if err != nil {
		return err
	}

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
	_, err = fmt.Fprintf(w, msg)
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
	msg += "X-Mailer: Gonder 0.2\r\n"
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



func Sender() {

	type campaignData struct {
		id, from, from_name, iface, host string
		attachments []attachmentData
		stream      int
		delay       int
	}

	type recipientData struct {
		id, to, to_name	string
	}

	campaign, err := Db.Prepare("SELECT t1.`id`,t1.`from`,t1.`from_name`,t2.`iface`,t2.`host`,t2.`stream`,t2.`delay` FROM `campaign` t1 INNER JOIN `interface` t2 ON t2.`id`=t1.`interface_id` WHERE NOW() BETWEEN t1.`start_time` AND t1.`end_time`")
	checkErr(err)
	defer campaign.Close()

	for {
		camp, err := campaign.Query()
		checkErr(err)
		defer camp.Close()

		var wc sync.WaitGroup
		for camp.Next() {

			c := new(campaignData)

			err = camp.Scan(&c.id, &c.from, &c.from_name, &c.iface, &c.host, &c.stream, &c.delay)
			checkErr(err)

			wc.Add(1)
			go func(cPart *campaignData) {
				attachment, err := Db.Prepare("SELECT `path`, `file` FROM attachment WHERE campaign_id=?")
				checkErr(err)
				defer attachment.Close()

				attach, err := attachment.Query(cPart.id)
				checkErr(err)
				defer attach.Close()

				cPart.attachments = nil

				var location string
				var name string
				for attach.Next() {
					err = attach.Scan(&location, &name)
					checkErr(err)
					cPart.attachments = append(cPart.attachments, attachmentData{Location: location, Name: name})
				}

				recipient, err := Db.Prepare("SELECT `id`, `email`, `name` FROM recipient WHERE campaign_id=? AND status IS NULL LIMIT ?")
				checkErr(err)
				defer recipient.Close()

				recip, err := recipient.Query(cPart.id, cPart.stream)
				checkErr(err)
				defer recip.Close()

				var wr sync.WaitGroup
				for recip.Next() {

					r := new(recipientData)

					err = recip.Scan(&r.id, &r.to, &r.to_name)
					checkErr(err)

					wr.Add(1)
					go func(cData *campaignData, rData *recipientData ) {
						data := new(MailData)
						data.Iface = cData.iface
						data.From = cData.from
						data.From_name = cData.from_name
						data.Host = cData.host
						data.Attachments = cData.attachments

						d := getMailMessage(cData.id, rData.id)
						data.Subject = d.Subject
						data.Html = d.Body
						data.Extra_header = "List-Unsubscribe: " + getUnsubscribeUrl(cData.id, rData.id) + "\r\nPrecedence: bulk\r\n"

						data.To = r.to
						data.To_name = r.to_name

						// Send mail
						res := data.SendMail()

						var rs string
						if res == nil {
							rs = "Ok"
						} else {
							rs = res.Error()
						}

						log.Printf("Send mail for recipient id %s email %s is %s", rData.id, data.To, rs)
						rows, err := Db.Query("UPDATE recipient SET status=?, date=NOW() WHERE id=?", rs, rData.id)
						checkErr(err)
						defer rows.Close()
						defer wr.Done()
					}(cPart, r)
				}
				wr.Wait()
				time.Sleep(time.Second + time.Duration(c.delay)*time.Second)
				defer wc.Done()
			}(c)
		}
		wc.Wait()
		time.Sleep(15 * time.Second) // easy with database
	}
}


func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}
