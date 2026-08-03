package steam

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// InstallPath resolves the Steam install directory and optionally kills Steam.
func InstallPath(kill bool) (string, error) {
	if pid := findPID("steam.exe"); pid != 0 {
		exe, err := processExePath(pid)
		if err == nil && exe != "" {
			dir := filepath.Dir(exe)
			if kill {
				if err := KillSteam(); err != nil {
					return "", err
				}
			}
			return dir, nil
		}
	}

	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Classes\steam\Shell\Open\Command`, registry.QUERY_VALUE)
	if err != nil {
		// fallback common path
		pf := os.Getenv("ProgramFiles(x86)")
		if pf == "" {
			pf = os.Getenv("ProgramFiles")
		}
		candidate := filepath.Join(pf, "Steam")
		if _, err := os.Stat(filepath.Join(candidate, "steam.exe")); err == nil {
			if kill {
				_ = KillSteam()
			}
			return candidate, nil
		}
		return "", fmt.Errorf("steam not found in registry: %w", err)
	}
	defer k.Close()

	val, _, err := k.GetStringValue("")
	if err != nil {
		return "", fmt.Errorf("steam command value: %w", err)
	}
	val = strings.TrimSpace(val)
	val = strings.Trim(val, `"`)
	// typically: C:\...\steam.exe -- "%1"
	if i := strings.Index(strings.ToLower(val), "steam.exe"); i >= 0 {
		dir := filepath.Dir(val[:i+len("steam.exe")])
		if kill {
			_ = KillSteam()
		}
		return dir, nil
	}
	return "", fmt.Errorf("cannot parse steam path from %q", val)
}

// LocalVDFPath returns %LOCALAPPDATA%\Steam\local.vdf
func LocalVDFPath() (string, error) {
	local := os.Getenv("LOCALAPPDATA")
	if local == "" {
		return "", fmt.Errorf("LOCALAPPDATA is empty")
	}
	return filepath.Join(local, "Steam", "local.vdf"), nil
}

// KillSteam force-stops Steam processes.
func KillSteam() error {
	_ = exec.Command("taskkill", "/F", "/IM", "steam.exe").Run()
	_ = exec.Command("taskkill", "/F", "/IM", "steamwebhelper.exe").Run()
	time.Sleep(2 * time.Second)
	return nil
}

// LaunchSteam starts steam.exe from install dir.
func LaunchSteam(installDir string) error {
	exe := filepath.Join(installDir, "steam.exe")
	cmd := exec.Command(exe)
	cmd.Dir = installDir
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Start()
}

// SetAutoLoginUser writes HKCU\SOFTWARE\Valve\Steam AutoLoginUser.
func SetAutoLoginUser(account string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Valve\Steam`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	return k.SetStringValue("AutoLoginUser", account)
}

func findPID(name string) uint32 {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0
	}
	defer windows.CloseHandle(snap)

	var pe windows.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	if err := windows.Process32First(snap, &pe); err != nil {
		return 0
	}
	for {
		exe := windows.UTF16ToString(pe.ExeFile[:])
		if strings.EqualFold(exe, name) {
			return pe.ProcessID
		}
		if err := windows.Process32Next(snap, &pe); err != nil {
			break
		}
	}
	return 0
}

func processExePath(pid uint32) (string, error) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return "", err
	}
	defer windows.CloseHandle(h)

	var buf [windows.MAX_PATH]uint16
	size := uint32(len(buf))
	err = windows.QueryFullProcessImageName(h, 0, &buf[0], &size)
	if err != nil {
		return "", err
	}
	return windows.UTF16ToString(buf[:size]), nil
}
