package igdb

import "testing"

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
