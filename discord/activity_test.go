package discord

import (
	"strings"
	"testing"
)

func TestBuildDescriptionAndTag(t *testing.T) {
	a := Build(Presence{
		Title:       "X",
		Description: "   ",
		Tag:         "SW-1234-5678-9012",
		ShowTag:     true,
		CoverKey:    "https://example/cover.jpg",
		Start:       100,
	})
	if a.Details != "X  " {
		t.Fatalf("padded title %q", a.Details)
	}
	if a.State != "SW-1234-5678-9012" {
		t.Fatalf("tag as state %q", a.State)
	}
	if a.SmallImage != "" {
		t.Fatalf("no small icon without TagIcon: %+v", a)
	}
	if a.LargeImage != "https://example/cover.jpg" || a.LargeText != "X" {
		t.Fatalf("assets %+v", a)
	}

	a = Build(Presence{
		Title:   "Splatoon 3",
		Console: "Nintendo Switch",
	})
	if a.Details != "Splatoon 3" || a.LargeText != "Splatoon 3" {
		t.Fatalf("game title %+v", a)
	}

	a = Build(Presence{
		Title:       "Game",
		Description: "In a party",
		Tag:         "FC: 0000-0000-0000",
		ShowTag:     true,
		TagIcon:     true,
		CoverKey:    "default",
	})
	if a.State != "In a party" || a.SmallImage != "id" || a.SmallText != "FC: 0000-0000-0000" {
		t.Fatalf("custom+icon %+v", a)
	}
}

func TestBuildParty(t *testing.T) {
	a := Build(Presence{Title: "Game", Party: true, PartySize: 2, PartyMax: 4})
	if a.State != "  " {
		t.Fatalf("party needs a state: %q", a.State)
	}
	if a.PartySize != 2 || a.PartyMax != 4 {
		t.Fatalf("party %+v", a)
	}
	p := a.payload()
	party, _ := p["party"].(map[string]any)
	size, _ := party["size"].([]int)
	if len(size) != 2 || size[0] != 2 || size[1] != 4 {
		t.Fatalf("payload party %v", p["party"])
	}
}

func TestBuildButtons(t *testing.T) {
	a := Build(Presence{
		Title: "Game",
		Buttons: []Button{
			{Label: "Buy on eShop", URL: "https://ec.nintendo.com/US/en/titles/7001"},
		},
	})
	p := a.payload()
	btns, _ := p["buttons"].([]map[string]string)
	if len(btns) != 1 || btns[0]["label"] != "Buy on eShop" || !strings.Contains(btns[0]["url"], "nintendo.com") {
		t.Fatalf("buttons %v", p["buttons"])
	}
}

func TestFormatTag(t *testing.T) {
	if g := FormatTag("WUP", "ninstar", [3]string{}); g != "ID: ninstar" {
		t.Fatalf("nnid %q", g)
	}
	if g := FormatTag("HAC", "", [3]string{"1234", "5678", "9012"}); g != "SW-1234-5678-9012" {
		t.Fatalf("switch %q", g)
	}
	if g := FormatTag("CTR", "", [3]string{"1234", "5678", "9012"}); g != "FC: 1234-5678-9012" {
		t.Fatalf("3ds fallback %q", g)
	}
	if g := FormatTag("CTR", "SW-12", [3]string{}); g != "SW-12" {
		t.Fatalf("freeform 3ds %q", g)
	}
	if g := FormatTag("NES", "PSN-foo", [3]string{}); g != "PSN-foo" {
		t.Fatalf("other console %q", g)
	}
	if g := FormatTag("BEE", "", [3]string{"12", "5678", "9012"}); g != "" {
		t.Fatalf("short fc %q", g)
	}
}
