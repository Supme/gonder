package api

import (
	"encoding/json"
	"gonder/panel"
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

	for _, path := range []string{"assets/w2ui/locale", "assets/gonder/locale"} {
		files, err := panel.Assets.ReadDir(path)
		if err != nil {
			return l, err
		}

		for _, file := range files {
			name := strings.ToLower(strings.TrimRight(filepath.Base(file.Name()), filepath.Ext(file.Name())))
			var phrases language

			data, err := panel.Assets.ReadFile(filepath.Join(path, file.Name()))
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
