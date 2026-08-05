// Package gdrive uploads export files to the user's Google Drive via OAuth2 (desktop + PKCE).
package gdrive

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"nfa-tool/internal/storage"
)

const (
	scopeDriveFile = "https://www.googleapis.com/auth/drive.file"
	authURL        = "https://accounts.google.com/o/oauth2/v2/auth"
	tokenURL       = "https://oauth2.googleapis.com/token"
	uploadURL      = "https://www.googleapis.com/upload/drive/v3/files?uploadType=multipart&fields=id,name,webViewLink"
	credsFile      = "google-oauth.json"
	tokenFile      = "gdrive-token.sealed"
)

// Status is exposed to the UI.
type Status struct {
	HasCredentials bool   `json:"hasCredentials"`
	Connected      bool   `json:"connected"`
	ClientIDHint   string `json:"clientIdHint"`
}

// Client handles OAuth and Drive uploads. Files live next to accounts.db.
type Client struct {
	dir string
	mu  sync.Mutex

	authMu     sync.Mutex
	authCancel context.CancelFunc
	authActive bool
}

// New creates a client that stores credentials/tokens under dir.
func New(dir string) *Client {
	return &Client{dir: dir}
}

// CancelAuth aborts an in-progress browser OAuth wait (safe if idle).
func (c *Client) CancelAuth() {
	c.authMu.Lock()
	fn := c.authCancel
	c.authMu.Unlock()
	if fn != nil {
		fn()
	}
}

// AuthInProgress reports whether Connect is waiting on the browser.
func (c *Client) AuthInProgress() bool {
	c.authMu.Lock()
	defer c.authMu.Unlock()
	return c.authActive
}

type oauthCreds struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type tokenData struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
	TokenType    string    `json:"token_type"`
}

// GetStatus reports setup/connection state.
func (c *Client) GetStatus() Status {
	cr, err := c.loadCreds()
	st := Status{HasCredentials: err == nil && cr.ClientID != ""}
	if st.HasCredentials {
		st.ClientIDHint = maskID(cr.ClientID)
	}
	tok, err := c.loadToken()
	st.Connected = err == nil && tok != nil && (tok.RefreshToken != "" || tok.AccessToken != "")
	return st
}

// SaveCredentials writes Desktop OAuth client id/secret (or full Google JSON).
func (c *Client) SaveCredentials(clientID, clientSecret string) error {
	clientID = strings.TrimSpace(clientID)
	clientSecret = strings.TrimSpace(clientSecret)
	if clientID == "" {
		return fmt.Errorf("client_id required")
	}
	// allow pasting full downloaded JSON
	if strings.HasPrefix(clientID, "{") {
		parsed, err := parseCredsJSON([]byte(clientID))
		if err != nil {
			return err
		}
		return c.writeCreds(parsed)
	}
	return c.writeCreds(oauthCreds{ClientID: clientID, ClientSecret: clientSecret})
}

// ImportCredentialsFile copies/parses a Google Cloud OAuth client JSON into place.
func (c *Client) ImportCredentialsFile(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	cr, err := parseCredsJSON(b)
	if err != nil {
		return err
	}
	return c.writeCreds(cr)
}

// ClearCredentials removes OAuth client file (keeps or drops token separately).
func (c *Client) ClearCredentials() error {
	_ = os.Remove(filepath.Join(c.dir, credsFile))
	return c.Disconnect()
}

// Disconnect drops the stored refresh/access tokens.
func (c *Client) Disconnect() error {
	return os.Remove(filepath.Join(c.dir, tokenFile))
}

