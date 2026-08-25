package locales

import "testing"

func TestTableEnglishDefault(t *testing.T) {
	if Table("")["ABOUT_CLOSE"] != "Close" {
		t.Fatal(Table("")["ABOUT_CLOSE"])
	}
	if Table("en_US.UTF-8")["GAME_SYSTEM"] != "Game System" {
		t.Fatal(Table("en_US.UTF-8")["GAME_SYSTEM"])
	}
}

func TestTableGerman(t *testing.T) {
	de := Table("de_DE")
	if de["SETTINGS_BACK"] != "Zurück" {
		t.Fatalf("%q", de["SETTINGS_BACK"])
	}
	if de["ABOUT_CLOSE"] != "Schließen" {
		t.Fatalf("%q", de["ABOUT_CLOSE"])
	}
}

func TestNewlines(t *testing.T) {
	s := EnUS["CONNECTION_ERROR_HINT"]
	if !containsNL(s) {
		t.Fatalf("expected newline in hint: %q", s)
	}
}

func containsNL(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			return true
		}
	}
	return false
}
