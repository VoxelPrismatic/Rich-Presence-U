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
		"https://dev.twitch.tv/login",
		"https://www.twitch.tv/settings/security",
		"https://dev.twitch.tv/console/apps/create",
		"https://dev.twitch.tv/console/apps",
		"https://localhost",
		"Ensure that you have 2FA",
		"Use these settings:",
		"Input Order",
		"Application Integration",
		"OAuth Redirect URLs",
		"Confidential",
		"Create!",
		"[New Secret]",
		"Copy the Client ID and Secret into this settings page accordingly",
		"{{NAME}}",
	} {
		if !strings.Contains(s, link) {
			t.Fatalf("missing %s", link)
		}
	}
	for _, gone := range []string{
		"non-commercial",
		"partner@igdb.com",
		"developer-agreement",
		"throwaway-qi74ncw43udwynw",
		"api-docs.igdb.com",
	} {
		if strings.Contains(s, gone) {
			t.Fatalf("still contains %s", gone)
		}
	}
	if _, ok := Table("de")["IGDB_INSTRUCTIONS"]; ok {
		t.Fatal("translations should keep the official English instructions via fallback")
	}

	name := ThrowawayAppName()
	if !validThrowawayName(name) {
		t.Fatalf("name %q", name)
	}
	other := ThrowawayAppName()
	if name == other {
		t.Fatalf("expected a new name, got %q twice", name)
	}
	filled := GameSearchInstructions(s, name)
	if strings.Contains(filled, "{{NAME}}") || !strings.Contains(filled, "<tt>"+name+"</tt>") {
		t.Fatalf("name was not filled: %q", filled)
	}
}

func validThrowawayName(name string) bool {
	const prefix = "throwaway-"
	if !strings.HasPrefix(name, prefix) || len(name) != len(prefix)+16 {
		return false
	}
	for _, c := range name[len(prefix):] {
		if c < 'a' || c > 'z' {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}

func containsNL(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			return true
		}
	}
	return false
}
