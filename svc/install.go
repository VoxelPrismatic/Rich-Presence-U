package svc

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ApplicationsDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "applications")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".local", "share", "applications")
}

func LogoPath(configDir string) string {
	return filepath.Join(configDir, "logo.png")
}

// DesktopEntry is a launcher pointing at the installed binary and icon.
func DesktopEntry(execPath, iconPath string) string {
	return `[Desktop Entry]
Type=Application
Version=1.0
Name=Rich Presence Qt
Comment=Show Nintendo games on Discord
Exec=` + quotePath(execPath) + `
TryExec=` + execPath + `
Icon=` + iconPath + `
Terminal=false
Categories=Game;Network;
StartupNotify=false
StartupWMClass=rich-presence-u
Keywords=discord;nintendo;switch;presence;
`
}

func quotePath(path string) string {
	if path == "" {
		return path
	}
	if strings.ContainsAny(path, " \t") {
		return `"` + path + `"`
	}
	return path
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.Size() > 0
}

// Present is true when the launcher binary, logo, and desktop file exist.
func Present(configDir, appsDir string) bool {
	return fileExists(LauncherPath(configDir)) && fileExists(LogoPath(configDir)) && fileExists(filepath.Join(appsDir, DesktopFile))
}

func Installed(configDir string) bool {
	return Present(configDir, ApplicationsDir())
}

func copyExecutable(dst string) error {
	src, err := os.Executable()
	if err != nil {
		return err
	}
	if resolved, err := filepath.EvalSymlinks(src); err == nil {
		src = resolved
	}
	srcAbs, err := filepath.Abs(src)
	if err != nil {
		return err
	}
	dstAbs, err := filepath.Abs(dst)
	if err != nil {
		return err
	}
	if srcAbs == dstAbs {
		return os.Chmod(dstAbs, 0o755)
	}
	in, err := os.Open(srcAbs)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dstAbs), 0o755); err != nil {
		return err
	}
	tmp := dstAbs + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		os.Remove(tmp)
		return copyErr
	}
	if closeErr != nil {
		os.Remove(tmp)
		return closeErr
	}
	if err := os.Rename(tmp, dstAbs); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Chmod(dstAbs, 0o755)
}

func writeSidecars(ctx context.Context, configDir, appsDir, binPath, version string) error {
	logo, err := fetchAsset(ctx, version, "logo.png")
	if err != nil {
		return fmt.Errorf("logo.png: %w", err)
	}
	icon := LogoPath(configDir)
	if err := os.WriteFile(icon, logo, 0o644); err != nil {
		return err
	}
	if err := os.MkdirAll(appsDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(appsDir, DesktopFile), []byte(DesktopEntry(binPath, icon)), 0o644)
}

func Install(ctx context.Context, configDir string) error {
	return InstallTo(ctx, configDir, ApplicationsDir())
}

func InstallTo(ctx context.Context, configDir, appsDir string) error {
	if configDir == "" {
		return fmt.Errorf("missing config dir")
	}
	if appsDir == "" {
		return fmt.Errorf("missing applications dir")
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}
	bin := LauncherPath(configDir)
	if err := copyExecutable(bin); err != nil {
		return err
	}
	return writeSidecars(ctx, configDir, appsDir, bin, VERSION)
}

func ApplyUpdate(ctx context.Context, configDir, latest string) (string, error) {
	return ApplyUpdateTo(ctx, configDir, ApplicationsDir(), latest)
}

func ApplyUpdateTo(ctx context.Context, configDir, appsDir, latest string) (string, error) {
	tag := latest
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + tag
	}
	bin := LauncherPath(configDir)
	url := githubBase + "/releases/download/" + tag + "/" + LauncherName
	if err := downloadToFile(ctx, url, bin, 0o755); err != nil {
		return "", fmt.Errorf("binary: %w", err)
	}
	if err := writeSidecars(ctx, configDir, appsDir, bin, latest); err != nil {
		return "", err
	}
	return bin, nil
}
