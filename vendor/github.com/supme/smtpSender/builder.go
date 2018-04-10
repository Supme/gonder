package smtpSender

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"mime"
	"mime/quotedprintable"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Builder helper for create email
type Builder struct {
	From            string
	To              string
	Subject         string
	subjectFunc     func(io.Writer) error
	replyTo         string
	headers         []string
	textPlain       []byte
	textHTML        []byte
	htmlFunc        func(io.Writer) error
	textFunc        func(io.Writer) error
	textHTMLRelated []*os.File
	attachments     []*os.File
	markerGlobal    marker
	markerAlt       marker
	markerHTML      marker
	dkim            builderDKIM
}

type builderDKIM struct {
	domain     string
	selector   string
	privateKey []byte
}

// SetDKIM sign DKIM parameters
func (b *Builder) SetDKIM(domain, selector string, privateKey []byte) {
	b.dkim.domain = domain
	b.dkim.selector = selector
	b.dkim.privateKey = privateKey
}

// SetFrom email sender
func (b *Builder) SetFrom(name, email string) {
	b.From = mime.BEncoding.Encode("utf-8", name) + "<" + email + ">"
}

// SetTo email recipient
func (b *Builder) SetTo(name, email string) {
	b.To = mime.BEncoding.Encode("utf-8", name) + "<" + email + ">"
}

// SetSubject set email subject
func (b *Builder) SetSubject(text string) {
	b.Subject = text
}

// AddSubjectFunc add writer function for subject
func (b *Builder) AddSubjectFunc(f func(io.Writer) error) {
	b.subjectFunc = f
}

// AddReplyTo add Reply-To header
func (b *Builder) AddReplyTo(name, email string) {
	b.replyTo = email
}

// AddHTMLFunc add writer function for HTML
func (b *Builder) AddHTMLFunc(f func(io.Writer) error, file ...string) error {
	for i := range file {
		file, err := os.Open(file[i])
		if err != nil {
			return err
		}
		b.textHTMLRelated = append(b.textHTMLRelated, file)
	}
	b.htmlFunc = f
	return nil
}

// AddTextFunc add writer function for plain text
func (b *Builder) AddTextFunc(f func(io.Writer) error) {
	b.textFunc = f
}

// AddHeader add extra header to email
func (b *Builder) AddHeader(headers ...string) {
	for i := range headers {
		b.headers = append(b.headers, headers[i]+"\r\n")
	}
}

// AddTextHTML add text/html content with related file.
//
// Example use related file in html
//  AddTextHTML(
//  	`... <img src="cid:myImage.jpg" width="500px" height="250px" border="1px" alt="My image"/> ...`,
//  	"/path/to/attach/myImage.jpg",
//  )
func (b *Builder) AddTextHTML(html []byte, file ...string) (err error) {
	for i := range file {
		file, err := os.Open(file[i])
		if err != nil {
			return err
		}
		b.textHTMLRelated = append(b.textHTMLRelated, file)
	}
	b.textHTML = html
	return nil
}

// AddTextPlain add plain text
func (b *Builder) AddTextPlain(text []byte) {
	b.textPlain = text
}

// AddAttachment add attachment files to email
func (b *Builder) AddAttachment(file ...string) error {
	for i := range file {
		file, err := os.Open(file[i])
		if err != nil {
			return err
		}
		b.attachments = append(b.attachments, file)
	}
	return nil
}

// Email return Email struct with render function
func (b *Builder) Email(id string, resultFunc func(Result)) Email {
	email := new(Email)
	email.ID = id
	email.From = b.From
	email.To = b.To
	email.ResultFunc = resultFunc
	email.WriteCloser = func(w io.WriteCloser) (err error) {
		defer w.Close()
		if b.dkim.domain == "" {
			err = b.builder(w)
			return err
		}
		buf := &bytes.Buffer{}
		err = b.builder(buf)
		if err != nil {
			return nil
		}
		e := buf.Bytes()
		dkimSign(b.dkim, &e)
		_, err = w.Write(e)
		buf.Reset()
		return err
	}
	return *email
}

