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
	required := map[string][]string{
		"actions": {
			"action-unavailable-symbolic.svg",
			"albumfolder-user-trash.svg",
			"amarok_cart_view.svg",
			"checkmark.svg",
			"chronometer.svg",
			"compass.svg",
			"configure.svg",
			"document-save.svg",
			"edit-clear-history.svg",
			"edit-find.svg",
			"go-next.svg",
			"go-previous.svg",
			"help-contextual.svg",
			"network-connect.svg",
			"network-disconnect.svg",
			"player-time.svg",
			"view-refresh.svg",
			"view-visible-off.svg",
			"view-visible.svg",
		},
		"places": {
			"folder-games.svg",
			"folder-public.svg",
		},
	}
	for kind, names := range required {
		for _, name := range names {
			p := filepath.Join(root, kind, name)
			st, err := os.Stat(p)
			if err != nil || st.Size() == 0 {
				t.Fatalf("%s: %v", p, err)
			}
		}
	}
}
