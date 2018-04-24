package dkim

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"net"
	"strings"
)

// pubKeyRep represents a parsed version of public key record
type pubKeyRep struct {
	Version      string
	HashAlgo     []string
	KeyType      string
	Note         string
	PubKey       rsa.PublicKey
	ServiceType  []string
	FlagTesting  bool // flag y
	FlagIMustBeD bool // flag i
}

func newPubKeyFromDnsTxt(selector, domain string) (*pubKeyRep, verifyOutput, error) {
	txt, err := net.LookupTXT(selector + "._domainkey." + domain)
	if err != nil {
		if strings.HasSuffix(err.Error(), "no such host") {
			return nil, PERMFAIL, ErrVerifyNoKeyForSignature
		} else {
			return nil, TEMPFAIL, ErrVerifyKeyUnavailable
		}
	}

	// empty record
	if len(txt) == 0 {
		return nil, PERMFAIL, ErrVerifyNoKeyForSignature
	}

	pkr := new(pubKeyRep)
	pkr.Version = "DKIM1"
	pkr.HashAlgo = []string{"sha1", "sha256"}
	pkr.KeyType = "rsa"
	pkr.ServiceType = []string{"all"}
	pkr.FlagTesting = false
	pkr.FlagIMustBeD = false

	// parsing, we keep the first record
	// TODO: if there is multiple record

	p := strings.Split(txt[0], ";")
	for i, data := range p {
		keyVal := strings.SplitN(data, "=", 2)
		val := ""
		if len(keyVal) > 1 {
			val = strings.TrimSpace(keyVal[1])
		}
		switch strings.ToLower(strings.TrimSpace(keyVal[0])) {
		case "v":
			// RFC: is this tag is specified it MUST be the first in the record
			if i != 0 {
				return nil, PERMFAIL, ErrVerifyTagVMustBeTheFirst
			}
			pkr.Version = val
			if pkr.Version != "DKIM1" {
				return nil, PERMFAIL, ErrVerifyVersionMusBeDkim1
			}
		case "h":
			p := strings.Split(strings.ToLower(val), ":")
			pkr.HashAlgo = []string{}
			for _, h := range p {
				h = strings.TrimSpace(h)
				if h == "sha1" || h == "sha256" {
					pkr.HashAlgo = append(pkr.HashAlgo, h)
				}
			}
			// if empty switch back to default
			if len(pkr.HashAlgo) == 0 {
				pkr.HashAlgo = []string{"sha1", "sha256"}
			}
		case "k":
			if strings.ToLower(val) != "rsa" {
				return nil, PERMFAIL, ErrVerifyBadKeyType
			}
		case "n":
			pkr.Note = val
		case "p":
			rawkey := val
			if rawkey == "" {
				return nil, PERMFAIL, ErrVerifyRevokedKey
			}
			un64, err := base64.StdEncoding.DecodeString(rawkey)
			if err != nil {
				return nil, PERMFAIL, ErrVerifyBadKey
			}
			pk, err := x509.ParsePKIXPublicKey(un64)
			if pk, ok := pk.(*rsa.PublicKey); ok {
				pkr.PubKey = *pk
			}
		case "s":
			t := strings.Split(strings.ToLower(val), ":")
			for _, tt := range t {
				if tt == "*" {
					pkr.ServiceType = []string{"all"}
					break
				}
				if tt == "email" {
					pkr.ServiceType = []string{"email"}
				}
			}
		case "t":
			flags := strings.Split(strings.ToLower(val), ":")
			for _, flag := range flags {
				if flag == "y" {
					pkr.FlagTesting = true
					continue
				}
				if flag == "s" {
					pkr.FlagIMustBeD = true
				}
			}
		}
	}

	// if no pubkey
	if pkr.PubKey == (rsa.PublicKey{}) {
		return nil, PERMFAIL, ErrVerifyNoKey
	}

	return pkr, SUCCESS, nil
}
