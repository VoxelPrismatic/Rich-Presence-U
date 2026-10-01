package assets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBreezeDir(t *testing.T) {
	root := BreezeDir()
	if root == "" {
		t.Fatal("empty")
	}
	for _, name := range []string{
		"checkmark.svg",
		"player-time.svg",
		"chronometer.svg",
		"view-refresh.svg",
	} {
		p := filepath.Join(root, "actions", name)
		st, err := os.Stat(p)
		if err != nil || st.Size() == 0 {
			t.Fatalf("%s: %v", p, err)
		}
	}
}
