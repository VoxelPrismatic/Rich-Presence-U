package nso

import (
	"strings"
	"testing"
)

const sampleCSV = `ID,US,EU,JP,US TITLE,EU TITLE,JP TITLE
otsw,✓,,,1-2-Switch,,
swsmd,✓,✓,✓,SEGA Genesis - Nintendo Switch Online,SEGA Mega Drive - Nintendo Switch Online,セガ メガドライブ for Nintendo Switch Online
empty,,,,,,,
short,✓
custom,,✓,,,"Only Europe",
`

func TestParseCSV(t *testing.T) {
	games, err := ParseCSV(strings.NewReader(sampleCSV), HAC, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(games) != 3 {
		t.Fatalf("got %d games: %+v", len(games), ids(games))
	}
	g := games[0]
	if g.ID != "otsw" || g.AssetSystem != HAC {
		t.Fatalf("first game: %+v", g)
	}
	if !g.Icons[US] || g.Icons[EU] || g.Title(US) != "1-2-Switch" {
		t.Fatalf("otsw fields: %+v", g)
	}
	if g.Title(JP) != "1-2-Switch" {
		t.Fatalf("fallback title %q", g.Title(JP))
	}
	sega := games[1]
	if sega.Title(EU) != "SEGA Mega Drive - Nintendo Switch Online" {
		t.Fatalf("eu title %q", sega.Title(EU))
	}
	if r, ok := sega.IconRegion(JP); !ok || r != JP {
		t.Fatalf("jp icon region %q %v", r, ok)
	}
	onlyEU := games[2]
	if onlyEU.Title(US) != "Only Europe" {
		t.Fatalf("eu-only fallback %q", onlyEU.Title(US))
	}
	if r, ok := onlyEU.IconRegion(US); !ok || r != EU {
		t.Fatalf("icon fallback %q %v", r, ok)
	}
}

func TestParseCSVDedupes(t *testing.T) {
	raw := sampleCSV + "otsw,✓,,,1-2-Switch Deluxe,,\n"
	games, err := ParseCSV(strings.NewReader(raw), HAC, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(games) != 3 {
		t.Fatalf("got %d games: %v", len(games), ids(games))
	}
	if games[0].Title(US) != "1-2-Switch Deluxe" {
		t.Fatalf("last duplicate should win: %q", games[0].Title(US))
	}
}

func TestParseCSVPrefix(t *testing.T) {
	games, err := ParseCSV(strings.NewReader(sampleCSV), HAC, HACPrefix)
	if err != nil {
		t.Fatal(err)
	}
	if games[0].ID != "hac::otsw" {
		t.Fatalf("prefixed id %q", games[0].ID)
	}
	if games[0].NativeID() != "otsw" || !games[0].FromSwitch() {
		t.Fatalf("native id helpers: %+v", games[0])
	}
}

func ids(games []Game) []string {
	out := make([]string, len(games))
	for i, g := range games {
		out[i] = g.ID
	}
	return out
}
