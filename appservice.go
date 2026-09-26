package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"nfa-tool/internal/gdrive"
	"nfa-tool/internal/steam"
	"nfa-tool/internal/storage"
	"nfa-tool/internal/token"
	"nfa-tool/internal/update"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// AppName is the product display name shown in the UI title bar.
const AppName = "NFA Tool"

// AppVersion is the single source of truth: main.go, GetVersion and the updater read it.
const AppVersion = "3.1.0"

// GetAppName returns the product display name.
func (s *AppService) GetAppName() string {
	return AppName
}

// AppService is the Wails v3 backend service.
type AppService struct {
	store  *storage.Store
	gdrive *gdrive.Client
	base   string

	bulkMu   sync.Mutex
	lastBulk []bulkExportEntry
}

type bulkExportEntry struct {
	line string
	ok   bool
}

// AccountDTO is exposed to the frontend.
type AccountDTO struct {
	Name      string `json:"name"`
	SteamID   string `json:"steamId,omitempty"`
	Persona   string `json:"persona,omitempty"`
	ExpiresIn string `json:"expiresIn"`
	Valid     bool   `json:"valid"`
	Avatar    string `json:"avatar,omitempty"`
}

// Result is a generic UI response.
type Result struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

// NewAppService creates the backend service rooted at the executable's directory.
func NewAppService() *AppService {
	base, err := os.Executable()
	if err != nil {
		base, _ = os.Getwd()
	} else {
		base = filepath.Dir(base)
	}
	if _, err := os.Stat(filepath.Join(base, "go.mod")); err != nil {
		if wd, err := os.Getwd(); err == nil {
			if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
				base = wd
			}
		}
	}
	store, err := storage.New(base)
	if err != nil {
		store, _ = storage.New(".")
		base = "."
	}
	initDiagLog(base)
	return &AppService{
		store:  store,
		gdrive: gdrive.New(base),
		base:   base,
	}
}

// GetVersion returns the current app version.
func (s *AppService) GetVersion() string {
	return AppVersion
}

// CheckForUpdates queries GitHub Releases for a newer version.
func (s *AppService) CheckForUpdates() update.Info {
	return update.CheckLatest(AppVersion)
}

// InstallUpdate downloads the release exe and restarts into it.
func (s *AppService) InstallUpdate(downloadURL string) Result {
	if err := update.ApplyDownload(downloadURL); err != nil {
		return s.fail(err.Error())
	}
	goSafe("update-quit", func() {
		time.Sleep(600 * time.Millisecond)
		if app := application.Get(); app != nil {
			app.Quit()
		} else {
			os.Exit(0)
		}
	})
	return s.ok("Update downloaded. Restarting…")
}

// OpenURL opens a link in the default browser.
func (s *AppService) OpenURL(url string) {
	_ = openBrowser(url)
}

var avatarHTTP = &http.Client{Timeout: 8 * time.Second}

func (s *AppService) avatarPath(steamID string) string {
	return filepath.Join(s.base, "avatars", steamID+".jpg")
}

func (s *AppService) cacheAvatar(steamID, url string) bool {
	if steamID == "" || url == "" {
		return false
	}
	p := s.avatarPath(steamID)
	if _, err := os.Stat(p); err == nil {
		return true
	}
	resp, err := avatarHTTP.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil || len(data) == 0 {
		return false
	}
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, data, 0o644); err != nil {
		return false
	}
	return true
}

func emitAvatarReady(name, steamID, avatar, persona string) {
	if avatar == "" && persona == "" {
		return
	}
	if app := application.Get(); app != nil {
		app.Event.Emit("avatar:ready", map[string]string{"name": name, "steamId": steamID, "avatar": avatar, "persona": persona})
	}
}

var personasMu sync.Mutex

func (s *AppService) personasPath() string {
	return filepath.Join(s.base, "avatars", "personas.json")
}

func (s *AppService) loadPersonas() map[string]string {
	m := map[string]string{}
	if data, err := os.ReadFile(s.personasPath()); err == nil {
		_ = json.Unmarshal(data, &m)
	}
	return m
}

// cachePersona stores the Steam persona name for an account next to its avatar.
func (s *AppService) cachePersona(steamID, persona string) {
	if steamID == "" || persona == "" {
		return
	}
	personasMu.Lock()
	defer personasMu.Unlock()
	m := s.loadPersonas()
	if m[steamID] == persona {
		return
	}
	m[steamID] = persona
	data, err := json.Marshal(m)
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(s.personasPath()), 0o755)
	_ = os.WriteFile(s.personasPath(), data, 0o644)
}

