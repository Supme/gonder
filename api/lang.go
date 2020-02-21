package api

import (
	"encoding/json"
	"gonder/bindata"
	"path/filepath"
	"strings"
)

type languages struct {
	storage map[string]language
}

type language struct {
	Phrases map[string]string `json:"phrases"`
}

func newLang() (*languages, error) {
	l := new(languages)
	l.storage = map[string]language{}

	for _, path := range []string{"panel/assets/w2ui/locale", "panel/assets/gonder/locale"} {
		files, err := bindata.AssetDir(path)
		if err != nil {
			return l, err
		}

		for _, file := range files {
			name := strings.ToLower(strings.TrimRight(filepath.Base(file), filepath.Ext(file)))
			var phrases language

			data, err := bindata.Asset(filepath.Join(path, file))
			if err != nil {
				return l, err
			}

			err = json.Unmarshal(data, &phrases)
			if err != nil {
				return l, err
			}

			if _, ok := l.storage[name]; !ok {
				l.storage[name] = phrases
			} else {
				for ph, tr := range phrases.Phrases {
					l.storage[name].Phrases[ph] = tr
				}
			}
		}
	}

	return l, nil
}

func (l *languages) tr(lang, phrase string) string {
	if phrases, ok := l.storage[lang]; ok {
		if translate, ok := phrases.Phrases[phrase]; ok {
			return translate
		}
	}
	return phrase
}
