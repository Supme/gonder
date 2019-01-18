package api

import "testing"

func TestLang(t *testing.T) {
	l, err := loadLang()
	if err != nil {
		t.Fatal(err)
	}

	translate := map[string]string{
		"Campaign": "Кампания",
		"Name":     "Имя",
		"Yes":      "Да",
		"No":       "Нет",
	}
	for ph, tr := range translate {
		ltr := l.tr("ru-ru", ph)
		if ltr != tr {
			t.Errorf("'%s' not translate as '%s' must be '%s'", ph, ltr, tr)
		}
	}

	trQwerty := l.tr("ru-ru", "qwerty")
	if trQwerty != "qwerty" {
		t.Errorf("'%s' not equal'%s'", trQwerty, "qwerty")
	}
}

func BenchmarkLoadLang(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, err := loadLang()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTranslate(b *testing.B) {
	l, err := loadLang()
	if err != nil {
		b.Fatal(err)
	}
	testData := []string{
		"Search",
		"Search...",
		"Select Search Field",
		"selected",
		"Server Response",
		"Show",
		"Subject",
		"From name",
		"From email",
		"Start date",
		"Start time",
		"End date",
		"End time",
		"Compress HTML",
		"Send unsubscribe",
		"Settings",
		"Result",
		"Error",
		"Show/hide columns",
		"Skip",
		"Sorting took",
		"Toggle Line Numbers",
		"Yes",
		"Yesterday",
		"Save Grid State",
		"Restore Default State",
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for i := range testData {
			_ = l.tr("ru-ru", testData[i])
		}
	}
}

func loadLang() (*languages, error) {
	return newLang("../panel/assets/w2ui/locale/*.json", "../panel/assets/gonder/locale/*.json")
}
