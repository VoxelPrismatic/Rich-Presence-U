package discord

import "strings"

// FormatTag builds the status line for a Nintendo Network ID or friend code.
// Switch / Switch 2 use the 12-digit SW- boxes. Everything else is freeform
// text in nnid (with a 12-digit TagFC fallback for old 3DS prefs).
func FormatTag(system, nnid string, fc [3]string) string {
	system = strings.ToUpper(system)
	if system == "HAC" || system == "BEE" {
		a, b, c := digits(fc[0]), digits(fc[1]), digits(fc[2])
		if len(a)+len(b)+len(c) < 12 {
			return ""
		}
		return "SW-" + a + "-" + b + "-" + c
	}
	if id := strings.TrimSpace(nnid); id != "" {
		if system == "WUP" {
			return "ID: " + id
		}
		return id
	}
	a, b, c := digits(fc[0]), digits(fc[1]), digits(fc[2])
	if len(a)+len(b)+len(c) < 12 {
		return ""
	}
	return "FC: " + a + "-" + b + "-" + c
}

func digits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