func (s *AppService) cachedPersona(steamID string) string {
	if steamID == "" {
		return ""
	}
	personasMu.Lock()
	defer personasMu.Unlock()
	return s.loadPersonas()[steamID]
}

// cacheAvatarAndEmit caches the avatar and notifies the frontend so list rows
// can swap the placeholder without waiting for a full refresh. When a prefetch
// run is active and still expects this avatar, it also advances the counter —
// this is how avatars fetched via the info dialog get counted.
func (s *AppService) cacheAvatarAndEmit(name, steamID, url, persona string) {
	if persona != "" {
		s.cachePersona(steamID, persona)
	}
	avatar := ""
	if s.cacheAvatar(steamID, url) {
		avatar = s.cachedAvatar(steamID)
		if done, total, ok := markAvatarDone(steamID); ok {
			emitAvatarProgress(done, total)
		}
	}
	emitAvatarReady(name, steamID, avatar, persona)
}

func emitAvatarProgress(done, total int) {
	if app := application.Get(); app != nil {
		app.Event.Emit("avatar:progress", map[string]int{"done": done, "total": total})
	}
}

var (
	avatarPrefetchRunning atomic.Bool
	avatarPrefetchMu      sync.Mutex
	avatarPrefetchPending map[string]bool
	avatarPrefetchDone    int
	avatarPrefetchTotal   int
)

// markAvatarDone counts a steamID towards the active prefetch run. It is
// idempotent per run and reports whether a progress event should be emitted.
func markAvatarDone(steamID string) (done, total int, ok bool) {
	avatarPrefetchMu.Lock()
	defer avatarPrefetchMu.Unlock()
	if !avatarPrefetchRunning.Load() || avatarPrefetchPending == nil {
		return 0, 0, false
	}
	if !avatarPrefetchPending[steamID] {
		return 0, 0, false
	}
	delete(avatarPrefetchPending, steamID)
	avatarPrefetchDone++
	return avatarPrefetchDone, avatarPrefetchTotal, true
}

// PrefetchAvatars slowly downloads missing avatars in the background,
// one profile request every few seconds to keep load negligible.
// Runs in its own goroutine; progress is streamed as avatar:progress events.
func (s *AppService) PrefetchAvatars() {
	if !avatarPrefetchRunning.CompareAndSwap(false, true) {
		return
	}
	goSafe("avatar-prefetch", func() {
		defer avatarPrefetchRunning.Store(false)
		m, err := s.store.Load()
		if err != nil {
			return
		}
		type entry struct {
			name    string
			steamID string
		}
		var pending []entry
		for name, tok := range m {
			info, err := token.ParseAndValidate(tok)
			if err != nil || info.SteamID == "" {
				continue
			}
			_, avatarErr := os.Stat(s.avatarPath(info.SteamID))
			if avatarErr == nil && s.cachedPersona(info.SteamID) != "" {
				continue
			}
			pending = append(pending, entry{name, info.SteamID})
		}
		sort.Slice(pending, func(i, j int) bool { return pending[i].name < pending[j].name })
		total := len(pending)
		if total == 0 {
			return
		}
		avatarPrefetchMu.Lock()
		avatarPrefetchPending = make(map[string]bool, total)
		for _, e := range pending {
			avatarPrefetchPending[e.steamID] = true
		}
		avatarPrefetchDone = 0
		avatarPrefetchTotal = total
		avatarPrefetchMu.Unlock()
		emitAvatarProgress(0, total)
		defer emitAvatarProgress(total, total)
		for i, e := range pending {
			if i > 0 {
				time.Sleep(3 * time.Second)
			}
			// An avatar cached meanwhile (info dialog / key check) skips the fetch
			// and has already been counted by cacheAvatarAndEmit.
			needAvatar := false
			if _, err := os.Stat(s.avatarPath(e.steamID)); err != nil {
				needAvatar = true
			}
			needPersona := s.cachedPersona(e.steamID) == ""
			if needAvatar || needPersona {
				avatarURL, persona := steam.FetchProfileBasics(e.steamID)
				if !needAvatar {
					avatarURL = ""
				}
				s.cacheAvatarAndEmit(e.name, e.steamID, avatarURL, persona)
			}
			if done, total, ok := markAvatarDone(e.steamID); ok {
				emitAvatarProgress(done, total)
			}
		}
	})
}

