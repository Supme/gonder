package api

import (
	"encoding/json"
	"github.com/supme/gonder/models"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type languages struct {
	storage map[string]language
}

type language struct {
	Phrases map[string]string `json:"phrases"`
}

func newLang(paths ...string) (*languages, error) {
	l := new(languages)
	l.storage = map[string]language{}
	var (
		files, fi []string
		err       error
	)

	for _, path := range paths {
		fi, err = filepath.Glob(models.FromRootDir(path))
		if err != nil {
			return l, err
		}
		files = append(files, fi...)
	}

	for _, file := range files {
		name := strings.ToLower(strings.TrimRight(filepath.Base(file), filepath.Ext(file)))
		var phrases language
		f, err := os.Open(file)
		if err != nil {
			return l, err
		}
		defer f.Close()
		data, err := ioutil.ReadAll(f)
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

	return l, err
}

func (l *languages) tr(lang, phrase string) string {
	if phrases, ok := l.storage[lang]; ok {
		if translate, ok := phrases.Phrases[phrase]; ok {
			return translate
		}
	}
	return phrase
}
