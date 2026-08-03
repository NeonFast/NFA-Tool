package token

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Info holds validated Steam refresh token metadata.
type Info struct {
	AccountHint string
	SteamID     string
	ExpiresIn   time.Duration
	Raw         string
}

// ParseAndValidate checks Steam refresh JWT structure (signature not verified — same as original).
func ParseAndValidate(raw string) (*Info, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.ReplaceAll(raw, " ", "")
	raw = strings.ReplaceAll(raw, "\n", "")
	raw = strings.ReplaceAll(raw, "\r", "")
	// Normalize capitalized header variant from some sellers
	raw = strings.ReplaceAll(raw,
		"EyAidHlwIjogIkpXVCIsICJhbGciOiAiRWREU0EiIH0",
		"eyAidHlwIjogIkpXVCIsICJhbGciOiAiRWREU0EiIH0",
	)

	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	payload, err := decodeSegment(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid token payload: %w", err)
	}

	var claims struct {
		Iss string   `json:"iss"`
		Sub string   `json:"sub"`
		Aud any      `json:"aud"`
		Exp int64    `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("token json: %w", err)
	}
	if !strings.EqualFold(claims.Iss, "steam") {
		return nil, fmt.Errorf("token issuer is not steam")
	}
	if !audHasClient(claims.Aud) {
		return nil, fmt.Errorf("token audience missing client")
	}
	if claims.Sub == "" {
		return nil, fmt.Errorf("token missing steam id")
	}

	exp := time.Unix(claims.Exp, 0)
	remaining := time.Until(exp)
	if remaining <= 0 {
		return nil, fmt.Errorf("token expired")
	}

	return &Info{
		SteamID:   claims.Sub,
		ExpiresIn: remaining,
		Raw:       raw,
	}, nil
}

// ParseAccountKey accepts "login----token" or bare token.
func ParseAccountKey(input string) (account string, token string, err error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", "", fmt.Errorf("empty input")
	}

	// Common marketplace format: user----pass----token or user----token
	if strings.Contains(input, "----") {
		parts := strings.Split(input, "----")
		account = strings.ToLower(strings.TrimSpace(parts[0]))
		if i := strings.Index(account, "@"); i >= 0 {
			account = account[:i]
		}
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if looksLikeJWT(p) {
				return account, p, nil
			}
		}
		return "", "", fmt.Errorf("no JWT found in account key")
	}

	if looksLikeJWT(input) {
		return "", input, nil
	}
	return "", "", fmt.Errorf("invalid input format (use login----token)")
}

func looksLikeJWT(s string) bool {
	s = strings.TrimSpace(s)
	if strings.Count(s, ".") != 2 {
		return false
	}
	lower := strings.ToLower(s)
	return strings.HasPrefix(lower, "ey") || strings.Contains(s, "eyAidHlw")
}

func audHasClient(aud any) bool {
	switch v := aud.(type) {
	case string:
		return v == "client"
	case []any:
		for _, a := range v {
			if s, ok := a.(string); ok && s == "client" {
				return true
			}
		}
	}
	return false
}

func decodeSegment(seg string) ([]byte, error) {
	seg = strings.TrimSpace(seg)
	switch len(seg) % 4 {
	case 2:
		seg += "=="
	case 3:
		seg += "="
	}
	// JWT uses base64url
	if b, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(seg, "=")); err == nil {
		return b, nil
	}
	return base64.StdEncoding.DecodeString(seg)
}
