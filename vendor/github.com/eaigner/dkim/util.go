package dkim

import (
	"bytes"
	"errors"
)

func splitEML(eml []byte) (header, body []byte, err error) {
	if c := bytes.SplitN(eml, []byte("\r\n\r\n"), 2); len(c) == 2 {
		return c[0], c[1], nil
	}
	err = errors.New("could not read header block")
	return
}
