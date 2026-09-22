package steam

import (
	"encoding/hex"
	"fmt"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

var descriptionRaw = []byte{
	'B', 0, 'O', 0, 'b', 0, 'f', 0, 'u', 0, 's', 0, 'c', 0, 'a', 0, 't', 0, 'e', 0,
	'B', 0, 'u', 0, 'f', 0, 'f', 0, 'e', 0, 'r', 0, 0, 0,
}

const cryptProtectFlags = 0x11

var (
	crypt32                = windows.NewLazySystemDLL("crypt32.dll")
	procCryptProtectData   = crypt32.NewProc("CryptProtectData")
	procCryptUnprotectData = crypt32.NewProc("CryptUnprotectData")
)

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
	}
}

func descriptionWide() []uint16 {
	s := string(descriptionRaw)
	u := utf16Encode(s)
	u = append(u, 0)
	return u
}

func utf16Encode(s string) []uint16 {
	out := make([]uint16, 0, len(s)+1)
	for _, r := range s {
		if r < 0x10000 {
			out = append(out, uint16(r))
		} else {
			r -= 0x10000
			out = append(out, uint16(0xD800+(r>>10)), uint16(0xDC00+(r&0x3FF)))
		}
	}
	return out
}

// EncryptToken DPAPI-encrypts a refresh token with the account name as entropy.
func EncryptToken(token, accountName string) (string, error) {
	in := newBlob([]byte(token))
	entropy := newBlob([]byte(accountName))
	desc := descriptionWide()

	var out dataBlob
	r1, _, callErr := procCryptProtectData.Call(
		uintptr(unsafe.Pointer(in)),
		uintptr(unsafe.Pointer(&desc[0])),
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

// DecryptToken decrypts a ConnectCache blob, using the account name as entropy.
func DecryptToken(encryptedHex, accountName string) (string, error) {
	raw, err := hex.DecodeString(strings.TrimSpace(encryptedHex))
	if err != nil {
		return "", err
	}
	in := newBlob(raw)
	var entPtr uintptr
	var ent *dataBlob
	if accountName != "" {
		ent = newBlob([]byte(accountName))
		entPtr = uintptr(unsafe.Pointer(ent))
	}
	var out dataBlob
	r1, _, callErr := procCryptUnprotectData.Call(
		uintptr(unsafe.Pointer(in)),
		0,
		entPtr,
		0, 0, 0,
		uintptr(unsafe.Pointer(&out)),
	)
	if r1 == 0 {
		r1, _, callErr = procCryptUnprotectData.Call(
			uintptr(unsafe.Pointer(in)),
			0, 0, 0, 0, 0,
			uintptr(unsafe.Pointer(&out)),
		)
	}
	if r1 == 0 {
		if callErr != nil {
			return "", fmt.Errorf("CryptUnprotectData: %w", callErr)
		}
		return "", fmt.Errorf("CryptUnprotectData failed")
	}
	defer out.free()
	return strings.Trim(string(out.bytes()), "\x00"), nil
}
