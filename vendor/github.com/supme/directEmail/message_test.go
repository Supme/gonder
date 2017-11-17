package directEmail

import (
	"testing"
)

func BenchmarkRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		email := New()

		email.Host = "resolv.name.myhost.com"

		email.FromEmail = "sender@example.com"
		email.FromName = "Sender name"

		email.ToEmail = "user@example.com"
		email.ToName = "Reciver name"

		email.Subject = "Тест send email" // отправки email"

		err := email.TextPlain(`Hello!
		My email
		Текст моего TEXT сообщения
		`)
		if err != nil {
			b.Fatal(err)
		}

		err = email.TextHtml(`
		<h2>My email</h2>
		<p>Текст моего HTML</p>
		`)
		if err != nil {
			b.Fatal(err)
		}

		//err = email.TextHtmlWithRelated(`
		//<h2>My email</h2>
		//<p>Текст моего HTML with related files сообщения</p>
		//<p>Картинка: <img src="cid:StickAndCarrot.jpg" width="500px" height="415px" border="1px" alt="StickAndCarrot"/></p>
		//	`,
		//	"/home/aagafonov/Golang/src/github.com/supme/directEmail/example/StickAndCarrot.jpg",
		//)
		//if err !=nil {
		//	b.Fatal(err)
		//}

		err = email.Render()
		if err != nil {
			b.Fatal(err)
		}
	}
}
