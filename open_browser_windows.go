//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

func openBrowser(url string) error {
	cmd := exec.Command("cmd", "/C", "start", "", url)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Start()
}
