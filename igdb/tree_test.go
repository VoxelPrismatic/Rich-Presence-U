package igdb

import (
	"strings"
	"testing"
)

func TestDisplayName(t *testing.T) {
	cases := []struct {
		slug string
		want string
	}{
		{"NS1", "Nintendo Switch"},
		{"NS2", "Nintendo Switch 2"},
		{"Wii", "Wii"},
		{"FC-Disc", "Famicom Disc System"},
		{"GBA", "GameBoy Advance"},
		{"GCN", "GameCube"},
	}
	for _, tc := range cases {
		p, ok := BySlug(tc.slug)
		if !ok {
			t.Fatalf("BySlug(%q) missing", tc.slug)
		}
		if got := p.DisplayName(); got != tc.want {
			t.Errorf("%s DisplayName = %q, want %q", tc.slug, got, tc.want)
		}
	}
}

func TestShopNameAndPageURL(t *testing.T) {
	ns1, _ := BySlug("NS1")
	nes, _ := BySlug("NES")
	if ns1.ShopName() != "eShop" {
		t.Fatalf("switch shop %q", ns1.ShopName())
	}
	if nes.ShopName() != "" {
		t.Fatalf("NES should be IGDB, got %q", nes.ShopName())
	}
	if !strings.Contains(nes.PageURL(), "igdb.com") || !strings.Contains(nes.PageURL(), "Nintendo") {
		t.Fatalf("igdb url %q", nes.PageURL())
	}
}

func TestStoreCodeSwitchOnly(t *testing.T) {
	ns1, _ := BySlug("NS1")
	ns2, _ := BySlug("NS2")
	nes, _ := BySlug("NES")
	if ns1.StoreCode() != "HAC" || ns2.StoreCode() != "BEE" {
		t.Fatalf("switch store codes: %q %q", ns1.StoreCode(), ns2.StoreCode())
	}
	if nes.StoreCode() != "" {
		t.Fatalf("NES should not search the eShop, got %q", nes.StoreCode())
	}
}

func TestSlugForStoreCode(t *testing.T) {
	if SlugForStoreCode("HAC") != "NS1" || SlugForStoreCode("bee") != "NS2" {
		t.Fatal("switch mapping")
	}
	if SlugForStoreCode("CTR") != "3DS" || SlugForStoreCode("WUP") != "WiiU" {
		t.Fatal("legacy mapping")
	}
	if SlugForStoreCode("XBOX") != "" {
		t.Fatal("unknown should be empty")
	}
}

func TestByID(t *testing.T) {
	p, ok := ByID(130)
	if !ok || p.Slug != "NS1" {
		t.Fatalf("id 130: %+v %v", p, ok)
	}
	if _, ok := ByID(0); ok {
		t.Fatal("id 0 should miss")
	}
}

func TestPickerTree(t *testing.T) {
	tree := PickerTree()
	if len(tree) == 0 {
		t.Fatal("empty tree")
	}
	if tree[0].Label != "Nintendo" {
		t.Fatalf("first manufacturer %q", tree[0].Label)
	}
	if tree[0].Leaf() {
		t.Fatal("Nintendo should be a group")
	}

	var switchFam *Node
	var gamecube *Node
	for i := range tree[0].Children {
		c := &tree[0].Children[i]
		switch c.Label {
		case "Switch":
			switchFam = c
		case "GameCube":
			gamecube = c
		}
	}
	if switchFam == nil {
		t.Fatal("missing Switch family")
	}
	if tree[0].Children[0].Label != "Switch" {
		t.Fatalf("Switch should be first Nintendo row, got %q", tree[0].Children[0].Label)
	}
	if switchFam.Leaf() || len(switchFam.Children) != 2 {
		t.Fatalf("Switch should drill to 2 consoles, got %+v", switchFam)
	}
	if switchFam.Children[0].Platform.Slug != "NS1" || switchFam.Children[1].Platform.Slug != "NS2" {
		t.Fatalf("Switch children: %+v", switchFam.Children)
	}
	if gamecube == nil || !gamecube.Leaf() || gamecube.Platform.Slug != "GCN" {
		t.Fatalf("GameCube should be a collapsed leaf, got %+v", gamecube)
	}
}

