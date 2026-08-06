package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"nfa-tool/internal/gdrive"
	"nfa-tool/internal/steam"
	"nfa-tool/internal/storage"
	"nfa-tool/internal/token"
	"nfa-tool/internal/update"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// AppName / AppVersion shown in the UI title bar.
const AppName = "NFA Tool Recode v2"
const AppVersion = "2.1.1"

// GetAppName returns the product display name.
func (s *AppService) GetAppName() string {
	return AppName
}

// AppService is the Wails v3 backend service.
type AppService struct {
	store  *storage.Store
	gdrive *gdrive.Client
}

// AccountDTO is exposed to the frontend.
type AccountDTO struct {
	Name      string `json:"name"`
	ExpiresIn string `json:"expiresIn"`
	Valid     bool   `json:"valid"`
}

// Result is a generic UI response.
type Result struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

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
		// Fall back to cwd so the app still starts; operations will fail with clear errors if store is nil-checked.
		// storage.New rarely fails; panic is worse for a GUI app.
		store, _ = storage.New(".")
		base = "."
	}
	return &AppService{
		store:  store,
		gdrive: gdrive.New(base),
	}
}

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
	// quit so the bat can replace the file
	go func() {
		time.Sleep(600 * time.Millisecond)
		if app := application.Get(); app != nil {
			app.Quit()
		} else {
			os.Exit(0)
		}
	}()
	return s.ok("Update downloaded. Restarting…")
}

// OpenURL opens a link in the default browser.
func (s *AppService) OpenURL(url string) {
	_ = openBrowser(url)
}

func (s *AppService) ListAccounts() ([]AccountDTO, error) {
	// Do NOT harvest Steam ConnectCache here — that re-imported deleted accounts
	// after every refresh. Harvest only once when the local DB is empty.
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
			dto.ExpiresIn = formatExpiryUntil(info.ExpiresAt)
		} else {
			dto.ExpiresIn = "expired/invalid"
		}
		out = append(out, dto)
	}
	return out, nil
}

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

	err = steam.Login(steam.LoginOptions{
		AccountName:  account,
		Token:        info.Raw,
		SteamID:      info.SteamID,
		KeepExisting: keepExisting,
	})
	if err != nil {
		return s.fail(err.Error())
	}
	_ = s.store.Put(account, info.Raw)
	msg := fmt.Sprintf("Logged in as %s · token valid until %s", account, formatExpiryUntil(info.ExpiresAt))
	if !steam.IsSteamRunning() {
		msg += " · warning: steam.exe not detected after launch"
	}
	return s.ok(msg)
}

func (s *AppService) LoginSaved(account string, keepExisting bool) Result {
	tok, ok, err := s.store.Get(account)
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

	err = steam.Login(steam.LoginOptions{
		AccountName:  account,
		Token:        info.Raw,
		SteamID:      info.SteamID,
		KeepExisting: keepExisting,
	})
	if err != nil {
		return s.fail(err.Error())
	}
	_ = s.store.Put(account, info.Raw)
	msg := fmt.Sprintf("Logged in as %s · token valid until %s", account, formatExpiryUntil(info.ExpiresAt))
	if !steam.IsSteamRunning() {
		msg += " · warning: steam.exe not detected after launch"
	}
	return s.ok(msg)
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
	return Result{OK: false, Message: msg}
}

func (s *AppService) DeleteAccount(account string) Result {
	account = strings.TrimSpace(account)
	if account == "" {
		return s.fail("account name required")
	}
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

// ExportTokensToGoogleDrive uploads login----token text for selected (or all) accounts.
func (s *AppService) ExportTokensToGoogleDrive(names []string) Result {
	if s.gdrive == nil {
		return s.fail("google drive unavailable")
	}
	text, n, err := s.buildExport(names)
	if err != nil {
		return s.fail(err.Error())
	}
	if n == 0 {
		return s.fail("no accounts to export")
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
	name := fmt.Sprintf("nfa-tokens-%s.txt", time.Now().Format("2006-01-02_150405"))
	link, err := s.gdrive.UploadText(name, text)
	if err != nil {
		// one reconnect attempt
		if strings.Contains(err.Error(), "expired") || strings.Contains(err.Error(), "invalid_grant") {
			if err2 := s.gdrive.Connect(openBrowser); err2 == nil {
				link, err = s.gdrive.UploadText(name, text)
			}
		}
		if err != nil {
			return s.fail(err.Error())
		}
	}
	msg := fmt.Sprintf("Uploaded %d account(s) to Google Drive", n)
	if link != "" {
		msg += ": " + link
		_ = openBrowser(link)
	}
	return s.ok(msg)
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
	// case-insensitive stable-ish order
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

func (s *AppService) WindowMinimise() {
	if app := application.Get(); app != nil {
		if w := app.Window.Current(); w != nil {
			w.Minimise()
		}
	}
}

func (s *AppService) WindowClose() {
	if s.gdrive != nil {
		s.gdrive.CancelAuth()
	}
	if app := application.Get(); app != nil {
		// close guide window first if open
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

// formatExpiryUntil returns a stable UTC timestamp for UI/i18n, e.g. 2026-09-15 14:30 UTC
func formatExpiryUntil(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	return t.UTC().Format("2006-01-02 15:04 UTC")
}
