package ui

import (
	"embed"
	"encoding/csv"
	"io"
	"os"
	"strings"
	"sync"
)

//go:embed locales/*.csv
var localeFS embed.FS

var localeFiles = map[string]string{
	"de": "locales/german.csv",
	"en": "locales/english.csv",
	"es": "locales/spanish.csv",
	"fr": "locales/french.csv",
	"hu": "locales/hungarian.csv",
	"nl": "locales/dutch.csv",
	"pt": "locales/portuguese.csv",
}

var localeNames = map[string]string{
	"de": "Deutsch",
	"en": "English",
	"es": "Español",
	"fr": "Français",
	"hu": "Magyar",
	"nl": "Nederlands",
	"pt": "Português",
}

var fallback = map[string]string{
	"SHORT_DESC":        "Short description",
	"SHORT_DESC_CUSTOM": "Custom",
	"SHORT_DESC_EMPTY":  "Empty",
	"NO_PARTY":          "(no party)",
	"STATUS_ENABLED":    "Status enabled",
	"STATUS_DISABLED":   "Status disabled",
	"STATUS_CONNECTING": "Connecting to Discord...",
	"CONFIGURE":         "Configure",
	"DATA_SELECT":       "Select one",
	"ABOUT_CORE":        "Core",
	"ABOUT_QT":          "Qt UI",
	"ABOUT_INFO":        "Info",
	"SETTINGS_BACK":     "Back",
	"GAME_SYSTEM":       "Game System",
	"PLATFORM_SEARCH":   "Search platforms",
	"TIMER_REMOVE":      "Remove",
	"INSTALL_TITLE":     "Install Rich Presence Qt?",
	"INSTALL_HINT":      "Add a launcher to your application menu?\nThis copies the app into ~/.config/rich-presence-u and downloads the icon and desktop entry from GitHub.",
	"UPDATE_TITLE":      "Update available",
	"UPDATE_HINT":       "Version %s is available. Download and install it now?",
	"UPDATE_AVAILABLE":  "Version %s is available.\nhttps://github.com/VoxelPrismatic/Rich-Presence-U/releases/latest",
	"ESHOP_BUTTON":      "Buy on eShop",
}

type i18n struct {
	mu    sync.RWMutex
	lang  string
	table map[string]string
	en    map[string]string
}

func newI18n() *i18n {
	t := &i18n{en: loadLocale("en")}
	t.Set("")
	return t
}

func (t *i18n) Set(lang string) {
	if lang == "" {
		lang = os.Getenv("LANG")
		if i := strings.IndexAny(lang, "._"); i >= 0 {
			lang = lang[:i]
		}
		lang = strings.ToLower(lang)
	}
	if _, ok := localeFiles[lang]; !ok {
		lang = "en"
	}
	table := t.en
	if lang != "en" {
		table = loadLocale(lang)
	}
	t.mu.Lock()
	t.lang = lang
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
	if s, ok := t.en[key]; ok {
		return s
	}
	if s, ok := fallback[key]; ok {
		return s
	}
	return key
}

func loadLocale(lang string) map[string]string {
	name, ok := localeFiles[lang]
	if !ok {
		return map[string]string{}
	}
	f, err := localeFS.Open(name)
	if err != nil {
		return map[string]string{}
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.LazyQuotes = true
	r.FieldsPerRecord = -1
	out := map[string]string{}
	first := true
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(rec) < 2 {
			continue
		}
		if first {
			first = false
			continue
		}
		out[rec[0]] = rec[1]
	}
	return out
}

func localeCodes() []string {
	return []string{"de", "en", "es", "fr", "hu", "nl", "pt"}
}
