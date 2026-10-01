package locales

import (
	"strings"
	"testing"
)

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

func TestIGDBInstructions(t *testing.T) {
	s := EnUS["IGDB_INSTRUCTIONS"]
	for _, link := range []string{
		"https://api-docs.igdb.com/#account-creation",
		"https://dev.twitch.tv/login",
		"https://www.twitch.tv/settings/security",
		"https://dev.twitch.tv/console/apps/create",
		"https://dev.twitch.tv/console/apps",
		"In order to use the search games feature, you must have a Twitch account. Game Search is provided by IGDB.",
		"Confidential",
		"Client ID and Client Secret",
	} {
		if !strings.Contains(s, link) {
			t.Fatalf("missing %s", link)
		}
	}
	for _, gone := range []string{"non-commercial", "partner@igdb.com", "developer-agreement"} {
		if strings.Contains(s, gone) {
			t.Fatalf("still contains %s", gone)
		}
	}
	if _, ok := Table("de")["IGDB_INSTRUCTIONS"]; ok {
		t.Fatal("translations should keep the official English instructions via fallback")
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
