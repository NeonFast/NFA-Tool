package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	GitHubOwner = "NeonFast"
	GitHubRepo  = "NFA-Tool"
	UserAgent   = "NFA-Tool-Recode-v2"
)

// Info is returned to the UI.
type Info struct {
	UpdateAvailable bool   `json:"updateAvailable"`
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion"`
	ReleaseURL      string `json:"releaseUrl"`
	DownloadURL     string `json:"downloadUrl"`
	ReleaseNotes    string `json:"releaseNotes"`
	Error           string `json:"error,omitempty"`
}

type ghRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Body    string `json:"body"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// CheckLatest queries GitHub Releases API.
func CheckLatest(current string) Info {
	info := Info{
		CurrentVersion: normalizeVer(current),
		LatestVersion:  normalizeVer(current),
	}

	client := &http.Client{Timeout: 12 * time.Second}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", GitHubOwner, GitHubRepo)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		info.Error = err.Error()
		return info
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		info.Error = err.Error()
		return info
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		info.Error = "no releases yet"
		return info
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		info.Error = fmt.Sprintf("github %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
		return info
	}

	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		info.Error = err.Error()
		return info
	}

	latest := normalizeVer(rel.TagName)
	info.LatestVersion = latest
	info.ReleaseURL = rel.HTMLURL
	info.ReleaseNotes = strings.TrimSpace(rel.Body)
	info.DownloadURL = pickExeAsset(rel)
	if info.DownloadURL == "" && rel.HTMLURL != "" {
		info.DownloadURL = rel.HTMLURL
	}

	cmp := Compare(info.CurrentVersion, latest)
	info.UpdateAvailable = cmp < 0
	return info
}

func pickExeAsset(rel ghRelease) string {
	var fallback string
	for _, a := range rel.Assets {
		name := strings.ToLower(a.Name)
		if !strings.HasSuffix(name, ".exe") {
			continue
		}
		if fallback == "" {
			fallback = a.BrowserDownloadURL
		}
		if strings.Contains(name, "recode") || strings.Contains(name, "nfa-tool") {
			return a.BrowserDownloadURL
		}
	}
	return fallback
}

func normalizeVer(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(strings.ToLower(v), "v")
	// strip pre-release suffix for compare base: 2.0.0-beta -> 2.0.0 (still compared simply)
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}
	return v
}

// Compare returns -1 if a<b, 0 if equal, 1 if a>b (semver-ish major.minor.patch).
func Compare(a, b string) int {
	ap := parseParts(a)
	bp := parseParts(b)
	n := len(ap)
	if len(bp) > n {
		n = len(bp)
	}
	for i := 0; i < n; i++ {
		var av, bv int
		if i < len(ap) {
			av = ap[i]
		}
		if i < len(bp) {
			bv = bp[i]
		}
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
	}
	return 0
}

func parseParts(v string) []int {
	v = normalizeVer(v)
	segs := strings.Split(v, ".")
	out := make([]int, 0, len(segs))
	for _, s := range segs {
		n, _ := strconv.Atoi(s)
		out = append(out, n)
	}
	if len(out) == 0 {
		return []int{0}
	}
	return out
}

// ApplyDownload downloads exeURL and schedules replace+restart of the running binary.
func ApplyDownload(exeURL string) error {
	if exeURL == "" || !strings.HasPrefix(strings.ToLower(exeURL), "https://") {
		return fmt.Errorf("invalid download url")
	}
	if !strings.HasSuffix(strings.ToLower(exeURL), ".exe") {
		return fmt.Errorf("download is not an exe — open the release page instead")
	}

	cur, err := os.Executable()
	if err != nil {
		return err
	}
	cur, _ = filepath.Abs(cur)
	dir := filepath.Dir(cur)
	tmp := filepath.Join(dir, "NFA-Tool-Recode-v2.new.exe")
	bat := filepath.Join(dir, "nfa-update.bat")

	client := &http.Client{Timeout: 5 * time.Minute}
	req, err := http.NewRequest(http.MethodGet, exeURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		_ = os.Remove(tmp)
		return err
	}
	f.Close()

	// batch: wait for this pid to exit, replace exe, start new.
	// Self-delete is deferred to a child cmd so this script does not error with
	// "The batch file cannot be found" after deleting itself mid-run.
	pid := os.Getpid()
	script := fmt.Sprintf("@echo off\r\n"+
		"setlocal EnableExtensions\r\n"+
		"set \"PID=%d\"\r\n"+
		"set \"TARGET=%s\"\r\n"+
		"set \"NEW=%s\"\r\n"+
		":wait\r\n"+
		"tasklist /FI \"PID eq %%PID%%\" 2>NUL | findstr /C:\" %%PID%% \" >NUL\r\n"+
		"if not errorlevel 1 (\r\n"+
		"  >NUL 2>&1 timeout /t 1 /nobreak\r\n"+
		"  goto wait\r\n"+
		")\r\n"+
		">NUL 2>&1 timeout /t 1 /nobreak\r\n"+
		"del /F /Q \"%%TARGET%%\" >NUL 2>&1\r\n"+
		"move /Y \"%%NEW%%\" \"%%TARGET%%\" >NUL 2>&1\r\n"+
		"if exist \"%%TARGET%%\" start \"\" \"%%TARGET%%\"\r\n"+
		"start \"\" /B cmd /d /c \"ping -n 2 127.0.0.1 >nul & del /f /q \"\"%%~f0\"\"\"\r\n"+
		"exit /b 0\r\n",
		pid, cur, tmp)

	if err := os.WriteFile(bat, []byte(script), 0o755); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	// Run hidden — no console window for the user to close.
	cmd := exec.Command("cmd", "/d", "/c", bat)
	cmd.Dir = dir
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	if err := cmd.Start(); err != nil {
		_ = os.Remove(tmp)
		_ = os.Remove(bat)
		return err
	}
	return nil
}
