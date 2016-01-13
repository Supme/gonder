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
)

func Sender() {

	var data mailData

	campaign, err := Db.Prepare("SELECT t1.`id`,t1.`from`,t1.`from_name`,t2.`iface`,t2.`host`,t2.`stream`,t2.`delay` FROM `campaign` t1 INNER JOIN `interface` t2 ON t2.`id`=t1.`interface_id` WHERE NOW() BETWEEN t1.`start_time` AND t1.`end_time`")
	checkErr(err)
	defer campaign.Close()

	for {
		camp, err := campaign.Query()
		checkErr(err)
		defer camp.Close()

		var wc sync.WaitGroup
		for camp.Next() {
			err = camp.Scan(&campaignId, &data.From, &data.From_name, &iface, &data.Host, &stream, &delay)
			checkErr(err)

			wc.Add(1)
			go func(cId string, cData mailData, cIface string, cStream int, cDelay int) {

				attachment, err := Db.Prepare("SELECT `path`, `file` FROM attachment WHERE campaign_id=?")
				checkErr(err)
				defer attachment.Close()

				err = setIface(cIface)
				checkErr(err)

				attach, err := attachment.Query(cId)
				checkErr(err)
				defer attach.Close()

				cData.Attachments = nil

				var location string
				var name string
				for attach.Next() {
					err = attach.Scan(&location, &name)
					checkErr(err)
					cData.Attachments = append(cData.Attachments, attachmentData{Location: location, Name: name})
				}

				recipient, err := Db.Prepare("SELECT `id`, `email`, `name` FROM recipient WHERE campaign_id=? AND status IS NULL LIMIT ?")
				checkErr(err)
				defer recipient.Close()

				recip, err := recipient.Query(cId, cStream)
				checkErr(err)
				defer recip.Close()

				var wr sync.WaitGroup
				for recip.Next() {

					var rId string
					err = recip.Scan(&rId, &cData.To, &cData.To_name)
					checkErr(err)

					wr.Add(1)
					go func(cid string, rid string, rData mailData) {

						d := getMailMessage(cid, rid)
						rData.Subject = d.Subject
						rData.Html = d.Body
						rData.Extra_header = "List-Unsubscribe: " + getUnsubscribeUrl(cid, rid) + "\r\nPrecedence: bulk\r\n"

						// Send mail
						res := sendMail(rData)

						var r string
						if res == nil {
							r = "Ok"
						} else {
							r = res.Error()
						}

						log.Printf("Send mail for recipient id %s email %s is %s", rid, rData.To, r)
						rows, err := Db.Query("UPDATE recipient SET status=?, date=NOW() WHERE id=?", r, rid)
						checkErr(err)
						defer rows.Close()
						defer wr.Done()
					}(cId, rId, cData)
				}
				wr.Wait()
				time.Sleep(time.Second + time.Duration(cDelay)*time.Second)
				defer wc.Done()
			}(campaignId, data, iface, stream, delay)
		}
		wc.Wait()
	}
}

func setIface(iface string) error {
	if iface == "" {
		// default interface
		n = net.Dialer{}
	} else {
		if iface[0:8] == "socks://" {
			iface = iface[8:]
			err := socksConnect(iface)
			if err != nil {
				return err
			}
		} else {
			connectAddr := net.ParseIP(iface)
			tcpAddr := &net.TCPAddr{
				IP: connectAddr,
			}
			n = net.Dialer{LocalAddr: tcpAddr}
		}
	}
	return nil
}

func socksConnect(socks string) error {
	var err error
	s, err = proxy.SOCKS5("tcp", socks, nil, proxy.FromEnvironment())
	if err != nil {
		return errors.New(fmt.Sprintf("Socks failed: %v\r\n", err))
	}
	return nil
}

func sendMail(m mailData) error {
	var smx string
	var mx []*net.MX
	var conn net.Conn

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
			if s != nil {
				conn, err = s.Dial("tcp", smx)
			} else {
				conn, err = n.Dial("tcp", smx)
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

	msg := makeMail(m)

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

func makeMail(m mailData) (msg string) {
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

func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}
