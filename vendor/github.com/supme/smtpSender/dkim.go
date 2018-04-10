package smtpSender

import (
	"github.com/toorop/go-dkim"
)

func dkimSign(d builderDKIM, body *[]byte) error {
	options := dkim.NewSigOptions()
	options.PrivateKey = d.privateKey
	options.Domain = d.domain
	options.Selector = d.selector
	options.Algo = "rsa-sha1"
	options.Headers = []string{"from", "to", "subject"}
	options.AddSignatureTimestamp = true
	options.Canonicalization = "simple/simple"
	options.BodyLength = 30

	err := dkim.Sign(body, options)
	if err != nil {
		return err
	}

	return nil
}
