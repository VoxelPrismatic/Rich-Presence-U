package ui

import (
	"os"
	"testing"

	"github.com/voxelprismatic/richpresenceu/nso"
)

func TestPlatformRowsRoundTrip(t *testing.T) {
	st := defaultSystem()
	st.Game = "newer"
	st.History = []string{"older-recent", "newer"}
	st.TagFC = [3]string{"1111", "2222", "3333"}
	st.TagID = "nnid"
	st.TagIcon = true
	st.TimePreserve = true
	st.Library = map[string]GameState{
		"old":          {Mode: "custom", Description: "kept", Party: true, PartySize: 2, PartyMax: 4, Region: nso.EU},
		"older-recent": {Mode: "empty"},
		"newer":        {Mode: "friendcode", Description: "fc", Region: nso.US},
	}
	systems := map[string]*SystemState{"NS1": &st}
	prefs, hist := platformsToRows(systems)
	if len(prefs) != 1 || prefs[0].Platform != "NS1" || prefs[0].GameID != "newer" || !prefs[0].FcAsIcon {
		t.Fatalf("prefs %+v", prefs)
	}
	fc, tag := decodeFriendCode(prefs[0].FriendCode)
	if fc != st.TagFC || tag != "nnid" {
		t.Fatalf("code %v %q", fc, tag)
	}

	back := systemsFromDB(prefs, hist)
	got := back["NS1"]
	if got == nil || got.Game != "newer" || got.TagID != "nnid" || got.TagFC != st.TagFC || !got.TimePreserve {
		t.Fatalf("back %+v", got)
	}
	if len(got.History) != 2 || got.History[0] != "older-recent" || got.History[1] != "newer" {
		t.Fatalf("history %#v", got.History)
	}
	old := got.Library["old"]
	if old.Mode != "custom" || old.Description != "kept" || !old.Party || old.Region != nso.EU {
		t.Fatalf("old game %+v", old)
	}
	if got.Library["newer"].Mode != "friendcode" || got.Library["newer"].Region != nso.US {
		t.Fatalf("newer %+v", got.Library["newer"])
	}
}

func TestPlatformsFieldPopulated(t *testing.T) {
	if platformsFieldPopulated(nil) || platformsFieldPopulated(map[string]*SystemState{}) {
		t.Fatal("empty platforms field")
	}
	blank := defaultSystem()
	if platformsFieldPopulated(map[string]*SystemState{"": &blank}) {
		t.Fatal("blank key")
	}
	st := defaultSystem()
	if !platformsFieldPopulated(map[string]*SystemState{"NS1": &st}) {
		t.Fatal("platform row should count as populated")
	}
}

func TestMigrateSkipsEmptyPlatformsField(t *testing.T) {
	dir := t.TempDir()
	c, err := nso.New(dir, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	st := defaultSystem()
	st.Game = "from-db"
	prefs, hist := platformsToRows(map[string]*SystemState{"NS1": &st})
	if err := c.SavePlatforms(prefs, hist); err != nil {
		t.Fatal(err)
	}
	settings := defaultSettings()
	if err := savePrefs(dir, settings); err != nil {
		t.Fatal(err)
	}
	settings, fromFile := loadPrefs(dir)
	got := loadPlatforms(c, settings, fromFile)
	if got["NS1"] == nil || got["NS1"].Game != "from-db" {
		t.Fatalf("database row changed: %+v", got["NS1"])
	}
}

func TestMigratePlatformsOutOfPrefs(t *testing.T) {
	dir := t.TempDir()
	raw := []byte(`{
  "platform": "NS1",
  "region": "US",
  "platforms": {
    "HAC": {
      "game": "70010000012345",
      "history": ["70010000012345"],
      "tag_fc": ["1111", "2222", "3333"],
      "tag_icon": true,
      "library": {
        "70010000012345": {"mode": "custom", "description": "zelda", "region": "US"}
      }
    }
  }
}`)
	if err := os.WriteFile(prefsPath(dir), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := nso.New(dir, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	settings, fromFile := loadPrefs(dir)
	systems := loadPlatforms(c, settings, fromFile)
	st := systems["NS1"]
	if st == nil || st.Game != "70010000012345" || st.Library["70010000012345"].Description != "zelda" {
		t.Fatalf("migrated %+v", st)
	}
	if prefsHavePlatforms(dir) {
		b, _ := os.ReadFile(prefsPath(dir))
		t.Fatalf("platforms still in prefs: %s", b)
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}

	c2, err := nso.New(dir, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c2.Close() })
	settings, fromFile = loadPrefs(dir)
	if len(fromFile) != 0 {
		t.Fatalf("json platforms should be gone: %+v", fromFile)
	}
	systems = loadPlatforms(c2, settings, fromFile)
	st = systems["NS1"]
	if st == nil || st.Game != "70010000012345" || !st.TagIcon || st.TagFC[0] != "1111" {
		t.Fatalf("reloaded %+v", st)
	}
	if len(st.History) != 1 || st.History[0] != "70010000012345" {
		t.Fatalf("history %#v", st.History)
	}
}
