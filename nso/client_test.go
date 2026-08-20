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
			w.Write([]byte(`[dlc]
hac_titles="` + "http://" + r.Host + `/hac.csv"
hac_assets="` + "http://" + r.Host + `/hac/"
wup_titles="` + "http://" + r.Host + `/empty.csv"
ctr_titles="` + "http://" + r.Host + `/empty.csv"
bee_titles="` + "http://" + r.Host + `/bee.csv"
`))
		case strings.HasSuffix(r.URL.Path, ".csv"):
			if strings.HasSuffix(r.URL.Path, "hac.csv") {
				w.Write([]byte(sampleCSV))
			} else if strings.HasSuffix(r.URL.Path, "bee.csv") {
				w.Write([]byte("ID,US,EU,JP,US TITLE,EU TITLE,JP TITLE\namnesiareb,✓,,,Amnesia: Rebirth,,\n"))
			} else {
				w.Write([]byte("ID,US,EU,JP,US TITLE,EU TITLE,JP TITLE\n"))
			}
		case strings.HasSuffix(r.URL.Path, ".jpg"):
			w.Write(jpeg)
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
	hac := c.Games(HAC)
	if len(hac) != 3 {
		t.Fatalf("hac games %d", len(hac))
	}
	bee := c.Games(BEE)
	if len(bee) != 4 { // 1 native + 3 from hac
		t.Fatalf("bee games %d ids=%v", len(bee), ids(bee))
	}
	g, ok := c.Lookup(BEE, "hac::otsw")
	if !ok || g.AssetSystem != HAC {
		t.Fatalf("prefixed lookup: %v %+v", ok, g)
	}
	url := c.CoverURL(g, US)
	if !strings.HasSuffix(url, "/hac/otsw.us.jpg") {
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
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	before := len(hits)
	if _, err := c.CoverPath(context.Background(), g, US); err != nil {
		t.Fatal(err)
	}
	if len(hits) != before {
		t.Fatalf("cover was re-downloaded")
	}

	var stored GormGame
	if err := c.db.Where("system = ? AND catalog_id = ?", "HAC", "otsw").First(&stored).Error; err != nil {
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
	if len(c2.Games(HAC)) != 3 {
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
