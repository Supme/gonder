package directEmail

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/net/idna"
	"math/rand"
	"mime"
	"strings"
)

func (slf *Email) makeMarker() string {
	b := make([]byte, 30)
	rand.Read(b)
	en := base64.StdEncoding // or URLEncoding
	d := make([]byte, en.EncodedLen(len(b)))
	en.Encode(d, b)
	return "_" + string(d) + "_"
}

func (slf *Email) line76(target *bytes.Buffer, encoded string) (err error) {
	nbrLines := len(encoded) / 76
	for i := 0; i < nbrLines; i++ {
		_, err = target.WriteString(encoded[i*76 : (i+1)*76])
		if err != nil {
			return err
		}
		_, err = target.WriteString("\n")
		if err != nil {
			return err
		}
	}
	_, err = target.WriteString(encoded[nbrLines*76:])
	if err != nil {
		return err
	}
	_, err = target.WriteString("\n")
	if err != nil {
		return err
	}

	return nil
}

func (slf *Email) encodeRFC2045(s string) string {
	return mime.BEncoding.Encode("utf-8", s)
}

func (slf *Email) domainFromEmail(email string) (string, error) {
	splitEmail := strings.SplitN(email, "@", 2)
	if len(splitEmail) != 2 {
		return "", errors.New("Bad from email address")
	}

	domain, err := idna.ToASCII(strings.TrimRight(splitEmail[1], "."))
	if err != nil {
		return "", fmt.Errorf("Domain name failed: %v", err)
	}

	return domain, nil
}

func debug(args ...interface{}) {
	if debugIs {
		fmt.Print(args...)
	}
}
