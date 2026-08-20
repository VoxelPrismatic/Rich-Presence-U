package nso

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// CoverURL is the remote image key Discord accepts (a full https URL).
func (c *Client) CoverURL(game Game, preferred Region) string {
	if game.CoverArt != "" && storeArtID(game) {
		return game.CoverArt
	}
	if !game.Verified() {
		return ""
	}
	region, ok := game.IconRegion(preferred)
	if !ok {
		return ""
	}
	base := c.Meta.AssetsURL(game.AssetSystem)
	if base == "" {
		return ""
	}
	return base + game.NativeID() + "." + region.Lower() + ".jpg"
}

// CoverPath returns a cached local image, downloading it when missing.
// Files live in ~/.cache/rich-presence-u/{english-title}.ext
func (c *Client) CoverPath(ctx context.Context, game Game, preferred Region) (string, error) {
	if !game.Verified() {
		return "", fmt.Errorf("unverified game")
	}
	url := c.CoverURL(game, preferred)
	if url == "" {
		return "", fmt.Errorf("no cover url for %s", game.ID)
	}
	path := c.imagePath(game, url)
	if st, err := os.Stat(path); err == nil && st.Size() > 0 {
		return path, nil
	}
	body, err := c.get(ctx, url)
	if err != nil {
		return "", err
	}
	defer body.Close()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return "", err
	}
	_, copyErr := io.Copy(f, body)
	closeErr := f.Close()
	if copyErr != nil {
		os.Remove(tmp)
		return "", copyErr
	}
	if closeErr != nil {
		os.Remove(tmp)
		return "", closeErr
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return "", err
	}
	return path, nil
}

// CoverKey is the Discord asset key: a URL, or "default" when none exists.
func (c *Client) CoverKey(game Game, preferred Region) string {
	if u := c.CoverURL(game, preferred); u != "" {
		return u
	}
	return "default"
}

func (c *Client) imagePath(game Game, url string) string {
	ext := filepath.Ext(strings.ToLower(url))
	if ext == "" || len(ext) > 5 {
		ext = ".jpg"
	}
	name := sanitizeFile(game.EnglishTitle())
	if name == "" {
		name = sanitizeFile(game.NativeID())
	}
	return filepath.Join(c.CacheDir, name+ext)
}

func storeArtID(game Game) bool {
	if strings.HasPrefix(game.ID, "eu:") {
		return true
	}
	id := game.NativeID()
	if len(id) < 10 {
		return false
	}
	for _, r := range id {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func sanitizeFile(name string) string {
	name = strings.TrimSpace(name)
	var b strings.Builder
	b.Grow(len(name))
	lastUnderscore := false
	for _, r := range name {
		switch {
		case r == 0:
			continue
		case strings.ContainsRune(`<>:"/\|?*`, r):
			if !lastUnderscore {
				b.WriteByte('_')
				lastUnderscore = true
			}
		case unicode.IsControl(r):
			continue
		default:
			b.WriteRune(r)
			lastUnderscore = false
		}
	}
	return strings.Trim(b.String(), " ._")
}
