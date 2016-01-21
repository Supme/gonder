package dkim

import (
	"fmt"
	"regexp"
	"strings"
)

var headerRelaxRx = regexp.MustCompile(`\s+`)

type Header string

func NewHeader(key, value string) Header {
	return Header(fmt.Sprintf("%s: %s", key, value))
}

// Component returns the raw header component. Components are separated by colons.
func (h Header) Component(i int) (c string) {
	if c := strings.SplitN(string(h), ":", 2); len(c) > i {
		return c[i]
	}
	return ""
}

// Key returns the (relaxed) key with leading and trailing whitespace trimmed.
func (h Header) Key(relax bool) string {
	key := h.Component(0)
	if relax {
		return strings.ToLower(strings.TrimSpace(key))
	}
	return strings.TrimSpace(key)
}

// Value returns the (relaxed) value with leading and trailing whitespace trimmed
func (h Header) Value(relax bool) string {
	value := h.Component(1)
	if relax {
		return strings.TrimSpace(headerRelaxRx.ReplaceAllString(value, " "))
	}
	return strings.TrimSpace(value)
}

// Canonical returns the (relaxed) canonical header
func (h Header) Canonical(relax bool) string {
	if relax {
		return h.Key(relax) + ":" + h.Value(relax)
	}
	return string(h)
}
