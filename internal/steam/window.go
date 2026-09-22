package steam

import (
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                       = windows.NewLazyDLL("user32.dll")
	procEnumWindows              = user32.NewProc("EnumWindows")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")

	kernel32                         = windows.NewLazyDLL("kernel32.dll")
	procQueryFullProcessImageNameW   = kernel32.NewProc("QueryFullProcessImageNameW")
)

func queryFullProcessImageName(h windows.Handle, buf *[512]uint16, n *uint32) error {
	r, _, err := procQueryFullProcessImageNameW.Call(
		uintptr(h), 0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(n)),
	)
	if r == 0 {
		return err
	}
	return nil
}

// SteamWindowVisible reports whether steam.exe currently owns any visible
// top-level window — the main client, the account switcher or a login
// dialog all count (window classes differ between them, so we match by PID).
func SteamWindowVisible() bool {
	found := false
	cb := windows.NewCallback(func(hwnd uintptr, _ uintptr) uintptr {
		vis, _, _ := procIsWindowVisible.Call(hwnd)
		if vis == 0 {
			return 1
		}
		var pid uint32
		procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
		if pid == 0 || !pidHasExe(pid, "steam.exe") {
			return 1
		}
		found = true
		return 0
	})
	procEnumWindows.Call(cb, 0)
	return found
}

// WaitForSteamWindow polls until the Steam window appears or the timeout
// expires. Returns true when the window was seen.
func WaitForSteamWindow(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if SteamWindowVisible() {
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	return SteamWindowVisible()
}

func pidHasExe(pid uint32, exe string) bool {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return false
	}
	defer windows.CloseHandle(h)
	var buf [512]uint16
	n := uint32(len(buf))
	if err := queryFullProcessImageName(h, &buf, &n); err != nil {
		return false
	}
	name := windows.UTF16ToString(buf[:n])
	return strings.EqualFold(filepath.Base(name), exe)
}
