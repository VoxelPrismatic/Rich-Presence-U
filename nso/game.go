package nso

import "strings"

// Game is one eShop title (or a previously selected store hit saved locally).
type Game struct {
	// ID is the store nsuid, or "eu:" + fs_id when nsuid is missing. Switch
	// titles listed under Switch 2 are prefixed with hac::.
	ID string
	// AssetSystem is the platform the title originally shipped on.
	AssetSystem System
	Icons       map[Region]bool
	Titles      map[Region]string
	CoverArt    string
}

func (g Game) EnglishTitle() string {
	if t := g.Titles[US]; t != "" {
		return t
	}
	return g.Title(US)
}

func (g Game) Verified() bool {
	return len(g.Titles) > 0
}

func (g Game) NativeID() string {
	return strings.TrimPrefix(g.ID, HACPrefix)
}

func (g Game) FromSwitch() bool {
	return strings.HasPrefix(g.ID, HACPrefix)
}

// IsStoreID reports whether id looks like an eShop nsuid / Europe fs_id,
// as opposed to a slug from the old Rich Presence U CSV dump.
func IsStoreID(id string) bool {
	id = strings.TrimPrefix(id, HACPrefix)
	if strings.HasPrefix(strings.ToLower(id), "eu:") {
		return true
	}
	if len(id) < 10 {
		return false
	}
	for _, r := range id {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// Title returns the name for preferred, then US → EU → JP.
func (g Game) Title(preferred Region) string {
	if t := g.Titles[preferred]; t != "" {
		return t
	}
	for _, r := range Regions {
		if t := g.Titles[r]; t != "" {
			return t
		}
	}
	return g.ID
}

// IconRegion returns the cover region for preferred, then US → EU → JP.
func (g Game) IconRegion(preferred Region) (Region, bool) {
	if g.Icons[preferred] {
		return preferred, true
	}
	for _, r := range Regions {
		if g.Icons[r] {
			return r, true
		}
	}
	return "", false
}

func (g Game) HasTitle(region Region) bool {
	return g.Titles[region] != ""
}
