package assets

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

// Breeze icons (KDE Breeze, LGPL) used when a system icon theme is missing.

//go:embed breeze
var breezeFS embed.FS

var (
	once sync.Once
	dir  string
)

// BreezeDir extracts bundled Breeze SVGs into a temp directory and returns it.
func BreezeDir() string {
	once.Do(func() {
		root, err := os.MkdirTemp("", "rpqt-breeze-")
		if err != nil {
			return
		}
		err = fs.WalkDir(breezeFS, "breeze", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, err := filepath.Rel("breeze", path)
			if err != nil {
				return err
			}
			dst := filepath.Join(root, rel)
			if d.IsDir() {
				if rel == "." {
					return nil
				}
				return os.MkdirAll(dst, 0o755)
			}
			b, err := breezeFS.ReadFile(path)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
				return err
			}
			return os.WriteFile(dst, b, 0o644)
		})
		if err != nil {
			os.RemoveAll(root)
			return
		}
		dir = root
	})
	return dir
}
