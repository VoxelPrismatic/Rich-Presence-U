package nso

import (
	"strings"
	"testing"
)

func TestParseMetadata(t *testing.T) {
	raw := `[bin]
latest=1600
minimal=1300
version="1.6.0"
changes="- New language\n- Sorted list"
url="https://example.com/download"

[dlc]
hac_client="1259967215323840564"
hac_titles="https://example.com/hac.csv"
hac_assets="https://example.com/hac"
bee_client="1385689410502263016"

[url]
home="https://ninstars.blogspot.com/rpc"
help="https://example.com/help"
help_pt="https://example.com/help-pt"
`
	m, err := ParseMetadata(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	if m.Latest != 1600 || m.Version != "1.6.0" {
		t.Fatalf("bin: %+v", m)
	}
	if !strings.Contains(m.Changes, "Sorted list") {
		t.Fatalf("changes %q", m.Changes)
	}
	if m.ClientID(HAC) != "1259967215323840564" {
		t.Fatalf("client %q", m.ClientID(HAC))
	}
	if m.AssetsURL(HAC) != "https://example.com/hac/" {
		t.Fatalf("assets missing slash: %q", m.AssetsURL(HAC))
	}
	if m.HelpURL("pt") != "https://example.com/help-pt" {
		t.Fatalf("help pt %q", m.HelpURL("pt"))
	}
	if m.HelpURL("de") != "https://example.com/help" {
		t.Fatalf("help fallback %q", m.HelpURL("de"))
	}
	if m.ClientID(WUP) != defaultWUPClient {
		t.Fatalf("default wup client should remain")
	}
}

func TestDefaultMetadata(t *testing.T) {
	m := DefaultMetadata()
	if m.ClientID(BEE) == "" || m.ClientID(HAC) == "" {
		t.Fatalf("defaults incomplete: %+v", m)
	}
	if m.Titles[HAC] != "" || m.Assets[CTR] != "" {
		t.Fatalf("should not ship ninstar title urls: %+v", m)
	}
}
