package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/voxelprismatic/richpresenceu/nso"
)

func TestNormalizeSettingsMapsSwitch(t *testing.T) {
	s := normalizeSettings(Settings{System: nso.HAC})
	if s.Platform != "NS1" || s.System != nso.HAC {
		t.Fatalf("%+v", s)
	}
	s = normalizeSettings(Settings{System: nso.BEE})
	if s.Platform != "NS2" || s.System != nso.BEE {
		t.Fatalf("%+v", s)
	}
	s = normalizeSettings(Settings{Platform: "NES"})
	if s.Platform != "NES" || s.System != "" {
		t.Fatalf("NES should not bind an eShop system: %+v", s)
	}
	s = normalizeSettings(Settings{Platform: "nope", System: nso.HAC})
	if s.Platform != "NS1" || s.System != nso.HAC {
		t.Fatalf("unknown platform %+v", s)
	}
}

func TestPrefsMigrateHACKey(t *testing.T) {
	dir := t.TempDir()
	raw := []byte(`{
  "system": "HAC",
  "region": "US",
  "platforms": {
    "HAC": {"game": "70010000012345", "library": {}}
  }
}`)
	if err := os.WriteFile(filepath.Join(dir, "prefs.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	s, systems := loadPrefs(dir)
	if s.Platform != "NS1" || s.System != nso.HAC {
		t.Fatalf("settings %+v", s)
	}
	st := systems["NS1"]
	if st == nil || st.Game != "70010000012345" {
		t.Fatalf("migrated state %+v %v", st, systems)
	}
	if err := savePrefs(dir, s, systems); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "prefs.json"))
	if err != nil {
		t.Fatal(err)
	}
	var pf prefsFile
	if err := json.Unmarshal(b, &pf); err != nil {
		t.Fatal(err)
	}
	if _, ok := pf.Platforms["NS1"]; !ok {
		t.Fatalf("saved keys %v", pf.Platforms)
	}
}

func TestPlatformKey(t *testing.T) {
	if platformKey("HAC") != "NS1" || platformKey("NS1") != "NS1" {
		t.Fatal(platformKey("HAC"), platformKey("NS1"))
	}
}
