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
	p := filepath.Join(root, "actions", "checkmark.svg")
	st, err := os.Stat(p)
	if err != nil || st.Size() == 0 {
		t.Fatalf("%s: %v", p, err)
	}
}
