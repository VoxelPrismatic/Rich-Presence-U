package nso

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GormGame is one title row in ~/.config/rich-presence-u/games.db.
// gorm.Model supplies the integer primary key; CatalogID is the CSV slug.
type GormGame struct {
	gorm.Model
	CatalogID     string `gorm:"uniqueIndex:idx_game_sys_cat;size:128"`
	System        string `gorm:"uniqueIndex:idx_game_sys_cat;size:8"`
	AssetSystem   string `gorm:"size:8"`
	TitleAmericas string `gorm:"column:title_americas"`
	TitleEurope   string `gorm:"column:title_europe"`
	TitleJapan    string `gorm:"column:title_japan"`
	CoverArt      string
	IconAmericas  bool
	IconEurope    bool
	IconJapan     bool
}

type gormKV struct {
	Key   string `gorm:"primaryKey"`
	Value string `gorm:"type:text"`
}

func openDB(path string) (*gorm.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("open games.db: %w", err)
	}
	if err := db.AutoMigrate(&GormGame{}, &gormKV{}); err != nil {
		return nil, err
	}
	return db, nil
}

func (r GormGame) Game() Game {
	g := Game{
		ID:          r.CatalogID,
		AssetSystem: System(r.AssetSystem),
		Icons:       map[Region]bool{},
		Titles:      map[Region]string{},
		CoverArt:    r.CoverArt,
	}
	if g.AssetSystem == "" {
		g.AssetSystem = System(r.System)
	}
	setTitle(g.Titles, US, r.TitleAmericas)
	setTitle(g.Titles, EU, r.TitleEurope)
	setTitle(g.Titles, JP, r.TitleJapan)
	if r.IconAmericas {
		g.Icons[US] = true
	}
	if r.IconEurope {
		g.Icons[EU] = true
	}
	if r.IconJapan {
		g.Icons[JP] = true
	}
	return g
}

func toGormGame(system System, g Game, cover string) GormGame {
	return GormGame{
		CatalogID:     g.ID,
		System:        string(system),
		AssetSystem:   string(g.AssetSystem),
		TitleAmericas: g.Titles[US],
		TitleEurope:   g.Titles[EU],
		TitleJapan:    g.Titles[JP],
		CoverArt:      cover,
		IconAmericas:  g.Icons[US],
		IconEurope:    g.Icons[EU],
		IconJapan:     g.Icons[JP],
	}
}

func (c *Client) replaceSystem(system System, games []Game) error {
	seen := map[string]int{}
	rows := make([]GormGame, 0, len(games))
	for _, g := range games {
		cover := c.CoverURL(g, US)
		if cover == "" {
			cover = c.CoverURL(g, EU)
		}
		if cover == "" {
			cover = c.CoverURL(g, JP)
		}
		row := toGormGame(system, g, cover)
		if i, ok := seen[row.CatalogID]; ok {
			rows[i] = row
			continue
		}
		seen[row.CatalogID] = len(rows)
		rows = append(rows, row)
	}
	return c.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("system = ?", string(system)).Delete(&GormGame{}).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		return tx.CreateInBatches(rows, 200).Error
	})
}

func (c *Client) loadSystem(system System) (*Catalog, error) {
	cat := newCatalog()
	want := []System{system}
	if system == BEE {
		want = append(want, HAC)
	}
	for _, sys := range want {
		var rows []GormGame
		if err := c.db.Where("system = ?", string(sys)).Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			g := row.Game()
			if system == BEE && sys == HAC {
				g.ID = HACPrefix + g.NativeID()
			}
			cat.add(g)
		}
	}
	return cat, nil
}

func (c *Client) saveMeta() error {
	c.mu.RLock()
	raw, err := json.Marshal(c.Meta)
	c.mu.RUnlock()
	if err != nil {
		return err
	}
	row := gormKV{Key: "metadata", Value: string(raw)}
	return c.db.Save(&row).Error
}

func (c *Client) loadMeta() error {
	var row gormKV
	if err := c.db.First(&row, "key = ?", "metadata").Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}
	var m Metadata
	if err := json.Unmarshal([]byte(row.Value), &m); err != nil {
		return err
	}
	if m.Clients == nil {
		m.Clients = map[System]string{}
	}
	if m.Titles == nil {
		m.Titles = map[System]string{}
	}
	if m.Assets == nil {
		m.Assets = map[System]string{}
	}
	if m.HelpLang == nil {
		m.HelpLang = map[string]string{}
	}
	base := DefaultMetadata()
	for sys, v := range base.Clients {
		if m.Clients[sys] == "" {
			m.Clients[sys] = v
		}
	}
	for sys, v := range base.Titles {
		if m.Titles[sys] == "" {
			m.Titles[sys] = v
		}
	}
	for sys, v := range base.Assets {
		if m.Assets[sys] == "" {
			m.Assets[sys] = v
		}
	}
	c.mu.Lock()
	c.Meta = m
	c.mu.Unlock()
	return nil
}

func (c *Client) gameCount() (int64, error) {
	var n int64
	err := c.db.Model(&GormGame{}).Count(&n).Error
	return n, err
}
