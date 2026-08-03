package steam

import (
	"encoding/hex"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	crypt32              = windows.NewLazySystemDLL("crypt32.dll")
	procCryptProtectData = crypt32.NewProc("CryptProtectData")
	procCryptUnprotectData = crypt32.NewProc("CryptUnprotectData")
)

// Matches Python win32crypt flags: UI_FORBIDDEN (1) | AUDIT (16) = 17
const cryptProtectFlags = 0x11

type dataBlob struct {
	cbData uint32
	pbData *byte
}

func newBlob(data []byte) *dataBlob {
	if len(data) == 0 {
		return &dataBlob{}
	}
	return &dataBlob{
		cbData: uint32(len(data)),
		pbData: &data[0],
	}
}

func (b *dataBlob) bytes() []byte {
	if b == nil || b.cbData == 0 || b.pbData == nil {
		return nil
	}
	return unsafe.Slice(b.pbData, b.cbData)
}

func (b *dataBlob) free() {
	if b != nil && b.pbData != nil {
		_, _ = windows.LocalFree(windows.Handle(unsafe.Pointer(b.pbData)))
		b.pbData = nil
		b.cbData = 0
	}
}

// EncryptToken encrypts a Steam refresh token the same way the original tool does.
func EncryptToken(token, accountName string) (string, error) {
	in := newBlob([]byte(token))
	entropy := newBlob([]byte(accountName))
	descr, err := windows.UTF16PtrFromString("BObfuscateBuffer")
	if err != nil {
		return "", err
	}

	var out dataBlob
	r1, _, err := procCryptProtectData.Call(
		uintptr(unsafe.Pointer(in)),
		uintptr(unsafe.Pointer(descr)),
		uintptr(unsafe.Pointer(entropy)),
		0,
		0,
		uintptr(cryptProtectFlags),
		uintptr(unsafe.Pointer(&out)),
	)
	if r1 == 0 {
		if err != nil {
			return "", fmt.Errorf("CryptProtectData: %w", err)
		}
		return "", fmt.Errorf("CryptProtectData failed")
	}
	defer out.free()

	return fmt.Sprintf("%x", out.bytes()), nil
}

// DecryptToken decrypts a ConnectCache DPAPI blob (hex-encoded).
func DecryptToken(encryptedHex string) (string, error) {
	raw, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return "", fmt.Errorf("invalid hex: %w", err)
	}

	in := newBlob(raw)
	var out dataBlob
	r1, _, err := procCryptUnprotectData.Call(
		uintptr(unsafe.Pointer(in)),
		0,
		0,
		0,
		0,
		0,
		uintptr(unsafe.Pointer(&out)),
	)
	if r1 == 0 {
		if err != nil {
			return "", fmt.Errorf("CryptUnprotectData: %w", err)
		}
		return "", fmt.Errorf("CryptUnprotectData failed")
	}
	defer out.free()

	return string(out.bytes()), nil
}
