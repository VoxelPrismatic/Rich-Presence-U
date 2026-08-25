package igdb

import (
	"net/url"
	"sort"
	"strings"
)

// Node is one row in the platform picker: a manufacturer, a family, or a console.
type Node struct {
	Label    string
	Platform Platform // set when this node is a selectable console
	Children []Node
}

// Leaf reports whether this node is a selectable console.
func (n Node) Leaf() bool {
	return n.Platform.Slug != "" || n.Platform.ID != 0
}

// DisplayName is the label shown in the picker for a console.
// Empty Names fall back to the slug; Suffix is appended to the first name
// (the "dot product" described on Platform).
func (p Platform) DisplayName() string {
	name := p.Slug
	if len(p.Names) > 0 && p.Names[0] != "" {
		name = p.Names[0]
	}
	if len(p.Suffix) > 0 && p.Suffix[0] != "" {
		if name == "" {
			return p.Suffix[0]
		}
		return name + " " + p.Suffix[0]
	}
	return name
}

// StoreCode is the Nintendo title-database code used for eShop search.
// Only Switch consoles are wired; everything else returns "".
func (p Platform) StoreCode() string {
	switch p.Slug {
	case "NS1":
		return "HAC"
	case "NS2":
		return "BEE"
	default:
		return ""
	}
}

// ShopName is the storefront shown on the "Open … Page" button for consoles
// that search a shop. Empty means IGDB.
func (p Platform) ShopName() string {
	switch p.StoreCode() {
	case "HAC", "BEE", "CTR", "WUP":
		return "eShop"
	default:
		return ""
	}
}

// PageURL is the IGDB website page for this console.
func (p Platform) PageURL() string {
	return "https://www.igdb.com/search?q=" + url.QueryEscape(p.DisplayName())
}

// BySlug looks up a mapped console by its short slug (eg "NS1", "NES").
func BySlug(slug string) (Platform, bool) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return Platform{}, false
	}
	for _, families := range MappedPlatforms {
		for _, plats := range families {
			for _, p := range plats {
				if p.Slug == slug {
					return p, true
				}
			}
		}
	}
	return Platform{}, false
}

// ByID looks up a mapped console by IGDB platform id.
func ByID(id PlatformID) (Platform, bool) {
	if id == 0 {
		return Platform{}, false
	}
	for _, families := range MappedPlatforms {
		for _, plats := range families {
			for _, p := range plats {
				if p.ID == id {
					return p, true
				}
			}
		}
	}
	return Platform{}, false
}

// SlugForStoreCode maps a Nintendo title-database code (HAC, BEE, …) onto a
// mapped slug. Unknown codes return "".
func SlugForStoreCode(code string) string {
	switch strings.ToUpper(strings.TrimSpace(code)) {
	case "HAC":
		return "NS1"
	case "BEE":
		return "NS2"
	case "CTR":
		return "3DS"
	case "WUP":
		return "WiiU"
	default:
		return ""
	}
}

var topPriority = []string{
	"Nintendo",
	"PlayStation",
	"Xbox",
	"PC",
	"Home Computer",
	"Retro",
	"Arcade",
	"Virtual Reality",
	"Phone/Mobile",
	"Mini Computer",
}

func familyPriority(maker string) []string {
	switch maker {
	case "Nintendo":
		return []string{"Switch"}
	case "PlayStation":
		return []string{"Home", "Portable", "VR"}
	case "PC":
		return []string{"Windows", "macOS", "Linux", "Legacy"}
	case "Retro":
		return []string{"Atari", "Sega"}
	case "Mini Computer":
		return []string{"PDP"}
	case "Arcade":
		return []string{"Arcade", "Neo Geo"}
	default:
		return nil
	}
}

// PickerTree builds the manufacturer → family → console tree used by the
// platform selector. Families that only contain one console are collapsed so
// the console is selected directly from the manufacturer page, except Mini
// Computer brands which always keep the extra grouping level. A manufacturer
// with a single family (Xbox) skips that redundant family row.
func PickerTree() []Node {
	makers := mapKeys(MappedPlatforms)
	sortByPriority(makers, topPriority)

	out := make([]Node, 0, len(makers))
	for _, maker := range makers {
		families := MappedPlatforms[maker]
		famNames := mapKeys(families)
		sortByPriority(famNames, familyPriority(maker))

		node := Node{Label: maker}
		if len(famNames) == 1 {
			for _, p := range families[famNames[0]] {
				node.Children = append(node.Children, Node{Label: p.DisplayName(), Platform: p})
			}
			out = append(out, node)
			continue
		}

		keepFamily := maker == "Mini Computer"
		for _, fam := range famNames {
			plats := families[fam]
			if !keepFamily && len(plats) == 1 {
				p := plats[0]
				node.Children = append(node.Children, Node{Label: p.DisplayName(), Platform: p})
				continue
			}
			famNode := Node{Label: fam}
			for _, p := range plats {
				famNode.Children = append(famNode.Children, Node{Label: p.DisplayName(), Platform: p})
			}
			node.Children = append(node.Children, famNode)
		}
		out = append(out, node)
	}
	return out
}

func mapKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for name := range m {
		keys = append(keys, name)
	}
	return keys
}

func sortByPriority(names, priority []string) {
	rank := make(map[string]int, len(priority))
	for i, n := range priority {
		rank[n] = i
	}
	sort.SliceStable(names, func(i, j int) bool {
		ri, okI := rank[names[i]]
		rj, okJ := rank[names[j]]
		switch {
		case okI && okJ:
			return ri < rj
		case okI:
			return true
		case okJ:
			return false
		default:
			return names[i] < names[j]
		}
	})
}
