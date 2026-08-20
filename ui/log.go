package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type logger struct {
	mu      sync.Mutex
	dir     string
	enabled bool
	lines   []string
}

func (l *logger) SetEnabled(v bool) {
	l.mu.Lock()
	l.enabled = v
	l.mu.Unlock()
}

func (l *logger) Print(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	line := time.Now().Format("[15:04:05] ") + msg
	fmt.Println(line)
	l.mu.Lock()
	l.lines = append(l.lines, line)
	enabled := l.enabled
	dir := l.dir
	l.mu.Unlock()
	if enabled && dir != "" {
		_ = os.MkdirAll(dir, 0o755)
		f, err := os.OpenFile(filepath.Join(dir, "debug.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return
		}
		_, _ = f.WriteString(line + "\n")
		_ = f.Close()
	}
}
