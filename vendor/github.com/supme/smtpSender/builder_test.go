package smtpSender

import (
	"bytes"
	"encoding/base64"
	tmplHTML "html/template"
	"io"
	"io/ioutil"
	"testing"
	tmplText "text/template"
)

var (
	pkey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQC5exyEkt7y+xgJI63jgqGVb7bWmSNvZSfbXqXFLklVJcB70Sy4
zGb4/GbNULpCKFtOJItAJNN80GpRak2X465roTsmQvHI0rUQMaQQUJ9ZP3OdUO+/
MU1IR0T1WW5ZULxO6zo5oEpYxvflTgzwzEuJbYVsUvyA/fgMt+CkUKNJHwIDAQAB
AoGACieXVBrGYf8lPbraVk5cklXfaLhRnFOpbvUrljQGh8bdVuoIzMVEDfWjmzIE
QIL9HLYbeZOKkJbIe1SakupALkRj9BgS5TjC8EF3naIl6vcZDsaV7T+AwhZqDXF6
OyCzqCW99UzrFhs/dQ/j9b/uOM0leRS8bP+gS7Awp5RxgmECQQDq9fypFIDYvJbM
eq6QI/OE/phmqsaSlNbE1myDCAFx9OOdR0LkCxDVJx/aSe4hTYXen5gl+uWoMJ8z
sfGxovOdAkEAyhbk0ZayKwwYC9eJlNk2kMml512G7jPoAf5iV7URvTBiW39yZFFU
nvBtmMpQTat7eDBxjoEOD3q0i5QHlzfI6wJBAK7OMBnDDVEyjaa3p2PJu4U4vT20
1GN9pINxW+3oaNrVbPo4aEWtDernXsVSt33DZVOJvPKUxYPqGKenPcABEekCQQCM
jmvL0nJNOnYnFlxcqM8o2PeI+iX02ylM6a9grVGPMm3WkcfwOhkPCs5PbLd5rgGM
ULVKljw/S+rzAZxd8rDNAkANpbVUb3cGBPFK18PR7OsSAxcUNMfwbr5j1oI270i5
uCCZipHDZcGKVanHeBAV8lwRsKYHkkydniVmVfJIIr1z
-----END RSA PRIVATE KEY-----`)
	textPlain                = []byte("Привет, буфет\r\nЗдорова, колбаса!\r\nКак твои дела?\r\n0123456789\r\nabcdefgh\r\n")
	textHTML                 = []byte("<h1>Привет, буфет</h1><br/>\r\n<h2>Здорова, колбаса!</h2><br/>\r\n<h3>Как твои дела?</h3><br/>\r\n0123456789\r\nabcdefgh\r\n")
	discard   io.WriteCloser = devNull{}
)

type devNull struct{}

func (devNull) Write(p []byte) (int, error) { return len(p), nil }
func (devNull) Close() error                { return nil }

func TestBuilder(t *testing.T) {
	bldr := new(Builder)
	bldr.SetSubject("Test subject")
	bldr.SetFrom("Вася", "vasya@mail.tld")
	bldr.SetTo("Петя", "petya@mail.tld")
	bldr.AddHeader("Content-Language: ru", "Message-ID: <test_message>", "Precedence: bulk")
	bldr.AddTextPlain(textPlain)
	bldr.AddTextHTML(textHTML, "./testdata/prwoman.png")
	bldr.AddAttachment("./testdata/knwoman.png")
	email := bldr.Email("Id-123", func(Result) {})
	err := email.WriteCloser(discard)
	if err != nil {
		t.Error(err)
	}

}

//type writeCloser struct {
//	bytes.Buffer
//}
//func (wc *writeCloser) Close() error {
//	return nil
//}

func TestBuilderTemplate(t *testing.T) {
	bldr := new(Builder)
	data := map[string]string{"Name": "Вася"}

	subj := tmplText.New("Text")
	subj.Parse("Test subject for {{.Name}}")
	bldr.AddSubjectFunc(func(w io.Writer) error {
		return subj.Execute(w, data)
	})

	bldr.SetFrom("Вася", "vasya@mail.tld")
	bldr.SetTo("Петя", "petya@mail.tld")

	bldr.AddHeader("Content-Language: ru", "Message-ID: <test_message>", "Precedence: bulk")

	html := tmplHTML.New("HTML")
	html.Parse(`<h1>This 'HTML' template.</h1><h2>Hello {{.Name}}</h2>`)
	text := tmplText.New("Text")
	text.Parse("This 'Text' template. Hello {{.Name}}")

	bldr.AddTextFunc(func(w io.Writer) error {
		return text.Execute(w, data)
	})
	bldr.AddHTMLFunc(func(w io.Writer) error {
		return html.Execute(w, data)
	})

	email := bldr.Email("Id-123", func(Result) {})

	err := email.WriteCloser(discard)
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkBuilder(b *testing.B) {
	bldr := new(Builder)
	bldr.SetSubject("Test subject")
	bldr.SetFrom("Вася", "vasya@mail.tld")
	bldr.SetTo("Петя", "petya@mail.tld")
	bldr.AddHeader("Content-Language: ru", "Message-ID: <test_message>", "Precedence: bulk")
	bldr.AddTextPlain(textPlain)
	bldr.AddTextHTML(textHTML)
	var err error
	for n := 0; n < b.N; n++ {
		email := bldr.Email("Id-123", func(Result) {})
		err = email.WriteCloser(discard)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkBuilderTemplate(b *testing.B) {
	bldr := new(Builder)
	data := map[string]string{"Name": "Вася"}

	subj := tmplText.New("Text")
	subj.Parse("Test subject for {{.Name}}")
	bldr.AddSubjectFunc(func(w io.Writer) error {
		return subj.Execute(w, data)
	})

	bldr.SetFrom("Вася", "vasya@mail.tld")
	bldr.SetTo("Петя", "petya@mail.tld")

	bldr.AddHeader("Content-Language: ru", "Message-ID: <test_message>", "Precedence: bulk")

	html := tmplHTML.New("HTML")
	html.Parse(`<h1>This 'HTML' template.</h1><h2>Hello {{.Name}}</h2>`)
	text := tmplText.New("Text")
	text.Parse("This 'Text' template. Hello {{.Name}}")

	bldr.AddTextFunc(func(w io.Writer) error {
		return text.Execute(w, data)
	})
	bldr.AddHTMLFunc(func(w io.Writer) error {
		return html.Execute(w, data)
	})

	var err error
	for n := 0; n < b.N; n++ {
		email := bldr.Email("Id-123", func(Result) {})
		err = email.WriteCloser(discard)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkBuilderAttachment(b *testing.B) {
	bldr := new(Builder)
	bldr.SetSubject("Test subject")
	bldr.SetFrom("Вася", "vasya@mail.tld")
	bldr.SetTo("Петя", "petya@mail.tld")
	bldr.AddHeader("Content-Language: ru", "Message-ID: <test_message>", "Precedence: bulk")
	bldr.AddTextPlain(textPlain)
	bldr.AddTextHTML(textHTML, "./testdata/prwoman.png")
	bldr.AddAttachment("./testdata/knwoman.png")
	var err error
	for n := 0; n < b.N; n++ {
		email := bldr.Email("Id-123", func(Result) {})
		err = email.WriteCloser(discard)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkBuilderDKIM(b *testing.B) {
	bldr := new(Builder)
	bldr.SetDKIM("mail.ru", "test", pkey)
	bldr.SetSubject("Test subject")
	bldr.SetFrom("Вася", "vasya@mail.tld")
	bldr.SetTo("Петя", "petya@mail.tld")
	bldr.AddHeader("Content-Language: ru", "Message-ID: <test_message>", "Precedence: bulk")
	bldr.AddTextPlain(textPlain)
	bldr.AddTextHTML(textHTML)
	var err error
	for n := 0; n < b.N; n++ {
		email := bldr.Email("Id-123", func(Result) {})
		err = email.WriteCloser(discard)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkBuilderAttachmentDKIM(b *testing.B) {
	bldr := new(Builder)
	bldr.SetDKIM("mail.ru", "test", pkey)
	bldr.SetSubject("Test subject")
	bldr.SetFrom("Вася", "vasya@mail.tld")
	bldr.SetTo("Петя", "petya@mail.tld")
	bldr.AddHeader("Content-Language: ru", "Message-ID: <test_message>", "Precedence: bulk")
	bldr.AddTextPlain(textPlain)
	bldr.AddTextHTML(textHTML, "./testdata/prwoman.png")
	bldr.AddAttachment("./testdata/knwoman.png")
	var err error
	for n := 0; n < b.N; n++ {
		email := bldr.Email("Id-123", func(Result) {})
		err = email.WriteCloser(discard)
		if err != nil {
			b.Error(err)
		}
	}
}

func TestDelimitWriter(t *testing.T) {
	m := []byte(textHTML)
	w := &bytes.Buffer{}
	dwr := newDelimitWriter(w, []byte{0x0d, 0x0a}, 16)
	encoder := base64.NewEncoder(base64.StdEncoding, dwr)
	_, err := encoder.Write(m)
	if err != nil {
		t.Error(err)
	}
	err = encoder.Close()
	if err != nil {
		t.Error(err)
	}

	d, _ := base64.StdEncoding.DecodeString(w.String())
	if c := bytes.Compare(m, d); c != 0 {
		t.Error("Base64 encode/decode not equivalent")
	}
}

func BenchmarkBase64DelimitWriter(b *testing.B) {
	m := []byte("<h1>Hello, буфет</h1><br/>\r\n<h2>Здорова, колбаса!</h2><br/>\r\n<h3>Как твои дела?</h3><br/>\r\n0123456789\r\nabcdefgh\r\n")
	w := ioutil.Discard
	dwr := newDelimitWriter(w, []byte{0x0d, 0x0a}, 8)
	encoder := base64.NewEncoder(base64.StdEncoding, dwr)
	for n := 0; n < b.N; n++ {
		_, err := encoder.Write(m)
		if err != nil {
			b.Error(err)
		}
		err = encoder.Close()
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkDelimitWriter(b *testing.B) {
	m := []byte("<h1>Hello, буфет</h1><br/>\r\n<h2>Здорова, колбаса!</h2><br/>\r\n<h3>Как твои дела?</h3><br/>\r\n0123456789\r\nabcdefgh\r\n")
	w := ioutil.Discard
	dwr := newDelimitWriter(w, []byte{0x0d, 0x0a}, 8)
	for n := 0; n < b.N; n++ {
		_, err := dwr.Write(m)
		if err != nil {
			b.Error(err)
		}
	}
}
