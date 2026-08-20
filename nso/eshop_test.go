package nso

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const euJSON = `{
  "response": {
    "docs": [
      {
        "title": "PRAGMATA",
        "fs_id": 2987033,
        "image_url_sq_s": "https://example.com/pragmata.jpg",
        "originally_for_t": "BEE",
        "playable_on_txt": ["BEE"],
        "nsuid_txt": ["70010000101666"]
      },
      {
        "title": "Wavetale",
        "fs_id": "111",
        "image_url_sq_s": "https://example.com/wavetale.jpg",
        "originally_for_t": "HAC",
        "playable_on_txt": ["HAC"],
        "nsuid_txt": ["70010000052116"]
      }
    ]
  }
}`

func TestSearchStore(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(euJSON))
	}))
	t.Cleanup(srv.Close)

	c, err := New(t.TempDir(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	c.HTTP = srv.Client()
	c.StoreURL = srv.URL

	bee, err := c.SearchStore(context.Background(), "pragmata", BEE)
	if err != nil {
		t.Fatal(err)
	}
	if len(bee) != 2 {
		t.Fatalf("bee hits %d", len(bee))
	}
	if bee[0].Titles[US] != "PRAGMATA" || bee[0].AssetSystem != BEE {
		t.Fatalf("pragmata %+v", bee[0])
	}
	if !strings.HasPrefix(bee[1].ID, HACPrefix) {
		t.Fatalf("switch title on bee should be prefixed: %q", bee[1].ID)
	}
	if bee[0].CoverArt == "" {
		t.Fatal("missing cover")
	}

	hac, err := c.SearchStore(context.Background(), "wave", HAC)
	if err != nil {
		t.Fatal(err)
	}
	if len(hac) != 1 || hac[0].Titles[US] != "Wavetale" {
		t.Fatalf("hac hits %+v", hac)
	}
}

func TestMergeGames(t *testing.T) {
	local := []Hit{{Game: Game{ID: "a", Titles: map[Region]string{US: "Wavetale"}}, DisplayTitle: "Wavetale"}}
	extra := []Game{
		{ID: "a", Titles: map[Region]string{US: "Wavetale"}},
		{ID: "7001", Titles: map[Region]string{US: "PRAGMATA"}, CoverArt: "x"},
	}
	out := MergeGames(local, extra, 8)
	if len(out) != 2 || out[1].ID != "7001" {
		t.Fatalf("%+v", out)
	}
}