func (s *AppService) cachedAvatar(steamID string) string {
	if steamID == "" {
		return ""
	}
	data, err := os.ReadFile(s.avatarPath(steamID))
	if err != nil || len(data) == 0 || len(data) > 2<<20 {
		return ""
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(data)
}

func (s *AppService) dropAvatar(steamID string) {
	if steamID != "" {
		_ = os.Remove(s.avatarPath(steamID))
	}
}

func (s *AppService) dropAvatarFor(account string) {
	if tok, ok, err := s.store.Get(account); err == nil && ok {
		if ti, err := token.ParseAndValidate(tok); err == nil {
			s.dropAvatar(ti.SteamID)
		}
	}
}

// ListAccounts returns all stored accounts with validity, expiry and cached avatar.
func (s *AppService) ListAccounts() ([]AccountDTO, error) {
	if m0, err := s.store.Load(); err == nil && len(m0) == 0 {
		if harvested, err := steam.HarvestConnectCache(); err == nil && len(harvested) > 0 {
			_ = s.store.Merge(harvested)
		}
	}

	m, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	out := make([]AccountDTO, 0, len(m))
	for name, tok := range m {
		dto := AccountDTO{Name: name, Valid: false, ExpiresIn: "unknown"}
		if info, err := token.ParseAndValidate(tok); err == nil {
			dto.Valid = true
			dto.SteamID = info.SteamID
			dto.Persona = s.cachedPersona(info.SteamID)
			dto.ExpiresIn = formatExpiryUntil(info.ExpiresAt)
			dto.Avatar = s.cachedAvatar(info.SteamID)
		} else {
			dto.ExpiresIn = "expired/invalid"
		}
		out = append(out, dto)
	}
	return out, nil
}

func emitLoginStage(account, stage string) {
	if app := application.Get(); app != nil {
		app.Event.Emit("login:stage", map[string]string{"account": account, "stage": stage})
	}
}

func emitBulkItem(it steam.BulkCheckItem) {
	if app := application.Get(); app != nil {
		app.Event.Emit("bulk:item", it)
	}
}

func (s *AppService) rememberBulk(line string, ok bool) {
	if line == "" {
		return
	}
	s.bulkMu.Lock()
	s.lastBulk = append(s.lastBulk, bulkExportEntry{line: line, ok: ok})
	s.bulkMu.Unlock()
}

// bulkLines returns the stored lines of the last bulk check filtered by
// which ("ok" for working accounts, anything else for the rest).
func (s *AppService) bulkLines(which string) ([]string, error) {
	s.bulkMu.Lock()
	entries := append([]bulkExportEntry(nil), s.lastBulk...)
	s.bulkMu.Unlock()
	if len(entries) == 0 {
		return nil, fmt.Errorf("no batch check results to export")
	}
	wantOK := which == "ok"
	lines := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.ok == wantOK {
			lines = append(lines, e.line)
		}
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("nothing to export")
	}
	return lines, nil
}

// ExportBulkResults writes the lines of the last bulk check to a user-picked
// text file: which="ok" exports working accounts, anything else the rest.
// Returns the written file path, or an empty string if the dialog was
// cancelled.
func (s *AppService) ExportBulkResults(which string) (string, error) {
	lines, err := s.bulkLines(which)
	if err != nil {
		return "", err
	}
	wantOK := which == "ok"

	app := application.Get()
	if app == nil {
		return "", fmt.Errorf("dialog unavailable")
	}
	name := "failed-accounts.txt"
	if wantOK {
		name = "working-accounts.txt"
	}
	dlg := app.Dialog.SaveFile()
	if dlg == nil {
		return "", fmt.Errorf("dialog unavailable")
	}
	dlg.SetMessage("Export accounts")
	dlg.SetFilename(name)
	dlg.AddFilter("Text files", "*.txt")
	path, err := dlg.PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return "", nil
	}
	if !strings.HasSuffix(strings.ToLower(path), ".txt") {
		path += ".txt"
	}
	content := strings.Join(lines, "\r\n") + "\r\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func (s *AppService) resyncSteamTokens() {
	if harvested, err := steam.HarvestConnectCache(); err == nil && len(harvested) > 0 {
		_ = s.store.Merge(harvested)
	}
}

// LoginFromKey validates a pasted login----token key and logs into Steam with it.
func (s *AppService) LoginFromKey(accountKey string, keepExisting bool) Result {
	account, rawTok, err := token.ParseAccountKey(accountKey)
	if err != nil {
		return s.fail(err.Error())
	}
	info, err := token.ParseAndValidate(rawTok)
	if err != nil {
		return s.fail(err.Error())
	}
	if account == "" {
		return s.fail("account name required (use login----token)")
	}

	windowSeen, err := steam.Login(steam.LoginOptions{
		AccountName:  account,
		Token:        info.Raw,
		SteamID:      info.SteamID,
		KeepExisting: keepExisting,
		OnProgress:   func(stage string) { emitLoginStage(account, stage) },
	})
	if err != nil {
		return s.fail(err.Error())
	}
	_ = s.store.Put(account, info.Raw)
	s.resyncSteamTokens()
	msg := fmt.Sprintf("Logged in as %s · token valid until %s", account, formatExpiryUntil(info.ExpiresAt))
	if !windowSeen {
		msg += " · warning: steam.exe not detected after launch"
	}
	return s.ok(msg)
}

