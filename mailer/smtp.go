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
	var wg sync.WaitGroup

	campaign, err := Db.Prepare("SELECT t1.`id`,t1.`from`,t1.`from_name`,t2.`iface`,t2.`host`,t2.`stream`,t2.`delay` FROM `campaign` t1 INNER JOIN `interface` t2 ON t2.`id`=t1.`interface_id` WHERE NOW() BETWEEN t1.`start_time` AND t1.`end_time`")
	checkErr(err)
	defer campaign.Close()

	attachment, err := Db.Prepare("SELECT `path`, `file` FROM attachment WHERE campaign_id=?")
	checkErr(err)
	defer attachment.Close()

	// Select recipient
	recipient, err := Db.Prepare("SELECT `id`, `email`, `name` FROM recipient WHERE campaign_id=? AND status IS NULL LIMIT ?")
	checkErr(err)
	defer recipient.Close()

	for {
		camp, err := campaign.Query()
		checkErr(err)
		defer camp.Close()

		for camp.Next() {
			err = camp.Scan(&campaignId, &data.From, &data.From_name, &iface, &data.Host, &stream, &delay)
			checkErr(err)

			err = setIface(iface)
			checkErr(err)

			attach, err := attachment.Query(campaignId)
			checkErr(err)
			defer attach.Close()

			data.Attachments = nil

			var location string
			var name string
			for attach.Next() {
				err = attach.Scan(&location, &name)
				checkErr(err)
				data.Attachments = append(data.Attachments, attachmentData{Location: location, Name: name})
			}

			recip, err := recipient.Query(campaignId, stream)
			checkErr(err)
			defer recip.Close()

			for recip.Next() {

				err = recip.Scan(&recipientId, &data.To, &data.To_name)
				checkErr(err)

				d := getMailMessage(campaignId, recipientId)
				data.Subject = d.Subject
				data.Html = d.Body
				data.Extra_header = "List-Unsubscribe: " + getUnsubscribeUrl(campaignId, recipientId) + "\r\nPrecedence: bulk\r\n"

				// Send mail
				wg.Add(1)
				go func(id string, data mailData) {
					var r string

					res := sendMail(data)
					if res == nil {
						r = "Ok"
					} else {
						r = res.Error()
					}
					log.Printf("Send mail for recipient id %s is %s", id, r)
					rows, err := Db.Query("UPDATE recipient SET status=? WHERE id=?", r, id)
					checkErr(err)
					defer rows.Close()
					defer wg.Done()
				}(recipientId, data)
			}
			wg.Wait()
		}
		time.Sleep(time.Second + time.Duration(delay)*time.Second)
	}
}

func setIface(iface string) error {
	if iface == "" {
		// default interface
		n = net.Dialer{}
	} else {
/*		_, _, err := net.SplitHostPort(iface)
		if err == nil {
			err = socksConnect(iface)
		} else {
			err = netConnect(iface)
		}
		if err != nil {
			return err
		}
*/
// ToDo socks proxy connect
		connectAddr := net.ParseIP(iface)
		tcpAddr := &net.TCPAddr{
			IP: connectAddr,
		}
		n = net.Dialer{LocalAddr: tcpAddr}
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

func netConnect(iface string) error {
	var err error
	ief, err := net.InterfaceByName(iface)
	if err != nil {
		return errors.New(fmt.Sprintf("Set interface failed: %v\r\n", err))
	}
	addrs, err := ief.Addrs()
	if err != nil {
		return errors.New(fmt.Sprintf("Get interface address failed: %v\r\n", err))
	}
	tcpAddr := &net.TCPAddr{
		IP: addrs[0].(*net.IPNet).IP,
	}
	n = net.Dialer{LocalAddr: tcpAddr}
	return nil
}

func sendMail(m mailData) error {
	var smx string
	var mx []*net.MX
	var conn net.Conn

	//ToDo cache MX servers and punycode
	domain, err := idna.ToASCII(strings.Split(m.To, "@")[1])
	if err != nil {
		return errors.New(fmt.Sprintf("Domain name failed: %v\r\n", err))
	}
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
		return errors.New(fmt.Sprintf("Connect: %v\r\n", err))
	}

	host, _, _ := net.SplitHostPort(smx)
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return errors.New(fmt.Sprintf("New client failed: %v\r\n", err))
	}

	if err := c.Hello(m.Host); err != nil {
		return errors.New(fmt.Sprintf("Hello: %v\r\n", err))
	}

	// Set the sender and recipient first
	if err := c.Mail(m.From); err != nil {
		return errors.New(fmt.Sprintf("Mail: %v\r\n", err))
	}

	if err := c.Rcpt(m.To); err != nil {
		return errors.New(fmt.Sprintf("Rcpt: %v\r\n", err))
	}

	msg := makeMail(m)

	//dkim.New()

	w, err := c.Data()
	if err != nil {
		return errors.New(fmt.Sprintf("Data: %v", err))
	}
	_, err = fmt.Fprintf(w, msg)
	if err != nil {
		return errors.New(fmt.Sprintf("Fprintf: %v", err))
	}

	err = w.Close()
	if err != nil {
		return errors.New(fmt.Sprintf("Close: %v", err))
	}

	// Send the QUIT command and close the connection.
	err = c.Quit()
	if err != nil {
		return errors.New(fmt.Sprintf("Quit: %v", err))
	}

	return nil
}

func makeMail(m mailData) (msg string) {
	marker := makeMarker()

	msg = ""

	// -------------- head ----------------------------------------------------------
	msg += "From: " + encodeRFC2047(m.From_name) + " <" + m.From + ">" + "\r\n"
	msg += "To: " + encodeRFC2047(m.To_name) + " <" + m.To + ">" + "\r\n"
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
