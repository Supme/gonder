package models

import (
	"testing"
)

func TestPrepareHtmlTemplate(t *testing.T) {
	Config.UTMDefaultURL = "https://Site.Net"
	tmpl := []struct{src, must string}{
		{
			src:  `<a target="_blank" hRef=" https://Site.com/">`,
			must: `<a target="_blank" hRef="{{RedirectUrl . "https://Site.com/"}}">` + StatHTMLImgTag,
		},
		{
			src:  `<a target="_blank" hRef="{{RedirectUrl . "https://Site.com/"}}">`,
			must: `<a target="_blank" hRef="{{RedirectUrl . "https://Site.com/"}}">` + StatHTMLImgTag,
		},
		{
			src:  `<a target="_blank" hRef="[Url Label] https://Site.com/">`,
			must: `<a target="_blank" hRef="{{RedirectUrl . "[Url Label] https://Site.com/"}}">` + StatHTMLImgTag,
		},
		{
			src:  `<a href="./files/Марк Саммерфильд Программирование на языке Go. Разработка приложений XXI века (2013).pdf">очень полезный файл который нужно прочитать откорки до корки</a>`,
			must: `<a href="https:/Site.Net/files/Марк Саммерфильд Программирование на языке Go. Разработка приложений XXI века (2013).pdf">очень полезный файл который нужно прочитать откорки до корки</a>` + StatHTMLImgTag,
		},
		{
			src:  `<a href="/files/Марк Саммерфильд Программирование на языке Go. Разработка приложений XXI века (2013).pdf">очень полезный файл который нужно прочитать откорки до корки</a>`,
			must: `<a href="https:/Site.Net/files/Марк Саммерфильд Программирование на языке Go. Разработка приложений XXI века (2013).pdf">очень полезный файл который нужно прочитать откорки до корки</a>` + StatHTMLImgTag,
		},
		{
			src:  `<html><body></body></html>`,
			must: `<html><body>` + StatHTMLImgTag + `</body></html>`,
		},
		{
			src:  `Hello!`,
			must: `Hello!` + StatHTMLImgTag,
		},
	}

	for _, tmp := range tmpl {
			res, err := prepareHTMLTemplate(tmp.src, false)
			if err != nil {
				t.Errorf("prepare html: %s", err)
			}
			if res != tmp.must {
				t.Errorf("prepared html src: '%s' result: '%s', but need: '%s'", tmp.src, res, tmp.must)
			}
	}
}

func TestPrepareAmpTemplate(t *testing.T) {
	Config.UTMDefaultURL = "https://Site.Net"
	tmpl := []struct{src, must string}{
		{
			src:  `<a target="_blank" hRef="https://Site.com/">`,
			must: `<a target="_blank" hRef="{{RedirectUrl . "https://Site.com/"}}">` + StatAMPImgTag,
		},
		{
			src:  `<html><body></body></html>`,
			must: `<html><body>` + StatAMPImgTag + `</body></html>`,
		},
		{
			src:  `Hello!`,
			must: `Hello!` + StatAMPImgTag,
		},
	}

	for _, tmp := range tmpl {
			res, err := prepareAMPTemplate(tmp.src)
			if err != nil {
				t.Errorf("prepare amp: %s", err)
			}
			if res != tmp.must {
				t.Errorf("prepared amp src: '%s' result: '%s', but need: '%s'", tmp.src, res, tmp.must)
			}
	}
}

var (
	srcTmpl = `
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

	mustTmpl = `
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
  <img src="{{.StatUrl}}" border="0" width="10" height="10" alt=""/></body>
</html>
`
)

func TestHtmlStringPrepare(t *testing.T) {
	Config.UTMDefaultURL = "https://Site.Net"
	res, err := prepareHTMLTemplate(srcTmpl, false)
	if err != nil {
		t.Error(err)
	}
	// fmt.Printf("Html template result:\n%s", srcTmpl)
	if res != mustTmpl {
		t.Error("html result string prepare template is not equal")
	}
}

func BenchmarkHtmlStringPrepare(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, _ = prepareHTMLTemplate(srcTmpl, false)
	}
}

func BenchmarkHtmlStringPrepareCompress(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, _ = prepareHTMLTemplate(srcTmpl, true)
	}
}