// SaveAccountKey validates and stores a pasted key without logging into Steam.
func (s *AppService) SaveAccountKey(accountKey string) Result {
	account, rawTok, err := token.ParseAccountKey(accountKey)
	if err != nil {
		return s.fail(err.Error())
	}
	info, err := token.ParseAndValidate(rawTok)
	if err != nil {
		return s.fail(err.Error())
	}
	if account == "" {
		return s.fail("account name required (use login----token)")
	}
	if err := s.store.Put(account, info.Raw); err != nil {
		return s.fail(err.Error())
	}
	return s.ok(fmt.Sprintf("Saved as %s · token valid until %s", account, formatExpiryUntil(info.ExpiresAt)))
}

// LoginSaved logs into Steam using a token already stored under account.
func (s *AppService) LoginSaved(account string, keepExisting bool) Result {	tok, ok, err := s.store.Get(account)
	if err != nil {
		return s.fail(err.Error())
	}
	if !ok {
		return s.fail("account not found")
	}
	info, err := token.ParseAndValidate(tok)
	if err != nil {
		return s.fail(err.Error())
	}

	windowSeen, err := steam.Login(steam.LoginOptions{
		AccountName:  account,
		Token:        info.Raw,
		SteamID:      info.SteamID,
		KeepExisting: keepExisting,
		OnProgress:   func(stage string) { emitLoginStage(account, stage) },
	})
	if err != nil {
		return s.fail(err.Error())
	}
	_ = s.store.Put(account, info.Raw)
	s.resyncSteamTokens()
	msg := fmt.Sprintf("Logged in as %s · token valid until %s", account, formatExpiryUntil(info.ExpiresAt))
	if !windowSeen {
		msg += " · warning: steam.exe not detected after launch"
	}
	return s.ok(msg)
}

// PickTextFile opens a file dialog and returns the file's text content
// (used by the bulk checker to load a .txt with keys).
func (s *AppService) PickTextFile() (string, error) {
	app := application.Get()
	if app == nil {
		return "", fmt.Errorf("dialog unavailable")
	}
	dlg := app.Dialog.OpenFile()
	if dlg == nil {
		return "", fmt.Errorf("dialog unavailable")
	}
	dlg.SetTitle("Open text file")
	dlg.AddFilter("Text files", "*.txt")
	dlg.AddFilter("All files", "*.*")
	path, err := dlg.PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return "", nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// BulkCheckResult is the outcome of a bulk token check.
type BulkCheckResult struct {
	Total int                  `json:"total"`
	Items []steam.BulkCheckItem `json:"items"`
}

func splitProxyLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if line = strings.TrimSpace(strings.TrimSuffix(line, "\r")); line != "" {
			out = append(out, line)
		}
	}
	return out
}

// CheckAccountKeys bulk-checks pasted keys (login----token per line),
// optionally distributing requests over the given proxies. Each finished item
// is streamed to the frontend as a "bulk:item" event.
func (s *AppService) CheckAccountKeys(input, proxies string) BulkCheckResult {
	entries, errs := token.ParseBulkKeys(input)
	items := make([]steam.BulkCheckItem, 0, len(entries)+len(errs))

	s.bulkMu.Lock()
	s.lastBulk = s.lastBulk[:0]
	s.bulkMu.Unlock()

	for _, e := range errs {
		it := steam.BulkCheckItem{Account: e, Status: "invalid"}
		items = append(items, it)
		s.rememberBulk(e, false)
		emitBulkItem(it)
	}
	onItem := func(it steam.BulkCheckItem, src token.ParsedKey) {
		s.rememberBulk(src.Account+"----"+src.Token, it.Status == "ok")
		emitBulkItem(it)
	}
	items = append(items, steam.CheckTokens(entries, splitProxyLines(proxies), onItem)...)
	return BulkCheckResult{Total: len(items), Items: items}
}

