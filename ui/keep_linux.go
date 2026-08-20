//go:build linux

package ui

import (
	"os/exec"
	"sync"
)

type inhibitor struct {
	mu  sync.Mutex
	cmd *exec.Cmd
}

func (i *inhibitor) Set(on bool) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if on {
		if i.cmd != nil {
			return
		}
		cmd := exec.Command("systemd-inhibit", "--what=idle:sleep", "--who=Rich Presence Qt", "--why=Discord status visible", "sleep", "infinity")
		if err := cmd.Start(); err != nil {
			return
		}
		i.cmd = cmd
		return
	}
	if i.cmd != nil && i.cmd.Process != nil {
		_ = i.cmd.Process.Kill()
		_, _ = i.cmd.Process.Wait()
	}
	i.cmd = nil
}