func TestMakerOrder(t *testing.T) {
	tree := PickerTree()
	want := []string{"Nintendo", "PlayStation", "Xbox", "PC"}
	for i, w := range want {
		if i >= len(tree) || tree[i].Label != w {
			t.Fatalf("maker[%d]=%q want %q", i, labelAt(tree, i), w)
		}
	}
}

func TestXboxFlattened(t *testing.T) {
	xbox := mustFind(t, PickerTree(), "Xbox")
	if len(xbox.Children) == 0 || !xbox.Children[0].Leaf() {
		t.Fatalf("Xbox should list consoles directly, got %+v", xbox.Children)
	}
	if xbox.Children[0].Platform.Slug != "XB" {
		t.Fatalf("first Xbox row %+v", xbox.Children[0])
	}
}

func TestPCHierarchy(t *testing.T) {
	pc := mustFind(t, PickerTree(), "PC")
	got := childLabels(*pc)
	want := []string{"Windows", "macOS", "Linux", "Legacy"}
	if len(got) < 4 {
		t.Fatalf("PC children %v", got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Fatalf("PC[%d]=%q want %q in %v", i, got[i], w, got)
		}
	}
	for _, slug := range []string{"Win", "Mac", "Linux"} {
		name := map[string]string{"Win": "Windows", "Mac": "macOS", "Linux": "Linux"}[slug]
		n := mustFind(t, pc.Children, name)
		if !n.Leaf() || n.Platform.Slug != slug {
			t.Fatalf("%s should be a collapsed leaf, got %+v", name, n)
		}
	}
	legacy := mustFind(t, pc.Children, "Legacy")
	if legacy.Leaf() {
		t.Fatal("Legacy should be a group")
	}
	x68 := mustFind(t, legacy.Children, "Sharp X68000")
	if x68.Platform.Slug != "X68k" {
		t.Fatalf("X68000 %+v", x68)
	}
	mz := mustFind(t, legacy.Children, "Sharp MZ-2200")
	if mz.Platform.Slug != "MZ2200" {
		t.Fatalf("MZ-2200 %+v", mz)
	}
}

func TestRetroAtariSega(t *testing.T) {
	retro := mustFind(t, PickerTree(), "Retro")
	if len(retro.Children) < 2 || retro.Children[0].Label != "Atari" || retro.Children[1].Label != "Sega" {
		t.Fatalf("Retro should start Atari, Sega, got %v", childLabels(*retro))
	}
	atari := mustFind(t, retro.Children, "Atari")
	if atari.Leaf() {
		t.Fatal("Atari should be a family")
	}
	if mustFind(t, atari.Children, "Atari 2600").Platform.Slug != "A2600" {
		t.Fatal("missing Atari 2600")
	}
	sega := mustFind(t, retro.Children, "Sega")
	if mustFind(t, sega.Children, "Sega Mega Drive").Platform.Slug != "GEN" {
		t.Fatal("missing Genesis/Mega Drive")
	}
}

func TestMiniComputerBrandThenSystem(t *testing.T) {
	mini := mustFind(t, PickerTree(), "Mini Computer")
	if mini.Children[0].Label != "PDP" {
		t.Fatalf("PDP should be first Mini Computer brand, got %v", childLabels(*mini))
	}
	pdp := mustFind(t, mini.Children, "PDP")
	if pdp.Leaf() {
		t.Fatal("PDP should be a brand group")
	}
	pdp1 := mustFind(t, pdp.Children, "PDP-1")
	if pdp1.Platform.Slug != "PDP1" {
		t.Fatalf("PDP-1 %+v", pdp1)
	}
	cdc := mustFind(t, mini.Children, "CDC")
	if cdc.Leaf() {
		t.Fatal("single-system Mini Computer brands should still show the brand")
	}
	if len(cdc.Children) != 1 || cdc.Children[0].Platform.Slug != "Cyber70" {
		t.Fatalf("CDC children %+v", cdc.Children)
	}
}

