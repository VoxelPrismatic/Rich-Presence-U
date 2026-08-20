package nso

import "strings"

// System is a Nintendo platform in the title database.
type System string

const (
	CTR System = "CTR" // Nintendo 3DS
	WUP System = "WUP" // Wii U
	HAC System = "HAC" // Nintendo Switch
	BEE System = "BEE" // Nintendo Switch 2
)

// Systems is the consoles shown in the UI.
var Systems = []System{HAC, BEE}

// Region is a storefront region for titles and icons.
type Region string

const (
	US Region = "US"
	EU Region = "EU"
	JP Region = "JP"
)

// Regions is the fallback order used when a preferred region is missing.
var Regions = []Region{US, EU, JP}

func (s System) Valid() bool {
	switch s {
	case CTR, WUP, HAC, BEE:
		return true
	default:
		return false
	}
}

func (s System) Lower() string {
	return strings.ToLower(string(s))
}

func (s System) DisplayName() string {
	switch s {
	case CTR:
		return "Nintendo 3DS"
	case WUP:
		return "Nintendo Wii U"
	case HAC:
		return "Nintendo Switch"
	case BEE:
		return "Nintendo Switch 2"
	default:
		return string(s)
	}
}

func ParseSystem(v string) (System, bool) {
	s := System(strings.ToUpper(strings.TrimSpace(v)))
	if s.Valid() {
		return s, true
	}
	return "", false
}

func (r Region) Valid() bool {
	switch r {
	case US, EU, JP:
		return true
	default:
		return false
	}
}

func (r Region) Lower() string {
	return strings.ToLower(string(r))
}

func (r Region) DisplayName() string {
	switch r {
	case US:
		return "Americas"
	case EU:
		return "Europe"
	case JP:
		return "Japan"
	default:
		return string(r)
	}
}

func ParseRegion(v string) (Region, bool) {
	r := Region(strings.ToUpper(strings.TrimSpace(v)))
	if r.Valid() {
		return r, true
	}
	return "", false
}

// HACPrefix is applied to Switch titles when they are listed under Switch 2.
const HACPrefix = "hac::"
