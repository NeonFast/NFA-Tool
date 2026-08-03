package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"nfa-tool/internal/steam"
	"nfa-tool/internal/storage"
	"nfa-tool/internal/token"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const AppVersion = "1.0.0"

// App is the Wails backend.
type App struct {
	ctx   context.Context
	store *storage.Store
}

func NewApp() *App {
	base, err := os.Executable()
	if err != nil {
		base, _ = os.Getwd()
	} else {
		base = filepath.Dir(base)
	}
	// In `wails dev` exe lives in build/bin — keep backup next to project when possible.
	if _, err := os.Stat(filepath.Join(base, "go.mod")); err != nil {
		if wd, err := os.Getwd(); err == nil {
			if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
				base = wd
			}
		}
	}
	return &App{store: storage.New(base)}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetVersion returns app version string.
func (a *App) GetVersion() string {
	return AppVersion
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

// ListAccounts returns saved accounts with token validity.
func (a *App) ListAccounts() ([]AccountDTO, error) {
	// Harvest live steam cache into backup (best-effort)
	if harvested, err := steam.HarvestConnectCache(); err == nil && len(harvested) > 0 {
		_ = a.store.Merge(harvested)
	}

	m, err := a.store.Load()
	if err != nil {
		return nil, err
	}
	out := make([]AccountDTO, 0, len(m))
	for name, tok := range m {
		dto := AccountDTO{Name: name, Valid: false, ExpiresIn: "unknown"}
		if info, err := token.ParseAndValidate(tok); err == nil {
			dto.Valid = true
			dto.ExpiresIn = formatDuration(info.ExpiresIn)
		} else {
			dto.ExpiresIn = "expired/invalid"
		}
		out = append(out, dto)
	}
	return out, nil
}

// LoginFromKey parses account key and logs in.
func (a *App) LoginFromKey(accountKey string, keepExisting bool) Result {
	account, rawTok, err := token.ParseAccountKey(accountKey)
	if err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	info, err := token.ParseAndValidate(rawTok)
	if err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	if account == "" {
		return Result{OK: false, Message: "account name required (use login----token)"}
	}

	// Snapshot current cache before overwrite
	if harvested, err := steam.HarvestConnectCache(); err == nil {
		_ = a.store.Merge(harvested)
	}

	err = steam.Login(steam.LoginOptions{
		AccountName:  account,
		Token:        info.Raw,
		SteamID:      info.SteamID,
		KeepExisting: keepExisting,
	})
	if err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	_ = a.store.Put(account, info.Raw)
	return Result{
		OK:      true,
		Message: fmt.Sprintf("Logged in as %s · token valid %s", account, formatDuration(info.ExpiresIn)),
	}
}

// LoginSaved logs in with a previously saved account.
func (a *App) LoginSaved(account string, keepExisting bool) Result {
	m, err := a.store.Load()
	if err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	tok, ok := m[account]
	if !ok {
		return Result{OK: false, Message: "account not found"}
	}
	info, err := token.ParseAndValidate(tok)
	if err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	if harvested, err := steam.HarvestConnectCache(); err == nil {
		_ = a.store.Merge(harvested)
	}
	err = steam.Login(steam.LoginOptions{
		AccountName:  account,
		Token:        info.Raw,
		SteamID:      info.SteamID,
		KeepExisting: keepExisting,
	})
	if err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: fmt.Sprintf("Logged in as %s", account)}
}

// DeleteAccount removes a saved account.
func (a *App) DeleteAccount(account string) Result {
	if err := a.store.Delete(account); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "Account deleted"}
}

// ResetSteam wipes Steam config/userdata.
func (a *App) ResetSteam() Result {
	selection, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         "Reset Steam",
		Message:       "This will delete Steam config and userdata, then relaunch Steam. Continue?",
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
		CancelButton:  "No",
	})
	if err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	if selection != "Yes" {
		return Result{OK: false, Message: "Cancelled"}
	}
	if err := steam.ResetSteam(); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "Steam has been reset"}
}

// Minimise / close helpers for custom titlebar.
func (a *App) WindowMinimise() {
	runtime.WindowMinimise(a.ctx)
}

func (a *App) WindowClose() {
	runtime.Quit(a.ctx)
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		return "expired"
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}
