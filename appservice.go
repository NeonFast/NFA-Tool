package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"nfa-tool/internal/steam"
	"nfa-tool/internal/storage"
	"nfa-tool/internal/token"
	"nfa-tool/internal/update"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// AppName / AppVersion shown in the UI title bar.
const AppName = "NFA Tool Recode v2"
const AppVersion = "2.0.3"

// GetAppName returns the product display name.
func (s *AppService) GetAppName() string {
	return AppName
}

// AppService is the Wails v3 backend service.
type AppService struct {
	store *storage.Store
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
	}
	return &AppService{store: store}
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
	if harvested, err := steam.HarvestConnectCache(); err == nil && len(harvested) > 0 {
		_ = s.store.Merge(harvested)
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
	if err := s.store.Delete(account); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "Account deleted"}
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
	if app := application.Get(); app != nil {
		app.Quit()
	}
}

// formatExpiryUntil returns a stable UTC timestamp for UI/i18n, e.g. 2026-09-15 14:30 UTC
func formatExpiryUntil(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	return t.UTC().Format("2006-01-02 15:04 UTC")
}
