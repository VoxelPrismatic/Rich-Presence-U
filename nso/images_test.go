package nso

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCachedCover(t *testing.T) {
	dir := t.TempDir()
	c := &Client{CacheDir: dir}
	g := Game{
		ID:     "70010000000001",
		Titles: map[Region]string{US: "Test"},
		Covers: map[Region]string{US: "https://example.test/cover.jpg"},
		Icons:  map[Region]bool{US: true},
	}
	if _, ok := c.CachedCover(g, US); ok {
		t.Fatal("missing file reported as cached")
	}
	path := c.imagePath(g, g.Covers[US], US)
	if err := os.WriteFile(path, []byte("img"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok := c.CachedCover(g, US)
	if !ok || got != path {
		t.Fatalf("cached = %q %v, want %q", got, ok, path)
	}
	if filepath.Base(got) == "" {
		t.Fatal("empty base")
	}
}
