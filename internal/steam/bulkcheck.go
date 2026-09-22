package steam

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"

	"nfa-tool/internal/token"
)

// BulkCheckItem is the per-account verdict of a bulk token check.
// Status: "ok" (Steam CM accepted the token — login will work), "rejected",
// "expired", "invalid", "error" (the check itself failed: network, rate limit).
type BulkCheckItem struct {
	Account   string `json:"account"`
	SteamID   string `json:"steamId,omitempty"`
	Status    string `json:"status"`
	ExpiresAt string `json:"expiresAt,omitempty"`
	Detail    string `json:"detail,omitempty"`
}

// BuildProxyClient parses one proxy line and returns an HTTP client using it.
// Format: [scheme://][user:pass@]host:port, scheme = http | https | socks5
// (http is assumed when the scheme is omitted).
func BuildProxyClient(line string) (*http.Client, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("empty proxy")
	}
	if !strings.Contains(line, "://") {
		line = "http://" + line
	}
	u, err := url.Parse(line)
	if err != nil {
		return nil, err
	}
	tr := &http.Transport{}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
		tr.Proxy = http.ProxyURL(u)
	case "socks5":
		var auth *proxy.Auth
		if u.User != nil {
			pw, _ := u.User.Password()
			auth = &proxy.Auth{User: u.User.Username(), Password: pw}
		}
		dialer, err := proxy.SOCKS5("tcp", u.Host, auth, &net.Dialer{Timeout: 10 * time.Second})
		if err != nil {
			return nil, err
		}
		tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		}
	default:
		return nil, fmt.Errorf("unsupported proxy scheme %q", u.Scheme)
	}
	return &http.Client{Transport: tr, Timeout: 15 * time.Second}, nil
}

// CheckTokens bulk-checks refresh tokens: offline validation plus a
// non-consuming liveness probe — a Steam CM logon over WebSocket (unlike the
// web API, a CM logon does not rotate or kill the refresh token). onItem is
// invoked as each check completes (completion order); results keep the input
// order. Concurrency is capped low so mass checks stay gentle on the CM.
func CheckTokens(entries []token.ParsedKey, _ []string, onItem func(BulkCheckItem, token.ParsedKey)) []BulkCheckItem {
	items := make([]BulkCheckItem, len(entries))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 3)
	for i, e := range entries {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, e token.ParsedKey) {
			defer wg.Done()
			defer func() { <-sem }()
			it := checkOne(e)
			items[i] = it
			if onItem != nil {
				onItem(it, e)
			}
		}(i, e)
	}
	wg.Wait()
	return items
}

func eresultName(code int) string {
	names := map[int]string{
		1: "OK", 2: "Fail", 5: "InvalidPassword", 6: "AccountDisabled",
		8: "InvalidParam", 15: "AccessDenied", 27: "Expired", 48: "TryAnotherCM",
		63: "AccountNotFound", 73: "Pending", 84: "RateLimitExceeded",
		89: "AccountLoginDeniedNeedTwoFactor",
	}
	if n, ok := names[code]; ok {
		return n
	}
	return fmt.Sprintf("eresult %d", code)
}

func checkOne(e token.ParsedKey) BulkCheckItem {
	it := BulkCheckItem{Account: e.Account}
	info, err := token.ParseAndValidate(e.Token)
	if err != nil {
		if strings.Contains(err.Error(), "expired") {
			it.Status = "expired"
		} else {
			it.Status = "invalid"
		}
		it.Detail = err.Error()
		return it
	}
	it.SteamID = info.SteamID
	it.ExpiresAt = info.ExpiresAt.Format(dateLayout)

	eres, err := CheckTokenCM(info.Raw, 15*time.Second)
	if err == nil && (eres == 48 || eres == 84) {
		// TryAnotherCM / RateLimitExceeded — one retry
		eres, err = CheckTokenCM(info.Raw, 15*time.Second)
	}
	if err != nil {
		it.Status = "error"
		it.Detail = err.Error()
		return it
	}
	it.Detail = eresultName(eres)
	if eres == 1 {
		it.Status = "ok"
	} else {
		it.Status = "rejected"
	}
	return it
}
