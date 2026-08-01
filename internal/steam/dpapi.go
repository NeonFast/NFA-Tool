package steam

import (
	"encoding/hex"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	crypt32                = windows.NewLazySystemDLL("crypt32.dll")
	procCryptProtectData   = crypt32.NewProc("CryptProtectData")
	procCryptUnprotectData = crypt32.NewProc("CryptUnprotectData")
)

// Python: flags=17
const cryptProtectFlags = 0x11

type dataBlob struct {
	cbData uint32
	pbData *byte
}

func newBlob(data []byte) *dataBlob {
	if len(data) == 0 {
		return &dataBlob{}
	}
	cp := make([]byte, len(data))
	copy(cp, data)
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
		b.cbData = 0
	}
}

// EncryptToken = Python steam_encrypt
func EncryptToken(token, accountName string) (string, error) {
	in := newBlob([]byte(token))
	entropy := newBlob([]byte(accountName))
	descr, err := windows.UTF16PtrFromString("BObfuscateBuffer")
	if err != nil {
		return "", err
	}
	var out dataBlob
	r1, _, callErr := procCryptProtectData.Call(
		uintptr(unsafe.Pointer(in)),
		uintptr(unsafe.Pointer(descr)),
		uintptr(unsafe.Pointer(entropy)),
		0, 0,
		uintptr(cryptProtectFlags),
		uintptr(unsafe.Pointer(&out)),
	)
	if r1 == 0 {
		if callErr != nil {
			return "", fmt.Errorf("CryptProtectData: %w", callErr)
		}
		return "", fmt.Errorf("CryptProtectData failed")
	}
	defer out.free()
	return hex.EncodeToString(out.bytes()), nil
}

// DecryptToken = Python steam_decrypt (no entropy)
func DecryptToken(encryptedHex string) (string, error) {
	raw, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return "", err
	}
	in := newBlob(raw)
	var out dataBlob
	r1, _, callErr := procCryptUnprotectData.Call(
		uintptr(unsafe.Pointer(in)),
		0, 0, 0, 0, 0,
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
