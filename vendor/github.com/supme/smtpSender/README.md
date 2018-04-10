# smtpSender

[![Go Report Card](https://goreportcard.com/badge/github.com/Supme/smtpSender)](https://goreportcard.com/report/github.com/Supme/smtpSender)


Send email
```
bldr := &smtpSender.Builder{
	From: "Sender <sender@domain.tld>",
	To: "Me <me+test@mail.tld>",
	Subject: "Test subject",
}
bldr.SetDKIM("domain.tld", "test", myPrivateKey)
bldr.AddHeader("Content-Language: ru", "Message-ID: <Id-123>", "Precedence: bulk")
bldr.AddTextPlain("textPlain")
bldr.AddTextHTML("<h1>textHTML</h1><img src=\"cid:image.gif\"/>", "./image.gif")
bldr.AddAttachment("./file.zip", "./music.mp3")
email := bldr.Email("Id-123", func(result smtpSender.Result){
	fmt.Printf("Result for email id '%s' duration: %f sec result: %v\n", result.ID, result.Duration.Seconds(), result.Err)
})
	
conn := new(smtpSender.Connect)
conn.SetHostName("sender.domain.tld")
conn.SetMapIP("192.168.0.10", "31.32.33.34")
	
email.Send(conn, nil)

or

server = &smtpSender.SMTPserver{
	Host:     "smtp.server.tld",
	Port:     587,
	Username: "sender@domain.tld",
	Password: "password",
}
email.Send(conn, server)
```


Best way send email from pool
```
pipe := smtpSender.NewPipe(
	smtpSender.Config{
		Iface:  "31.32.33.34",
		Stream:   5,
	},
	smtpSender.Config{
		Iface:  "socks5://222.222.222.222:7080",
		Stream: 2,
	})
pipe.Start()

for i := 1; i <= 50; i++ {
    bldr := new(smtpSender.Builder)
    bldr.SetFrom("Sender", "sender@domain.tld")
    bldr.SetTo("Me", "me+test@mail.tld")
    bldr.SetSubject("Test subject " + id)
    bldr.AddTextHTML("<h1>textHTML</h1><img src=\"cid:image.gif\"/>", "./image.gif")
    email := bldr.Email(id, func(result smtpSender.Result) {
       	fmt.Printf("Result for email id '%s' duration: %f sec result: %v\n", result.ID, result.Duration.Seconds(), result.Err)
       	wg.Done()
    })
	err := pipe.Send(email)
	if err != nil {
		fmt.Printf("Send email id '%d' error %+v\n", i, err)
		break
	}
	if i == 35 {
		pipe.Stop()
	}
}
```


Use template for email
```
import (
	...
	tmplHTML "html/template"
	tmplText "text/template"
)
    ...
	subj := tmplText.New("Subject")
	subj.Parse("{{.Name}} this template subject text.")
	html := tmplHTML.New("HTML")
	html.Parse(`<h1>This 'HTML' template.</h1><img src="cid:image.gif"><h2>Hello {{.Name}}!</h2>`)
	text := tmplText.New("Text")
	text.Parse("This 'Text' template. Hello {{.Name}}!")
	data := map[string]string{"Name": "Вася"}
	bldr.AddSubjectFunc(func(w io.Writer) error {
		return subj.Execute(w, &data)
	})
	bldr.AddTextFunc(func(w io.Writer) error {
		return text.Execute(w, &data)
	})
	bldr.AddHTMLFunc(func(w io.Writer) error {
		return html.Execute(w, &data)
	}, "./image.gif")
    ...
```


One more method send email from pool (Depricated)
```
emailPipe := smtpSender.NewEmailPipe(
	smtpSender.Config{
		Iface:  "31.32.33.34",
		Stream:   5,
	},
	smtpSender.Config{
		Iface:  "socks5://222.222.222.222:7080",
		Stream: 2,
	})

start := time.Now()
wg := &sync.WaitGroup{}
for i := 1; i <= 15; i++ {
	id := "Id-" + strconv.Itoa(i)
	bldr := new(smtpSender.Builder)
	bldr.SetFrom("Sender", "sender@domain.tld")
	bldr.SetTo("Me", "me+test@mail.tld")
	bldr.SetSubject("Test subject " + id)
	bldr.SetDKIM("domain.tld", "test", myPrivateKey)
	bldr.AddHeader("Content-Language: ru", "Message-ID: <Id-123>", "Precedence: bulk")
	bldr.AddTextPlain("textPlain")
	bldr.AddTextHTML("<h1>textHTML</h1><img src=\"cid:image.gif\"/>", "./image.gif")
	bldr.AddAttachment("./file.zip", "./music.mp3")
	wg.Add(1)
	email := bldr.Email(id, func(result smtpSender.Result) {
		fmt.Printf("Result for email id '%s' duration: %f sec result: %v\n", result.ID, result.Duration.Seconds(), result.Err)
		wg.Done()
	})
	emailPipe <- email
}
wg.Wait()

fmt.Printf("Stream send duration: %s\r\n", time.Now().Sub(start).String())
```