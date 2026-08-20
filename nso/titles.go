package nso

import "strings"

// Catalog is the in-memory title list for one system (BEE includes HAC rows).
type Catalog struct {
	Games []Game
	Total map[Region]int
	ByID  map[string]int
}

func newCatalog() *Catalog {
	return &Catalog{
		Total: map[Region]int{},
		ByID:  map[string]int{},
	}
}

func (c *Catalog) add(g Game) {
	c.ByID[g.ID] = len(c.Games)
	c.Games = append(c.Games, g)
	for _, r := range Regions {
		if g.Titles[r] != "" {
			c.Total[r]++
		}
	}
}

func (c *Catalog) Lookup(id string) (Game, bool) {
	if c == nil {
		return Game{}, false
	}
	i, ok := c.ByID[id]
	if !ok {
		return Game{}, false
	}
	return c.Games[i], true
}

func setTitle(dst map[Region]string, r Region, v string) {
	v = strings.TrimSpace(v)
	if v != "" {
		dst[r] = v
	}
}
