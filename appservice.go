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

	"github.com/wailsapp/wails/v3/pkg/application"
)

// AppVersion is shown in the UI title bar.
const AppVersion = "1.0.0"

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
	return &AppService{store: storage.New(base)}
}

func (s *AppService) GetVersion() string {
	return AppVersion
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
			dto.ExpiresIn = formatDuration(info.ExpiresIn)
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
		return Result{OK: false, Message: err.Error()}
	}
	info, err := token.ParseAndValidate(rawTok)
	if err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	if account == "" {
		return Result{OK: false, Message: "account name required (use login----token)"}
	}

	if harvested, err := steam.HarvestConnectCache(); err == nil {
		_ = s.store.Merge(harvested)
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
	_ = s.store.Put(account, info.Raw)
	return Result{
		OK:      true,
		Message: fmt.Sprintf("Logged in as %s · token valid %s", account, formatDuration(info.ExpiresIn)),
	}
}

func (s *AppService) LoginSaved(account string, keepExisting bool) Result {
	m, err := s.store.Load()
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
		_ = s.store.Merge(harvested)
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
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "Steam has been reset"}
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
