package directEmail

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mime/quotedprintable"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// SetRawMessageBytes
func (slf *Email) SetRawMessageBytes(data []byte) error {
	slf.raw.Reset()
	_, err := slf.raw.Write(data)
	return err
}

// SetRawMessageString
func (slf *Email) SetRawMessageString(data string) error {
	slf.raw.Reset()
	_, err := slf.raw.WriteString(data)
	return err
}

// GetRawMessageBytes
func (slf *Email) GetRawMessageBytes() []byte {
	return slf.raw.Bytes()
}

// GetRawMessageString
func (slf *Email) GetRawMessageString() string {
	return slf.raw.String()
}

// Header add extra headers to email
func (slf *Email) Header(headers ...string) {
	for i := range headers {
		slf.headers = append(slf.headers, headers[i])
	}
}

// TextPlain add text/plain content to email
func (slf *Email) TextPlain(content string) (err error) {
	var part bytes.Buffer
	defer part.Reset()
	_, err = part.WriteString("Content-Type: text/plain;\r\n\t charset=\"utf-8\"\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n")
	if err != nil {
		return err
	}
	w := quotedprintable.NewWriter(&part)
	w.Write([]byte(strings.TrimLeft(content, "\r\n")))
	w.Close()
	slf.textPlain = part.Bytes()
	return nil
}

// TextHtml add text/html content to email
func (slf *Email) TextHtml(content string) (err error) {
	var part bytes.Buffer
	defer part.Reset()
	_, err = part.WriteString("Content-Type: text/html;\r\n\t charset=\"utf-8\"\r\nContent-Transfer-Encoding: base64\r\n\r\n")
	if err != nil {
		return err
	}
	err = slf.line76(&part, base64.StdEncoding.EncodeToString([]byte(content)))
	if err != nil {
		return err
	}
	slf.textHTML = part.Bytes()
	return nil
}

// TextHtmlWithRelated add text/html content with related file.
//
// Example use file in html
//  email.TextHtmlWithRelated(
//  	`... <img src="cid:myImage.jpg" width="500px" height="250px" border="1px" alt="My image"/> ...`,
//  	"/path/to/attach/myImage.jpg",
//  )
func (slf *Email) TextHtmlWithRelated(content string, files ...string) (err error) {
	var (
		part bytes.Buffer
		data []byte
	)
	defer part.Reset()

	marker := slf.makeMarker()
	_, err = part.WriteString("Content-Type: multipart/related;\r\n\tboundary=\"" + marker + "\"\r\n")
	if err != nil {
		return err
	}

	_, err = part.WriteString("\r\n--" + marker + "\r\n")
	if err != nil {
		return err
	}
	_, err = part.WriteString("Content-Type: text/html;\r\n\t charset=\"utf-8\"\r\nContent-Transfer-Encoding: base64\r\n\r\n")
	if err != nil {
		return err
	}
	err = slf.line76(&part, base64.StdEncoding.EncodeToString([]byte(content)))
	if err != nil {
		return err
	}

	for i := range files {
		_, err = part.WriteString("\r\n--" + marker + "\r\n")
		if err != nil {
			return err
		}
		data, err = ioutil.ReadFile(files[i])
		if err != nil {
			return err
		}
		_, err = part.WriteString(fmt.Sprintf("Content-Type: %s;\r\n\tname=\"%s\"\r\nContent-Transfer-Encoding: base64\r\nContent-ID: <%s>\r\nContent-Disposition: inline;\r\n\tfilename=\"%s\"; size=%d;\r\n\r\n", http.DetectContentType(data), filepath.Base(files[i]), filepath.Base(files[i]), filepath.Base(files[i]), len(data)))
		if err != nil {
			return err
		}
		err = slf.line76(&part, base64.StdEncoding.EncodeToString(data))
		if err != nil {
			return err
		}
	}
	_, err = part.WriteString("\r\n--" + marker + "--\r\n")

	slf.textHTML = part.Bytes()
	return nil
}

// Attachment attach files to email message
func (slf *Email) Attachment(files ...string) (err error) {
	var (
		part bytes.Buffer
		data []byte
	)

	for i := range files {
		data, err = ioutil.ReadFile(files[i])
		if err != nil {
			return err
		}
		_, err = part.WriteString(fmt.Sprintf("Content-Type: %s;\r\n\tname=\"%s\"\r\nContent-Transfer-Encoding: base64\r\nContent-Disposition: attachment;\r\n\tfilename=\"%s\"; size=%d;\r\n\r\n", http.DetectContentType(data), filepath.Base(files[i]), filepath.Base(files[i]), len(data)))
		if err != nil {
			return err
		}
		err = slf.line76(&part, base64.StdEncoding.EncodeToString(data))
		if err != nil {
			return err
		}
		slf.attachments = append(slf.attachments, part.Bytes())
		part.Reset()
	}

	return nil
}

// Render added text/html, text/plain, attachments part to raw view
func (slf *Email) Render() (err error) {
	return slf.RenderWithDkim("", []byte{})
}

