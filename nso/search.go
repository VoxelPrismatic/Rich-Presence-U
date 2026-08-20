package nso

import (
	"sort"
	"strings"
)

// Hit is one search suggestion.
type Hit struct {
	Game         Game
	DisplayTitle string
	Exact        bool
}

// Search matches query against every regional title (and the id). Exact title
// matches are flagged; substring matches are sorted by display title.
func Search(games []Game, query string, display Region) []Hit {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}
	q := strings.ToLower(query)
	hits := make([]Hit, 0)
	seen := map[string]bool{}
	for _, g := range games {
		exact := strings.ToLower(g.NativeID()) == q || strings.ToLower(g.ID) == q
		matched := exact
		if !matched {
			for _, title := range g.Titles {
				lt := strings.ToLower(title)
				if lt == q {
					exact = true
					matched = true
					break
				}
				if strings.Contains(lt, q) {
					matched = true
				}
			}
		}
		if !matched || seen[g.ID] {
			continue
		}
		seen[g.ID] = true
		hits = append(hits, Hit{
			Game:         g,
			DisplayTitle: g.Title(display),
			Exact:        exact,
		})
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].Exact != hits[j].Exact {
			return hits[i].Exact
		}
		return strings.ToLower(hits[i].DisplayTitle) < strings.ToLower(hits[j].DisplayTitle)
	})
	return hits
}

// Resolve typed input to a catalog game. Exact title or id wins; otherwise the
// text is treated as a custom (unverified) entry.
func Resolve(games []Game, typed string, display Region) (game Game, verified bool) {
	typed = strings.TrimSpace(typed)
	if typed == "" {
		return Game{}, false
	}
	hits := Search(games, typed, display)
	for _, h := range hits {
		if h.Exact {
			return h.Game, true
		}
	}
	return Game{ID: typed, Titles: map[Region]string{}, Icons: map[Region]bool{}}, false
}