func (b *Builder) builder(w io.Writer) (err error) {
	// Headers
	err = b.writeHeaders(w)
	if err != nil {
		return
	}

	if len(b.attachments) != 0 {
		b.markerGlobal.new()
		_, err = w.Write([]byte("Content-Type: multipart/mixed;\r\n\tboundary=\"" + b.markerGlobal.string() + "\"\r\n"))
		if err != nil {
			return
		}
	}

	// Plain text this Text HTML
	if (len(b.textPlain) != 0 || b.textFunc != nil) && (len(b.textHTML) != 0 || b.htmlFunc != nil) {
		if b.markerGlobal.isset() {
			_, err = w.Write([]byte("\r\n"))
			if err != nil {
				return
			}
			_, err = w.Write(b.markerGlobal.delimiter())
			if err != nil {
				return
			}
		}
		b.markerAlt.new()
		_, err = w.Write([]byte("Content-Type: multipart/alternative;\r\n\tboundary=\"" + b.markerAlt.string() + "\"\r\n\r\n"))
		if err != nil {
			return
		}

		_, err = w.Write(b.markerAlt.delimiter())
		if err != nil {
			return
		}
		err = b.writeTextPlain(w)
		if err != nil {
			return
		}

		_, err = w.Write(b.markerAlt.delimiter())
		if err != nil {
			return
		}
		err = b.writeTextHTML(w)
		if err != nil {
			return
		}

		_, err = w.Write(b.markerAlt.finish())
		if err != nil {
			return
		}

	} else if len(b.textHTML) != 0 || b.htmlFunc != nil {
		if b.markerGlobal.isset() {
			_, err = w.Write([]byte("\r\n"))
			if err != nil {
				return
			}
			_, err = w.Write(b.markerGlobal.delimiter())
			if err != nil {
				return
			}
		}
		err = b.writeTextHTML(w)
		if err != nil {
			return
		}
	} else if len(b.textPlain) != 0 || b.textFunc != nil {
		if b.markerGlobal.isset() {
			_, err = w.Write([]byte("\r\n"))
			if err != nil {
				return
			}
			_, err = w.Write(b.markerGlobal.delimiter())
			if err != nil {
				return
			}
		}
		err = b.writeTextPlain(w)
		if err != nil {
			return
		}
	}

	// Attachments
	err = b.writeAttachment(w)
	if err != nil {
		return
	}

	if b.markerGlobal.isset() {
		_, err = w.Write(b.markerGlobal.finish())
		if err != nil {
			return
		}
	}

	return
}

func (b *Builder) writeHeaders(w io.Writer) (err error) {
	_, err = w.Write([]byte("From: " + b.From + "\r\n"))
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("To: " + b.To + "\r\n"))
	if err != nil {
		return err
	}
	if b.replyTo != "" {
		_, err = w.Write([]byte("Reply-To: <" + b.replyTo + ">\r\n"))
		if err != nil {
			return err
		}
	}
	_, err = w.Write([]byte("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n"))
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("MIME-Version: 1.0\r\n"))
	for i := range b.headers {
		_, err = w.Write([]byte(b.headers[i]))
		if err != nil {
			return err
		}
	}

	_, err = w.Write([]byte("Subject: "))
	if err != nil {
		return err
	}
	subj := bytes.NewBufferString(b.Subject)
	if b.subjectFunc != nil {
		err = b.subjectFunc(subj)
		if err != nil {
			return err
		}
	}
	_, err = w.Write([]byte(mime.BEncoding.Encode("utf-8", subj.String()) + "\r\n"))

	return err
}

// Text block
func (b *Builder) writeTextPlain(w io.Writer) (err error) {
	_, err = w.Write([]byte("Content-Type: text/plain;\r\n\t charset=\"utf-8\"\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n"))
	if err != nil {
		return
	}
	q := quotedprintable.NewWriter(w)

	_, err = q.Write(b.textPlain)
	if err != nil {
		return
	}

	if b.textFunc != nil {
		err = b.textFunc(q)
		if err != nil {
			return
		}
	}
	err = q.Close()
	if err != nil {
		return
	}
	_, err = w.Write([]byte("\r\n\r\n"))
	return
}

