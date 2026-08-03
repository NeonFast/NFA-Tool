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

func GetSteamInstallPath() (string, error)       { return getSteamInstallPath(true) }
func GetSteamInstallPathNoKill() (string, error) { return getSteamInstallPath(false) }
func InstallPath(kill bool) (string, error)      { return getSteamInstallPath(kill) }

func getSteamInstallPath(kill bool) (string, error) {
	// Prefer SteamPath (roster / shefu)
	if k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Valve\Steam`, registry.QUERY_VALUE); err == nil {
		if sp, _, err := k.GetStringValue("SteamPath"); err == nil && sp != "" {
			k.Close()
			dir := filepath.Clean(sp)
			if _, err := os.Stat(filepath.Join(dir, "steam.exe")); err == nil {
				if kill {
					_ = KillSteam()
				}
				return dir, nil
			}
		} else {
			k.Close()
		}
	}

	if pid := findPID("steam.exe"); pid != 0 {
		exe, err := processExePath(pid)
		if err != nil {
			return "", err
		}
		if kill {
			_ = KillSteam()
		}
		return filepath.Dir(exe), nil
	}

	return "", fmt.Errorf("steam not found — is it installed?")
}

func GetLocalVDFPath() (string, error) {
	app := os.Getenv("LOCALAPPDATA")
	if app == "" {
		app = os.Getenv("localappdata")
	}
	if app == "" {
		return "", fmt.Errorf("LOCALAPPDATA empty")
	}
	return filepath.Join(app, "Steam", "local.vdf"), nil
}

func LocalVDFPath() (string, error) { return GetLocalVDFPath() }

// KillSteam — roster style: steam.exe + steamwebhelper with /T
func KillSteam() error {
	runHidden("taskkill", "/F", "/IM", "steam.exe", "/T")
	time.Sleep(400 * time.Millisecond)
	runHidden("taskkill", "/F", "/IM", "steamwebhelper.exe", "/T")
	time.Sleep(800 * time.Millisecond)
	return nil
}

func runHidden(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000} // CREATE_NO_WINDOW
	_ = cmd.Run()
}

// LaunchSteam — roster: DETACHED_PROCESS, direct steam.exe (not child chain via cmd)
func LaunchSteam(installDir string) error {
	exe := filepath.Join(installDir, "steam.exe")
	if _, err := os.Stat(exe); err != nil {
		return fmt.Errorf("steam.exe not found")
	}
	// Prefer unelevated via explorer when we are admin (UI otherwise broken)
	if isProcessElevated() {
		if err := launchViaExplorer(exe); err == nil {
			time.Sleep(1 * time.Second)
			if findPID("steam.exe") != 0 {
				return nil
			}
		}
	}
	// roster fallback: detached direct spawn
	cmd := exec.Command(exe)
	cmd.Dir = installDir
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x00000008, // DETACHED_PROCESS
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	return nil
}

func isProcessElevated() bool {
	var sid *windows.SID
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0, &sid,
	)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)
	member, err := windows.Token(0).IsMember(sid)
	return err == nil && member
}

func launchViaExplorer(target string) error {
	windir := os.Getenv("SystemRoot")
	if windir == "" {
		windir = `C:\Windows`
	}
	explorer := filepath.Join(windir, "explorer.exe")
	verb, _ := windows.UTF16PtrFromString("open")
	file, _ := windows.UTF16PtrFromString(explorer)
	params, _ := windows.UTF16PtrFromString(`"` + target + `"`)
	ret, _, err := windows.NewLazySystemDLL("shell32.dll").NewProc("ShellExecuteW").Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(file)),
		uintptr(unsafe.Pointer(params)),
		0,
		uintptr(windows.SW_SHOWNORMAL),
	)
	if ret <= 32 {
		if err != nil {
			return err
		}
		return fmt.Errorf("ShellExecute %d", ret)
	}
	return nil
}

func IsSteamRunning() bool { return findPID("steam.exe") != 0 }

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
