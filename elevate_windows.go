//go:build windows

package main

import (
	"os"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ensureAdmin exits the current process after launching an elevated copy when needed.
// Production builds also request admin via the exe manifest (UAC).
func ensureAdmin() {
	if isAdmin() {
		return
	}
	exe, err := os.Executable()
	if err != nil {
		os.Exit(1)
	}
	// Re-launch elevated
	verb, _ := syscall.UTF16PtrFromString("runas")
	file, _ := syscall.UTF16PtrFromString(exe)
	// preserve args (skip argv0)
	args := strings.Join(os.Args[1:], " ")
	param, _ := syscall.UTF16PtrFromString(args)
	dir, _ := syscall.UTF16PtrFromString("")
	var show int32 = 1 // SW_SHOWNORMAL

	ret, _, _ := windows.NewLazySystemDLL("shell32.dll").NewProc("ShellExecuteW").Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(file)),
		uintptr(unsafe.Pointer(param)),
		uintptr(unsafe.Pointer(dir)),
		uintptr(show),
	)
	// ShellExecute returns >32 on success
	if ret <= 32 {
		os.Exit(1)
	}
	os.Exit(0)
}

func isAdmin() bool {
	var sid *windows.SID
	// BUILTIN\Administrators
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid,
	)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)

	token := windows.Token(0)
	member, err := token.IsMember(sid)
	if err != nil {
		return false
	}
	return member
}