// CheckSavedAccounts bulk-checks every stored token.
func (s *AppService) CheckSavedAccounts(proxies string) BulkCheckResult {
	m, err := s.store.Load()
	if err != nil {
		return BulkCheckResult{}
	}
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Strings(names)
	entries := make([]token.ParsedKey, 0, len(m))
	for _, name := range names {
		entries = append(entries, token.ParsedKey{Account: name, Token: m[name]})
	}

	s.bulkMu.Lock()
	s.lastBulk = s.lastBulk[:0]
	s.bulkMu.Unlock()

	onItem := func(it steam.BulkCheckItem, src token.ParsedKey) {
		s.rememberBulk(src.Account+"----"+src.Token, it.Status == "ok")
		emitBulkItem(it)
	}
	items := steam.CheckTokens(entries, splitProxyLines(proxies), onItem)
	return BulkCheckResult{Total: len(items), Items: items}
}

// HarvestSteamAccounts re-reads Steam's local session files (ConnectCache +
// loginusers.vdf) and merges found accounts into the store.
func (s *AppService) HarvestSteamAccounts() Result {
	harvested, err := steam.HarvestConnectCache()
	if err != nil {
		return s.fail(err.Error())
	}
	if len(harvested) == 0 {
		return s.fail("no accounts found in Steam files")
	}
	if err := s.store.Merge(harvested); err != nil {
		return s.fail(err.Error())
	}
	return s.ok(fmt.Sprintf("Harvested %d account(s)", len(harvested)))
}

// Notify shows a native Windows message box. Frontend should pass already-translated title/message.
func (s *AppService) Notify(success bool, title, message string) {
	app := application.Get()
	if app == nil {
		return
	}
	if title == "" {
		if success {
			title = "Success"
		} else {
			title = "Error"
		}
	}
	var dlg *application.MessageDialog
	if success {
		dlg = app.Dialog.Info().SetTitle(title).SetMessage(message)
	} else {
		dlg = app.Dialog.Error().SetTitle(title).SetMessage(message)
	}
	dlg.AddButton("OK")
	dlg.Show()
}

func (s *AppService) ok(msg string) Result {
	return Result{OK: true, Message: msg}
}

func (s *AppService) fail(msg string) Result {
	diagf("ошибка: %s", msg)
	return Result{OK: false, Message: msg}
}

// SystemStatus is the diagnostics snapshot for the Logs tab.
type SystemStatus struct {
	Version        string `json:"version"`
	SteamRunning   bool   `json:"steamRunning"`
	SteamPath      string `json:"steamPath"`
	AccountsTotal  int    `json:"accountsTotal"`
	AccountsValid  int    `json:"accountsValid"`
	DriveConnected bool   `json:"driveConnected"`
	SteamAPIOnline bool   `json:"steamApiOnline"`
}

// GetSystemStatus collects local + network diagnostics (no proxy).
func (s *AppService) GetSystemStatus() SystemStatus {
	st := SystemStatus{Version: AppVersion, SteamRunning: steam.IsSteamRunning()}
	if p, err := steam.GetSteamInstallPathNoKill(); err == nil {
		st.SteamPath = p
	}
	if m, err := s.store.Load(); err == nil {
		st.AccountsTotal = len(m)
		for _, tok := range m {
			if _, err := token.ParseAndValidate(tok); err == nil {
				st.AccountsValid++
			}
		}
	}
	if s.gdrive != nil {
		st.DriveConnected = s.gdrive.GetStatus().Connected
	}
	st.SteamAPIOnline = steam.APIReachable()
	return st
}

// CheckAccountKey runs the full info check for an arbitrary pasted key
// (login----token or bare token), without saving it.
func (s *AppService) CheckAccountKey(accountKey string) (steam.AccountInfo, error) {
	_, rawTok, err := token.ParseAccountKey(accountKey)
	if err != nil {
		return steam.AccountInfo{}, err
	}
	info, err := token.ParseAndValidate(rawTok)
	if err != nil {
		return steam.AccountInfo{}, err
	}
	out := steam.FetchAccountInfo(info.Raw, info.SteamID)
	if out.AvatarFull != "" {
		goSafe("avatar-cache", func() { s.cacheAvatarAndEmit("", out.SteamID, out.AvatarFull, out.PersonaName) })
	} else if out.PersonaName != "" {
		s.cachePersona(out.SteamID, out.PersonaName)
	}
	return *out, nil
}

