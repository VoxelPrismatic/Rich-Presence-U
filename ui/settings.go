package ui

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/voxelprismatic/richpresenceu/nso"
)

type Settings struct {
	System          nso.System `json:"system"`
	Language        string     `json:"language"`
	Refresh         int        `json:"refresh"`
	RefreshLast     int64      `json:"refresh_last"`
	AutoConnect     bool       `json:"auto_connect"`
	KeepOn          bool       `json:"keep_on"`
	DebugLog        bool       `json:"debug_log"`
	Activity        bool       `json:"activity"`
	Timer           int        `json:"timer"`
	Region          nso.Region `json:"region"`
	WindowW         int        `json:"window_w"`
	WindowH         int        `json:"window_h"`
	WindowX         int        `json:"window_x"`
	WindowY         int        `json:"window_y"`
	InstallDeclined string     `json:"install_declined,omitempty"`
	UpdateDeclined  string     `json:"update_declined,omitempty"`
}

type GameState struct {
	Mode        string     `json:"mode"` // friendcode, custom, empty
	Description string     `json:"description"`
	Party       bool       `json:"party"`
	PartySize   int        `json:"party_size"`
	PartyMax    int        `json:"party_max"`
	Region      nso.Region `json:"region"`
}

type SystemState struct {
	Game         string               `json:"game"`
	History      []string             `json:"history"`
	TagFC        [3]string            `json:"tag_fc"`
	TagID        string               `json:"tag_id"`
	TagIcon      bool                 `json:"tag_icon"`
	TimePreserve bool                 `json:"time_preserve"`
	Library      map[string]GameState `json:"library"`
}

type prefsFile struct {
	Settings
	Platforms map[string]SystemState `json:"platforms"`
}

func defaultSettings() Settings {
	return Settings{
		System:   nso.HAC,
		Refresh:  604800,
		KeepOn:   true,
		DebugLog: true,
		Activity: true,
		Region:   nso.US,
		WindowW:  560,
		WindowH:  640,
	}
}

func defaultGame() GameState {
	return GameState{Mode: "empty", PartySize: 1, PartyMax: 1}
}

func defaultSystem() SystemState {
	return SystemState{Library: map[string]GameState{}}
}

func prefsPath(dir string) string {
	return filepath.Join(dir, "prefs.json")
}

func loadPrefs(dir string) (Settings, map[nso.System]*SystemState) {
	s := defaultSettings()
	systems := map[nso.System]*SystemState{}
	for _, sys := range nso.Systems {
		st := defaultSystem()
		systems[sys] = &st
	}

	b, err := os.ReadFile(prefsPath(dir))
	if err != nil {
		s = migrateGodotSettings(legacyDataDir(), s)
		migrateLegacyPlatforms(legacyDataDir(), systems)
		return normalizeSettings(s), systems
	}
	var pf prefsFile
	if err := json.Unmarshal(b, &pf); err != nil {
		return normalizeSettings(s), systems
	}
	s = pf.Settings
	for key, st := range pf.Platforms {
		sys, ok := nso.ParseSystem(key)
		if !ok {
			continue
		}
		if st.Library == nil {
			st.Library = map[string]GameState{}
		}
		copy := st
		systems[sys] = &copy
	}
	return normalizeSettings(s), systems
}

func savePrefs(dir string, s Settings, systems map[nso.System]*SystemState) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	pf := prefsFile{Settings: s, Platforms: map[string]SystemState{}}
	for sys, st := range systems {
		if st != nil {
			pf.Platforms[string(sys)] = *st
		}
	}
	b, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(prefsPath(dir), b, 0o644)
}

func normalizeSettings(s Settings) Settings {
	if !s.System.Valid() || s.System == nso.CTR || s.System == nso.WUP {
		s.System = nso.HAC
	}
	if !s.Region.Valid() {
		s.Region = nso.US
	}
	return s
}

func legacyDataDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "rich_presence_u")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".local", "share", "rich_presence_u")
}

func migrateLegacyPlatforms(dir string, systems map[nso.System]*SystemState) {
	if dir == "" {
		return
	}
	for _, sys := range nso.Systems {
		b, err := os.ReadFile(filepath.Join(dir, "platforms", sys.Lower()+".json"))
		if err != nil {
			continue
		}
		st := defaultSystem()
		_ = json.Unmarshal(b, &st)
		if st.Library == nil {
			st.Library = map[string]GameState{}
		}
		type oldLib struct {
			Description string `json:"description"`
			Party       bool   `json:"party"`
			PartySize   int    `json:"party_size"`
			PartyMax    int    `json:"party_max"`
			Region      string `json:"region"`
		}
		var raw struct {
			Library map[string]oldLib `json:"library"`
		}
		if json.Unmarshal(b, &raw) == nil {
			for id, g := range raw.Library {
				cur := st.Library[id]
				if cur.Mode == "" {
					if strings.TrimSpace(g.Description) != "" {
						cur.Mode = "custom"
						cur.Description = g.Description
					} else {
						cur.Mode = "empty"
					}
				}
				if cur.PartySize == 0 {
					cur.PartySize = g.PartySize
				}
				if cur.PartyMax == 0 {
					cur.PartyMax = g.PartyMax
				}
				if !cur.Party {
					cur.Party = g.Party
				}
				if cur.Region == "" {
					cur.Region, _ = nso.ParseRegion(g.Region)
				}
				st.Library[id] = cur
			}
		}
		copy := st
		systems[sys] = &copy
	}
}

func migrateGodotSettings(dir string, s Settings) Settings {
	if dir == "" {
		return s
	}
	f, err := os.Open(filepath.Join(dir, "settings.cfg"))
	if err != nil {
		if b, err := os.ReadFile(filepath.Join(dir, "settings.json")); err == nil {
			_ = json.Unmarshal(b, &s)
		}
		return s
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "[") || !strings.Contains(line, "=") {
			continue
		}
		k, v, _ := strings.Cut(line, "=")
		k = strings.TrimSpace(k)
		v = strings.Trim(strings.TrimSpace(v), `"`)
		switch k {
		case "system":
			if sys, ok := nso.ParseSystem(v); ok {
				s.System = sys
			}
		case "language":
			s.Language = v
		case "refresh":
			s.Refresh, _ = strconv.Atoi(v)
		case "refresh_last":
			s.RefreshLast, _ = strconv.ParseInt(v, 10, 64)
		case "auto_connect":
			s.AutoConnect = v == "true"
		case "keep_on":
			s.KeepOn = v == "true"
		case "debug_log":
			s.DebugLog = v == "true"
		case "activity":
			s.Activity = v == "true"
		case "timer":
			s.Timer, _ = strconv.Atoi(v)
		}
	}
	if b, err := os.ReadFile(filepath.Join(dir, "settings.json")); err == nil {
		_ = json.Unmarshal(b, &s)
	}
	return s
}

func (s Settings) RefreshEvery() time.Duration {
	if s.Refresh <= 0 {
		return 0
	}
	return time.Duration(s.Refresh) * time.Second
}
