package nso

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

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

func ParseCSV(r io.Reader, assetSystem System, idPrefix string) ([]Game, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	cr.LazyQuotes = true
	cr.TrimLeadingSpace = true

	var games []Game
	byID := map[string]int{}
	row := 0
	for {
		rec, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("csv row %d: %w", row+1, err)
		}
		row++
		if row == 1 {
			continue
		}
		if len(rec) < 7 {
			continue
		}
		id := strings.TrimSpace(rec[0])
		if id == "" {
			continue
		}
		g := Game{
			ID:          idPrefix + id,
			AssetSystem: assetSystem,
			Icons:       map[Region]bool{},
			Titles:      map[Region]string{},
		}
		setFlag(g.Icons, US, rec[1])
		setFlag(g.Icons, EU, rec[2])
		setFlag(g.Icons, JP, rec[3])
		setTitle(g.Titles, US, rec[4])
		setTitle(g.Titles, EU, rec[5])
		setTitle(g.Titles, JP, rec[6])
		if !g.Verified() {
			continue
		}
		if i, ok := byID[g.ID]; ok {
			games[i] = g
			continue
		}
		byID[g.ID] = len(games)
		games = append(games, g)
	}
	return games, nil
}

func setFlag(dst map[Region]bool, r Region, v string) {
	if strings.TrimSpace(v) != "" {
		dst[r] = true
	}
}

func setTitle(dst map[Region]string, r Region, v string) {
	v = strings.TrimSpace(v)
	if v != "" {
		dst[r] = v
	}
}
