package nso

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"gorm.io/gorm"
)

const UserAgent = "RichPresenceU/2.0 (+https://github.com/VoxelPrismatic/Rich-Presence-U)"

// Progress reports title-database download steps.
type Progress struct {
	Stage   string // metadata, titles, done
	System  System
	Current int
	Total   int
}

// Client fetches and caches the Nintendo title database and cover art.
type Client struct {
	ConfigDir string
	CacheDir  string
	HTTP      *http.Client
	MetaURL   string

	db       *gorm.DB
	mu       sync.RWMutex
	Meta     Metadata
	catalogs map[System]*Catalog
}

func New(configDir, cacheDir string) (*Client, error) {
	var err error
	if configDir == "" {
		configDir, err = ConfigDir()
		if err != nil {
			return nil, err
		}
	}
	if cacheDir == "" {
		cacheDir, err = CacheDir()
		if err != nil {
			return nil, err
		}
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, err
	}
	c := &Client{
		ConfigDir: configDir,
		CacheDir:  cacheDir,
		HTTP:      &http.Client{Timeout: 30 * time.Second},
		MetaURL:   DefaultMetadataURL,
		Meta:      DefaultMetadata(),
		catalogs:  map[System]*Catalog{},
	}
	c.db, err = openDB(c.dbPath())
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) Close() error {
	if c.db == nil {
		return nil
	}
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (c *Client) Catalog(system System) *Catalog {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.catalogs[system]
}

func (c *Client) Games(system System) []Game {
	cat := c.Catalog(system)
	if cat == nil {
		return nil
	}
	out := make([]Game, len(cat.Games))
	copy(out, cat.Games)
	return out
}

func (c *Client) Lookup(system System, id string) (Game, bool) {
	cat := c.Catalog(system)
	if cat == nil {
		return Game{}, false
	}
	return cat.Lookup(id)
}

func (c *Client) Totals(system System) map[Region]int {
	cat := c.Catalog(system)
	if cat == nil {
		return map[Region]int{}
	}
	out := map[Region]int{}
	for k, v := range cat.Total {
		out[k] = v
	}
	return out
}

// LoadCache reads titles and metadata from games.db.
func (c *Client) LoadCache() error {
	if err := c.loadMeta(); err != nil {
		return err
	}
	return c.loadAllTitles()
}

func (c *Client) loadAllTitles() error {
	next := map[System]*Catalog{}
	for _, sys := range Systems {
		cat, err := c.loadSystem(sys)
		if err != nil {
			return err
		}
		next[sys] = cat
	}
	c.mu.Lock()
	c.catalogs = next
	c.mu.Unlock()
	return nil
}

// Refresh downloads metadata and every platform CSV, then writes games.db.
func (c *Client) Refresh(ctx context.Context, progress func(Progress)) error {
	report := func(p Progress) {
		if progress != nil {
			progress(p)
		}
	}
	report(Progress{Stage: "metadata", Current: 0, Total: 1 + len(Systems)})
	if err := c.fetchMetadata(ctx); err != nil {
		return err
	}
	c.mu.RLock()
	meta := c.Meta
	c.mu.RUnlock()
	for i, sys := range Systems {
		report(Progress{Stage: "titles", System: sys, Current: i + 1, Total: 1 + len(Systems)})
		url := meta.TitlesURL(sys)
		if url == "" {
			continue
		}
		body, err := c.get(ctx, url)
		if err != nil {
			return fmt.Errorf("%s titles: %w", sys, err)
		}
		data, err := io.ReadAll(body)
		body.Close()
		if err != nil {
			return err
		}
		games, err := ParseCSV(bytes.NewReader(data), sys, "")
		if err != nil {
			return fmt.Errorf("%s titles: %w", sys, err)
		}
		if err := c.replaceSystem(sys, games); err != nil {
			return err
		}
	}
	if err := c.loadAllTitles(); err != nil {
		return err
	}
	report(Progress{Stage: "done", Current: 1 + len(Systems), Total: 1 + len(Systems)})
	return nil
}

func (c *Client) fetchMetadata(ctx context.Context) error {
	url := c.MetaURL
	if url == "" {
		url = DefaultMetadataURL
	}
	body, err := c.get(ctx, url)
	if err != nil {
		return fmt.Errorf("metadata: %w", err)
	}
	defer body.Close()
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	m, err := ParseMetadata(bytes.NewReader(data))
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.Meta = m
	c.mu.Unlock()
	return c.saveMeta()
}

func (c *Client) NeedsRefresh(interval time.Duration, last time.Time) bool {
	if interval <= 0 {
		return false
	}
	if last.IsZero() {
		return true
	}
	return time.Since(last) > interval
}

func (c *Client) CachePresent() bool {
	n, err := c.gameCount()
	return err == nil && n > 0
}

func (c *Client) ClearCache() error {
	if err := c.db.Unscoped().Where("1 = 1").Delete(&GormGame{}).Error; err != nil {
		return err
	}
	c.mu.Lock()
	c.catalogs = map[System]*Catalog{}
	c.mu.Unlock()
	return os.RemoveAll(c.CacheDir)
}

func (c *Client) ResetAll() error {
	_ = c.Close()
	if err := os.RemoveAll(c.ConfigDir); err != nil {
		return err
	}
	return os.RemoveAll(c.CacheDir)
}

func (c *Client) get(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("%s: %s", url, resp.Status)
	}
	return resp.Body, nil
}
