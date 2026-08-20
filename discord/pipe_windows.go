//go:build windows

package discord

import (
	"fmt"
	"net"
	"os"
	"time"
)

func ipcPaths() []string {
	var paths []string
	for i := 0; i < 10; i++ {
		paths = append(paths, fmt.Sprintf(`\\.\pipe\discord-ipc-%d`, i))
	}
	return paths
}

func dialIPC(path string, timeout time.Duration) (net.Conn, error) {
	deadline := time.Now().Add(timeout)
	for {
		f, err := os.OpenFile(path, os.O_RDWR, 0)
		if err == nil {
			return &fileConn{File: f}, nil
		}
		if time.Now().After(deadline) {
			return nil, err
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// fileConn lets a Windows named-pipe handle satisfy net.Conn.
type fileConn struct {
	*os.File
}

func (c *fileConn) LocalAddr() net.Addr                { return pipeAddr(c.Name()) }
func (c *fileConn) RemoteAddr() net.Addr               { return pipeAddr(c.Name()) }
func (c *fileConn) SetDeadline(t time.Time) error      { return nil }
func (c *fileConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *fileConn) SetWriteDeadline(t time.Time) error { return nil }

type pipeAddr string

func (a pipeAddr) Network() string { return "pipe" }
func (a pipeAddr) String() string  { return string(a) }
