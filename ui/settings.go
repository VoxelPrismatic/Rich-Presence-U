package ui

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/voxelprismatic/richpresenceu/igdb"
	"github.com/voxelprismatic/richpresenceu/nso"
)

type Settings struct {
	System           nso.System  `json:"system"`
	Platform         string      `json:"platform,omitempty"`
	Language         string      `json:"language"`
	Refresh          int         `json:"refresh"`
	RefreshLast      int64       `json:"refresh_last"`
	AutoConnect      connectMode `json:"auto_connect"`
	KeepOn           bool        `json:"keep_on"`
	DebugLog         bool        `json:"debug_log"`
	Activity         bool        `json:"activity"`
	PauseTimer       bool        `json:"pause_timer"`
	HideDiscord      bool        `json:"hide_discord"`
	AutoUnhide       bool        `json:"auto_unhide,omitempty"`
	Region           nso.Region  `json:"region"`
	WindowW          int         `json:"window_w"`
	WindowH          int         `json:"window_h"`
	WindowX          int         `json:"window_x"`
	WindowY          int         `json:"window_y"`
	InstallDeclined  string      `json:"install_declined,omitempty"`
	UpdateDeclined   string      `json:"update_declined,omitempty"`
	IGDBClientID     string      `json:"igdb_client_id,omitempty"`
	IGDBClientSecret string      `json:"igdb_client_secret,omitempty"`
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
		System:      nso.HAC,
		Platform:    "NS1",
		Refresh:     604800,
		KeepOn:      true,
		DebugLog:    true,
		Activity:    true,
		AutoConnect: connectNever,
		HideDiscord: true,
		Region:      nso.US,
		WindowW:     560,
		WindowH:     640,
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

func platformKey(key string) string {
	if slug := igdb.SlugForStoreCode(key); slug != "" {
		return slug
	}
	return strings.TrimSpace(key)
}

func loadPrefs(dir string) (Settings, map[string]*SystemState) {
	s := defaultSettings()
	systems := map[string]*SystemState{}

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
	applyLegacyHide(&s, b)
	for key, st := range pf.Platforms {
		slug := platformKey(key)
		if slug == "" {
			continue
		}
		if st.Library == nil {
			st.Library = map[string]GameState{}
		}
		copy := st
		systems[slug] = &copy
	}
	return normalizeSettings(s), systems
}

func savePrefs(dir string, s Settings, systems map[string]*SystemState) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	pf := prefsFile{Settings: s, Platforms: map[string]SystemState{}}
	for slug, st := range systems {
		if st != nil && slug != "" {
			pf.Platforms[slug] = *st
		}
	}
	b, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(prefsPath(dir), b, 0o644)
}

func normalizeSettings(s Settings) Settings {
	if s.Platform == "" {
		if !s.System.Valid() || s.System == nso.CTR || s.System == nso.WUP {
			s.System = nso.HAC
		}
		s.Platform = igdb.SlugForStoreCode(string(s.System))
	}
	if p, ok := igdb.BySlug(s.Platform); ok {
		if sys, ok := nso.ParseSystem(p.StoreCode()); ok {
			s.System = sys
		} else {
			s.System = ""
		}
	} else {
		s.Platform = "NS1"
		s.System = nso.HAC
	}
	if !s.Region.Valid() {
		s.Region = nso.US
	}
	switch s.AutoConnect {
	case connectNever, connectStartup, connectPoll:
	default:
		s.AutoConnect = connectNever
	}
	return s
}

const (
	connectNever   connectMode = "never"
	connectStartup connectMode = "startup"
	connectPoll    connectMode = "poll"
)

// connectMode is how the app connects to Discord on its own.
// Older prefs stored auto_connect as a bool: true is startup, false is never.
type connectMode string

func (m *connectMode) UnmarshalJSON(b []byte) error {
	switch strings.TrimSpace(string(b)) {
	case "true":
		*m = connectStartup
	case "false", "null":
		*m = connectNever
	default:
		var text string
		if err := json.Unmarshal(b, &text); err != nil {
			*m = connectNever
			return nil
		}
		*m = connectMode(text)
	}
	return nil
}

func autoConnectHint(mode connectMode) string {
	switch mode {
	case connectStartup:
		return "AUTOCONNECT_HINT_STARTUP"
	case connectPoll:
		return "AUTOCONNECT_HINT_POLL"
	default:
		return "AUTOCONNECT_HINT_NEVER"
	}
}

// applyLegacyHide maps the old hide_behavior string onto the checkboxes when
// a prefs file has neither pause_timer nor hide_discord.
func applyLegacyHide(s *Settings, raw []byte) {
	var probe struct {
		PauseTimer   *bool  `json:"pause_timer"`
		HideDiscord  *bool  `json:"hide_discord"`
		HideBehavior string `json:"hide_behavior"`
	}
	if json.Unmarshal(raw, &probe) != nil {
		return
	}
	if probe.PauseTimer != nil || probe.HideDiscord != nil {
		return
	}
	if probe.HideBehavior == "pause" {
		s.PauseTimer = true
		s.HideDiscord = false
		return
	}
	s.PauseTimer = false
	s.HideDiscord = true
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

func migrateLegacyPlatforms(dir string, systems map[string]*SystemState) {
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
		systems[platformKey(string(sys))] = &copy
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
			if v == "true" {
				s.AutoConnect = connectStartup
			} else {
				s.AutoConnect = connectNever
			}
		case "keep_on":
			s.KeepOn = v == "true"
		case "debug_log":
			s.DebugLog = v == "true"
		case "activity":
			s.Activity = v == "true"
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
