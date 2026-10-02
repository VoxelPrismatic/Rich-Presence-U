package svc

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	githubBase = "https://github.com/" + GitHubRepo
	githubHTTP = &http.Client{Timeout: 45 * time.Second}
)

func versionFromLocation(loc string) string {
	loc = strings.TrimSpace(loc)
	if loc == "" {
		return ""
	}
	if i := strings.Index(loc, "?"); i >= 0 {
		loc = loc[:i]
	}
	tag := loc
	if i := strings.LastIndex(loc, "/"); i >= 0 {
		tag = loc[i+1:]
	}
	return strings.TrimPrefix(tag, "v")
}

func versionParts(v string) (major, minor, patch int, ok bool) {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if v == "" {
		return 0, 0, 0, false
	}
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}
	bits := strings.Split(v, ".")
	if len(bits) < 1 {
		return 0, 0, 0, false
	}
	n := func(s string) int {
		x, _ := strconv.Atoi(s)
		return x
	}
	major = n(bits[0])
	if len(bits) > 1 {
		minor = n(bits[1])
	}
	if len(bits) > 2 {
		patch = n(bits[2])
	}
	return major, minor, patch, true
}

// Newer reports whether latest is a higher semver than current.
func Newer(latest, current string) bool {
	lm, ln, lp, okL := versionParts(latest)
	cm, cn, cp, okC := versionParts(current)
	if !okL || !okC {
		return false
	}
	if lm != cm {
		return lm > cm
	}
	if ln != cn {
		return ln > cn
	}
	return lp > cp
}

// LatestVersion reads GitHub's /releases/latest redirect target.
func LatestVersion(ctx context.Context) (string, error) {
	c := *githubHTTP
	c.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubBase+"/releases/latest", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", UserAgent())
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	loc := resp.Header.Get("Location")
	if loc == "" {
		return "", fmt.Errorf("latest release: HTTP %s, no Location", resp.Status)
	}
	ver := versionFromLocation(loc)
	if ver == "" {
		return "", fmt.Errorf("latest release: bad Location %q", loc)
	}
	return ver, nil
}

func getBytes(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent())
	resp, err := githubHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", rawURL, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func fetchAsset(ctx context.Context, version, name string) ([]byte, error) {
	version = strings.TrimPrefix(version, "v")
	urls := []string{
		githubBase + "/releases/download/v" + version + "/" + name,
		githubBase + "/releases/latest/download/" + name,
	}
	var last error
	for _, u := range urls {
		b, err := getBytes(ctx, u)
		if err != nil {
			last = err
			continue
		}
		if len(b) == 0 {
			last = fmt.Errorf("%s: empty", u)
			continue
		}
		return b, nil
	}
	if last == nil {
		last = fmt.Errorf("asset %s not found", name)
	}
	return nil, last
}

// DownloadProgress reports bytes written. total is negative when the server
// does not send a length.
type DownloadProgress func(written, total int64)

func downloadToFile(ctx context.Context, rawURL, dst string, mode os.FileMode, progress DownloadProgress) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", UserAgent())
	resp, err := githubHTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", rawURL, resp.Status)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	tmp := dst + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	fail := func(cause error) error {
		f.Close()
		os.Remove(tmp)
		return cause
	}
	total := resp.ContentLength
	var written int64
	if progress != nil {
		progress(0, total)
	}
	buf := make([]byte, 32*1024)
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				return fail(werr)
			}
			written += int64(n)
			if progress != nil {
				progress(written, total)
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return fail(rerr)
		}
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, dst); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Chmod(dst, mode)
}
