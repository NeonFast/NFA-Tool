//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// openBrowser opens url in the default browser without going through cmd.exe.
// cmd "start" splits on "&", which breaks OAuth query strings (response_type etc.).
func openBrowser(url string) error {
	if url == "" {
		return fmt.Errorf("empty url")
	}
	if err := shellExecuteOpen(url); err == nil {
		return nil
	}
	// fallback: rundll32 does not interpret &
	cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Start()
}

func shellExecuteOpen(url string) error {
	verb, err := windows.UTF16PtrFromString("open")
	if err != nil {
		return err
	}
	file, err := windows.UTF16PtrFromString(url)
	if err != nil {
		return err
	}
	// SW_SHOWNORMAL = 1
	r, _, callErr := windows.NewLazySystemDLL("shell32.dll").NewProc("ShellExecuteW").Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(file)),
		0,
		0,
		1,
	)
	// ShellExecute returns > 32 on success
	if r <= 32 {
		if callErr != nil {
			return fmt.Errorf("ShellExecute: %w", callErr)
		}
		return fmt.Errorf("ShellExecute failed: %d", r)
	}
	return nil
}