// GetAccountInfo collects public account details using the saved token.
func (s *AppService) GetAccountInfo(account string) (steam.AccountInfo, error) {
	tok, ok, err := s.store.Get(account)
	if err != nil {
		return steam.AccountInfo{}, err
	}
	if !ok {
		return steam.AccountInfo{}, fmt.Errorf("account not found")
	}
	info, err := token.ParseAndValidate(tok)
	if err != nil {
		return steam.AccountInfo{}, err
	}
	out := steam.FetchAccountInfo(info.Raw, info.SteamID)
	if out.AvatarFull != "" {
		goSafe("avatar-cache", func() { s.cacheAvatarAndEmit(account, out.SteamID, out.AvatarFull, out.PersonaName) })
	} else if out.PersonaName != "" {
		s.cachePersona(out.SteamID, out.PersonaName)
	}
	return *out, nil
}

// DeleteAccount removes a saved account and its cached avatar.
func (s *AppService) DeleteAccount(account string) Result {
	account = strings.TrimSpace(account)
	if account == "" {
		return s.fail("account name required")
	}
	s.dropAvatarFor(account)
	if err := s.store.Delete(account); err != nil {
		return s.fail(err.Error())
	}
	return s.ok("Account deleted")
}

// DeleteAccounts removes multiple saved accounts from the local DB.
func (s *AppService) DeleteAccounts(names []string) Result {
	okN := 0
	var lastErr string
	seen := map[string]bool{}
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		key := strings.ToLower(n)
		if seen[key] {
			continue
		}
		seen[key] = true
		s.dropAvatarFor(n)
		if err := s.store.Delete(n); err != nil {
			lastErr = err.Error()
			continue
		}
		okN++
	}
	if okN == 0 {
		if lastErr != "" {
			return s.fail(lastErr)
		}
		return s.fail("no accounts to delete")
	}
	msg := fmt.Sprintf("Deleted %d account(s)", okN)
	if lastErr != "" {
		msg += " · some failed"
	}
	return s.ok(msg)
}

// ImportTokens imports multi-line login----token text into the local store (no Steam login).
func (s *AppService) ImportTokens(text string) Result {
	keys, errs := token.ParseBulkKeys(text)
	if len(keys) == 0 {
		if len(errs) > 0 {
			return s.fail(fmt.Sprintf("import failed: %s", errs[0]))
		}
		return s.fail("no accounts to import")
	}
	okN := 0
	for _, k := range keys {
		if err := s.store.Put(k.Account, k.Token); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", k.Account, err))
			continue
		}
		okN++
	}
	if okN == 0 {
		msg := "import failed"
		if len(errs) > 0 {
			msg = errs[0]
		}
		return s.fail(msg)
	}
	msg := fmt.Sprintf("Imported %d account(s)", okN)
	if len(errs) > 0 {
		msg += fmt.Sprintf(" · %d skipped", len(errs))
	}
	return s.ok(msg)
}

// ImportTokensFromFile opens a .txt file and imports login----token lines.
func (s *AppService) ImportTokensFromFile() Result {
	app := application.Get()
	if app == nil {
		return s.fail("dialog unavailable")
	}
	dlg := app.Dialog.OpenFile()
	if dlg == nil {
		return s.fail("dialog unavailable")
	}
	dlg.SetTitle("Import tokens")
	dlg.AddFilter("Text files", "*.txt")
	dlg.AddFilter("All files", "*.*")
	path, err := dlg.PromptForSingleSelection()
	if err != nil {
		return s.fail(err.Error())
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return s.fail("Cancelled")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return s.fail(err.Error())
	}
	return s.ImportTokens(string(raw))
}

// ExportTokens returns login----token lines for the given names (empty = all).
func (s *AppService) ExportTokens(names []string) (string, error) {
	text, n, err := s.buildExport(names)
	if err != nil {
		return "", err
	}
	if n == 0 {
		return "", fmt.Errorf("no accounts to export")
	}
	return text, nil
}

// GoogleDriveStatus returns OAuth setup / connection state.
func (s *AppService) GoogleDriveStatus() gdrive.Status {
	if s.gdrive == nil {
		return gdrive.Status{}
	}
	return s.gdrive.GetStatus()
}

// SaveGoogleCredentials stores Desktop OAuth client id + secret.
func (s *AppService) SaveGoogleCredentials(clientID, clientSecret string) Result {
	if s.gdrive == nil {
		return s.fail("google drive unavailable")
	}
	if err := s.gdrive.SaveCredentials(clientID, clientSecret); err != nil {
		return s.fail(err.Error())
	}
	return s.ok("Google OAuth saved")
}

