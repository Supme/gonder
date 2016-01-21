package dkim

import (
	"crypto"
	_ "crypto/sha256"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Conf map[string]string

const (
	VersionKey          = "v"
	AlgorithmKey        = "a"
	DomainKey           = "d"
	SelectorKey         = "s"
	CanonicalizationKey = "c"
	QueryMethodKey      = "q"
	BodyLengthKey       = "l"
	TimestampKey        = "t"
	ExpireKey           = "x"
	FieldsKey           = "h"
	BodyHashKey         = "bh"
	SignatureDataKey    = "b"
	AUIDKey             = "i"
	CopiedFieldsKey     = "z"
)

const (
	AlgorithmSHA256 = "rsa-sha256"
)

func NewConf(domain string, selector string) (Conf, error) {
	if domain == "" {
		return nil, fmt.Errorf("domain invalid")
	}
	if selector == "" {
		return nil, fmt.Errorf("selector invalid")
	}
	return Conf{
		VersionKey:          "1",
		AlgorithmKey:        AlgorithmSHA256,
		DomainKey:           domain,
		SelectorKey:         selector,
		CanonicalizationKey: "relaxed/simple",
		QueryMethodKey:      "dns/txt",
		TimestampKey:        strconv.FormatInt(time.Now().Unix(), 10),
		FieldsKey:           "",
		BodyHashKey:         "",
		SignatureDataKey:    "",
	}, nil
}

func (c Conf) Validate() error {
	minRequired := []string{
		VersionKey,
		AlgorithmKey,
		DomainKey,
		SelectorKey,
		CanonicalizationKey,
		QueryMethodKey,
		TimestampKey,
	}
	for _, v := range minRequired {
		if _, ok := c[v]; !ok {
			return fmt.Errorf("key '%s' missing", v)
		}
	}
	return nil
}

func (c Conf) Algorithm() string {
	if a := c[AlgorithmKey]; a != "" {
		return a
	}
	return AlgorithmSHA256
}

func (c Conf) Hash() crypto.Hash {
	if c.Algorithm() == AlgorithmSHA256 {
		return crypto.SHA256
	}
	panic("algorithm not implemented")
}

func (c Conf) RelaxedHeader() bool {
	can := strings.ToLower(c[CanonicalizationKey])
	if strings.HasPrefix(can, "relaxed") {
		return true
	}
	return false
}

func (c Conf) RelaxedBody() bool {
	can := strings.ToLower(c[CanonicalizationKey])
	if strings.HasSuffix(can, "/relaxed") {
		return true
	}
	return false
}

func (c Conf) String() string {
	keyOrder := []string{
		VersionKey,
		AlgorithmKey,
		CanonicalizationKey,
		DomainKey,
		QueryMethodKey,
		SelectorKey,
		TimestampKey,
		BodyHashKey,
		FieldsKey,
		CopiedFieldsKey,
		AUIDKey,
		BodyLengthKey,
		SignatureDataKey,
	}
	pairs := make([]string, 0, len(keyOrder))
	for _, k := range keyOrder {
		v, ok := c[k]
		if ok {
			pairs = append(pairs, k+"="+v)
		}
	}
	return strings.Join(pairs, "; ")
}
