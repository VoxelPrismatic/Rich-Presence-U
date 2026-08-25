package igdb

import (
	"strings"
	"testing"
)

func TestGameSearchURLKeepsBrackets(t *testing.T) {
	u := GameSearchURL("zelda", 18)
	if strings.Contains(u, "f%5B") || strings.Contains(u, "%5D") {
		t.Fatalf("brackets were encoded: %s", u)
	}
	if strings.Contains(u, "%s") || strings.Contains(u, "%d") || strings.Contains(u, "{search") || strings.Contains(u, "{game") {
		t.Fatalf("placeholders left in URL: %s", u)
	}
	want := []string{
		"https://www.igdb.com/advanced_search?",
		"d=1",
		"f[type]=games",
		"q=zelda",
		"s=score",
		"f[platforms.id_in]=18",
	}
	for _, part := range want {
		if !strings.Contains(u, part) {
			t.Errorf("missing %q in %s", part, u)
		}
	}
}

func TestGameSearchURLEscapesQueryOnly(t *testing.T) {
	u := GameSearchURL("super mario", 19)
	if !strings.Contains(u, "f[type]=games") || !strings.Contains(u, "f[platforms.id_in]=19") {
		t.Fatalf("filters changed: %s", u)
	}
	if !strings.Contains(u, "q=super+mario") && !strings.Contains(u, "q=super%20mario") {
		t.Fatalf("query not escaped: %s", u)
	}
}

func TestGameSearchURLOmitsZeroPlatform(t *testing.T) {
	u := GameSearchURL("halo", 0)
	if strings.Contains(u, "platforms.id_in") {
		t.Fatalf("unexpected platform filter: %s", u)
	}
}

func TestPageURLUsesAdvancedSearch(t *testing.T) {
	nes, ok := BySlug("NES")
	if !ok {
		t.Fatal("NES")
	}
	u := nes.PageURL()
	if !strings.Contains(u, "f[platforms.id_in]=18") || !strings.Contains(u, "f[type]=games") {
		t.Fatalf("%s", u)
	}
}

func TestCoverAndGamePageURL(t *testing.T) {
	if got := CoverImageURL("co1"); got != "https://images.igdb.com/igdb/image/upload/t_cover_big/co1.jpg" {
		t.Fatalf("%s", got)
	}
	if GamePageURL("the-legend-of-zelda") != "https://www.igdb.com/games/the-legend-of-zelda" {
		t.Fatal(GamePageURL("the-legend-of-zelda"))
	}
}

func TestFilterIDsIncludes(t *testing.T) {
	nes, _ := BySlug("NES")
	ids := nes.FilterIDs()
	if len(ids) < 2 || ids[0] != 18 {
		t.Fatalf("%v", ids)
	}
	found := false
	for _, id := range ids {
		if id == 99 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected FC id 99 in %v", ids)
	}
}
