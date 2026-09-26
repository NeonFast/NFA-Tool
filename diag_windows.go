//go:build windows

package main

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

type memoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

// platformSysInfo appends Windows-specific details: OS build, RAM,
// WebView2 Runtime version and elevation state.
func platformSysInfo() string {
	var s string
	if k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.READ); err == nil {
		cv, _, _ := k.GetStringValue("DisplayVersion")
		cb, _, _ := k.GetStringValue("CurrentBuild")
		ubr, _, _ := k.GetIntegerValue("UBR")
		s += fmt.Sprintf("Windows: %s (сборка %s.%d)\n", cv, cb, ubr)
		k.Close()
	}
	var mem memoryStatusEx
	mem.Length = uint32(unsafe.Sizeof(mem))
	ret, _, _ := windows.NewLazySystemDLL("kernel32.dll").NewProc("GlobalMemoryStatusEx").Call(uintptr(unsafe.Pointer(&mem)))
	if ret != 0 {
		s += fmt.Sprintf("RAM: всего %d МБ, свободно %d МБ\n", mem.TotalPhys>>20, mem.AvailPhys>>20)
	}
	for _, path := range []string{
		`SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}`,
		`SOFTWARE\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}`,
	} {
		if k, err := registry.OpenKey(registry.LOCAL_MACHINE, path, registry.READ); err == nil {
			if v, _, err := k.GetStringValue("pv"); err == nil {
				s += fmt.Sprintf("WebView2 Runtime: %s\n", v)
			}
			k.Close()
			break
		}
	}
	s += fmt.Sprintf("Права администратора: %v\n", isAdmin())
	return s
}
