package storage

// SealString encrypts a short secret for local storage (Windows DPAPI).
func SealString(plain string) (string, error) {
	return sealToken(plain)
}

// OpenString decrypts a value produced by SealString.
func OpenString(sealed string) (string, error) {
	return openToken(sealed)
}
