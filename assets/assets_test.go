package assets

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestIconRoot(t *testing.T) {
	root := IconRoot()
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
			"media-playback-pause.svg",
			"media-playback-start.svg",
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
	for _, scheme := range []string{"light", "dark"} {
		for kind, names := range required {
			for _, name := range names {
				p := filepath.Join(root, scheme, kind, name)
				st, err := os.Stat(p)
				if err != nil || st.Size() == 0 {
					t.Fatalf("%s: %v", p, err)
				}
			}
		}
	}
	light, err := os.ReadFile(filepath.Join(root, "light", "actions", "checkmark.svg"))
	if err != nil {
		t.Fatal(err)
	}
	dark, err := os.ReadFile(filepath.Join(root, "dark", "actions", "checkmark.svg"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(light, dark) {
		t.Fatal("dark checkmark matches light")
	}
	if !bytes.Contains(light, []byte("#232629")) {
		t.Fatal("light checkmark is missing the Breeze light text color")
	}
	if !bytes.Contains(dark, []byte("#fcfcfc")) {
		t.Fatal("dark checkmark is missing the Breeze dark text color")
	}
}
