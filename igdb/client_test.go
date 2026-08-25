package igdb

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGamesQuery(t *testing.T) {
	q := gamesQuery(`zelda "1"`, []int{18, 99})
	if !strings.HasPrefix(q, `search "zelda \"1\""`) {
		t.Fatalf("search: %s", q)
	}
	if !strings.Contains(q, "where platforms = (18,99)") {
		t.Fatalf("platforms: %s", q)
	}
	if !strings.Contains(q, "fields id,name,slug,url,cover.image_id") {
		t.Fatalf("fields: %s", q)
	}
}

func TestSearchGames(t *testing.T) {
	var gamesBody string
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.Form.Get("client_id") != "cid" || r.Form.Get("client_secret") != "sec" {
			http.Error(w, "bad creds", 401)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"access_token":"tok","expires_in":3600}`)
	})
	mux.HandleFunc("/games", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Client-ID") != "cid" || r.Header.Get("Authorization") != "Bearer tok" {
			http.Error(w, "auth", 401)
			return
		}
		b, _ := io.ReadAll(r.Body)
		gamesBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `[{"id":1026,"name":"The Legend of Zelda","slug":"the-legend-of-zelda","url":"https://www.igdb.com/games/the-legend-of-zelda","cover":{"image_id":"co1"}}]`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := NewClient()
	c.TokenURL = srv.URL + "/token"
	c.APIURL = srv.URL
	c.SetCredentials("cid", "sec")
	nes, _ := BySlug("NES")
	hits, err := c.SearchGames(context.Background(), "zelda", nes)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gamesBody, `search "zelda"`) || !strings.Contains(gamesBody, "platforms = (18,99)") {
		t.Fatalf("query %q", gamesBody)
	}
	if len(hits) != 1 || hits[0].CatalogID() != "igdb:1026" || hits[0].CoverURL == "" {
		t.Fatalf("%+v", hits)
	}
}

func TestPingRequiresCredentials(t *testing.T) {
	c := NewClient()
	if c.Configured() {
		t.Fatal("empty client")
	}
	if err := c.Ping(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}
