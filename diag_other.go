//go:build !windows

package main

// platformSysInfo is a stub for non-Windows builds.
func platformSysInfo() string { return "" }
