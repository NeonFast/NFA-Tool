//go:build windows

package storage

import (
	"encoding/base64"
	"fmt"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// App-specific DPAPI entropy so blobs are not generic user DPAPI dumps.
var protectEntropy = []byte("NFA-Tool-Recode-v2\x00accounts\x00v1")

const (
	cryptProtectUIForbidden = 0x1
	tokenPrefix             = "dpapi:"
)

var (
	crypt32              = windows.NewLazySystemDLL("crypt32.dll")
	procCryptProtectData = crypt32.NewProc("CryptProtectData")
	procCryptUnprotect   = crypt32.NewProc("CryptUnprotectData")
)

type dataBlob struct {
	cbData uint32
	pbData *byte
}

func blobFrom(b []byte) *dataBlob {
	if len(b) == 0 {
		return &dataBlob{}
	}
	cp := make([]byte, len(b))
	copy(cp, b)
	return &dataBlob{cbData: uint32(len(cp)), pbData: &cp[0]}
}

func (b *dataBlob) bytes() []byte {
	if b == nil || b.cbData == 0 || b.pbData == nil {
		return nil
	}
	out := make([]byte, b.cbData)
	copy(out, unsafe.Slice(b.pbData, b.cbData))
	return out
}

func (b *dataBlob) free() {
	if b != nil && b.pbData != nil {
		_, _ = windows.LocalFree(windows.Handle(unsafe.Pointer(b.pbData)))
		b.pbData = nil
	}
}

func sealToken(plain string) (string, error) {
	plain = strings.TrimSpace(plain)
	if plain == "" {
		return "", fmt.Errorf("empty token")
	}
	if strings.HasPrefix(plain, tokenPrefix) {
		return plain, nil // already sealed
	}
	in := blobFrom([]byte(plain))
	ent := blobFrom(protectEntropy)
	descr, err := windows.UTF16PtrFromString("NFA-Tool-Recode-v2")
	if err != nil {
		return "", err
	}
	var out dataBlob
	r1, _, callErr := procCryptProtectData.Call(
		uintptr(unsafe.Pointer(in)),
		uintptr(unsafe.Pointer(descr)),
		uintptr(unsafe.Pointer(ent)),
		0, 0,
		uintptr(cryptProtectUIForbidden),
		uintptr(unsafe.Pointer(&out)),
	)
	if r1 == 0 {
		if callErr != nil {
			return "", fmt.Errorf("CryptProtectData: %w", callErr)
		}
		return "", fmt.Errorf("CryptProtectData failed")
	}
	defer out.free()
	return tokenPrefix + base64.StdEncoding.EncodeToString(out.bytes()), nil
}

func openToken(stored string) (string, error) {
	stored = strings.TrimSpace(stored)
	if stored == "" {
		return "", fmt.Errorf("empty token")
	}
	// plaintext JWT (pre-encryption DB) — migrate path
	if !strings.HasPrefix(stored, tokenPrefix) {
		if looksLikeJWT(stored) {
			return stored, nil
		}
		return "", fmt.Errorf("unknown token encoding")
	}
	rawB64 := strings.TrimPrefix(stored, tokenPrefix)
	raw, err := base64.StdEncoding.DecodeString(rawB64)
	if err != nil {
		return "", fmt.Errorf("token decode: %w", err)
	}
	in := blobFrom(raw)
	ent := blobFrom(protectEntropy)
	var out dataBlob
	r1, _, callErr := procCryptUnprotect.Call(
		uintptr(unsafe.Pointer(in)),
		0,
		uintptr(unsafe.Pointer(ent)),
		0, 0, 0,
		uintptr(unsafe.Pointer(&out)),
	)
	if r1 == 0 {
		if callErr != nil {
			return "", fmt.Errorf("CryptUnprotectData: %w", callErr)
		}
		return "", fmt.Errorf("CryptUnprotectData failed")
	}
	defer out.free()
	return string(out.bytes()), nil
}

func looksLikeJWT(s string) bool {
	return strings.Count(s, ".") == 2 && (strings.HasPrefix(strings.ToLower(s), "ey") || strings.Contains(s, "eyAidHlw"))
}
