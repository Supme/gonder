package models

import (
	"testing"
)

func TestCrypto(t *testing.T) {
	data := []struct {
		key   string
		input string
	}{
		{"Foo", "Boo"},
		{"Bar", "Car"},
		{"", "Blank key"},
		{"pa55", ""},
		{"5uper5ecretKey", `{"id":"57116867","email":"user@tld.mail","data":"[Кнопка / Подписаться] https://site.tld/models/teaser/?utm_source=email"}`},
	}

	for _, alg := range []CryptoAlg{CryptoAlgNone, CryptoAlgAES} {
		for _, d := range data {
			Config.secretString = d.key
			enc, err := encrypt(alg, []byte(d.input))
			if err != nil {
				t.Errorf("Unable to encrypt algoritm '%s' '%v' with key '%v': %v", alg, d.input, d.key, err)
				continue
			}
			dec, err := Decrypt(enc)
			if err != nil {
				t.Errorf("Unable to decrypt '%v' with key '%v': %v", enc, d.key, err)
				continue
			}
			//t.Errorf("'%s', '%s':'%s'\r\n", alg, string(enc), string(dec))
			if string(dec) != d.input {
				t.Errorf("Decrypt Key %v\n  Input: %v\n  Expect: %v\n  Actual: %v", d.key, enc, d.input, enc)
			}
		}

	}
}
