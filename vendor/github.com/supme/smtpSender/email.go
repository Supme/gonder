package smtpSender

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/smtp"
	"regexp"
	"strings"
	"time"
)

// Email struct
type Email struct {
	// ID is id for return result
	ID string
	// From emailField has format
	// example
	//  "Name <emailField@domain.tld>"
	//  "<emailField@domain.tld>"
	//  "emailField@domain.tld"
	From                            string
	fromName, fromEmail, fromDomain string
	// To emailField has format as From
	To                        string
	toName, toEmail, toDomain string
	// ResultFunc exec after send emil
	ResultFunc func(Result)
	// WriteCloser email body data writer function
	WriteCloser func(io.WriteCloser) error
}

// Result struct for return send emailField result
type Result struct {
	ID       string
	Duration time.Duration
	Err      error
}

// SMTPserver use for send email from server
type SMTPserver struct {
	Host     string
	Port     int
	Username string
	Password string
}

// Send sending this email
func (e *Email) Send(connect *Connect, server *SMTPserver) {
	if connect == nil {
		connect = &Connect{}
	}

	var (
		client *smtp.Client
		auth   smtp.Auth
		err    error
	)
	start := time.Now()
	e.parseEmail()

	if server == nil {
		client, err = connect.newClient(e.toDomain, true)
	} else {
		auth = smtp.PlainAuth(
			"",
			server.Username,
			server.Password,
			server.Host,
		)
		connect.SetSMTPport(server.Port)
		client, err = connect.newClient(server.Host, false)
	}
	if err != nil {
		e.ResultFunc(Result{ID: e.ID, Err: fmt.Errorf("421 %v", err), Duration: time.Now().Sub(start)})
		return
	}

	err = e.send(auth, client)
	e.ResultFunc(Result{ID: e.ID, Err: err, Duration: time.Now().Sub(start)})
	return
}

func (e *Email) send(auth smtp.Auth, client *smtp.Client) error {
	var (
		err              error
		testHookStartTLS func(*tls.Config)
	)
	if ok, _ := client.Extension("STARTTLS"); ok {
		config := &tls.Config{ServerName: e.toDomain, InsecureSkipVerify: true}
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

	if err := client.Mail(e.from()); err != nil {
		return err
	}

	if err := client.Rcpt(e.to()); err != nil {
		return err
	}

	defer func() {
		client.Quit()
		client.Close()
	}()

	w, err := client.Data()
	if err != nil {
		return err
	}

	return e.WriteCloser(w)
}

func (e *Email) from() string {
	return e.fromEmail + "@" + e.fromDomain
}

func (e *Email) to() string {
	return e.toEmail + "@" + e.toDomain
}

func (e *Email) parseEmail() (err error) {
	e.fromName, e.fromEmail, e.fromDomain, err = splitEmail(e.From)
	if err != nil {
		return fmt.Errorf("Field From has %s", err)
	}
	e.toName, e.toEmail, e.toDomain, err = splitEmail(e.To)
	if err != nil {
		return fmt.Errorf("Field To has %s", err)
	}
	return
}

var (
	splitEmailFullStringRe = regexp.MustCompile(`(.+)<(.+)@(.+\..{2,9})>`)
	splitEmailOnlyStringRe = regexp.MustCompile(`<(.+)@(.+\..{2,9})>`)
	splitEmailRe           = regexp.MustCompile(`(.+)@(.+\..{2,9})`)
)

func splitEmail(e string) (name, email, domain string, err error) {
	s := strings.TrimSpace(e)
	if m := splitEmailFullStringRe.FindStringSubmatch(s); m != nil && len(m) == 4 {
		name = strings.TrimSpace(m[1])
		email = strings.ToLower(strings.TrimSpace(m[2]))
		domain = strings.TrimRight(strings.ToLower(strings.TrimSpace(m[3])), ".")
	} else if m := splitEmailOnlyStringRe.FindStringSubmatch(s); m != nil && len(m) == 3 {
		email = strings.ToLower(strings.TrimSpace(m[1]))
		domain = strings.TrimRight(strings.ToLower(strings.TrimSpace(m[2])), ".")
	} else if m := splitEmailRe.FindStringSubmatch(s); m != nil && len(m) == 3 {
		email = strings.ToLower(strings.TrimSpace(m[1]))
		domain = strings.TrimRight(strings.ToLower(strings.TrimSpace(m[2])), ".")
	} else {
		err = fmt.Errorf("bad email format")
	}
	return
}
