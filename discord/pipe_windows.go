//go:build windows

package discord

import (
	"fmt"
	"net"
	"time"

	"github.com/Microsoft/go-winio"
)

func ipcPaths() []string {
	var paths []string
	for i := 0; i < 10; i++ {
		paths = append(paths, fmt.Sprintf(`\\.\pipe\discord-ipc-%d`, i))
	}
	return paths
}

func dialIPC(path string, timeout time.Duration) (net.Conn, error) {
	// os.OpenFile opens the pipe in synchronous mode. A blocked Read then
	// holds the handle, and the next Write waits until that Read finishes.
	// Discord does not answer until the write arrives, so Apply sits there
	// until the pipe is closed. Overlapped I/O lets the read loop and the
	// write run together, which is what a Unix socket already does.
	return winio.DialPipe(path, &timeout)
}