// Connect runs the browser OAuth flow (loopback + PKCE) and stores tokens.
// Call CancelAuth() to abort while waiting on the browser.
func (c *Client) Connect(openURL func(string) error) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cr, err := c.loadCreds()
	if err != nil {
		return err
	}

	// cancel previous hang if any
	c.CancelAuth()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	c.authMu.Lock()
	c.authCancel = cancel
	c.authActive = true
	c.authMu.Unlock()
	defer func() {
		cancel()
		c.authMu.Lock()
		c.authCancel = nil
		c.authActive = false
		c.authMu.Unlock()
	}()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port
	redirect := fmt.Sprintf("http://127.0.0.1:%d/", port)

	verifier, err := randomString(64)
	if err != nil {
		return err
	}
	challenge := pkceS256(verifier)
	state, err := randomString(24)
	if err != nil {
		return err
	}

	q := url.Values{}
	q.Set("client_id", cr.ClientID)
	q.Set("redirect_uri", redirect)
	q.Set("response_type", "code")
	q.Set("scope", scopeDriveFile)
	q.Set("access_type", "offline")
	q.Set("prompt", "consent")
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	q.Set("state", state)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	srv := &http.Server{ReadHeaderTimeout: 10 * time.Second}
	srv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		if e := r.URL.Query().Get("error"); e != "" {
			msg := e
			if d := r.URL.Query().Get("error_description"); d != "" {
				msg = d
			}
			_, _ = io.WriteString(w, "<html><body><h2>Authorization failed</h2><p>"+htmlEscape(msg)+"</p><p>You can close this window.</p></body></html>")
			select {
			case errCh <- fmt.Errorf("oauth: %s", msg):
			default:
			}
			return
		}
		if r.URL.Query().Get("state") != state {
			http.Error(w, "invalid state", http.StatusBadRequest)
			select {
			case errCh <- fmt.Errorf("invalid oauth state"):
			default:
			}
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			select {
			case errCh <- fmt.Errorf("missing oauth code"):
			default:
			}
			return
		}
		_, _ = io.WriteString(w, "<html><body style=\"font-family:system-ui;padding:2rem\"><h2>Connected</h2><p>Return to NFA Tool. You can close this window.</p></body></html>")
		select {
		case codeCh <- code:
		default:
		}
	})
	go func() { _ = srv.Serve(ln) }()
	defer func() {
		shCtx, shCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer shCancel()
		_ = srv.Shutdown(shCtx)
	}()

	// stop listener immediately on cancel
	go func() {
		<-ctx.Done()
		_ = ln.Close()
		shCtx, shCancel := context.WithTimeout(context.Background(), time.Second)
		defer shCancel()
		_ = srv.Shutdown(shCtx)
	}()

	if openURL == nil {
		return fmt.Errorf("no browser opener")
	}
	if err := openURL(authURL + "?" + q.Encode()); err != nil {
		return fmt.Errorf("open browser: %w", err)
	}

	var code string
	select {
	case code = <-codeCh:
	case err = <-errCh:
		return err
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("authorization timed out — press Cancel and try again")
		}
		return fmt.Errorf("Cancelled")
	}

	// exchange can take a moment; still respect cancel
	type tokRes struct {
		t   *tokenData
		err error
	}
	done := make(chan tokRes, 1)
	go func() {
		t, e := exchangeCode(cr, code, verifier, redirect)
		done <- tokRes{t, e}
	}()
	select {
	case r := <-done:
		if r.err != nil {
			return r.err
		}
		return c.saveToken(r.t)
	case <-ctx.Done():
		return fmt.Errorf("Cancelled")
	}
}

