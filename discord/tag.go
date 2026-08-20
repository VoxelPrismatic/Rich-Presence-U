package discord

import "strings"

// FormatTag builds the status line for a Nintendo Network ID or friend code.
// system is CTR / WUP / HAC / BEE.
func FormatTag(system, nnid string, fc [3]string) string {
	system = strings.ToUpper(system)
	if system == "WUP" {
		id := strings.TrimSpace(nnid)
		if id == "" {
			return ""
		}
		return "ID: " + id
	}
	a, b, c := digits(fc[0]), digits(fc[1]), digits(fc[2])
	if len(a)+len(b)+len(c) < 12 {
		return ""
	}
	code := a + "-" + b + "-" + c
	if system == "HAC" || system == "BEE" {
		return "SW-" + code
	}
	return "FC: " + code
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
