package steam

import (
	"fmt"
	"hash/crc32"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// LoginOptions mirrors Python login_game inputs.
// KeepExisting is optional UI feature (not in original Python).
type LoginOptions struct {
	AccountName  string
	Token        string
	SteamID      string
	KeepExisting bool
}

// Login = Python login_game (stable path).
func Login(opt LoginOptions) error {
	if opt.AccountName == "" || opt.Token == "" || opt.SteamID == "" {
		return fmt.Errorf("account, token and steam id are required")
	}
	// Python: strip @ suffix
	if i := strings.Index(opt.AccountName, "@"); i >= 0 {
		opt.AccountName = opt.AccountName[:i]
	}
	opt.AccountName = strings.ToLower(strings.TrimSpace(opt.AccountName))

	// Python get_steam_install_path kills steam if running
	install, err := InstallPath(true)
	if err != nil {
		return err
	}
	if _, err := os.Stat(install); err != nil {
		return fmt.Errorf("directory not recognized, please open steam first")
	}

	// Python: crc32(account)+"1", encrypt, mtbf
	crcKey := AccountCRCKey(opt.AccountName)
	encrypted, err := EncryptToken(opt.Token, opt.AccountName)
	if err != nil {
		return fmt.Errorf("encrypt token: %w", err)
	}
	mtbf := randomDigits(9)

	localPath, err := LocalVDFPath()
	if err != nil {
		return err
	}
	// Python: remove local.vdf first
	if _, err := os.Stat(localPath); err == nil {
		_ = os.Remove(localPath)
	}

	configDir := filepath.Join(install, "config")
	_ = os.MkdirAll(configDir, 0o755)

	// Python: AutoLoginUser
	if err := SetAutoLoginUser(opt.AccountName); err != nil {
		return fmt.Errorf("set AutoLoginUser: %w", err)
	}

	configPath := filepath.Join(configDir, "config.vdf")
	usersPath := filepath.Join(configDir, "loginusers.vdf")

	var configData, usersData, localData string
	if opt.KeepExisting {
		// optional merge (UI checkbox) — base still Python structures
		configData, err = mergeConfig(configPath, mtbf, opt.SteamID, opt.AccountName)
		if err != nil {
			return err
		}
		usersData, err = mergeLoginUsers(usersPath, opt.SteamID, opt.AccountName)
		if err != nil {
			return err
		}
		localData, err = mergeLocal(localPath, crcKey, encrypted)
		if err != nil {
			return err
		}
	} else {
		// Python exact: always full replace
		configData = buildConfig(mtbf, opt.SteamID, opt.AccountName)
		usersData = buildLoginUsers(opt.SteamID, opt.AccountName)
		localData = buildLocal(crcKey, encrypted)
	}

	// Python: remove then write config.vdf / loginusers.vdf / local.vdf
	removeWritable(configPath)
	if err := os.WriteFile(configPath, []byte(configData), 0o644); err != nil {
		return err
	}
	removeWritable(usersPath)
	if err := os.WriteFile(usersPath, []byte(usersData), 0o644); err != nil {
		return err
	}
	if _, err := os.Stat(localPath); err == nil {
		removeWritable(localPath)
		_ = os.Remove(localPath)
	}
	_ = os.MkdirAll(filepath.Dir(localPath), 0o755)
	if err := os.WriteFile(localPath, []byte(localData), 0o644); err != nil {
		return err
	}

	// Python: Popen steam.exe
	return LaunchSteam(install)
}

// ResetSteam = Python reset_steam
func ResetSteam() error {
	install, err := InstallPath(true)
	if err != nil {
		return err
	}
	for _, dir := range []string{
		filepath.Join(install, "userdata"),
		filepath.Join(install, "config"),
	} {
		if _, err := os.Stat(dir); err == nil {
			_ = filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
				if err == nil {
					_ = os.Chmod(p, 0o666)
				}
				return nil
			})
			_ = os.RemoveAll(dir)
		}
	}
	if lp, err := LocalVDFPath(); err == nil {
		_ = os.Remove(lp)
	}
	return LaunchSteam(install)
}

// ReadLoginUsers = Python get_all_login_users names
func ReadLoginUsers() ([]string, error) {
	install, err := InstallPath(false)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(install, "config", "loginusers.vdf")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	root, err := Parse(string(data))
	if err != nil {
		return nil, err
	}
	users := asMap(root["users"])
	if users == nil {
		return nil, nil
	}
	var names []string
	for _, v := range users {
		if u := asMap(v); u != nil {
			if n := asString(u["AccountName"]); n != "" {
				names = append(names, n)
			}
		}
	}
	return names, nil
}

// HarvestConnectCache = Python save_current_cache decrypt loop
func HarvestConnectCache() (map[string]string, error) {
	names, err := ReadLoginUsers()
	if err != nil || len(names) == 0 {
		return map[string]string{}, err
	}
	mapping := map[string]string{}
	for _, n := range names {
		mapping[AccountCRCKey(n)] = n
	}

	lp, err := LocalVDFPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(lp)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	root, err := Parse(string(data))
	if err != nil {
		return nil, err
	}
	cache := dig(root, "MachineUserConfigStore", "Software", "Valve", "Steam", "ConnectCache")
	if cache == nil {
		return map[string]string{}, nil
	}
	out := map[string]string{}
	for key, val := range cache {
		account := mapping[key]
		if account == "" {
			continue
		}
		tok, err := DecryptToken(asString(val))
		if err != nil {
			continue
		}
		out[account] = tok
	}
	return out, nil
}

