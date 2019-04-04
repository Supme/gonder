package campaign

import (
	"gonder/models"
	"testing"
)

var (
	tmpl = `
<!DOCTYPE html>
<html>
<head><meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
	<title>Прелесть</title>
</head>
  <body>
    <!-- WEB-VERSION {{if .WebUrl}} -->
      <p>Если письмо отображается некорректно, пожалуйста, воспользуйтесь <a href="{{.WebUrl}}" style="color: #9e9e9e;" target="_blank">этой ссылкой</a> для просмотра.</p>
    <!-- /WEB-VERSION {{end}} -->

	Привет, {{.Name}}! Ты {{if .Sex eq "male"}}мужчина{{else}}женщина{{endif}}. {{.Age}} лет.

    <p>
      Приложим файлы:<br/>
      <a href="files/Марк Саммерфильд Программирование на языке Go. Разработка приложений XXI века (2013).pdf">очень полезный файл который нужно прочитать откорки до корки</a><br/>
      <a href="/files/Марк Саммерфильд Программирование на языке Go. Разработка приложений XXI века (2013).pdf">очень полезный файл который нужно прочитать откорки до корки</a><br/>
      <a href="./files/Марк Саммерфильд Программирование на языке Go. Разработка приложений XXI века (2013).pdf">очень полезный файл который нужно прочитать откорки до корки</a><br/>
    </p>

    <p>
      <a href="#">Это решетка</a>
      <a href="mailto:aaa@ddd">Это мыло</a>
      <a href="tel:+7906457412">Это телефон</a>
      <a href="sms:+74566441234">Это смс</a>
      <a href="[My Site link] HttPs://www.Site.Net/index.php?a=test1&amp;b=test2&amp;c=test3&amp;d=test4" target="_blank">Ссылка на сайт</a>
    </p>

    <p>
      <img alt="" src="files/content/прелесть.jpg" style="width: 600px; height: 450px;" /><br/>
      <img alt="" src="/files/content/прелесть.jpg" style="width: 600px; height: 450px;" /><br/>
      <img alt="" src="./files/content/прелесть.jpg" style="width: 600px; height: 450px;" /><br/>
    </p>

    <p>
      Если Вы не желаете получать информацию, пожалуйста, <a href="{{.UnsubscribeUrl}}" style="color: #9e9e9e;" target="_blank"> откажитесь от подписки</a>
    </p>
  </body>
</html>
`

	goodTmpl = `
<!DOCTYPE html>
<html>
<head><meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
	<title>Прелесть</title>
</head>
  <body>
    <!-- WEB-VERSION {{if .WebUrl}} -->
      <p>Если письмо отображается некорректно, пожалуйста, воспользуйтесь <a href="{{.WebUrl}}" style="color: #9e9e9e;" target="_blank">этой ссылкой</a> для просмотра.</p>
    <!-- /WEB-VERSION {{end}} -->

	Привет, {{.Name}}! Ты {{if .Sex eq "male"}}мужчина{{else}}женщина{{endif}}. {{.Age}} лет.

    <p>
      Приложим файлы:<br/>
      <a href="https:/Site.Net/files/Марк Саммерфильд Программирование на языке Go. Разработка приложений XXI века (2013).pdf">очень полезный файл который нужно прочитать откорки до корки</a><br/>
      <a href="https:/Site.Net/files/Марк Саммерфильд Программирование на языке Go. Разработка приложений XXI века (2013).pdf">очень полезный файл который нужно прочитать откорки до корки</a><br/>
      <a href="https:/Site.Net/files/Марк Саммерфильд Программирование на языке Go. Разработка приложений XXI века (2013).pdf">очень полезный файл который нужно прочитать откорки до корки</a><br/>
    </p>

    <p>
      <a href="#">Это решетка</a>
      <a href="mailto:aaa@ddd">Это мыло</a>
      <a href="tel:+7906457412">Это телефон</a>
      <a href="sms:+74566441234">Это смс</a>
      <a href="{{RedirectUrl . "[My Site link] HttPs://www.Site.Net/index.php?a=test1&amp;b=test2&amp;c=test3&amp;d=test4"}}" target="_blank">Ссылка на сайт</a>
    </p>

    <p>
      <img alt="" src="https:/Site.Net/files/content/прелесть.jpg" style="width: 600px; height: 450px;" /><br/>
      <img alt="" src="https:/Site.Net/files/content/прелесть.jpg" style="width: 600px; height: 450px;" /><br/>
      <img alt="" src="https:/Site.Net/files/content/прелесть.jpg" style="width: 600px; height: 450px;" /><br/>
    </p>

    <p>
      Если Вы не желаете получать информацию, пожалуйста, <a href="{{.UnsubscribeUrl}}" style="color: #9e9e9e;" target="_blank"> откажитесь от подписки</a>
    </p>
  <img src='{{.StatPng}}' border='0px' width='10px' height='10px'/></body>
</html>
`
)

var r = Recipient{
	ID:         "test",
	CampaignID: "testCampaign",
	Email:      "test@site.tld",
	Name:       "Вася",
	Params: map[string]interface{}{
		"Sex": "male",
		"Age": 32,
	},
}

var c = campaign{
	ID:           "testCampaign",
	templateHTML: tmpl,
}

func init() {
	models.Config.URL = "https://Site.Net"
}

func TestHtmlStringPrepare(t *testing.T) {
	prepareHTMLTemplate(&tmpl, false)
	//fmt.Printf("Html template result:\n%s", tmpl)
	if tmpl != goodTmpl {
		t.Error("html result string prepare template is not equal")
	}
}

func BenchmarkHtmlStringPrepare(b *testing.B) {
	for n := 0; n < b.N; n++ {
		prepareHTMLTemplate(&tmpl, false)
	}
}

func BenchmarkHtmlStringPrepareCompress(b *testing.B) {
	for n := 0; n < b.N; n++ {
		prepareHTMLTemplate(&tmpl, true)
	}
}
