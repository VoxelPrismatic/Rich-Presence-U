package svc

import (
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// VERSION is the running app version. Bump this for releases.
const VERSION = "2.10.0"

var VERSION_PARTS = strings.Split(VERSION, ".")
var MAJOR, _ = strconv.Atoi(VERSION_PARTS[0])
var MINOR, _ = strconv.Atoi(VERSION_PARTS[1])
var PATCH, _ = strconv.Atoi(VERSION_PARTS[2])

// BUILD is VERSION as an integer (2.6.0 -> 2600).
var BUILD = MAJOR*10000 + MINOR*1000 + PATCH

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
