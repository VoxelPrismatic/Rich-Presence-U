package locales

import "strings"

// Names is the language picker label for each UI language code.
var Names = map[string]string{
	"de": "Deutsch",
	"en": "English",
	"es": "Español",
	"fr": "Français",
	"hu": "Magyar",
	"nl": "Nederlands",
	"pt": "Português",
}

var tables = map[string]map[string]string{
	"de":    De,
	"en":    EnUS,
	"en_us": EnUS,
	"es":    Es,
	"fr":    Fr,
	"hu":    Hu,
	"nl":    Nl,
	"pt":    Pt,
}

// Codes is the order of languages in the settings picker.
func Codes() []string {
	return []string{"de", "en", "es", "fr", "hu", "nl", "pt"}
}

// Table returns the translation map for lang (e.g. "de", "en_US", "en_US.UTF-8").
// Unknown languages fall back to English.
func Table(lang string) map[string]string {
	key := normalize(lang)
	if t, ok := tables[key]; ok {
		return t
	}
	if i := strings.IndexByte(key, '_'); i > 0 {
		if t, ok := tables[key[:i]]; ok {
			return t
		}
	}
	return EnUS
}

func normalize(lang string) string {
	lang = strings.TrimSpace(lang)
	if i := strings.IndexAny(lang, ".@"); i >= 0 {
		lang = lang[:i]
	}
	return strings.ToLower(strings.ReplaceAll(lang, "-", "_"))
}
