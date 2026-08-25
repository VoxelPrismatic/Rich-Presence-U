//go:build portable

package ui

import (
	"path/filepath"

	"github.com/voxelprismatic/richpresenceu/assets"
)

func useSystemIcons() bool { return false }

func iconFile(kind, name string) string {
	root := assets.BreezeDir()
	if root == "" {
		return ""
	}
	return filepath.Join(root, kind, name+".svg")
}
