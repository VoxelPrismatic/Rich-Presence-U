package svc

import (
	"fmt"
	"path/filepath"
	"runtime"
)

// VERSION is the running app version. Bump this for releases.
const MAJOR = 2
const MINOR = 9
const PATCH = 0

var VERSION = fmt.Sprintf("%d.%d.%d", MAJOR, MINOR, PATCH)

// BUILD is VERSION as an integer (2.6.0 -> 2600).
const BUILD = MAJOR*10000 + MINOR*1000 + PATCH

// GitHubRepo is owner/name used for releases and the User-Agent URL.
const GitHubRepo = "VoxelPrismatic/Rich-Presence-U"

// DesktopFile is the launcher filename under applications/.
const DesktopFile = "rich-presence-u.desktop"

// LauncherName is the unversioned binary name in the config dir.
const LauncherName = "app"

const (
	releaseLinux   = "rich-presence-qt_linux"
	releaseMacOS   = "rich-presence-qt_macOS.app.zip"
	releaseWindows = "rich-presence-qt_windows.zip"
)

func UserAgent() string {
	return "RichPresenceQt/" + VERSION + " (+https://github.com/" + GitHubRepo + ")"
}

func LauncherPath(configDir string) string {
	return filepath.Join(configDir, LauncherName)
}

// ReleaseAsset is the GitHub release filename for this OS.
func ReleaseAsset() string {
	switch runtime.GOOS {
	case "darwin":
		return releaseMacOS
	case "windows":
		return releaseWindows
	default:
		return releaseLinux
	}
}
