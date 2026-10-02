package ui

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

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

func TestConnectPollInterval(t *testing.T) {
	if connectPollEvery != 5*time.Second {
		t.Fatal(connectPollEvery)
	}
}

func TestAutoConnectLegacyBool(t *testing.T) {
	var s Settings
	if err := json.Unmarshal([]byte(`{"auto_connect": true}`), &s); err != nil {
		t.Fatal(err)
	}
	if s.AutoConnect != connectStartup {
		t.Fatalf("true = %q", s.AutoConnect)
	}
	if err := json.Unmarshal([]byte(`{"auto_connect": false}`), &s); err != nil {
		t.Fatal(err)
	}
	if s.AutoConnect != connectNever {
		t.Fatalf("false = %q", s.AutoConnect)
	}
	if err := json.Unmarshal([]byte(`{"auto_connect": "poll"}`), &s); err != nil {
		t.Fatal(err)
	}
	if s.AutoConnect != connectPoll {
		t.Fatalf("poll = %q", s.AutoConnect)
	}
	s = normalizeSettings(Settings{Platform: "NS1", AutoConnect: "nope"})
	if s.AutoConnect != connectNever {
		t.Fatalf("unknown = %q", s.AutoConnect)
	}
}

func TestTestUpdatePref(t *testing.T) {
	dir := t.TempDir()
	raw := []byte(`{"test_update": true, "platform": "NS1", "region": "US", "platforms": {"NS1": {"game": "1", "library": {}}}}`)
	if err := os.WriteFile(prefsPath(dir), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	s, systems := loadPrefs(dir)
	if !s.TestUpdate {
		t.Fatal("test_update should load")
	}
	if err := savePrefs(dir, s, systems); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(prefsPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(b, []byte(`"test_update": true`)) {
		t.Fatalf("saved prefs: %s", b)
	}
	s.TestUpdate = false
	if err := savePrefs(dir, s, systems); err != nil {
		t.Fatal(err)
	}
	b, err = os.ReadFile(prefsPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(b, []byte("test_update")) {
		t.Fatalf("false test_update should be omitted: %s", b)
	}
}

func TestPrefsKeepPlatformsWhenAutoConnectIsBool(t *testing.T) {
	dir := t.TempDir()
	raw := []byte(`{"auto_connect": true, "platform": "NS1", "region": "US", "platforms": {"NS1": {"game": "1", "library": {}}}}`)
	if err := os.WriteFile(prefsPath(dir), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	s, systems := loadPrefs(dir)
	if s.AutoConnect != connectStartup {
		t.Fatalf("mode %q", s.AutoConnect)
	}
	st := systems["NS1"]
	if st == nil || st.Game != "1" {
		t.Fatalf("platforms %+v", systems)
	}
}

func TestApplyLegacyHide(t *testing.T) {
	s := defaultSettings()
	applyLegacyHide(&s, []byte(`{"hide_behavior":"pause"}`))
	if !s.PauseTimer || s.HideDiscord {
		t.Fatalf("pause %+v", s.PauseTimer)
	}
	s = defaultSettings()
	applyLegacyHide(&s, []byte(`{"hide_behavior":"discord"}`))
	if s.PauseTimer || !s.HideDiscord {
		t.Fatalf("discord pause %v hide %v", s.PauseTimer, s.HideDiscord)
	}
	s = Settings{PauseTimer: true, HideDiscord: true}
	applyLegacyHide(&s, []byte(`{"pause_timer":true,"hide_discord":true,"hide_behavior":"pause"}`))
	if !s.PauseTimer || !s.HideDiscord {
		t.Fatalf("both %+v", s)
	}
	s = Settings{}
	applyLegacyHide(&s, []byte(`{"pause_timer":false,"hide_discord":false}`))
	if s.PauseTimer || s.HideDiscord {
		t.Fatalf("explicit off %+v", s)
	}
}

func TestAutoConnectHint(t *testing.T) {
	if autoConnectHint(connectNever) != "AUTOCONNECT_HINT_NEVER" {
		t.Fatal(autoConnectHint(connectNever))
	}
	if autoConnectHint(connectStartup) != "AUTOCONNECT_HINT_STARTUP" {
		t.Fatal(autoConnectHint(connectStartup))
	}
	if autoConnectHint(connectPoll) != "AUTOCONNECT_HINT_POLL" {
		t.Fatal(autoConnectHint(connectPoll))
	}
}

func TestPlatformKey(t *testing.T) {
	if platformKey("HAC") != "NS1" || platformKey("NS1") != "NS1" {
		t.Fatal(platformKey("HAC"), platformKey("NS1"))
	}
}
