package dkim

import (
	"encoding/base64"
	"testing"
)

var dkimSampleEML1 string = "A: X \r\n" +
	"B : Y\t\r\n" +
	"\tZ  \r\n" +
	"\r\n" +
	" C \r\n" +
	"D \t E\r\n" +
	"\r\n" +
	"\r\n"

var dkimSampleEML2 string = "From: Joe SixPack <joe@football.example.com>\r\n" +
	"To: Suzie Q <suzie@shopping.example.net>\r\n" +
	"Subject: Is dinner ready?\r\n" +
	"Date: Fri, 11 Jul 2003 21:00:37 -0700 (PDT)\r\n" +
	"Message-ID: <20030712040037.46341.5F8J@football.example.com>\r\n" +
	"\r\n" +
	"Hi.\r\n" +
	"\r\n" +
	"We lost the game. Are you hungry yet?\r\n" +
	"\r\n" +
	"Joe.\r\n"

var dkimSampleEML3 string = "Return-Path: aws@s3ig.com\r\n" +
	"MIME-Version: 1.0\r\n" +
	"From: aws@s3ig.com\r\n" +
	"To: check-auth@verifier.port25.com\r\n" +
	"Reply-To: aws@s3ig.com\r\n" +
	"Date: 10 Mar 2011 10:41:56 +0000\r\n" +
	"Subject: dkim test email\r\n" +
	"Content-Type: text/plain; charset=us-ascii\r\n" +
	"Content-Transfer-Encoding: quoted-printable\r\n" +
	"\r\n" +
	"This is the body of \t the message.=0D=0AThis is the second line\r\n" +
	"\r\n"

