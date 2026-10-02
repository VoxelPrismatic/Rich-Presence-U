package svc

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestDesktopEntry(t *testing.T) {
	out := DesktopEntry("/tmp/rich-presence-u/app", "/tmp/rich-presence-u/logo.png")
	if strings.Contains(out, "__EXEC__") || strings.Contains(out, "app-v") {
		t.Fatal(out)
	}
	if !strings.Contains(out, "Exec=/tmp/rich-presence-u/app") {
		t.Fatal(out)
	}
	if !strings.Contains(out, "Icon=/tmp/rich-presence-u/logo.png") {
		t.Fatal(out)
	}
	if !strings.Contains(out, "StartupWMClass=wl.float") {
		t.Fatalf("class: %s", out)
	}
	if !strings.Contains(out, "TryExec=/tmp/rich-presence-u/app") {
		t.Fatal(out)
	}
}

func TestPresent(t *testing.T) {
	cfg := t.TempDir()
	apps := t.TempDir()
	if Present(cfg, apps) {
		t.Fatal("empty dirs should not count as installed")
	}
	if err := os.WriteFile(LauncherPath(cfg), []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(LogoPath(cfg), []byte("y"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(apps, DesktopFile), []byte("[Desktop Entry]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !Present(cfg, apps) {
		t.Fatal("expected installed")
	}
}

func TestInstallTo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/releases/download/v" + VERSION + "/logo.png":
			w.Write([]byte("png"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	prevBase, prevHTTP := githubBase, githubHTTP
	githubBase, githubHTTP = srv.URL, srv.Client()
	t.Cleanup(func() {
		githubBase, githubHTTP = prevBase, prevHTTP
	})

	cfg := t.TempDir()
	apps := t.TempDir()
	if err := InstallTo(context.Background(), cfg, apps); err != nil {
		t.Fatal(err)
	}
	if !Present(cfg, apps) {
		t.Fatal("files missing after write")
	}
	body, err := os.ReadFile(filepath.Join(apps, DesktopFile))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), LauncherPath(cfg)) {
		t.Fatalf("desktop exec: %s", body)
	}
	if strings.Contains(string(body), "app-v") {
		t.Fatalf("desktop should not mention a versioned binary: %s", body)
	}
}

func TestReleaseAsset(t *testing.T) {
	if got := ReleaseAsset(); got == "" {
		t.Fatal("empty")
	}
}

func TestNewer(t *testing.T) {
	if !Newer("2.7.0", "2.6.0") || Newer("2.6.0", "2.6.0") || Newer("2.5.9", "2.6.0") {
		t.Fatal("compare")
	}
	if !Newer("v3.0.0", "2.9.9") {
		t.Fatal("major")
	}
}

func TestVersionFromLocation(t *testing.T) {
	if g := versionFromLocation("https://github.com/VoxelPrismatic/Rich-Presence-U/releases/tag/v2.7.0"); g != "2.7.0" {
		t.Fatalf("%q", g)
	}
	if g := versionFromLocation("/VoxelPrismatic/Rich-Presence-U/releases/tag/2.6.1"); g != "2.6.1" {
		t.Fatalf("%q", g)
	}
}

func TestLatestVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/releases/latest" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Location", "/VoxelPrismatic/Rich-Presence-U/releases/tag/v2.8.0")
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(srv.Close)
	prevBase, prevHTTP := githubBase, githubHTTP
	githubBase, githubHTTP = srv.URL, srv.Client()
	t.Cleanup(func() {
		githubBase, githubHTTP = prevBase, prevHTTP
	})
	got, err := LatestVersion(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got != "2.8.0" {
		t.Fatalf("got %q", got)
	}
}

func TestDownloadToFileProgress(t *testing.T) {
	body := bytes.Repeat([]byte("abc"), 1000)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		w.Write(body)
	}))
	t.Cleanup(srv.Close)

	dst := filepath.Join(t.TempDir(), "app")
	var lastW, lastT int64
	var calls int
	err := downloadToFile(context.Background(), srv.URL+"/rich-presence-qt_linux", dst, 0o755, func(written, total int64) {
		calls++
		lastW, lastT = written, total
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls < 2 || lastW != int64(len(body)) || lastT != int64(len(body)) {
		t.Fatalf("progress calls %d written %d total %d", calls, lastW, lastT)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, body) {
		t.Fatalf("file length %d", len(got))
	}
	st, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm() != 0o755 {
		t.Fatalf("mode %o", st.Mode().Perm())
	}
}

func TestDownloadToFileCancel(t *testing.T) {
	started := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100000")
		w.Write(bytes.Repeat([]byte("a"), 64))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		close(started)
		<-r.Context().Done()
	}))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithCancel(context.Background())
	dst := filepath.Join(t.TempDir(), "app")
	errCh := make(chan error, 1)
	go func() {
		errCh <- downloadToFile(ctx, srv.URL+"/rich-presence-qt_linux", dst, 0o755, nil)
	}()
	<-started
	cancel()
	if err := <-errCh; err == nil {
		t.Fatal("expected cancel")
	}
}
