package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"

	"nfa-tool/internal/steam"

	"github.com/wailsapp/wails/v3/pkg/application"
)

var (
	diagMu   sync.Mutex
	diagFile *os.File
	diagPath string
)

// initDiagLog opens the append-only diagnostics log next to the executable
// (logs/app.log) and routes fatal runtime crashes into it. Keeps one rotated
// backup (app.prev.log) when the current file grows past 4 MB.
func initDiagLog(base string) {
	dir := filepath.Join(base, "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	p := filepath.Join(dir, "app.log")
	if st, err := os.Stat(p); err == nil && st.Size() > 4<<20 {
		_ = os.Rename(p, filepath.Join(dir, "app.prev.log"))
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	diagFile = f
	diagPath = p
	_ = debug.SetCrashOutput(f, debug.CrashOptions{})
	pruneCrashDumps()
	diagf("=== запуск %s v%s (%s/%s, %s) ===", AppName, AppVersion, runtime.GOOS, runtime.GOARCH, runtime.Version())
}

// crashpadDir resolves the WebView2 Crashpad reports directory for this exe.
func crashpadDir() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	roaming := os.Getenv("APPDATA")
	if roaming == "" {
		return ""
	}
	return filepath.Join(roaming, filepath.Base(exe), "EBWebView", "Crashpad", "reports")
}

// pruneCrashDumps keeps only the two newest renderer crash dumps — each dump
// is ~10 MB and they pile up in the WebView2 profile forever otherwise.
func pruneCrashDumps() {
	dir := crashpadDir()
	if dir == "" {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) <= 2 {
		return
	}
	type dump struct {
		name string
		mt   time.Time
	}
	var dumps []dump
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".dmp") {
			continue
		}
		if info, err := e.Info(); err == nil {
			dumps = append(dumps, dump{e.Name(), info.ModTime()})
		}
	}
	sort.Slice(dumps, func(i, j int) bool { return dumps[i].mt.After(dumps[j].mt) })
	removed := 0
	for _, d := range dumps[2:] {
		if os.Remove(filepath.Join(dir, d.name)) == nil {
			removed++
		}
	}
	if removed > 0 {
		diagf("crashpad: удалено старых дампов: %d", removed)
	}
}

// diagf appends one timestamped line to the diagnostics log.
func diagf(format string, args ...any) {
	diagMu.Lock()
	defer diagMu.Unlock()
	if diagFile == nil {
		return
	}
	fmt.Fprintf(diagFile, "%s %s\n", time.Now().Format("2006-01-02 15:04:05.000"), fmt.Sprintf(format, args...))
}

// goSafe runs fn in a goroutine, logging panics instead of killing the process.
func goSafe(label string, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				diagf("ПАНИКА [%s]: %v\n%s", label, r, debug.Stack())
			}
		}()
		fn()
	}()
}

// diagFlush forces buffered crash output to disk (best effort).
func diagFlush() {
	diagMu.Lock()
	defer diagMu.Unlock()
	if diagFile != nil {
		_ = diagFile.Sync()
	}
}

// tailLog reads the last max bytes of the diagnostics log.
func tailLog(max int64) string {
	if diagPath == "" {
		return "(лог недоступен)"
	}
	diagFlush()
	st, err := os.Stat(diagPath)
	if err != nil {
		return "(лог недоступен)"
	}
	off := int64(0)
	if st.Size() > max {
		off = st.Size() - max
	}
	f, err := os.Open(diagPath)
	if err != nil {
		return "(лог недоступен)"
	}
	defer f.Close()
	if _, err := f.Seek(off, 0); err != nil {
		return "(лог недоступен)"
	}
	buf := make([]byte, st.Size()-off)
	n, _ := f.Read(buf)
	return string(buf[:n])
}

// collectSysInfo builds the hardware/environment header for the dump.
func collectSysInfo() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s v%s — дамп диагностики\n", AppName, AppVersion)
	fmt.Fprintf(&b, "Дата: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&b, "Платформа: %s/%s, Go %s\n", runtime.GOOS, runtime.GOARCH, runtime.Version())
	fmt.Fprintf(&b, "CPU: %d потоков\n", runtime.NumCPU())
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Fprintf(&b, "Процесс: heap %d МБ, горутин %d\n", m.Alloc>>20, runtime.NumGoroutine())
	b.WriteString(platformSysInfo())
	return b.String()
}

// SaveDiagnosticsDump lets the user save a single text file with system info,
// the recent UI journal and the tail of the diagnostics log.
func (s *AppService) SaveDiagnosticsDump(journal string) Result {
	app := application.Get()
	if app == nil {
		return s.fail("dialog unavailable")
	}
	dlg := app.Dialog.SaveFile()
	if dlg == nil {
		return s.fail("dialog unavailable")
	}
	dlg.SetMessage("Save diagnostics dump")
	dlg.SetFilename(fmt.Sprintf("nfa-diagnostics-%s.txt", time.Now().Format("2006-01-02_150405")))
	dlg.AddFilter("Text files", "*.txt")
	path, err := dlg.PromptForSingleSelection()
	if err != nil {
		return s.fail(err.Error())
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return s.fail("cancelled")
	}
	if !strings.HasSuffix(strings.ToLower(path), ".txt") {
		path += ".txt"
	}

	var b strings.Builder
	b.WriteString(collectSysInfo())
	if m, err := s.store.Load(); err == nil {
		fmt.Fprintf(&b, "Сохранённых аккаунтов: %d\n", len(m))
	}
	if s.gdrive != nil {
		st := s.gdrive.GetStatus()
		fmt.Fprintf(&b, "Google Drive: credentials=%v connected=%v\n", st.HasCredentials, st.Connected)
	}
	if p, err := steam.GetSteamInstallPathNoKill(); err == nil {
		fmt.Fprintf(&b, "Steam: %s\n", p)
	} else {
		fmt.Fprintf(&b, "Steam: не найден (%v)\n", err)
	}
	if dir := crashpadDir(); dir != "" {
		if entries, err := os.ReadDir(dir); err == nil {
			var dumps []string
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".dmp") {
					if info, err := e.Info(); err == nil {
						dumps = append(dumps, fmt.Sprintf("%s (%s, %.1f МБ)", e.Name(), info.ModTime().Format("2006-01-02 15:04"), float64(info.Size())/1048576))
					}
				}
			}
			if len(dumps) > 0 {
				fmt.Fprintf(&b, "Краш-дампы рендерера (%s):\n  %s\n", dir, strings.Join(dumps, "\n  "))
			} else {
				b.WriteString("Краш-дампов рендерера нет\n")
			}
		}
	}
	b.WriteString("\n=== Журнал интерфейса (последние записи) ===\n")
	if strings.TrimSpace(journal) == "" {
		b.WriteString("(пусто)\n")
	} else {
		b.WriteString(journal)
		b.WriteString("\n")
	}
	b.WriteString("\n=== Лог приложения (хвост logs/app.log) ===\n")
	b.WriteString(tailLog(64 << 10))
	b.WriteString("\n")

	diagf("дамп диагностики сохранён: %s", path)
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return s.fail(err.Error())
	}
	return s.ok("Diagnostics saved")
}

// LogFrontendError records a JS error/unhandled rejection from the UI.
func (s *AppService) LogFrontendError(msg string) {
	diagf("ОШИБКА UI: %s", msg)
}
