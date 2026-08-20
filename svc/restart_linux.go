//go:build linux

package svc

import (
	"os"
	"syscall"
)

func Restart(bin string) error {
	return syscall.Exec(bin, append([]string{bin}, os.Args[1:]...), os.Environ())
}