// AccountCRCKey = Python compute_crc32 + "1"
func AccountCRCKey(account string) string {
	sum := crc32.ChecksumIEEE([]byte(account))
	h := fmt.Sprintf("%08x", sum)
	h = strings.TrimLeft(h, "0")
	return h + "1"
}

// buildConfig = Python build_config (core fields that matter for login)
func buildConfig(mtbf, steamID, account string) string {
	return Dump(Map{
		"InstallConfigStore": Map{
			"Software": Map{
				"Valve": Map{
					"Steam": Map{
						"AutoUpdateWindowEnabled": "0",
						"ipv6check_http_state":    "bad",
						"ipv6check_udp_state":     "bad",
						"Accounts": Map{
							account: Map{"SteamID": steamID},
						},
						"CellIDServerOverride": "170",
						"MTBF":                 mtbf,
						"cip":                  "02000000507a6c24d6e96c6b00004021a356",
						"SurveyDate":           "2017-10-22",
						"SurveyDateVersion":    "-1724767764117155760",
						"SurveyDateType":       "3",
						"Rate":                 "30000",
					},
				},
			},
		},
	})
}

// buildLoginUsers = Python build_login_users
func buildLoginUsers(steamID, account string) string {
	return Dump(Map{
		"users": Map{
			steamID: Map{
				"AccountName":            account,
				"PersonaName":            account,
				"RememberPassword":       "1",
				"WantsOfflineMode":       "0",
				"SkipOfflineModeWarning": "0",
				"AllowAutoLogin":         "1",
				"MostRecent":             "1",
				"Timestamp":              strconv.FormatInt(time.Now().Unix(), 10),
			},
		},
	})
}

// buildLocal = Python build_local
func buildLocal(crcKey, encryptedJWT string) string {
	return Dump(Map{
		"MachineUserConfigStore": Map{
			"Software": Map{
				"Valve": Map{
					"Steam": Map{
						"ConnectCache": Map{
							crcKey: encryptedJWT,
						},
					},
				},
			},
		},
	})
}

func mergeConfig(path, mtbf, steamID, account string) (string, error) {
	root := Map{}
	if data, err := os.ReadFile(path); err == nil {
		if p, err := Parse(string(data)); err == nil {
			root = p
		}
	}
	if len(root) == 0 {
		return buildConfig(mtbf, steamID, account), nil
	}
	st := digEnsure(root, "InstallConfigStore", "Software", "Valve", "Steam")
	acc := ensureMap(st, "Accounts")
	acc[account] = Map{"SteamID": steamID}
	st["MTBF"] = mtbf
	return Dump(root), nil
}

func mergeLoginUsers(path, steamID, account string) (string, error) {
	root := Map{}
	if data, err := os.ReadFile(path); err == nil {
		if p, err := Parse(string(data)); err == nil {
			root = p
		}
	}
	if len(root) == 0 {
		return buildLoginUsers(steamID, account), nil
	}
	users := ensureMap(root, "users")
	for id, v := range users {
		if u := asMap(v); u != nil {
			u["MostRecent"] = "0"
			users[id] = u
		}
	}
	users[steamID] = Map{
		"AccountName":            account,
		"PersonaName":            account,
		"RememberPassword":       "1",
		"WantsOfflineMode":       "0",
		"SkipOfflineModeWarning": "0",
		"AllowAutoLogin":         "1",
		"MostRecent":             "1",
		"Timestamp":              strconv.FormatInt(time.Now().Unix(), 10),
	}
	return Dump(root), nil
}

func mergeLocal(path, crcKey, encryptedJWT string) (string, error) {
	root := Map{}
	if data, err := os.ReadFile(path); err == nil {
		if p, err := Parse(string(data)); err == nil {
			root = p
		}
	}
	if len(root) == 0 {
		return buildLocal(crcKey, encryptedJWT), nil
	}
	cache := digEnsure(root, "MachineUserConfigStore", "Software", "Valve", "Steam", "ConnectCache")
	cache[crcKey] = encryptedJWT
	return Dump(root), nil
}

func ensureMap(parent Map, key string) Map {
	if m := asMap(parent[key]); m != nil {
		return m
	}
	m := Map{}
	parent[key] = m
	return m
}

func dig(root Map, keys ...string) Map {
	cur := root
	for _, k := range keys {
		next := asMap(cur[k])
		if next == nil {
			return nil
		}
		cur = next
	}
	return cur
}

func digEnsure(root Map, keys ...string) Map {
	cur := root
	for _, k := range keys {
		cur = ensureMap(cur, k)
	}
	return cur
}

func removeWritable(path string) {
	if _, err := os.Stat(path); err != nil {
		return
	}
	_ = os.Chmod(path, 0o666)
	_ = os.Remove(path)
}

func randomDigits(n int) string {
	var b strings.Builder
	b.Grow(n)
	for i := 0; i < n; i++ {
		b.WriteByte(byte('0' + rand.Intn(10)))
	}
	return b.String()
}
