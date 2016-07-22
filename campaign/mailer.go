package campaign

import (
	"github.com/supme/gonder/models"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
//	"github.com/eaigner/dkim"
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
	// Mail parameters and data
	//
	// Iface: send email from selected interface
	//       "": default interface
	//       "12.34.56.78": ip address interface
	//       "socks://12.34.56.78:1234": send from remote socks server
	//       "1,4,6,8,9": id`s for send rotate by selected profile (set Host: "group")
	// Host: host name of selected interface ip
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
		conn         net.Conn
		socksConn    proxy.Dialer
		netConn      net.Dialer
	}

	// Attach file to mail
	//
	// Location: folder
	// Name: file
	Attachment struct {
		Location string
		Name     string
	}
)

// Send email
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
		return errors.New(fmt.Sprintf("Domain name failed: %v", err))
	}
	m.To_email = strings.Split(m.To_email, "@")[0] + "@" + domain

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
	if err != nil {
		return err
	}
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
			fmt.Printf("Connect time to %s %s. Lookup time %s.\n\r", domain, connTime, lookupTime)
			break
		}
	}
	if err != nil {
		return err
	}
	defer m.conn.Close()

	// 5 minute by RFC
	m.conn.SetDeadline(time.Now().Add(5 * time.Minute))

	c, err := smtp.NewClient(m.conn, serverMx)
	if err != nil {
		return errors.New(fmt.Sprintf("%v (NewClient)", err))
	}

	if err := c.Hello(m.Host); err != nil {
		return errors.New(fmt.Sprintf("%v (Hello)", err))
	}

	// Set the sender and recipient first
	if err := c.Mail(m.From_email); err != nil {
		return errors.New(fmt.Sprintf("%v (Mail)", err))
	}

	if err := c.Rcpt(m.To_email); err != nil {
		return errors.New(fmt.Sprintf("%v (Rcpt)", err))
	}

	//dkim.New()

	w, err := c.Data()
	if err != nil {
		return errors.New(fmt.Sprintf("%v (Data)", err))
	}
	_, err = fmt.Fprint(w, m.makeMail())
	if err != nil {
		return errors.New(fmt.Sprintf("%v (SendData)", err))
	}

	err = w.Close()
	if err != nil {
		return errors.New(fmt.Sprintf("%v (Close)", err))
	}

	return c.Quit()
}

func (m *MailData) makeMail() string {
	marker := makeMarker()

	var msg bytes.Buffer

	if m.From_name == "" {
		msg.WriteString(`From: ` + m.From_email + "\n")
	} else {
		msg.WriteString(`From: "` + encodeRFC2047(m.From_name) + `" <` + m.From_email + `>` +"\n")
	}

	if m.To_name == "" {
		msg.WriteString(`To: ` + m.To_email + "\n")
	} else {
		msg.WriteString(`To: "` + encodeRFC2047(m.To_name) + `" <` + m.To_email + `>` + "\n")
	}

	// -------------- head ----------------------------------------------------------

	msg.WriteString("Subject: " + encodeRFC2047(m.Subject) + "\n")
	msg.WriteString("MIME-Version: 1.0\n")
	msg.WriteString("X-Mailer: Gonder v" + models.Config.Version + "\n")
	msg.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\n")
	msg.WriteString("Content-Type: multipart/mixed;\n	boundary=\"" + marker + "\"\n")
	msg.WriteString(m.Extra_header + "\n\n")
	// ------------- /head ---------------------------------------------------------

	// ------------- body ----------------------------------------------------------
	msg.WriteString("\n--" + marker + "\nContent-Type: text/html; charset=\"utf-8\"\nContent-Transfer-Encoding: 8bit\n\n")
	msg.WriteString(m.Html)
	// ------------ /body ---------------------------------------------------------

	// ----------- attachments ----------------------------------------------------
	for _, file := range m.Attachments {

		msg.WriteString("\n--" + marker)
		//read and encode attachment
		content, err := ioutil.ReadFile(file.Location + file.Name)
		if err != nil {
			fmt.Println(err)
		}
		encoded := base64.StdEncoding.EncodeToString(content)

		//part 3 will be the attachment
		msg.WriteString(fmt.Sprintf("\nContent-Type: %s;\n	name=\"%s\"\nContent-Transfer-Encoding: base64\nContent-Disposition: attachment;\n	filename=\"%s\"\n\n", http.DetectContentType(content), file.Name, file.Name))
		//split the encoded file in lines (doesn't matter, but low enough not to hit a max limit)
		lineMaxLength := 500
		nbrLines := len(encoded) / lineMaxLength
		for i := 0; i < nbrLines; i++ {
			msg.WriteString(encoded[i*lineMaxLength:(i+1)*lineMaxLength] + "\n")
		}

		//append last line in buffer
		msg.WriteString(encoded[nbrLines*lineMaxLength:])
		msg.WriteString("\n")

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
