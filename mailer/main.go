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
	"strconv"
)

var (
	Send bool
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
		Attachments  []attachmentData
		s	     proxy.Dialer
		n            net.Dialer
	}


	attachmentData struct {
		Location string
		Name     string
	}


    campaignData struct {
		id, from, from_name, subject, body, iface, host, send_unsubscribe string
		attachments []attachmentData
		stream      int
		delay       int
	}

	recipientData struct {
		id, to, to_name	string
	}
)

func Run() {

	campaign, err := models.Db.Prepare("SELECT t1.`id`,t1.`from`,t1.`from_name`,t1.`subject`,t1.`body`,t2.`iface`,t2.`host`,t2.`stream`,t2.`delay`, t1.`send_unsubscribe`  FROM `campaign` t1 INNER JOIN `profile` t2 ON t2.`id`=t1.`profile_id` WHERE NOW() BETWEEN t1.`start_time` AND t1.`end_time`")
	checkErr(err)
	defer campaign.Close()

	for {
		camp, err := campaign.Query()
		checkErr(err)
		defer camp.Close()

		var wc sync.WaitGroup
		for camp.Next() {

			var id, from, from_name, subject, body, iface, host, send_unsubscribe string
			var stream, delay int

			err = camp.Scan(&id, &from, &from_name, &subject, &body, &iface, &host, &stream, &delay, &send_unsubscribe)
			checkErr(err)

			wc.Add(1)
			go func(c campaignData) {
				attachment, err := models.Db.Prepare("SELECT `path`, `file` FROM attachment WHERE campaign_id=?")
				checkErr(err)
				defer attachment.Close()

				attach, err := attachment.Query(c.id)
				checkErr(err)
				defer attach.Close()

				c.attachments = nil

				var location string
				var name string
				for attach.Next() {
					err = attach.Scan(&location, &name)
					checkErr(err)
					c.attachments = append(c.attachments, attachmentData{Location: location, Name: name})
				}

				sendCampaign(c)

				time.Sleep(time.Second + time.Duration(c.delay) * time.Second)
				defer wc.Done()

			}(campaignData{
				id: id,
				from: from,
				from_name: from_name,
				subject: subject,
				body: body,
				iface: iface,
				host: host,
				stream: stream,
				delay: delay,
				send_unsubscribe: send_unsubscribe,
			})
		}
		wc.Wait()
		time.Sleep(1 * time.Second) // easy with database
	}
}

func sendCampaign(c campaignData) {

	recipient, err := models.Db.Prepare("SELECT `id`, `email`, `name` FROM `recipient` WHERE campaign_id=? AND status IS NULL LIMIT ?")
	checkErr(err)
	defer recipient.Close()

	// stream send
	recipientCount := 1
	for recipientCount > 0 {

		recip, err := recipient.Query(c.id, c.stream)
		checkErr(err)
		defer recip.Close()

		var wr sync.WaitGroup
		for recip.Next() {
			r := new(recipientData)
			err = recip.Scan(&r.id, &r.to, &r.to_name)
			checkErr(err)

			var unsubscribeCount int
			models.Db.QueryRow("SELECT COUNT(*) FROM `unsubscribe` t1 INNER JOIN `campaign` t2 ON t1.group_id = t2.group_id WHERE t2.id = ? AND t1.email = ?", c.id, r.to).Scan(&unsubscribeCount)

			if unsubscribeCount == 0 || c.send_unsubscribe == "y" {
				wr.Add(1)
				go func(cData *campaignData, rData *recipientData ) {
					sendRecipient(cData, rData)
					defer wr.Done()
				}(&c, r)
			} else {
				log.Printf("Recipient id %s email %s is unsubscribed", r.id, r.to)
				statSend(r.id, "Unsubscribed")
			}
		}

		wr.Wait()
		models.Db.QueryRow("SELECT COUNT(*) FROM `unsubscribe` WHERE id = ?", c.id).Scan(&recipientCount)
	}

	// send one per cycle
	//SELECT DISTINCT r.`id`, r.`email`, r.`name`, r.`status` FROM `recipient` as r,`status` as s WHERE r.`campaign_id`=15 AND s.`bounce_id`=2 AND UPPER(`r`.`status`) LIKE CONCAT("%",s.`pattern`,"%")
	recipient, err = models.Db.Prepare("SELECT DISTINCT r.`id`, r.`email`, r.`name` FROM `recipient` as r,`status` as s WHERE r.`campaign_id`=? AND s.`bounce_id`=2 AND UPPER(`r`.`status`) LIKE CONCAT(\"%\",s.`pattern`,\"%\")")
	checkErr(err)
	defer recipient.Close()

	for i := 0; i < 3; i++ {
		time.Sleep(900 * time.Second)

		recip, err := recipient.Query(c.id)
		checkErr(err)
		defer recip.Close()

		for recip.Next() {
			r := new(recipientData)
			err = recip.Scan(&r.id, &r.to, &r.to_name)
			checkErr(err)

			sendRecipient(&c, r)
		}
	}
}

func sendRecipient(cData *campaignData, rData *recipientData )  {

		data := new(MailData)
		data.Iface = cData.iface
		data.From = cData.from
		data.From_name = cData.from_name
		data.Host = cData.host
		data.Attachments = cData.attachments
		data.To = rData.to
		data.To_name = rData.to_name

		var rs string
		d, e := models.MailMessage(cData.id, rData.id, cData.subject, cData.body)
		if e == nil {
			data.Subject = d.Subject
			data.Html = d.Body
			data.Extra_header = "List-Unsubscribe: " + models.UnsubscribeUrl(cData.id, rData.id) + "\r\nPrecedence: bulk\r\n"
			data.Extra_header += "Message-ID: <" + strconv.FormatInt(time.Now().Unix(), 10) + cData.id + "." + rData.id +"@" + cData.host + ">" + "\r\n"
			var res error
			if Send {
				// Send mail
				res = data.Send()
			} else {
				res = errors.New("Test send")
			}

			if res == nil {
				rs = "Ok"
			} else {
				rs = res.Error()
			}
		} else {
			rs = "Error " + e.Error()
		}

		log.Printf("Send mail for recipient id %s email %s is %s", rData.id, data.To, rs)
		statSend(rData.id, rs)
}

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
	msg += "X-Mailer: " + models.Version + "\r\n"
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

func statSend(id, result string) {
	models.Db.Exec("UPDATE recipient SET status=?, date=NOW() WHERE id=?", result, id)
}

func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}
