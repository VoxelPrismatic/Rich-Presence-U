package nso

import (
	"os"
	"path/filepath"
	"runtime"
)

const appDirName = "rich-presence-u"

// ConfigDir is ~/.config/rich-presence-u (games.db, prefs.json).
func ConfigDir() (string, error) {
	if override := os.Getenv("RICHPRESENCEU_CONFIG"); override != "" {
		return override, nil
	}
	if override := os.Getenv("RICHPRESENCEU_DATA"); override != "" {
		return override, nil
	}
	if runtime.GOOS == "windows" {
		if app := os.Getenv("APPDATA"); app != "" {
			return filepath.Join(app, appDirName), nil
		}
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, appDirName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support", appDirName), nil
	}
	return filepath.Join(home, ".config", appDirName), nil
}

// CacheDir is ~/.cache/rich-presence-u (downloaded cover art).
func CacheDir() (string, error) {
	if override := os.Getenv("RICHPRESENCEU_CACHE"); override != "" {
		return override, nil
	}
	if runtime.GOOS == "windows" {
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			return filepath.Join(local, appDirName), nil
		}
	}
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, appDirName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Caches", appDirName), nil
	}
	return filepath.Join(home, ".cache", appDirName), nil
}

func (c *Client) dbPath() string {
	return filepath.Join(c.ConfigDir, "games.db")
}
