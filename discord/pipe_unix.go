//go:build unix

package discord

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

func ipcPaths() []string {
	var roots []string
	seen := map[string]bool{}
	add := func(dir string) {
		if dir == "" || seen[dir] {
			return
		}
		seen[dir] = true
		roots = append(roots, dir)
	}
	if xdg := os.Getenv("XDG_RUNTIME_DIR"); xdg != "" {
		add(xdg)
		add(filepath.Join(xdg, "snap.discord"))
		add(filepath.Join(xdg, "app", "com.discordapp.Discord"))
		add(filepath.Join(xdg, "app", "com.discord.Discord"))
	}
	add(os.TempDir())
	for _, env := range []string{"TMPDIR", "TMP", "TEMP"} {
		add(os.Getenv(env))
	}

	var paths []string
	for _, root := range roots {
		for i := 0; i < 10; i++ {
			paths = append(paths, filepath.Join(root, fmt.Sprintf("discord-ipc-%d", i)))
		}
	}
	return paths
}

func dialIPC(path string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("unix", path, timeout)
}
