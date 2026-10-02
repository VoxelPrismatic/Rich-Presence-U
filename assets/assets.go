package assets

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

// Breeze icons (KDE Breeze, LGPL). light/ is for a light UI and dark/ is for a dark UI.

//go:embed light dark
var iconFS embed.FS

var (
	once sync.Once
	dir  string
)

// IconRoot extracts the bundled light and dark Breeze SVGs into a temp directory.
// The directory contains light/ and dark/, each with actions/ and places/.
func IconRoot() string {
	once.Do(func() {
		root, err := os.MkdirTemp("", "rpqt-icons-")
		if err != nil {
			return
		}
		err = fs.WalkDir(iconFS, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if path == "." {
				return nil
			}
			dst := filepath.Join(root, path)
			if d.IsDir() {
				return os.MkdirAll(dst, 0o755)
			}
			b, err := iconFS.ReadFile(path)
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