func TestSinclairIsHomeComputer(t *testing.T) {
	home := mustFind(t, PickerTree(), "Home Computer")
	sinclair := mustFind(t, home.Children, "Sinclair")
	if sinclair.Leaf() {
		t.Fatal("Sinclair should be a family")
	}
	if mustFind(t, sinclair.Children, "ZX Spectrum").Platform.Slug != "ZX" {
		t.Fatal("missing ZX Spectrum")
	}
	for _, c := range home.Children {
		if strings.Contains(strings.ToLower(c.Label), "sharp") {
			t.Fatalf("Sharp should be PC Legacy, not Home Computer: %q", c.Label)
		}
	}
	legacy := mustFind(t, mustFind(t, PickerTree(), "PC").Children, "Legacy")
	for _, c := range legacy.Children {
		if strings.Contains(strings.ToLower(c.Label), "sinclair") || c.Label == "ZX Spectrum" {
			t.Fatalf("Sinclair should not be PC Legacy: %q", c.Label)
		}
	}
}

func TestPSVRInPlayStationAndVR(t *testing.T) {
	tree := PickerTree()
	ps := mustFind(t, tree, "PlayStation")
	psVR := mustFind(t, ps.Children, "VR")
	if mustFind(t, psVR.Children, "PlayStation VR").Platform.Slug != "PSVR" {
		t.Fatal("PlayStation > VR missing PSVR")
	}
	if mustFind(t, psVR.Children, "PlayStation VR2").Platform.Slug != "PSVR2" {
		t.Fatal("PlayStation > VR missing PSVR2")
	}

	vr := mustFind(t, tree, "Virtual Reality")
	vrPS := mustFind(t, vr.Children, "PlayStation")
	if mustFind(t, vrPS.Children, "PlayStation VR").Platform.Slug != "PSVR" {
		t.Fatal("Virtual Reality > PlayStation missing PSVR")
	}
	if mustFind(t, vrPS.Children, "PlayStation VR2").Platform.Slug != "PSVR2" {
		t.Fatal("Virtual Reality > PlayStation missing PSVR2")
	}
}

func TestUnmappedBacklog(t *testing.T) {
	for mfr, consoles := range Platforms {
		for name, id := range consoles {
			if id != 203 || name != "DUPLICATE Stadia" {
				t.Errorf("unmapped %s %q id=%d — add to MappedPlatforms", mfr, name, id)
			}
		}
	}
}

func TestMappedIDsUnique(t *testing.T) {
	seen := map[PlatformID]string{}
	for maker, families := range MappedPlatforms {
		for fam, plats := range families {
			for _, p := range plats {
				if p.ID == 0 || p.Slug == "" {
					t.Errorf("%s/%s missing id or slug: %+v", maker, fam, p)
				}
				if prev, ok := seen[p.ID]; ok {
					if prev != p.Slug {
						t.Errorf("duplicate ID %d with different slugs (%s vs %s/%s)", p.ID, prev, maker, p.Slug)
					}
					continue
				}
				seen[p.ID] = p.Slug
			}
		}
	}
	if _, ok := seen[203]; ok {
		t.Fatal("DUPLICATE Stadia 203 should stay unmapped")
	}
}

func labelAt(nodes []Node, i int) string {
	if i < 0 || i >= len(nodes) {
		return ""
	}
	return nodes[i].Label
}

func childLabels(n Node) []string {
	out := make([]string, len(n.Children))
	for i, c := range n.Children {
		out[i] = c.Label
	}
	return out
}

func mustFind(t *testing.T, nodes []Node, label string) *Node {
	t.Helper()
	for i := range nodes {
		if nodes[i].Label == label {
			return &nodes[i]
		}
	}
	labels := make([]string, len(nodes))
	for i, c := range nodes {
		labels[i] = c.Label
	}
	t.Fatalf("missing %q in %v", label, labels)
	return nil
}
