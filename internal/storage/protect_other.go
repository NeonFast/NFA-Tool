//go:build !windows

package storage

import "fmt"

func sealToken(plain string) (string, error) {
	return "", fmt.Errorf("token encryption is only supported on Windows")
}

func openToken(stored string) (string, error) {
	return "", fmt.Errorf("token encryption is only supported on Windows")
}
