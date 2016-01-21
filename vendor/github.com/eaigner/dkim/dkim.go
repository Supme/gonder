package dkim

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"regexp"
	"strings"
)

const (
	SignatureHeaderKey = "DKIM-Signature"
)

var StdSignableHeaders = []string{
	"Cc",
	"Content-Type",
	"Date",
	"From",
	"Reply-To",
	"Subject",
	"To",
	SignatureHeaderKey,
}

type DKIM struct {
	signableHeaders []string
	conf            Conf
	privateKey      *rsa.PrivateKey
}

func New(conf Conf, keyPEM []byte) (d *DKIM, err error) {
	err = conf.Validate()
	if err != nil {
		return
	}
	if len(keyPEM) == 0 {
		return nil, errors.New("invalid key PEM data")
	}
	dkim := &DKIM{
		signableHeaders: StdSignableHeaders,
		conf:            conf,
	}
	der, _ := pem.Decode(keyPEM)
	key, err := x509.ParsePKCS1PrivateKey(der.Bytes)
	if err != nil {
		return nil, err
	}
	dkim.privateKey = key

	return dkim, nil
}

var (
	rxWsCompress = regexp.MustCompile(`[ \t]+`)
	rxWsCRLF     = regexp.MustCompile(` \r\n`)
)

func (d *DKIM) canonicalBody(body []byte) []byte {
	if d.conf.RelaxedBody() {
		if len(body) == 0 {
			return nil
		}
		// Reduce WSP sequences to single WSP
		body = rxWsCompress.ReplaceAll(body, []byte(" "))

		// Ignore all whitespace at end of lines.
		// Implementations MUST NOT remove the CRLF
		// at the end of the line
		body = rxWsCRLF.ReplaceAll(body, []byte("\r\n"))
	} else {
		if len(body) == 0 {
			return []byte("\r\n")
		}
	}

	// Ignore all empty lines at the end of the message body
	for i := len(body) - 1; i >= 0; i-- {
		if body[i] != '\r' && body[i] != '\n' && body[i] != ' ' {
			body = body[:i+1]
			break
		}
	}

	return append(body, '\r', '\n')
}

func (d *DKIM) canonicalBodyHash(body []byte) []byte {
	b := d.canonicalBody(body)
	digest := d.conf.Hash().New()
	digest.Write([]byte(b))

	return digest.Sum(nil)
}

func (d *DKIM) signableHeaderBlock(header, body []byte) string {
	headerList := ParseHeaderList(header)
	signableHeaderList := make(HeaderList, 0, len(headerList)+1)

	for _, k := range d.signableHeaders {
		if h := headerList.Get(k); h != "" {
			signableHeaderList = append(signableHeaderList, h)
		}
	}

	d.conf[BodyHashKey] = base64.StdEncoding.EncodeToString(d.canonicalBodyHash(body))
	d.conf[FieldsKey] = signableHeaderList.Fields()

	signableHeaderList = append(signableHeaderList, NewHeader(SignatureHeaderKey, d.conf.String()))

	// According to RFC6376 http://tools.ietf.org/html/rfc6376#section-3.7
	// the DKIM header must be inserted without a trailing <CRLF>.
	// That's why we have to trim the space from the canonical header.
	return strings.TrimSpace(signableHeaderList.Canonical(d.conf.RelaxedHeader()))
}

func (d *DKIM) signature(header, body []byte) (string, error) {
	block := d.signableHeaderBlock(header, body)
	hash := d.conf.Hash()
	digest := hash.New()
	digest.Write([]byte(block))

	sig, err := rsa.SignPKCS1v15(rand.Reader, d.privateKey, hash, digest.Sum(nil))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(sig), nil
}

func (d *DKIM) Sign(eml []byte) (signed []byte, err error) {
	header, body, err := splitEML(eml)
	if err != nil {
		return
	}
	sig, err := d.signature(header, body)
	if err != nil {
		return
	}
	d.conf[SignatureDataKey] = sig
	headerList := ParseHeaderList(header)

	// Append the signature header. Keep in mind these are raw values,
	// so we add a <SP> character before the key-value list
	headerList = append(headerList, NewHeader(SignatureHeaderKey, d.conf.String()))
	signedHeader := headerList.Canonical(false)

	signed = []byte(strings.Join([]string{
		signedHeader,
		string(body),
	}, "\r\n"))

	return
}
