package nso

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRefreshAndCover(t *testing.T) {
	jpeg := []byte{0xff, 0xd8, 0xff, 0xd9}
	var hits []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits = append(hits, r.URL.Path)
		switch {
		case r.URL.Path == "/meta":
			w.Write([]byte("[dlc]\nhac_client=\"1259967215323840564\"\n"))
		case strings.HasSuffix(r.URL.Path, ".jpg"):
			w.Write(jpeg)
		case strings.HasSuffix(r.URL.Path, ".csv"):
			t.Errorf("should not download title csv: %s", r.URL.Path)
			http.NotFound(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	config := t.TempDir()
	cache := t.TempDir()
	c, err := New(config, cache)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	c.HTTP = srv.Client()
	c.MetaURL = srv.URL + "/meta"

	if err := c.Refresh(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(config, "games.db")); err != nil {
		t.Fatal(err)
	}
	if len(c.Games(HAC)) != 0 {
		t.Fatalf("catalog should start empty, got %d", len(c.Games(HAC)))
	}

	g := Game{
		ID:          "70010000001234",
		AssetSystem: HAC,
		Icons:       map[Region]bool{US: true, EU: true},
		Titles:      map[Region]string{US: "1-2-Switch", EU: "1-2-Switch"},
		CoverArt:    srv.URL + "/1-2-switch.jpg",
	}
	if err := c.Remember(HAC, g); err != nil {
		t.Fatal(err)
	}
	got, ok := c.Lookup(HAC, g.ID)
	if !ok || got.Title(US) != "1-2-Switch" {
		t.Fatalf("remembered lookup: %v %+v", ok, got)
	}
	url := c.CoverURL(g, US)
	if url != g.CoverArt {
		t.Fatalf("cover url %q", url)
	}
	path, err := c.CoverPath(context.Background(), g, US)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "1-2-Switch.jpg" {
		t.Fatalf("cover filename %q", path)
	}
	if !strings.HasPrefix(path, cache) {
		t.Fatalf("cover not in cache dir: %q", path)
	}
	before := len(hits)
	if _, err := c.CoverPath(context.Background(), g, US); err != nil {
		t.Fatal(err)
	}
	if len(hits) != before {
		t.Fatalf("cover was re-downloaded")
	}

	var stored GormGame
	if err := c.db.Where("system = ? AND catalog_id = ?", "HAC", g.ID).First(&stored).Error; err != nil {
		t.Fatal(err)
	}
	if stored.TitleAmericas != "1-2-Switch" || stored.CoverArt == "" {
		t.Fatalf("gorm row %+v", stored)
	}

	c2, err := New(config, cache)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c2.Close() })
	if err := c2.LoadCache(); err != nil {
		t.Fatal(err)
	}
	if len(c2.Games(HAC)) != 1 {
		t.Fatalf("reload %d", len(c2.Games(HAC)))
	}

	if c.NeedsRefresh(time.Hour, time.Now()) {
		t.Fatal("fresh timestamp should not refresh")
	}
	if !c.NeedsRefresh(time.Hour, time.Time{}) {
		t.Fatal("zero last should refresh")
	}

	if err := c.ClearCache(); err != nil {
		t.Fatal(err)
	}
	if c.CachePresent() {
		t.Fatal("games still in db after clear")
	}
	if entries, _ := os.ReadDir(cache); len(entries) != 0 {
		t.Fatalf("cache dir not empty: %v", entries)
	}
}

func TestPurgeLegacyCatalog(t *testing.T) {
	c, err := New(t.TempDir(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })

	slug := Game{
		ID:          "spyroreignited",
		AssetSystem: HAC,
		Icons:       map[Region]bool{US: true},
		Titles:      map[Region]string{US: "Spyro(TM) Reignited Trilogy"},
	}
	store := Game{
		ID:          "70010000012345",
		AssetSystem: HAC,
		Icons:       map[Region]bool{US: true, EU: true},
		Titles:      map[Region]string{US: "Spyro Reignited Trilogy"},
		CoverArt:    "https://example.com/spyro.jpg",
	}
	if err := c.upsertGame(HAC, slug); err != nil {
		t.Fatal(err)
	}
	if err := c.Remember(HAC, store); err != nil {
		t.Fatal(err)
	}
	if err := c.LoadCache(); err != nil {
		t.Fatal(err)
	}
	games := c.Games(HAC)
	if len(games) != 1 || games[0].ID != store.ID {
		t.Fatalf("legacy dump should be gone: %+v", ids(games))
	}
	if _, ok := c.Lookup(HAC, slug.ID); ok {
		t.Fatal("csv slug still in catalog")
	}
}
