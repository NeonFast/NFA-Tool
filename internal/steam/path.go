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

// InstallPath = Python get_steam_install_path (always kills if steam running when kill=true)
func InstallPath(kill bool) (string, error) {
	if pid := findPID("steam.exe"); pid != 0 {
		exe, err := processExePath(pid)
		if err != nil {
			return "", err
		}
		dir := filepath.Dir(exe)
		if kill {
			// Python: taskkill /f /im steam.exe + steamwebhelper, sleep 2
			_ = exec.Command("taskkill", "/F", "/IM", "steam.exe").Run()
			_ = exec.Command("taskkill", "/F", "/IM", "steamwebhelper.exe").Run()
			time.Sleep(2 * time.Second)
		}
		return dir, nil
	}

	// Python registry fallback
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Classes\steam\Shell\Open\Command`, registry.QUERY_VALUE)
	if err != nil {
		return "", fmt.Errorf("steam not found: %w", err)
	}
	defer k.Close()
	val, _, err := k.GetStringValue("")
	if err != nil {
		return "", err
	}
	val = strings.ReplaceAll(val, `"`, "")
	// Python: path[:len-6] then [:len-9] → strip ' -- %1' and '\steam.exe'
	if len(val) > 6 {
		val = val[:len(val)-6]
	}
	if len(val) > 9 {
		val = val[:len(val)-9]
	}
	dir := filepath.Clean(val)
	if _, err := os.Stat(filepath.Join(dir, "steam.exe")); err != nil {
		// safer parse fallback
		if d, err2 := steamDirFromCommand(val); err2 == nil {
			dir = d
		} else {
			return "", fmt.Errorf("steam directory invalid: %s", dir)
		}
	}
	if kill {
		_ = exec.Command("taskkill", "/F", "/IM", "steam.exe").Run()
		_ = exec.Command("taskkill", "/F", "/IM", "steamwebhelper.exe").Run()
		time.Sleep(2 * time.Second)
	}
	return dir, nil
}

func steamDirFromCommand(val string) (string, error) {
	low := strings.ToLower(val)
	if i := strings.Index(low, "steam.exe"); i >= 0 {
		return filepath.Dir(val[:i+len("steam.exe")]), nil
	}
	return "", fmt.Errorf("cannot parse")
}

// LocalVDFPath = Python get_local_vdf_path
func LocalVDFPath() (string, error) {
	local := os.Getenv("LOCALAPPDATA")
	if local == "" {
		return "", fmt.Errorf("LOCALAPPDATA is empty")
	}
	return filepath.Join(local, "Steam", "local.vdf"), nil
}

// KillSteam = Python taskkill pair + sleep
func KillSteam() error {
	_ = exec.Command("taskkill", "/F", "/IM", "steam.exe").Run()
	_ = exec.Command("taskkill", "/F", "/IM", "steamwebhelper.exe").Run()
	time.Sleep(2 * time.Second)
	return nil
}

// LaunchSteam = Python subprocess.Popen(steam.exe, shell=True)
func LaunchSteam(installDir string) error {
	exe := filepath.Join(installDir, "steam.exe")
	cmd := exec.Command(exe)
	cmd.Dir = installDir
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Start()
}

// SetAutoLoginUser = Python winreg AutoLoginUser
func SetAutoLoginUser(account string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Valve\Steam`, registry.SET_VALUE)
	if err != nil {
		k, _, err = registry.CreateKey(registry.CURRENT_USER, `SOFTWARE\Valve\Steam`, registry.SET_VALUE)
		if err != nil {
			return err
		}
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
		if strings.EqualFold(windows.UTF16ToString(pe.ExeFile[:]), name) {
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
	buf := make([]uint16, 32768)
	size := uint32(len(buf))
	if err := windows.QueryFullProcessImageName(h, 0, &buf[0], &size); err != nil {
		return "", err
	}
	return windows.UTF16ToString(buf[:size]), nil
}
