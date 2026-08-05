package token

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Info holds validated Steam refresh token metadata.
type Info struct {
	AccountHint string
	SteamID     string
	ExpiresIn   time.Duration
	ExpiresAt   time.Time
	Raw         string
}

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

	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("token json: %w", err)
	}

	iss, _ := claims["iss"].(string)
	if !strings.EqualFold(iss, "steam") {
		return nil, fmt.Errorf("token issuer is not steam")
	}
	if !audHasClient(claims["aud"]) {
		return nil, fmt.Errorf("token audience missing client")
	}

	steamID := anyToString(claims["sub"])
	if steamID == "" {
		return nil, fmt.Errorf("token missing steam id")
	}

	var expUnix int64
	switch v := claims["exp"].(type) {
	case float64:
		expUnix = int64(v)
	case json.Number:
		expUnix, _ = v.Int64()
	case string:
		expUnix, _ = strconv.ParseInt(v, 10, 64)
	}
	if expUnix == 0 {
		return nil, fmt.Errorf("token missing exp")
	}

	remaining := time.Until(time.Unix(expUnix, 0))
	if remaining <= 0 {
		return nil, fmt.Errorf("token expired")
	}

	return &Info{
		SteamID:   steamID,
		ExpiresIn: remaining,
		ExpiresAt: time.Unix(expUnix, 0).UTC(),
		Raw:       raw,
	}, nil
}

// ParsedKey is one account line from bulk import.
type ParsedKey struct {
	Account string
	Token   string
}

// ParseAccountKey accepts "login----token" or bare token (single line).
func ParseAccountKey(input string) (account string, token string, err error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", "", fmt.Errorf("empty input")
	}
	// bulk paste: take first non-empty line only for single-login path
	if i := strings.IndexAny(input, "\r\n"); i >= 0 {
		input = strings.TrimSpace(input[:i])
	}

	// Common marketplace format: user----pass----token or user----token
	if strings.Contains(input, "----") {
		parts := strings.Split(input, "----")
		account = strings.ToLower(strings.TrimSpace(parts[0]))
		if i := strings.Index(account, "@"); i >= 0 {
			account = account[:i]
		}
		// Prefer the last JWT-looking segment (token is usually last)
		for i := len(parts) - 1; i >= 0; i-- {
			p := strings.TrimSpace(parts[i])
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

// ParseBulkKeys parses multi-line export text (login----token per line).
// Empty lines and lines starting with # are skipped.
func ParseBulkKeys(input string) ([]ParsedKey, []string) {
	var out []ParsedKey
	var errs []string
	seen := map[string]bool{}
	for _, line := range strings.Split(input, "\n") {
		line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		acc, tok, err := ParseAccountKey(line)
		if err != nil {
			errs = append(errs, truncateErr(line, err))
			continue
		}
		if acc == "" {
			errs = append(errs, truncateErr(line, fmt.Errorf("account name required")))
			continue
		}
		if _, err := ParseAndValidate(tok); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", acc, err))
			continue
		}
		key := strings.ToLower(acc)
		if seen[key] {
			// last wins — replace previous
			for i := range out {
				if strings.EqualFold(out[i].Account, acc) {
					out[i].Token = tok
					break
				}
			}
			continue
		}
		seen[key] = true
		out = append(out, ParsedKey{Account: acc, Token: tok})
	}
	return out, errs
}

func truncateErr(line string, err error) string {
	s := line
	if len(s) > 48 {
		s = s[:48] + "…"
	}
	return fmt.Sprintf("%s: %v", s, err)
}

func looksLikeJWT(s string) bool {
	s = strings.TrimSpace(s)
	if strings.Count(s, ".") != 2 {
		return false
	}
	lower := strings.ToLower(s)
	return strings.HasPrefix(lower, "ey") || strings.Contains(s, "eyAidHlw") || strings.Contains(s, "eyJ")
}

func audHasClient(aud any) bool {
	switch v := aud.(type) {
	case string:
		return v == "client" || strings.Contains(v, "client")
	case []any:
		for _, a := range v {
			if s, ok := a.(string); ok && (s == "client" || strings.Contains(s, "client")) {
				return true
			}
		}
	}
	return false
}

func anyToString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatInt(int64(t), 10)
	case json.Number:
		return t.String()
	case fmt.Stringer:
		return t.String()
	default:
		if v == nil {
			return ""
		}
		return fmt.Sprint(v)
	}
}

func decodeSegment(seg string) ([]byte, error) {
	seg = strings.TrimSpace(seg)
	// try raw URL first (no padding)
	if b, err := base64.RawURLEncoding.DecodeString(seg); err == nil {
		return b, nil
	}
	// padded URL
	if b, err := base64.URLEncoding.DecodeString(padB64(seg)); err == nil {
		return b, nil
	}
	// standard
	return base64.StdEncoding.DecodeString(padB64(seg))
}

func padB64(seg string) string {
	switch len(seg) % 4 {
	case 2:
		return seg + "=="
	case 3:
		return seg + "="
	default:
		return seg
	}
}
