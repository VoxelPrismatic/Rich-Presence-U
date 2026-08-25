package ui

import (
	"os"
	"strings"
	"sync"

	"github.com/voxelprismatic/richpresenceu/locales"
)

var localeNames = locales.Names

type i18n struct {
	mu    sync.RWMutex
	lang  string
	table map[string]string
}

func newI18n() *i18n {
	t := &i18n{}
	t.Set("")
	return t
}

func (t *i18n) Set(lang string) {
	if lang == "" {
		lang = os.Getenv("LANG")
	}
	table := locales.Table(lang)
	t.mu.Lock()
	t.lang = uiLang(lang)
	t.table = table
	t.mu.Unlock()
}

func (t *i18n) Lang() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.lang
}

func (t *i18n) T(key string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.table != nil {
		if s, ok := t.table[key]; ok {
			return s
		}
	}
	if s, ok := locales.EnUS[key]; ok {
		return s
	}
	return key
}

func localeCodes() []string {
	return locales.Codes()
}

func uiLang(lang string) string {
	lang = strings.TrimSpace(lang)
	if i := strings.IndexAny(lang, ".@"); i >= 0 {
		lang = lang[:i]
	}
	lang = strings.ToLower(strings.ReplaceAll(lang, "-", "_"))
	if _, ok := locales.Names[lang]; ok {
		return lang
	}
	if i := strings.IndexByte(lang, '_'); i > 0 {
		if _, ok := locales.Names[lang[:i]]; ok {
			return lang[:i]
		}
	}
	return "en"
}