// ImportGoogleCredentials opens a file picker for google-oauth.json from Cloud Console.
func (s *AppService) ImportGoogleCredentials() Result {
	if s.gdrive == nil {
		return s.fail("google drive unavailable")
	}
	app := application.Get()
	if app == nil {
		return s.fail("dialog unavailable")
	}
	dlg := app.Dialog.OpenFile()
	if dlg == nil {
		return s.fail("dialog unavailable")
	}
	dlg.SetTitle("Select google-oauth.json")
	dlg.AddFilter("JSON", "*.json")
	path, err := dlg.PromptForSingleSelection()
	if err != nil {
		return s.fail(err.Error())
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return s.fail("Cancelled")
	}
	if err := s.gdrive.ImportCredentialsFile(path); err != nil {
		return s.fail(err.Error())
	}
	return s.ok("Google OAuth imported")
}

// ConnectGoogleDrive runs browser OAuth (drive.file) and stores the refresh token.
func (s *AppService) ConnectGoogleDrive() Result {
	if s.gdrive == nil {
		return s.fail("google drive unavailable")
	}
	if err := s.gdrive.Connect(openBrowser); err != nil {
		return s.fail(err.Error())
	}
	return s.ok("Google Drive connected")
}

// CancelGoogleAuth aborts waiting for Google browser login (unblocks the UI).
func (s *AppService) CancelGoogleAuth() Result {
	if s.gdrive != nil {
		s.gdrive.CancelAuth()
	}
	return s.ok("Cancelled")
}

// GoogleAuthBusy is true while Connect waits on the browser.
func (s *AppService) GoogleAuthBusy() bool {
	if s.gdrive == nil {
		return false
	}
	return s.gdrive.AuthInProgress()
}

// DisconnectGoogleDrive removes the stored Google session.
func (s *AppService) DisconnectGoogleDrive() Result {
	if s.gdrive == nil {
		return s.fail("google drive unavailable")
	}
	s.gdrive.CancelAuth()
	_ = s.gdrive.Disconnect()
	return s.ok("Google Drive disconnected")
}

// uploadToDrive uploads text as a file to Google Drive, connecting first if
// needed and retrying once on an expired token.
func (s *AppService) uploadToDrive(name, text, okMsg string) Result {
	if s.gdrive == nil {
		return s.fail("google drive unavailable")
	}
	st := s.gdrive.GetStatus()
	if !st.HasCredentials {
		return s.fail("Google OAuth not configured")
	}
	if !st.Connected {
		if err := s.gdrive.Connect(openBrowser); err != nil {
			return s.fail(err.Error())
		}
	}
	link, err := s.gdrive.UploadText(name, text)
	if err != nil {
		if strings.Contains(err.Error(), "expired") || strings.Contains(err.Error(), "invalid_grant") {
			if err2 := s.gdrive.Connect(openBrowser); err2 == nil {
				link, err = s.gdrive.UploadText(name, text)
			}
		}
		if err != nil {
			return s.fail(err.Error())
		}
	}
	if link != "" {
		okMsg += ": " + link
		_ = openBrowser(link)
	}
	return s.ok(okMsg)
}

// ExportTokensToGoogleDrive uploads login----token text for selected (or all) accounts.
func (s *AppService) ExportTokensToGoogleDrive(names []string) Result {
	text, n, err := s.buildExport(names)
	if err != nil {
		return s.fail(err.Error())
	}
	if n == 0 {
		return s.fail("no accounts to export")
	}
	name := fmt.Sprintf("nfa-tokens-%s.txt", time.Now().Format("2006-01-02_150405"))
	return s.uploadToDrive(name, text, fmt.Sprintf("Uploaded %d account(s) to Google Drive", n))
}

// ExportBulkResultsToGoogleDrive uploads the lines of the last bulk check to
// Google Drive: which="ok" uploads working accounts, anything else the rest.
func (s *AppService) ExportBulkResultsToGoogleDrive(which string) Result {
	lines, err := s.bulkLines(which)
	if err != nil {
		return s.fail(err.Error())
	}
	name := fmt.Sprintf("nfa-batch-failed-%s.txt", time.Now().Format("2006-01-02_150405"))
	if which == "ok" {
		name = fmt.Sprintf("nfa-batch-working-%s.txt", time.Now().Format("2006-01-02_150405"))
	}
	content := strings.Join(lines, "\r\n") + "\r\n"
	return s.uploadToDrive(name, content, fmt.Sprintf("Uploaded %d account(s) to Google Drive", len(lines)))
}