// HTML block
func (b *Builder) writeTextHTML(w io.Writer) (err error) {
	if len(b.textHTMLRelated) != 0 {
		b.markerHTML.new()
		_, err = w.Write([]byte("Content-Type: multipart/related;\r\n\tboundary=\"" + b.markerHTML.string() + "\"\r\n\r\n"))
		if err != nil {
			return
		}
		_, err = w.Write(b.markerHTML.delimiter())
		if err != nil {
			return
		}
	}
	_, err = w.Write([]byte("Content-Type: text/html;\r\n\t charset=\"utf-8\"\r\nContent-Transfer-Encoding: base64\r\n\r\n"))
	if err != nil {
		return
	}

	dwr := newDelimitWriter(w, []byte{0x0d, 0x0a}, 76) // 76 from RFC
	b64Enc := base64.NewEncoder(base64.StdEncoding, dwr)
	defer b64Enc.Close()

	_, err = b64Enc.Write(b.textHTML)
	if err != nil {
		return
	}

	if b.htmlFunc != nil {
		err = b.htmlFunc(b64Enc)
		if err != nil {
			return
		}
	}
	_, err = w.Write([]byte("\r\n"))
	if err != nil {
		return
	}

	// related files
	for i := range b.textHTMLRelated {
		_, err = w.Write([]byte("\r\n"))
		if err != nil {
			return
		}
		_, err = w.Write(b.markerHTML.delimiter())
		if err != nil {
			return
		}

		err = fileBase64Writer(w, b.textHTMLRelated[i], "inline")
		if err != nil {
			return
		}
		_, err = w.Write([]byte("\r\n\r\n"))
		if err != nil {
			return
		}
	}

	if b.markerHTML.isset() {
		_, err = w.Write(b.markerHTML.finish())
		if err != nil {
			return
		}
	}
	return
}

func (b *Builder) writeAttachment(w io.Writer) (err error) {
	for i := range b.attachments {
		if !b.markerGlobal.isset() {
			b.markerGlobal.new()
			_, err = w.Write([]byte("Content-Type: multipart/mixed;\r\n\tboundary=\"" + b.markerGlobal.string() + "\"\r\n"))
			if err != nil {
				return
			}
		}
		_, err = w.Write([]byte("\r\n"))
		if err != nil {
			return
		}
		_, err = w.Write(b.markerGlobal.delimiter())
		if err != nil {
			return
		}
		err = fileBase64Writer(w, b.attachments[i], "attachment")
		if err != nil {
			return
		}
		_, err = w.Write([]byte("\r\n\r\n"))
		if err != nil {
			return
		}
	}
	return
}

type delimitWriter struct {
	n      int
	cnt    int
	dr     []byte
	writer io.Writer
}

func newDelimitWriter(writer io.Writer, dr []byte, cnt int) *delimitWriter {
	return &delimitWriter{n: 0, cnt: cnt, dr: dr, writer: writer}
}

func (w *delimitWriter) Write(p []byte) (n int, err error) {
	for i := range p {
		_, err = w.writer.Write(p[i : i+1])
		if err != nil {
			break
		}
		if w.n++; w.n%w.cnt == 0 {
			w.writer.Write(w.dr)

		}
	}
	return w.n, err
}

func fileBase64Writer(w io.Writer, f *os.File, disposition string) (err error) {
	name := filepath.Base(f.Name())
	var info os.FileInfo
	info, err = f.Stat()
	if err != nil {
		return
	}
	size := info.Size()
	buf := make([]byte, 512)
	_, err = f.Read(buf)
	if err != nil && err != io.EOF {
		return
	}
	content := http.DetectContentType(buf)
	_, err = f.Seek(0, 0)
	var contentID string
	if disposition == "inline" {
		contentID = "Content-ID: <" + name + ">\r\n"
	}
	_, err = w.Write([]byte(fmt.Sprintf(
		"Content-Type: %s;\r\n\tname=\"%s\"\r\nContent-Transfer-Encoding: base64\r\n%sContent-Disposition: %s;\r\n\tfilename=\"%s\"; size=%d;\r\n\r\n",
		content,
		name,
		contentID,
		disposition,
		name,
		size)))
	if err != nil {
		return
	}

	dwr := newDelimitWriter(w, []byte{0x0d, 0x0a}, 76) // 76 from RFC
	b64Enc := base64.NewEncoder(base64.StdEncoding, dwr)

	_, err = io.Copy(b64Enc, f)
	if err != nil {
		return err
	}

	return b64Enc.Close()
}

type marker []byte

func (m *marker) new() {
	b := make([]byte, 30)
	rand.Read(b)
	en := base64.StdEncoding // or URLEncoding
	d := make([]byte, en.EncodedLen(len(b)))
	en.Encode(d, b)
	*m = []byte(string(d))
}

func (m *marker) delimiter() []byte {
	return []byte("--" + string(*m) + "\r\n")
}

func (m *marker) finish() []byte {
	return []byte("\r\n--" + string(*m) + "--\r\n")
}

func (m *marker) isset() bool {
	return string(*m) != ""
}

func (m *marker) string() string {
	return string(*m)
}
