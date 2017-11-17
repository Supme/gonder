package directEmail

import (
	"testing"
)

const privateKey = `
-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQCkVU7ot+WG+FejaV0F8u1TtK2Ln0fRK6l8ekvF8yh6ZgKPsCxd
eE5xO0Jx0kYqXc69LQrDDR6DIZVIB7+CoQgJzO39mAoV4AGslWO5am1Urof68I67
hiYaPjA4XZBFbqcwvN7auo7K+7tSIUkFpoelj+rbVY0Ps+j0VAgImzZkEwIDAQAB
AoGAKFsCu8edSB3od6rCO1nCylGOZMFCw60zO+xUe1IRWK2AZ4TeAD4xFUF2Obln
nbPXt0E+aVPpcE5o+H1enFerP0gdHTo8/JxfoUmidLLR5ufdSX4ggHL2DSiCOOGz
okCUxtW+esx/264TGHgmveWGDlVA6KZf+bU3kVF3yZ7r0AECQQDSOQ8lvAThwB0I
x3uRu+7ut1fKbKbDZkwqpr4kyWhLnAladtvKN2cPvvktYGiCV/adgq6zjLSwcDmr
KMRKWKwBAkEAyB4dswJmvVqB8ehFBBOTuRFKr6EiKZ/1Sv9CW/QyvhWQ5nEFquos
UEe0PIztfzxxVf52AVXWBA5GJMQ9fyGgEwJBAKl+p9/cwHLj2oUBkXfm9rYxzO7A
u5RAHpkk55nxac3MeR4fRwa7tLTVXUJgwOKW2ZgVjZXmlKjNUzHVJK5s4AECQQCu
cmJdbBh3tHBWmq2fMhmyWLqMg6CuPHyuFfqZAjVBsrcPyzKvnVdn3DnoFsnqApyh
5CKmY1cfTfojjtY0/vD1AkBwddvlHcVOeq37euIZolab/RCzN0M/0Y33f04RHxLL
HNzVKGgQBjjkYXU65FPXQT7Ot34uUusjyZYDFX5KeXKK
-----END RSA PRIVATE KEY-----
`

func emailMessage() (email Email, err error) {
	email = Email{
		Host: "resolv.name.myhost.com",

		FromEmail: "sender@example.com",
		FromName:  "Sender name",

		ToEmail: "user@example.com",
		ToName:  "Reciver name",

		Subject: "Тест send email",
	}

	err = email.TextPlain(`Hello!
		My email
		Текст моего TEXT сообщения
		`)
	if err != nil {
		return
	}

	err = email.TextHtml(`
		<h2>My email</h2>
		<p>Текст моего HTML</p>
		`)
	if err != nil {
		return
	}

	//err = email.TextHtmlWithRelated(`
	//<h2>My email</h2>
	//<p>Текст моего HTML with related files сообщения</p>
	//<p>Картинка: <img src="cid:StickAndCarrot.jpg" width="500px" height="415px" border="1px" alt="StickAndCarrot"/></p>
	//	`,
	//	"/home/aagafonov/Golang/src/github.com/supme/directEmail/example/StickAndCarrot.jpg",
	//)
	//if err !=nil {
	//	return
	//}

	err = email.Render()
	if err != nil {
		return
	}

	return
}

func BenchmarkDkim(b *testing.B) {
	for i := 0; i < b.N; i++ {
		email, err := emailMessage()
		if err != nil {
			b.Fatal(err)
		}

		err = email.dkimSign("test", []byte(privateKey))
		if err != nil {
			b.Fatal(err)
		}
	}
}
