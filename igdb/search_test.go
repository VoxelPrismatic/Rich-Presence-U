package igdb

import "testing"

func findHit(hits []Hit, slug string) (Hit, bool) {
	for _, h := range hits {
		if h.Platform.Slug == slug {
			return h, true
		}
	}
	return Hit{}, false
}

func TestSearchLabels(t *testing.T) {
	cases := []struct {
		q, slug, label string
	}{
		{"gamecube", "GCN", "Nintendo" + chevron + "GameCube"},
		{"GAME CUBE", "GCN", "Nintendo" + chevron + "GameCube"},
		{"amiga", "Amiga", "Commodore" + chevron + "Commodore Amiga"},
		{"64dd", "64DD", "Nintendo 64" + chevron + "N64 DD"},
		{"ns2", "NS2", "Switch" + chevron + "Nintendo Switch 2"},
	}
	for _, tc := range cases {
		hits := Search(tc.q)
		h, ok := findHit(hits, tc.slug)
		if !ok {
			t.Fatalf("%q: missing %s in %#v", tc.q, tc.slug, labels(hits))
		}
		if h.Label() != tc.label {
			t.Errorf("%q %s: got %q, want %q", tc.q, tc.slug, h.Label(), tc.label)
		}
	}
}

func TestSearchAndTokens(t *testing.T) {
	hits := Search("switch 2")
	if _, ok := findHit(hits, "NS2"); !ok {
		t.Fatalf("switch 2 should hit NS2, got %v", labels(hits))
	}
	if _, ok := findHit(hits, "NS1"); ok {
		t.Fatal("switch 2 should not hit NS1")
	}
}

func TestSearchFuzzy(t *testing.T) {
	hits := Search("gcmc")
	if _, ok := findHit(hits, "GCN"); !ok {
		t.Fatalf("fuzzy gcmc should hit GameCube, got %v", labels(hits))
	}
	if hits := Search("zzzzqqq"); len(hits) != 0 {
		t.Fatalf("expected no hits, got %v", labels(hits))
	}
}

func TestSearchEmpty(t *testing.T) {
	if hits := Search("   "); hits != nil {
		t.Fatalf("blank query: %v", hits)
	}
}

func TestSearchDedupesDuplicates(t *testing.T) {
	for _, q := range []string{"psvr", "playstation vr", "ps vr"} {
		hits := Search(q)
		n := 0
		for _, h := range hits {
			if h.Platform.ID == 165 || h.Platform.Slug == "PSVR" {
				n++
			}
		}
		if n != 1 {
			t.Errorf("%q: PSVR should appear once, got %d in %v", q, n, labels(hits))
		}
	}
}

func labels(hits []Hit) []string {
	out := make([]string, len(hits))
	for i, h := range hits {
		out[i] = h.Label()
	}
	return out
}
