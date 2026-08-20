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
// gorm.Model supplies the integer primary key; CatalogID is the store id.
type GormGame struct {
	gorm.Model
	CatalogID     string `gorm:"uniqueIndex:idx_game_sys_cat;size:128"`
	System        string `gorm:"uniqueIndex:idx_game_sys_cat;size:8"`
	AssetSystem   string `gorm:"size:8"`
	TitleAmericas string `gorm:"column:title_americas"`
	TitleEurope   string `gorm:"column:title_europe"`
	TitleJapan    string `gorm:"column:title_japan"`
	CoverArt      string
	CoverAmericas string `gorm:"column:cover_americas"`
	CoverEurope   string `gorm:"column:cover_europe"`
	CoverJapan    string `gorm:"column:cover_japan"`
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
		Covers:      map[Region]string{},
		CoverArt:    r.CoverArt,
	}
	if g.AssetSystem == "" {
		g.AssetSystem = System(r.System)
	}
	setTitle(g.Titles, US, r.TitleAmericas)
	setTitle(g.Titles, EU, r.TitleEurope)
	setTitle(g.Titles, JP, r.TitleJapan)
	g.setCover(US, r.CoverAmericas)
	g.setCover(EU, r.CoverEurope)
	g.setCover(JP, r.CoverJapan)
	if r.CoverArt != "" && coverOf(g, US) == "" && coverOf(g, EU) == "" && coverOf(g, JP) == "" {
		for _, region := range Regions {
			if g.Titles[region] != "" {
				g.setCover(region, r.CoverArt)
			}
		}
		if coverOf(g, US) == "" && coverOf(g, EU) == "" && coverOf(g, JP) == "" {
			g.setCover(US, r.CoverArt)
		}
	}
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

func coverOf(g Game, region Region) string {
	if g.Covers != nil {
		return g.Covers[region]
	}
	return ""
}

func toGormGame(system System, g Game, cover string) GormGame {
	if cover == "" {
		cover, _ = g.Cover("")
	}
	return GormGame{
		CatalogID:     g.ID,
		System:        string(system),
		AssetSystem:   string(g.AssetSystem),
		TitleAmericas: g.Titles[US],
		TitleEurope:   g.Titles[EU],
		TitleJapan:    g.Titles[JP],
		CoverArt:      cover,
		CoverAmericas: coverOf(g, US),
		CoverEurope:   coverOf(g, EU),
		CoverJapan:    coverOf(g, JP),
		IconAmericas:  g.Icons[US] || coverOf(g, US) != "",
		IconEurope:    g.Icons[EU] || coverOf(g, EU) != "",
		IconJapan:     g.Icons[JP] || coverOf(g, JP) != "",
	}
}

func (c *Client) purgeLegacyCatalog() error {
	var rows []GormGame
	if err := c.db.Find(&rows).Error; err != nil {
		return err
	}
	ids := make([]uint, 0)
	for _, row := range rows {
		if IsStoreID(row.CatalogID) {
			continue
		}
		ids = append(ids, row.ID)
	}
	if len(ids) == 0 {
		return nil
	}
	return c.db.Unscoped().Where("id IN ?", ids).Delete(&GormGame{}).Error
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
			if !IsStoreID(row.CatalogID) {
				continue
			}
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

func (c *Client) upsertGame(system System, g Game) error {
	cover := g.CoverArt
	if cover == "" {
		cover = c.CoverURL(g, US)
	}
	row := toGormGame(system, g, cover)
	var existing GormGame
	err := c.db.Where("system = ? AND catalog_id = ?", string(system), g.ID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return c.db.Create(&row).Error
	}
	if err != nil {
		return err
	}
	row.ID = existing.ID
	return c.db.Save(&row).Error
}

func (c *Client) gameCount() (int64, error) {
	var n int64
	err := c.db.Model(&GormGame{}).Count(&n).Error
	return n, err
}
