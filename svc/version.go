package svc

import "path/filepath"

// VERSION is the running app version. Bump this for releases.
const VERSION = "2.6.0"

// BUILD is VERSION as an integer (2.6.0 -> 2600).
const BUILD = 2600

// GitHubRepo is owner/name used for releases and the User-Agent URL.
const GitHubRepo = "VoxelPrismatic/Rich-Presence-U"

// DesktopFile is the launcher filename under applications/.
const DesktopFile = "rich-presence-u.desktop"

// LauncherName is the unversioned binary name in the config dir.
const LauncherName = "app"

func UserAgent() string {
	return "RichPresenceQt/" + VERSION + " (+https://github.com/" + GitHubRepo + ")"
}

func LauncherPath(configDir string) string {
	return filepath.Join(configDir, LauncherName)
}
