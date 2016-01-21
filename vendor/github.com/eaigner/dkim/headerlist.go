package dkim

import (
	"bytes"
	"regexp"
	"strings"
)

type HeaderList []Header

var continuedHeaderRx = regexp.MustCompile(`^[ \t]`)

func ParseHeaderList(header []byte) (list HeaderList) {
	lines := bytes.Split(header, []byte("\r\n"))
	list = make(HeaderList, 0, len(lines))
	lastHeader := -1
	for _, v := range lines {
		if continuedHeaderRx.Match(v) && lastHeader >= 0 {
			list[lastHeader] += Header("\r\n" + string(v))
		} else {
			lastHeader = len(list)
			list = append(list, Header(v))
		}
	}
	return
}

func (l HeaderList) Get(key string) (h Header) {
	for _, v := range l {
		if v.Key(true) == strings.ToLower(key) {
			h = v
			return
		}
	}
	return
}

func (l HeaderList) Fields() string {
	a := make([]string, 0, len(l))
	for _, v := range l {
		k := v.Key(false)
		if strings.ToLower(k) != strings.ToLower(SignatureHeaderKey) {
			a = append(a, k)
		}
	}
	return strings.Join(a, ":")
}

func (l HeaderList) Canonical(relaxed bool) string {
	a := make([]string, 0, len(l))
	for _, v := range l {
		a = append(a, v.Canonical(relaxed))
	}
	return strings.Join(a, "\r\n") + "\r\n"
}