// Render added text/html, text/plain, attachments part to raw view
// If dkim selector not blank add DKIM signature email
// Generate private key:
//  openssl genrsa -out /path/to/key/example.com.key 2048
// Generate public key:
//  openssl rsa -in /path/to/key/example.com.key -pubout
// Add public key to DNS myselector._domainkey.example.com TXT record
//  k=rsa; p=MIGfMA0GC...
func (slf *Email) RenderWithDkim(dkimSelector string, dkimPrivateKey []byte) (err error) {
	var (
		marker string
	)

	// -------------- head ----------------------------------------------------------
	_, err = slf.raw.WriteString("Return-path: <" + slf.FromEmail + ">\r\n")
	if err != nil {
		return err
	}

	_, err = slf.raw.WriteString("From: ")
	if err != nil {
		return err
	}
	if slf.FromName != "" {
		_, err = slf.raw.WriteString(slf.encodeRFC2045(slf.FromName) + " ")
		if err != nil {
			return err
		}
	}
	_, err = slf.raw.WriteString("<" + slf.FromEmail + ">\r\n")
	if err != nil {
		return err
	}
	_, err = slf.raw.WriteString("To: ")
	if err != nil {
		return err
	}
	if slf.ToName != "" {
		_, err = slf.raw.WriteString(slf.encodeRFC2045(slf.ToName) + " ")
		if err != nil {
			return err
		}
	}
	_, err = slf.raw.WriteString("<" + slf.ToEmail + ">\r\n")
	if err != nil {
		return err
	}

	_, err = slf.raw.WriteString("Subject: " + slf.encodeRFC2045(slf.Subject) + "\r\n")
	if err != nil {
		return err
	}
	_, err = slf.raw.WriteString("MIME-Version: 1.0\r\n")
	if err != nil {
		return err
	}
	_, err = slf.raw.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")
	if err != nil {
		return err
	}

	// Email has attachment?
	if len(slf.attachments) > 0 {
		marker = slf.makeMarker()
		_, err = slf.raw.WriteString("Content-Type: multipart/mixed;\r\n\tboundary=\"" + marker + "\"\r\n")
		if err != nil {
			return err
		}
	}

	// add extra headers
	_, err = slf.raw.WriteString(strings.Join(slf.headers, "\r\n"))
	if err != nil {
		return err
	}
	if len(slf.headers) > 0 {
		_, err = slf.raw.WriteString("\r\n")
	}

	startBody := slf.raw.Len()
	// ------------- /head ---------------------------------------------------------

	if len(slf.textPlain) > 0 || len(slf.textHTML) > 0 {
		if marker != "" {
			_, err = slf.raw.WriteString("\r\n--" + marker + "\r\n")
		}
		err = slf.renderText()
		if err != nil {
			return err
		}
	}

	// ------------- attachments ----------------------------------------------------------
	for i := range slf.attachments {
		_, err = slf.raw.WriteString("\r\n--" + marker + "\r\n")
		if err != nil {
			return err
		}
		_, err = slf.raw.Write(slf.attachments[i])
		if err != nil {
			return err
		}
		if err != nil {
			return err
		}
	}
	// ------------- /attachments ----------------------------------------------------------

	if marker != "" {
		_, err = slf.raw.WriteString("\r\n--" + marker + "--\r\n")
		if err != nil {
			return err
		}
	}

	// ------------ blank email ------------------------------------------------------------
	if len(slf.textPlain) == 0 && len(slf.textHTML) == 0 && len(slf.attachments) == 0 {
		_, err = slf.raw.WriteString("Content-Type: text/plain;\r\n\t charset=\"utf-8\"\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\nBlank email")
		if err != nil {
			return err
		}
	}
	// ------------ /blank email -----------------------------------------------------------

	// ------------ DKIM -------------------------------------------------------------------
	slf.bodyLenght = slf.raw.Len() - startBody

	if dkimSelector != "" {
		err = slf.dkimSign(dkimSelector, dkimPrivateKey)
		if err != nil {
			return err
		}
	}
	// ------------ /DKIM ------------------------------------------------------------------

	return nil
}

func (slf *Email) renderText() error {
	var (
		marker string
		err    error
	)
	if len(slf.textPlain) > 0 && len(slf.textHTML) > 0 {
		marker = slf.makeMarker()
		_, err = slf.raw.WriteString("Content-Type: multipart/alternative;\r\n\tboundary=\"" + marker + "\"\r\n")
		if err != nil {
			return err
		}
	}

	if marker != "" {
		_, err = slf.raw.WriteString("\r\n--" + marker + "\r\n")
	}

	if len(slf.textPlain) > 0 {
		_, err = slf.raw.Write(slf.textPlain)
		if err != nil {
			return err
		}
	}

	if marker != "" {
		_, err = slf.raw.WriteString("\r\n--" + marker + "\r\n")
	}

	if len(slf.textHTML) > 0 {
		_, err = slf.raw.Write(slf.textHTML)
		if err != nil {
			return err
		}
	}

	if marker != "" {
		_, err = slf.raw.WriteString("\r\n--" + marker + "--\r\n")
	}

	return err
}
