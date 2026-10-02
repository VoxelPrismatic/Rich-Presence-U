package ui

import "testing"

func TestSchemeForLightness(t *testing.T) {
	if schemeForLightness(30, 240) != schemeDark {
		t.Fatal("dark window")
	}
	if schemeForLightness(240, 30) != schemeLight {
		t.Fatal("light window")
	}
	if schemeForLightness(128, 128) != schemeLight {
		t.Fatal("equal lightness stays light")
	}
}
