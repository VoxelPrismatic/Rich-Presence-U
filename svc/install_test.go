package svc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
