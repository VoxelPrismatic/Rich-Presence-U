package nso

import "testing"

func sampleGames() []Game {
	return []Game{
		{
			ID:          "70010000000001",
			AssetSystem: HAC,
			Icons:       map[Region]bool{US: true},
			Titles:      map[Region]string{US: "1-2-Switch"},
			CoverArt:    "https://example.com/otsw.jpg",
		},
		{
			ID:          "70010000000002",
			AssetSystem: HAC,
			Icons:       map[Region]bool{US: true, EU: true, JP: true},
			Titles: map[Region]string{
				US: "SEGA Genesis - Nintendo Switch Online",
				EU: "SEGA Mega Drive - Nintendo Switch Online",
				JP: "セガ メガドライブ for Nintendo Switch Online",
			},
			CoverArt: "https://example.com/sega.jpg",
		},
		{
			ID:          "70010000000003",
			AssetSystem: HAC,
			Icons:       map[Region]bool{EU: true},
			Titles:      map[Region]string{EU: "Only Europe"},
			CoverArt:    "https://example.com/eu.jpg",
		},
	}
}

func TestCatalogLookup(t *testing.T) {
	cat := newCatalog()
	for _, g := range sampleGames() {
		cat.add(g)
	}
	g, ok := cat.Lookup("70010000000001")
	if !ok || g.Title(US) != "1-2-Switch" {
		t.Fatalf("lookup: %v %+v", ok, g)
	}
	if g.Title(JP) != "1-2-Switch" {
		t.Fatalf("fallback title %q", g.Title(JP))
	}
	sega, _ := cat.Lookup("70010000000002")
	if sega.Title(EU) != "SEGA Mega Drive - Nintendo Switch Online" {
		t.Fatalf("eu title %q", sega.Title(EU))
	}
	if r, ok := sega.IconRegion(JP); !ok || r != JP {
		t.Fatalf("jp icon region %q %v", r, ok)
	}
	onlyEU, _ := cat.Lookup("70010000000003")
	if onlyEU.Title(US) != "Only Europe" {
		t.Fatalf("eu-only fallback %q", onlyEU.Title(US))
	}
}

func TestIsStoreID(t *testing.T) {
	if !IsStoreID("70010000012345") || !IsStoreID("hac::70010000012345") || !IsStoreID("eu:2987033") {
		t.Fatal("store ids should match")
	}
	if IsStoreID("spyroreignited") || IsStoreID("otsw") || IsStoreID("hac::otsw") {
		t.Fatal("csv slugs should not match")
	}
}

func ids(games []Game) []string {
	out := make([]string, len(games))
	for i, g := range games {
		out[i] = g.ID
	}
	return out
}
