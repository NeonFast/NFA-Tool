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

// LoginOptions controls how files are written.
type LoginOptions struct {
	AccountName string
	Token       string
	SteamID     string
	// KeepExisting merges into current Steam accounts instead of wiping loginusers.
	KeepExisting bool
}

// AccountSummary is a saved account row for the UI.
type AccountSummary struct {
	Name string `json:"name"`
}

// Login injects the refresh token into Steam ConnectCache and starts Steam.
func Login(opt LoginOptions) error {
	if opt.AccountName == "" || opt.Token == "" || opt.SteamID == "" {
		return fmt.Errorf("account, token and steam id are required")
	}
	if i := strings.Index(opt.AccountName, "@"); i >= 0 {
		opt.AccountName = opt.AccountName[:i]
	}
	opt.AccountName = strings.ToLower(strings.TrimSpace(opt.AccountName))

	install, err := InstallPath(true)
	if err != nil {
		return err
	}
	if _, err := os.Stat(install); err != nil {
		return fmt.Errorf("steam directory not found: %w", err)
	}

	encrypted, err := EncryptToken(opt.Token, opt.AccountName)
	if err != nil {
		return fmt.Errorf("encrypt token: %w", err)
	}

	crcKey := AccountCRCKey(opt.AccountName)
	configDir := filepath.Join(install, "config")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}

	if err := SetAutoLoginUser(opt.AccountName); err != nil {
		return fmt.Errorf("set AutoLoginUser: %w", err)
	}

	// config.vdf
	configPath := filepath.Join(configDir, "config.vdf")
	configData, err := buildConfigVDF(configPath, opt.SteamID, opt.AccountName, opt.KeepExisting)
	if err != nil {
		return err
	}
	if err := writeFile(configPath, configData); err != nil {
		return err
	}

	// loginusers.vdf
	usersPath := filepath.Join(configDir, "loginusers.vdf")
	usersData, err := buildLoginUsersVDF(usersPath, opt.SteamID, opt.AccountName, opt.KeepExisting)
	if err != nil {
		return err
	}
	if err := writeFile(usersPath, usersData); err != nil {
		return err
	}

	// local.vdf ConnectCache
	localPath, err := LocalVDFPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
		return err
	}
	localData, err := buildLocalVDF(localPath, crcKey, encrypted, opt.KeepExisting)
	if err != nil {
		return err
	}
	if err := writeFile(localPath, localData); err != nil {
		return err
	}

	return LaunchSteam(install)
}

// ResetSteam removes userdata/config and local.vdf, then launches Steam.
func ResetSteam() error {
	install, err := InstallPath(true)
	if err != nil {
		return err
	}
	for _, dir := range []string{
		filepath.Join(install, "userdata"),
		filepath.Join(install, "config"),
	} {
		_ = removeAllWritable(dir)
	}
	if lp, err := LocalVDFPath(); err == nil {
		_ = os.Remove(lp)
	}
	return LaunchSteam(install)
}

// ReadLoginUsers returns account names from loginusers.vdf.
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
		u := asMap(v)
		if u == nil {
			continue
		}
		if n := asString(u["AccountName"]); n != "" {
			names = append(names, n)
		}
	}
	return names, nil
}

// HarvestConnectCache decrypts tokens from local.vdf into account->token map.
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
	cache := digMap(root, "MachineUserConfigStore", "Software", "Valve", "Steam", "ConnectCache")
	if cache == nil {
		return map[string]string{}, nil
	}

	out := map[string]string{}
	for key, val := range cache {
		account := mapping[key]
		if account == "" {
			continue
		}
		hexVal := asString(val)
		tok, err := DecryptToken(hexVal)
		if err != nil {
			continue
		}
		out[account] = tok
	}
	return out, nil
}

// AccountCRCKey CRC32 hex (no leading zeros) + "1"
func AccountCRCKey(account string) string {
	sum := crc32.ChecksumIEEE([]byte(account))
	hex := fmt.Sprintf("%08x", sum)
	hex = strings.TrimLeft(hex, "0")
	if hex == "" {
		hex = "0"
	}
	return hex + "1"
}

func buildConfigVDF(path, steamID, account string, keep bool) (string, error) {
	mtbf := randomDigits(9)
	var root Map
	if keep {
		if data, err := os.ReadFile(path); err == nil {
			if parsed, err := Parse(string(data)); err == nil {
				root = parsed
			}
		}
	}
	if root == nil {
		root = Map{}
	}

	ics := ensureMap(root, "InstallConfigStore")
	soft := ensureMap(ics, "Software")
	valve := ensureMap(soft, "Valve")
	steam := ensureMap(valve, "Steam")
	accounts := ensureMap(steam, "Accounts")
	accounts[account] = Map{"SteamID": steamID}
	steam["MTBF"] = mtbf
	if _, ok := steam["AutoUpdateWindowEnabled"]; !ok {
		steam["AutoUpdateWindowEnabled"] = "0"
	}

	return Dump(root), nil
}

func buildLoginUsersVDF(path, steamID, account string, keep bool) (string, error) {
	var root Map
	if keep {
		if data, err := os.ReadFile(path); err == nil {
			if parsed, err := Parse(string(data)); err == nil {
				root = parsed
			}
		}
	}
	if root == nil {
		root = Map{}
	}
	users := ensureMap(root, "users")

	// clear MostRecent on others
	for id, v := range users {
		u := asMap(v)
		if u == nil {
			continue
		}
		u["MostRecent"] = "0"
		users[id] = u
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

func buildLocalVDF(path, crcKey, encryptedJWT string, keep bool) (string, error) {
	var root Map
	if keep {
		if data, err := os.ReadFile(path); err == nil {
			if parsed, err := Parse(string(data)); err == nil {
				root = parsed
			}
		}
	}
	if root == nil {
		root = Map{}
	}
	mucs := ensureMap(root, "MachineUserConfigStore")
	soft := ensureMap(mucs, "Software")
	valve := ensureMap(soft, "Valve")
	st := ensureMap(valve, "Steam")
	cache := ensureMap(st, "ConnectCache")
	if !keep {
		// replace cache entirely
		for k := range cache {
			delete(cache, k)
		}
	}
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

func digMap(root Map, keys ...string) Map {
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

func writeFile(path, content string) error {
	_ = makeWritable(path)
	_ = os.Remove(path)
	return os.WriteFile(path, []byte(content), 0o644)
}

func makeWritable(path string) error {
	if _, err := os.Stat(path); err != nil {
		return err
	}
	return os.Chmod(path, 0o666)
}

func removeAllWritable(path string) error {
	_ = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		_ = os.Chmod(p, 0o666)
		return nil
	})
	return os.RemoveAll(path)
}

func randomDigits(n int) string {
	var b strings.Builder
	b.Grow(n)
	for i := 0; i < n; i++ {
		b.WriteByte(byte('0' + rand.Intn(10)))
	}
	return b.String()
}
