package ui

import (
	"os"
	"path/filepath"

	"github.com/voxelprismatic/richpresenceu/assets"
)

func iconFile(kind, name string) string {
	root := assets.IconRoot()
	if root == "" {
		return ""
	}
	path := filepath.Join(root, iconScheme, kind, name+".svg")
	if iconScheme != schemeLight {
		if st, err := os.Stat(path); err != nil || st.Size() == 0 {
			return filepath.Join(root, schemeLight, kind, name+".svg")
		}
	}
	return path
}
