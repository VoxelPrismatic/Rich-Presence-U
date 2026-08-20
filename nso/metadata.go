package nso

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

const DefaultMetadataURL = "https://gist.github.com/ninstar/19c664a823d3a0312f47f5ac5e52a915/raw"

const DefaultClientID = "985449859299565649"

const (
	defaultWUPClient = "1259966953573847130"
	defaultHACClient = "1259967215323840564"
	defaultCTRClient = "1259967368000569394"
	defaultBEEClient = "1385689410502263016"
)

// Metadata is the remote config for Discord client ids and help/home URLs.
// Title CSV / asset fields are still parsed from the gist if present, but are
// not used: search is eShop-only.
type Metadata struct {
	Latest  int
	Minimal int
	Version string
	Changes string
	DLC     string
	BinURL  string

	Clients map[System]string
	Titles  map[System]string
	Assets  map[System]string

	Home     string
	Code     string
	Contact  string
	Group    string
	Help     string
	HelpLang map[string]string
}

func DefaultMetadata() Metadata {
	m := Metadata{
		Latest:  1600,
		Minimal: 1600,
		Clients: map[System]string{
			WUP: defaultWUPClient,
			HAC: defaultHACClient,
			CTR: defaultCTRClient,
			BEE: defaultBEEClient,
		},
		Titles:   map[System]string{},
		Assets:   map[System]string{},
		Home:     "https://ninstars.blogspot.com/rpc",
		Code:     "https://github.com/ninstar/Rich-Presence-U",
		Contact:  "https://ninstar.carrd.co",
		Group:    "https://discord.gg/N9bMDEcrX4",
		Help:     "https://discord.gg/N9bMDEcrX4",
		HelpLang: map[string]string{},
	}
	return m
}

func (m Metadata) ClientID(system System) string {
	if id := m.Clients[system]; id != "" {
		return id
	}
	return DefaultClientID
}

func (m Metadata) TitlesURL(system System) string {
	return m.Titles[system]
}

func (m Metadata) AssetsURL(system System) string {
	return m.Assets[system]
}

func (m Metadata) HelpURL(lang string) string {
	if lang != "" {
		if u := m.HelpLang[lang]; u != "" {
			return u
		}
	}
	return m.Help
}

func ParseMetadata(r io.Reader) (Metadata, error) {
	m := DefaultMetadata()
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	section := ""
	lineNo := 0
	for s.Scan() {
		lineNo++
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			continue
		}
		key, val, ok := splitKV(line)
		if !ok {
			return m, fmt.Errorf("metadata:%d: expected key=value", lineNo)
		}
		applyMeta(&m, section, key, val)
	}
	if err := s.Err(); err != nil {
		return m, err
	}
	return m, nil
}

func splitKV(line string) (string, string, bool) {
	i := strings.IndexByte(line, '=')
	if i <= 0 {
		return "", "", false
	}
	key := strings.ToLower(strings.TrimSpace(line[:i]))
	raw := strings.TrimSpace(line[i+1:])
	return key, unquote(raw), true
}

func unquote(s string) string {
	if len(s) >= 2 && s[0] == '"' {
		out, err := strconv.Unquote(s)
		if err == nil {
			return out
		}
		// Godot often writes "foo\nbar" without full Go quoting. Strip
		// surrounding quotes and interpret common escapes.
		inner := s[1:]
		if strings.HasSuffix(inner, `"`) {
			inner = inner[:len(inner)-1]
		}
		inner = strings.ReplaceAll(inner, `\n`, "\n")
		inner = strings.ReplaceAll(inner, `\t`, "\t")
		inner = strings.ReplaceAll(inner, `\"`, `"`)
		return inner
	}
	return s
}

func applyMeta(m *Metadata, section, key, val string) {
	switch section {
	case "bin":
		switch key {
		case "latest":
			m.Latest = atoi(val, m.Latest)
		case "minimal":
			m.Minimal = atoi(val, m.Minimal)
		case "version":
			m.Version = val
		case "changes":
			m.Changes = val
		case "dlc":
			m.DLC = val
		case "url":
			m.BinURL = val
		}
	case "dlc":
		if sys, rest, ok := splitSystemKey(key); ok {
			switch rest {
			case "client":
				m.Clients[sys] = val
			case "titles":
				m.Titles[sys] = val
			case "assets":
				if val != "" && !strings.HasSuffix(val, "/") {
					val += "/"
				}
				m.Assets[sys] = val
			}
		}
	case "url":
		switch {
		case key == "home":
			m.Home = val
		case key == "code":
			m.Code = val
		case key == "contact":
			m.Contact = val
		case key == "group":
			m.Group = val
		case key == "help":
			m.Help = val
		case strings.HasPrefix(key, "help_"):
			m.HelpLang[strings.TrimPrefix(key, "help_")] = val
		}
	}
}

func splitSystemKey(key string) (System, string, bool) {
	i := strings.IndexByte(key, '_')
	if i <= 0 {
		return "", "", false
	}
	sys, ok := ParseSystem(key[:i])
	if !ok {
		return "", "", false
	}
	return sys, key[i+1:], true
}

func atoi(s string, fallback int) int {
	n, err := strconv.Atoi(strings.TrimFunc(s, unicode.IsSpace))
	if err != nil {
		return fallback
	}
	return n
}