// ExportTokensToFile writes selected (or all) tokens to a user-chosen .txt file.
func (s *AppService) ExportTokensToFile(names []string) Result {
	text, n, err := s.buildExport(names)
	if err != nil {
		return s.fail(err.Error())
	}
	if n == 0 {
		return s.fail("no accounts to export")
	}

	app := application.Get()
	path := ""
	dialogShown := false
	if app != nil {
		if dlg := app.Dialog.SaveFile(); dlg != nil {
			dialogShown = true
			dlg.SetFilename("nfa-tokens.txt")
			dlg.AddFilter("Text files", "*.txt")
			p, err := dlg.PromptForSingleSelection()
			if err != nil {
				return s.fail(err.Error())
			}
			path = strings.TrimSpace(p)
			if path == "" {
				return s.fail("Cancelled")
			}
		}
	}
	if path == "" && !dialogShown {
		base := filepath.Dir(s.store.Path())
		path = filepath.Join(base, fmt.Sprintf("nfa-tokens-%d.txt", time.Now().Unix()))
	}
	if path == "" {
		return s.fail("Cancelled")
	}
	if !strings.HasSuffix(strings.ToLower(path), ".txt") {
		path += ".txt"
	}
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		return s.fail(err.Error())
	}
	return s.ok(fmt.Sprintf("Exported %d account(s) to %s", n, path))
}

func (s *AppService) buildExport(names []string) (string, int, error) {
	all, err := s.store.Load()
	if err != nil {
		return "", 0, err
	}
	want := map[string]bool{}
	if len(names) == 0 {
		for k := range all {
			want[strings.ToLower(k)] = true
		}
	} else {
		for _, n := range names {
			n = strings.ToLower(strings.TrimSpace(n))
			if n != "" {
				want[n] = true
			}
		}
	}
	var lines []string
	keys := make([]string, 0, len(all))
	for k := range all {
		keys = append(keys, k)
	}
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if strings.ToLower(keys[j]) < strings.ToLower(keys[i]) {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	for _, name := range keys {
		if !want[strings.ToLower(name)] {
			continue
		}
		tok := strings.TrimSpace(all[name])
		if tok == "" {
			continue
		}
		lines = append(lines, name+"----"+tok)
	}
	return strings.Join(lines, "\n") + "\n", len(lines), nil
}

// ResetSteam asks for confirmation, then wipes Steam config/userdata and relaunches it.
func (s *AppService) ResetSteam() Result {
	app := application.Get()
	if app == nil {
		return Result{OK: false, Message: "app not ready"}
	}

	confirmed := make(chan bool, 1)
	var once sync.Once
	done := func(v bool) {
		once.Do(func() { confirmed <- v })
	}

	dlg := app.Dialog.Question().
		SetTitle("Reset Steam").
		SetMessage("This will delete Steam config and userdata, then relaunch Steam. Continue?")
	yes := dlg.AddButton("Yes").OnClick(func() { done(true) })
	no := dlg.AddButton("No").OnClick(func() { done(false) })
	dlg.SetDefaultButton(no)
	dlg.SetCancelButton(no)
	_ = yes
	dlg.Show()

	if !<-confirmed {
		return Result{OK: false, Message: "Cancelled"}
	}
	if err := steam.ResetSteam(); err != nil {
		return s.fail(err.Error())
	}
	return s.ok("Steam has been reset")
}

// WindowMinimise minimises the current window.
func (s *AppService) WindowMinimise() {
	if app := application.Get(); app != nil {
		if w := app.Window.Current(); w != nil {
			w.Minimise()
		}
	}
}

// WindowClose cancels pending auth, closes the guide window and quits the app.
func (s *AppService) WindowClose() {
	if s.gdrive != nil {
		s.gdrive.CancelAuth()
	}
	if app := application.Get(); app != nil {
		if w, ok := app.Window.GetByName("drive-guide"); ok && w != nil {
			w.Close()
		}
		app.Quit()
	}
}

// OpenDriveGuide opens (or focuses) the Google Drive setup guide window.
func (s *AppService) OpenDriveGuide() {
	app := application.Get()
	if app == nil {
		return
	}
	if w, ok := app.Window.GetByName("drive-guide"); ok && w != nil {
		w.Show()
		w.Focus()
		return
	}
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "drive-guide",
		Title:            "Google Drive — setup",
		Width:            720,
		Height:           640,
		MinWidth:         720,
		MinHeight:        640,
		MaxWidth:         720,
		MaxHeight:        640,
		Frameless:        true,
		BackgroundColour: application.NewRGB(15, 15, 26),
		URL:              "/?page=drive-guide",
		DisableResize:    true,
	})
}

// CloseDriveGuide closes the guide window only (main app stays open).
func (s *AppService) CloseDriveGuide() {
	app := application.Get()
	if app == nil {
		return
	}
	if w, ok := app.Window.GetByName("drive-guide"); ok && w != nil {
		w.Close()
	}
}

func formatExpiryUntil(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	return t.UTC().Format("2006-01-02 15:04 UTC")
}
