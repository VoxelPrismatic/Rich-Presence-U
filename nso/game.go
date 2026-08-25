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
	// Covers is the eShop image URL per region. CoverArt is a legacy fallback.
	Covers   map[Region]string
	CoverArt string
	Stores   map[Region]string
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

// IGDBPrefix marks titles remembered from the IGDB games search.
const IGDBPrefix = "igdb:"

// IsIGDBID reports whether id was stored from an IGDB search hit.
func IsIGDBID(id string) bool {
	return strings.HasPrefix(id, IGDBPrefix)
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

// Cover returns the image URL for preferred, then any other region, then CoverArt.
func (g Game) Cover(preferred Region) (url string, region Region) {
	if g.Covers != nil {
		if u := g.Covers[preferred]; u != "" {
			return u, preferred
		}
		for _, r := range Regions {
			if u := g.Covers[r]; u != "" {
				return u, r
			}
		}
	}
	if g.CoverArt != "" {
		if preferred.Valid() {
			return g.CoverArt, preferred
		}
		return g.CoverArt, US
	}
	return "", ""
}

func (g *Game) setCover(region Region, url string) {
	url = strings.TrimSpace(url)
	if url == "" || !region.Valid() {
		return
	}
	if g.Covers == nil {
		g.Covers = map[Region]string{}
	}
	if g.Icons == nil {
		g.Icons = map[Region]bool{}
	}
	g.Covers[region] = url
	g.Icons[region] = true
	if g.CoverArt == "" {
		g.CoverArt = url
	}
}

func (g *Game) setStore(region Region, u string) {
	u = strings.TrimSpace(u)
	if u == "" || !region.Valid() {
		return
	}
	if g.Stores == nil {
		g.Stores = map[Region]string{}
	}
	g.Stores[region] = u
}

// Store is the eShop product page for preferred, then any other region, then
// a constructed ec.nintendo.com titles URL from the nsuid.
func (g Game) Store(preferred Region) string {
	if g.Stores != nil {
		if u := g.Stores[preferred]; u != "" {
			return u
		}
		for _, r := range Regions {
			if u := g.Stores[r]; u != "" {
				return u
			}
		}
	}
	return StorePage(g.ID, preferred)
}

// StorePage is Nintendo's regional product URL for an nsuid.
func StorePage(id string, region Region) string {
	nsuid := strings.TrimPrefix(id, HACPrefix)
	if nsuid == "" || strings.HasPrefix(nsuid, "eu:") || !IsStoreID(nsuid) {
		return ""
	}
	switch region {
	case JP:
		return "https://ec.nintendo.com/JP/ja/titles/" + nsuid
	case EU:
		return "https://ec.nintendo.com/GB/en/titles/" + nsuid
	default:
		return "https://ec.nintendo.com/US/en/titles/" + nsuid
	}
}
