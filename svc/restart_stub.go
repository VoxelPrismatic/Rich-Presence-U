//go:build !linux

package svc

import "fmt"

func Restart(bin string) error {
	return fmt.Errorf("restart not supported")
}