var dkimSamplePEMData []byte = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDptLAWO6YJyVGKwayFzVCEIhY0bjE4ObCr7JeILVWi9ELaWwdm
PEKW0afsqB/jR+yhMIdeyGbW9li4P2X8oFDLFwAkAng857ngz2RYNjf+7bGgvH8n
gbKXuFtnO1kI5+ou3C1+Tixk8aY44uWCFamXueAW1ZzlzexViyG4gdVXlQIBEQKB
gQDb9VpvRzLcCMU3TN6cDIgD49ip0R9D+g+w3qy8Zucv9POgVaycdPNgxVLAnjwh
NKJ5lxX+2rskq57LhvaTabVyDNkRjp8Oks1+nCbklaCEdkJ0S9z/IqjmmqCY4bP1
kJ0Eu18NHBId8dXoK3EQ8nzRCKE+tjkiMl5jC1kJzVdg8QJBAPrDkKMURlLRnMEA
LzXHwoNAPdK94IFwdypq62yDVm4u2rSMUtYlKmJQFiSkytQ2vjBb5HRxqL+p4mnq
30ZHWdUCQQDulfC32vcY7e2IetYhda+sysdZJnfrbquJpdlfBn2QFH8gjC2KM/q+
YtwQGJU/zjtwWN+/joi4vinlKD7RYSbBAkEAk4IY2GZHfALUrcPfiQwYEPic1lGT
Hvbcr4owIbarT99TeUN8BX9GG7ajnRWkfNToWK6GYp02FmPumKhHGkgWuQJBAKhp
1xheVBGY4+fePMxTEpgWqtWEkOJsPNmiPxXmdsAOd9q9TVJ/C1k2uXTGDv/c3qmo
JXgoYIJoHZKy/ypisfECQQC9FxTwEfzFLTANQVnUQKEbRUk3slaigQb3QoBKXOZr
oCyn+rxDcflW1RbZnsilaMbpN/PMw/IbqRjXA2Tg3Ty6
-----END RSA PRIVATE KEY-----`)

func TestNew(t *testing.T) {
	dkim, err := New(Conf{}, nil)
	if err == nil {
		t.Fatal(err)
	}
	if dkim != nil {
		t.Fatal(dkim)
	}
	conf, err := NewConf("domain", "selector")
	if err != nil {
		t.Fatal(err)
	}
	dkim, err = New(conf, nil)
	if err == nil {
		t.Fatal(err)
	}
	dkim, err = New(conf, dkimSamplePEMData)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCanonicalBody(t *testing.T) {
	dkim := &DKIM{}
	body := dkim.canonicalBody([]byte(""))
	if x := string(body); x != "\r\n" {
		t.Fatal("uninitialized struct should yield CRLF body", x)
	}

	conf, _ := NewConf("domain", "selector")
	conf[CanonicalizationKey] = "relaxed/relaxed"
	dkim, _ = New(conf, dkimSamplePEMData)

	_, body, err := splitEML([]byte(dkimSampleEML1))
	if err != nil {
		t.Fatal(err)
	}
	if x := string(dkim.canonicalBody(body)); x != " C\r\nD E\r\n" {
		t.Fatal(x)
	}

	conf[CanonicalizationKey] = "relaxed/simple"
	if x := string(dkim.canonicalBody(body)); x != " C \r\nD \t E\r\n" {
		t.Fatal(x)
	}

	_, body, err = splitEML([]byte(dkimSampleEML2))
	if err != nil {
		t.Fatal(err)
	}
	conf[CanonicalizationKey] = "relaxed/simple"
	if x := string(dkim.canonicalBody(body)); x != "Hi.\r\n\r\nWe lost the game. Are you hungry yet?\r\n\r\nJoe.\r\n" {
		t.Fatal(x)
	}

	_, body, err = splitEML([]byte(dkimSampleEML3))
	if err != nil {
		t.Fatal(err)
	}
	conf[CanonicalizationKey] = "relaxed/relaxed"
	if x := string(dkim.canonicalBody(body)); x != "This is the body of the message.=0D=0AThis is the second line\r\n" {
		t.Fatal(x)
	}
}

func TestCanonicalBodyHash(t *testing.T) {
	conf, _ := NewConf("domain", "selector")
	conf[CanonicalizationKey] = "relaxed/simple"

	dkim, _ := New(conf, dkimSamplePEMData)

	_, body2, err := splitEML([]byte(dkimSampleEML2))
	if err != nil {
		t.Fatal(err)
	}
	enc := base64.StdEncoding
	if x := enc.EncodeToString(dkim.canonicalBodyHash(body2)); x != "2jUSOH9NhtVGCQWNr9BrIAPreKQjO6Sn7XIkfJVOzv8=" {
		t.Fatal(x)
	}

	_, body, err := splitEML([]byte(dkimSampleEML3))
	if err != nil {
		t.Fatal(err)
	}
	conf[CanonicalizationKey] = "relaxed/relaxed"
	if x := enc.EncodeToString(dkim.canonicalBodyHash(body)); x != "vrfP/4tQvd9QIewLlBjIlqsKMPwXXKj66neZg/smWSc=" {
		t.Fatal(x)
	}

	// Simple canonical empty body
	conf[CanonicalizationKey] = "relaxed/simple"
	if x := enc.EncodeToString(dkim.canonicalBodyHash([]byte(""))); x != "frcCV1k9oG9oKj3dpUqdJg1PxRT2RSN/XKdLCPjaYaY=" {
		t.Fatal(x)
	}

	// Relaxed canonical empty body
	conf[CanonicalizationKey] = "relaxed/relaxed"
	if x := enc.EncodeToString(dkim.canonicalBodyHash([]byte(""))); x != "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=" {
		t.Fatal(x)
	}
}

func dkimSignable() *DKIM {
	conf, _ := NewConf("s3ig.com", "dkim")
	conf[CanonicalizationKey] = "relaxed/relaxed"
	conf[TimestampKey] = "1299753716"
	dkim, _ := New(conf, dkimSamplePEMData)
	dkim.signableHeaders = []string{
		"Content-Type",
		"From",
		"Subject",
		"To",
	}
	return dkim
}

func TestSignableHeaderBlock(t *testing.T) {
	header, body, err := splitEML([]byte(dkimSampleEML3))
	if err != nil {
		t.Fatal(err)
	}
	block := dkimSignable().signableHeaderBlock(header, body)
	expect := "content-type:text/plain; charset=us-ascii\r\n" +
		"from:aws@s3ig.com\r\n" +
		"subject:dkim test email\r\n" +
		"to:check-auth@verifier.port25.com\r\n" +
		"dkim-signature:v=1; a=rsa-sha256; c=relaxed/relaxed;" +
		" d=s3ig.com; q=dns/txt; s=dkim; t=1299753716;" +
		" bh=vrfP/4tQvd9QIewLlBjIlqsKMPwXXKj66neZg/smWSc=;" +
		" h=Content-Type:From:Subject:To; b="
	if block != expect {
		t.Fatal("signable header block invalid", block)
	}
}

func TestSignature(t *testing.T) {
	header, body, err := splitEML([]byte(dkimSampleEML3))
	if err != nil {
		t.Fatal(err)
	}
	sig, err := dkimSignable().signature(header, body)
	if err != nil {
		t.Fatal("error not nil", err)
	}
	if sig != "enIert1AWY8K9AIxTw0qQLOO3TKuRENfJvwYWDXi6xM7IWaz+Bb83xi5YnjBH0Q8opLn643qIaXGVIU2+LBA2a44PZGtTRXYMG3sbQpcEMjfJRPAhAQOazsSlVdq4SmAChAU3g8uPj4r71JdROucZSdm/mW8IoT4IympoCiLKdQ=" {
		t.Fatal("signature invalid", sig)
	}
}

func TestSignedEML(t *testing.T) {
	signed, err := dkimSignable().Sign([]byte(dkimSampleEML3))
	if err != nil {
		t.Fatal("error not nil", err)
	}
	expect := "Return-Path: aws@s3ig.com\r\n" +
		"MIME-Version: 1.0\r\n" +
		"From: aws@s3ig.com\r\n" +
		"To: check-auth@verifier.port25.com\r\n" +
		"Reply-To: aws@s3ig.com\r\n" +
		"Date: 10 Mar 2011 10:41:56 +0000\r\n" +
		"Subject: dkim test email\r\n" +
		"Content-Type: text/plain; charset=us-ascii\r\n" +
		"Content-Transfer-Encoding: quoted-printable\r\n" +
		"DKIM-Signature: v=1; a=rsa-sha256; c=relaxed/relaxed; d=s3ig.com; q=dns/txt; s=dkim;" +
		" t=1299753716; bh=vrfP/4tQvd9QIewLlBjIlqsKMPwXXKj66neZg/smWSc=;" +
		" h=Content-Type:From:Subject:To;" +
		" b=enIert1AWY8K9AIxTw0qQLOO3TKuRENfJvwYWDXi6xM7IWaz+Bb83xi5YnjBH0Q8opLn643qIaXGVIU2+LBA2a44PZGtTRXYMG3sbQpcEMjfJRPAhAQOazsSlVdq4SmAChAU3g8uPj4r71JdROucZSdm/mW8IoT4IympoCiLKdQ=\r\n" +
		"\r\n" +
		"This is the body of \t the message.=0D=0AThis is the second line\r\n" +
		"\r\n"
	if x := string(signed); x != expect {
		t.Fatal(x, "\n----\n", expect)
	}
}
