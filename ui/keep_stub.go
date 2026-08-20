//go:build !linux

package ui

type inhibitor struct{}

func (i *inhibitor) Set(on bool) {}