// UploadText creates a new text file in the user's Drive (drive.file scope).
func (c *Client) UploadText(filename, content string) (webViewLink string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	access, err := c.validAccessToken()
	if err != nil {
		return "", err
	}
	filename = strings.TrimSpace(filename)
	if filename == "" {
		filename = "nfa-tokens.txt"
	}

	boundary := "nfa_" + strings.ReplaceAll(fmt.Sprintf("%d", time.Now().UnixNano()), "-", "")
	meta, _ := json.Marshal(map[string]string{
		"name":     filename,
		"mimeType": "text/plain",
	})
	var body strings.Builder
	body.WriteString("--" + boundary + "\r\n")
	body.WriteString("Content-Type: application/json; charset=UTF-8\r\n\r\n")
	body.Write(meta)
	body.WriteString("\r\n--" + boundary + "\r\n")
	body.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
	body.WriteString(content)
	if !strings.HasSuffix(content, "\n") {
		body.WriteString("\n")
	}
	body.WriteString("\r\n--" + boundary + "--")

	req, err := http.NewRequest(http.MethodPost, uploadURL, strings.NewReader(body.String()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+access)
	req.Header.Set("Content-Type", "multipart/related; boundary="+boundary)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("drive upload HTTP %d: %s", resp.StatusCode, truncate(string(raw), 300))
	}
	var out struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		WebViewLink string `json:"webViewLink"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	if out.WebViewLink == "" && out.ID != "" {
		out.WebViewLink = "https://drive.google.com/file/d/" + out.ID + "/view"
	}
	return out.WebViewLink, nil
}

func (c *Client) validAccessToken() (string, error) {
	tok, err := c.loadToken()
	if err != nil || tok == nil {
		return "", fmt.Errorf("not connected to Google Drive")
	}
	if tok.AccessToken != "" && time.Now().Before(tok.Expiry.Add(-60*time.Second)) {
		return tok.AccessToken, nil
	}
	if tok.RefreshToken == "" {
		return "", fmt.Errorf("Google session expired — connect again")
	}
	cr, err := c.loadCreds()
	if err != nil {
		return "", err
	}
	fresh, err := refreshAccess(cr, tok.RefreshToken)
	if err != nil {
		return "", err
	}
	if fresh.RefreshToken == "" {
		fresh.RefreshToken = tok.RefreshToken
	}
	if err := c.saveToken(fresh); err != nil {
		return "", err
	}
	return fresh.AccessToken, nil
}

func exchangeCode(cr oauthCreds, code, verifier, redirect string) (*tokenData, error) {
	form := url.Values{}
	form.Set("client_id", cr.ClientID)
	if cr.ClientSecret != "" {
		form.Set("client_secret", cr.ClientSecret)
	}
	form.Set("code", code)
	form.Set("code_verifier", verifier)
	form.Set("grant_type", "authorization_code")
	form.Set("redirect_uri", redirect)
	return postToken(form)
}

func refreshAccess(cr oauthCreds, refresh string) (*tokenData, error) {
	form := url.Values{}
	form.Set("client_id", cr.ClientID)
	if cr.ClientSecret != "" {
		form.Set("client_secret", cr.ClientSecret)
	}
	form.Set("refresh_token", refresh)
	form.Set("grant_type", "refresh_token")
	return postToken(form)
}

func postToken(form url.Values) (*tokenData, error) {
	resp, err := http.PostForm(tokenURL, form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("token HTTP %d: %s", resp.StatusCode, truncate(string(raw), 300))
	}
	var tr struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}
	if err := json.Unmarshal(raw, &tr); err != nil {
		return nil, err
	}
	if tr.Error != "" {
		return nil, fmt.Errorf("%s: %s", tr.Error, tr.ErrorDesc)
	}
	if tr.AccessToken == "" {
		return nil, fmt.Errorf("empty access_token")
	}
	exp := time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	if tr.ExpiresIn <= 0 {
		exp = time.Now().Add(55 * time.Minute)
	}
	return &tokenData{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		Expiry:       exp,
		TokenType:    tr.TokenType,
	}, nil
}

func (c *Client) loadCreds() (oauthCreds, error) {
	b, err := os.ReadFile(filepath.Join(c.dir, credsFile))
	if err != nil {
		return oauthCreds{}, fmt.Errorf("Google OAuth not configured — add Client ID (google-oauth.json)")
	}
	return parseCredsJSON(b)
}

func (c *Client) writeCreds(cr oauthCreds) error {
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(cr, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(c.dir, credsFile), b, 0o600)
}

func parseCredsJSON(b []byte) (oauthCreds, error) {
	var flat oauthCreds
	if err := json.Unmarshal(b, &flat); err == nil && flat.ClientID != "" {
		return flat, nil
	}
	var wrap struct {
		Installed *oauthCreds `json:"installed"`
		Web       *oauthCreds `json:"web"`
	}
	if err := json.Unmarshal(b, &wrap); err != nil {
		return oauthCreds{}, fmt.Errorf("invalid oauth json")
	}
	if wrap.Installed != nil && wrap.Installed.ClientID != "" {
		return *wrap.Installed, nil
	}
	if wrap.Web != nil && wrap.Web.ClientID != "" {
		return *wrap.Web, nil
	}
	return oauthCreds{}, fmt.Errorf("client_id not found in oauth json")
}

func (c *Client) loadToken() (*tokenData, error) {
	b, err := os.ReadFile(filepath.Join(c.dir, tokenFile))
	if err != nil {
		return nil, err
	}
	plain, err := storage.OpenString(strings.TrimSpace(string(b)))
	if err != nil {
		// try plain json (dev)
		var t tokenData
		if json.Unmarshal(b, &t) == nil && t.AccessToken != "" {
			return &t, nil
		}
		return nil, err
	}
	var t tokenData
	if err := json.Unmarshal([]byte(plain), &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (c *Client) saveToken(t *tokenData) error {
	if t == nil {
		return fmt.Errorf("nil token")
	}
	raw, err := json.Marshal(t)
	if err != nil {
		return err
	}
	sealed, err := storage.SealString(string(raw))
	if err != nil {
		// fallback plain (non-windows)
		sealed = string(raw)
	}
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(c.dir, tokenFile), []byte(sealed), 0o600)
}

func pkceS256(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func randomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func maskID(id string) string {
	id = strings.TrimSpace(id)
	if len(id) <= 12 {
		return id
	}
	return id[:6] + "…" + id[len(id)-4:]
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func htmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return r.Replace(s)
}
