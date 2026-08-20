package nso

import "testing"

func TestSearchSubstringAndExact(t *testing.T) {
	games := sampleGames()
	hits := Search(games, "switch", US)
	if len(hits) != 2 {
		t.Fatalf("got %d hits", len(hits))
	}
	exact := Search(games, "1-2-Switch", US)
	if len(exact) != 1 || !exact[0].Exact || exact[0].Game.ID != "70010000000001" {
		t.Fatalf("exact: %+v", exact)
	}
	eu := Search(games, "mega drive", US)
	if len(eu) != 1 || eu[0].DisplayTitle != "SEGA Genesis - Nintendo Switch Online" {
		t.Fatalf("eu match should still display US title: %+v", eu)
	}
	id := Search(games, "70010000000001", US)
	if len(id) != 1 || !id[0].Exact {
		t.Fatalf("id search: %+v", id)
	}
	if Search(games, "   ", US) != nil {
		t.Fatal("blank query should be empty")
	}
}

func TestNormTitle(t *testing.T) {
	if normTitle("Splatoon™ 3") != "splatoon 3" {
		t.Fatalf("%q", normTitle("Splatoon™ 3"))
	}
}

func TestResolve(t *testing.T) {
	games := sampleGames()
	g, ok := Resolve(games, "1-2-Switch", US)
	if !ok || g.ID != "70010000000001" {
		t.Fatalf("verified resolve: %v %+v", ok, g)
	}
	g, ok = Resolve(games, "My Homebrew", US)
	if ok || g.ID != "My Homebrew" {
		t.Fatalf("custom resolve: %v %+v", ok, g)
	}
	g, ok = Resolve(games, "", US)
	if ok || g.ID != "" {
		t.Fatalf("empty resolve: %v %+v", ok, g)
	}
}
