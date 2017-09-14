package main

import (
	"github.com/supme/directEmail"
	"time"
	"fmt"
)

func main() {
	email := directEmail.New()

	// if use socks5 proxy
	email.Ip = "socks://123.124.125.126:8080"

	// if use specified interface
	email.Ip = "192.168.0.10"

	// if use NAT
	email.MapIp = map[string]string {
		"192.168.0.10": "31.33.34.35",
	}

	// if left blank, then auto resolved (for socks use IP for connecting server)
	email.Host = "resolv.name.myhost.com"

	email.FromEmail = "sender@example.com"
	email.FromName = "Sender name"

	email.ToEmail = "user@example.com"
	email.ToName = "Reciver name"

	// add extra headers if need
	email.Header(fmt.Sprintf("Message-ID: <test_message_%d>", time.Now().Unix()))
	email.Header("Content-Language: ru")

	email.Subject = "Тест отправки email"

	// plain text version
	email.TextPlain(`
My email
Текст моего TEXT сообщения
	`)

	// html version
	email.TextHtml(`
<h2>My email</h2>
<p>Текст моего HTML сообщения</p>
	`)

	// or html version with related files
	email.TextHtmlWithRelated(`
<h2>My email</h2>
<p>Текст моего HTML with related files сообщения</p>
<p>Картинка: <img src="cid:myImage.jpg" width="500px" height="250px" border="1px" alt="My image"/></p>
	`,
		"/path/to/attach/myImage.jpg",
	)

	// attach file if need
	email.Attachment("/path/to/attach/file.jpg")

	// Render email message
	email.Render()

	// if dkimSelector not blank, then add DKIM signature to message
	email.RenderWithDkim("myDKIMselector", []byte("DKIMprivateKey"))

	print("\n", string(email.GetRawMessageString()), "\n\n\n")

	err := email.Send()
	if err != nil {
		print("Send email with error:", err.Error())
	}

	// send from SMTP server use login and password
	err = email.SendThroughServer("smtp.server.tld", 587, "username", "password")
	if err != nil {
		print("Send email with error:\n", err.Error(), "\n")
	}

}