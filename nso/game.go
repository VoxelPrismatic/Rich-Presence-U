package nso

import "strings"

// Game is one row from a platform title CSV.
type Game struct {
	// ID is the catalog id. Switch titles merged into the Switch 2 list are
	// prefixed with hac:: so they do not collide with native Switch 2 ids.
	ID string
	// AssetSystem is the platform folder used for cover art (hac when the id
	// carries the hac:: prefix).
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
