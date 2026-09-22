package steam

import (
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// LoginOptions configures a token login run.
type LoginOptions struct {
	AccountName  string
	Token        string
	SteamID      string
	KeepExisting bool
	OnProgress   func(stage string)
}

// Login injects a refresh token into the Steam client session files.
func Login(opt LoginOptions) (bool, error) {
	report := func(stage string) {
		if opt.OnProgress != nil {
			opt.OnProgress(stage)
		}
	}
	account := strings.ToLower(strings.TrimSpace(opt.AccountName))
	if i := strings.Index(account, "@"); i >= 0 {
		account = account[:i]
	}
	token := strings.TrimSpace(opt.Token)
	steamID := strings.TrimSpace(opt.SteamID)
	if account == "" || token == "" || steamID == "" {
		return false, fmt.Errorf("account, token and steam id are required")
	}

	report("find_steam")
	install, err := GetSteamInstallPathNoKill()
	if err != nil {
		return false, err
	}
	configDir := filepath.Join(install, "config")
	configPath := filepath.Join(configDir, "config.vdf")
	usersPath := filepath.Join(configDir, "loginusers.vdf")
	localDir := filepath.Join(os.Getenv("LOCALAPPDATA"), "Steam")
	if os.Getenv("LOCALAPPDATA") == "" {
		localDir = filepath.Join(os.Getenv("localappdata"), "Steam")
	}
	localPath := filepath.Join(localDir, "local.vdf")

	if !fileExists(configPath) || !fileExists(usersPath) {
		return false, fmt.Errorf("open Steam and sign in once so it can create its config files, then try again")
	}

	report("stop_steam")
	_ = KillSteam()

	report("loginusers")
	if err := setLoginUsersActive(usersPath, account, steamID); err != nil {
		return false, fmt.Errorf("loginusers.vdf: %w", err)
	}

	report("connectcache")
	if err := storeConnectCacheToken(localPath, account, token); err != nil {
		return false, fmt.Errorf("local.vdf: %w", err)
	}

	report("config_vdf")
	_ = injectConfigAccount(configPath, account, steamID)

	report("registry")
	if err := SetAutoLoginUser(account); err != nil {
		return false, fmt.Errorf("AutoLoginUser: %w", err)
	}

	report("acl")
	fixACL(usersPath)
	fixACL(localPath)
	fixACL(configPath)

	time.Sleep(400 * time.Millisecond)
	report("launch")
	if err := LaunchSteam(install); err != nil {
		return false, err
	}
	report("wait_window")
	return WaitForSteamWindow(25 * time.Second), nil
}

// ResetSteam wipes Steam's userdata and config directories and relaunches the client.
func ResetSteam() error {
	install, err := GetSteamInstallPath()
	if err != nil {
		return err
	}
	_ = KillSteam()
	for _, directory := range []string{
		filepath.Join(install, "userdata"),
		filepath.Join(install, "config"),
	} {
		if fileExists(directory) {
			_ = filepath.Walk(directory, func(p string, info os.FileInfo, err error) error {
				if err == nil {
					_ = os.Chmod(p, 0o666)
				}
				return nil
			})
			_ = os.RemoveAll(directory)
		}
	}
	if lp, err := GetLocalVDFPath(); err == nil {
		_ = os.Remove(lp)
	}
	return LaunchSteam(install)
}

// ReadLoginUsers returns the account names listed in loginusers.vdf.
func ReadLoginUsers() ([]string, error) {
	install, err := GetSteamInstallPathNoKill()
	if err != nil {
		return nil, err
	}
	p := filepath.Join(install, "config", "loginusers.vdf")
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, line := range strings.Split(string(data), "\n") {
		fields := quotedFields(line)
		if len(fields) >= 2 && fields[0] == "AccountName" {
			names = append(names, fields[1])
		}
	}
	return names, nil
}

// HarvestConnectCache decrypts ConnectCache tokens from local.vdf for every
// account listed in loginusers.vdf.
func HarvestConnectCache() (map[string]string, error) {
	names, err := ReadLoginUsers()
	if err != nil || len(names) == 0 {
		return map[string]string{}, err
	}
	lp, err := GetLocalVDFPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(lp)
	if err != nil {
		return map[string]string{}, nil
	}
	enc := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		fields := quotedFields(line)
		if len(fields) >= 2 && strings.HasSuffix(fields[0], "1") && len(fields[1]) > 32 {
			enc[fields[0]] = fields[1]
		}
	}
	out := map[string]string{}
	for _, n := range names {
		key := AccountCRCKey(n)
		if e, ok := enc[key]; ok {
			if tok, err := DecryptToken(e, n); err == nil && tok != "" {
				out[n] = tok
			}
		}
	}
	return out, nil
}

// AccountCRCKey derives the ConnectCache store key for an account name.
func AccountCRCKey(account string) string {
	crc := crc32.ChecksumIEEE([]byte(account))
	hex := fmt.Sprintf("%08x", crc)
	trimmed := strings.TrimLeft(hex, "0")
	if trimmed == "" {
		return "01"
	}
	return trimmed + "1"
}

func setLoginUsersActive(path, username, steamID string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	s := string(content)
	s = strings.ReplaceAll(s, `"MostRecent"		"1"`, `"MostRecent"		"0"`)
	s = strings.ReplaceAll(s, "\"MostRecent\"\t\t\"1\"", "\"MostRecent\"\t\t\"0\"")
	s = strings.ReplaceAll(s, `"AutoLogin"		"1"`, `"AutoLogin"		"0"`)
	s = strings.ReplaceAll(s, "\"AutoLogin\"\t\t\"1\"", "\"AutoLogin\"\t\t\"0\"")
	s = strings.ReplaceAll(s, `"AllowAutoLogin"		"1"`, `"AllowAutoLogin"		"0"`)
	s = strings.ReplaceAll(s, "\"AllowAutoLogin\"\t\t\"1\"", "\"AllowAutoLogin\"\t\t\"0\"")

	if strings.Contains(s, `"`+steamID+`"`) {
		s = refreshUserBlock(s, username, steamID)
	} else {
		block := newUserBlock(username, steamID)
		idx := strings.LastIndex(s, "}")
		if idx < 0 {
			return fmt.Errorf("loginusers.vdf malformed")
		}
		s = s[:idx] + block + s[idx:]
	}
	return os.WriteFile(path, []byte(s), 0o644)
}

func refreshUserBlock(content, username, steamID string) string {
	lines := strings.Split(content, "\n")
	var out strings.Builder
	i := 0
	for i < len(lines) {
		line := lines[i]
		out.WriteString(line)
		out.WriteByte('\n')
		if isSteamIDHeader(line, steamID) {
			i++
			seenRemember, seenAuto, seenMost, seenAutoLogin := false, false, false, false
			var body []string
			for i < len(lines) {
				inner := lines[i]
				trim := strings.TrimSpace(inner)
				closing := trim == "}"
				body = append(body, rewriteLoginField(inner, username, &seenRemember, &seenAuto, &seenMost, &seenAutoLogin))
				i++
				if closing {
					break
				}
			}
			if len(body) > 0 {
				closing := body[len(body)-1]
				body = body[:len(body)-1]
				if !seenRemember {
					body = append(body, "\t\t\"RememberPassword\"\t\t\"1\"")
				}
				if !seenAuto {
					body = append(body, "\t\t\"AllowAutoLogin\"\t\t\"1\"")
				}
				if !seenMost {
					body = append(body, "\t\t\"MostRecent\"\t\t\"1\"")
				}
				if !seenAutoLogin {
					body = append(body, "\t\t\"AutoLogin\"\t\t\"1\"")
				}
				body = append(body, closing)
			}
			for _, b := range body {
				out.WriteString(b)
				if !strings.HasSuffix(b, "\n") {
					out.WriteByte('\n')
				}
			}
			continue
		}
		i++
	}
	return out.String()
}

func rewriteLoginField(line, username string, seenRemember, seenAuto, seenMost, seenAutoLogin *bool) string {
	switch {
	case strings.Contains(line, `"AccountName"`):
		return "\t\t\"AccountName\"\t\t\"" + username + "\""
	case strings.Contains(line, `"PersonaName"`):
		return line
	case strings.Contains(line, `"MostRecent"`):
		*seenMost = true
		return "\t\t\"MostRecent\"\t\t\"1\""
	case strings.Contains(line, `"Timestamp"`):
		return "\t\t\"Timestamp\"\t\t\"" + strconv.FormatInt(time.Now().Unix(), 10) + "\""
	case strings.Contains(line, `"RememberPassword"`):
		*seenRemember = true
		return "\t\t\"RememberPassword\"\t\t\"1\""
	case strings.Contains(line, `"AllowAutoLogin"`):
		*seenAuto = true
		return "\t\t\"AllowAutoLogin\"\t\t\"1\""
	case strings.Contains(line, `"AutoLogin"`):
		*seenAutoLogin = true
		return "\t\t\"AutoLogin\"\t\t\"1\""
	default:
		return line
	}
}

func newUserBlock(username, steamID string) string {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	return fmt.Sprintf(
		"\t\"%s\"\n\t{\n"+
			"\t\t\"AccountName\"\t\t\"%s\"\n"+
			"\t\t\"PersonaName\"\t\t\"%s\"\n"+
			"\t\t\"RememberPassword\"\t\t\"1\"\n"+
			"\t\t\"WantsOfflineMode\"\t\t\"0\"\n"+
			"\t\t\"SkipOfflineModeWarning\"\t\t\"0\"\n"+
			"\t\t\"AllowAutoLogin\"\t\t\"1\"\n"+
			"\t\t\"AutoLogin\"\t\t\"1\"\n"+
			"\t\t\"MostRecent\"\t\t\"1\"\n"+
			"\t\t\"Timestamp\"\t\t\"%s\"\n"+
			"\t}\n",
		steamID, username, username, ts,
	)
}

func isSteamIDHeader(line, steamID string) bool {
	fields := quotedFields(line)
	return len(fields) == 1 && fields[0] == steamID
}

func storeConnectCacheToken(path, username, token string) error {
	key := AccountCRCKey(username)
	encrypted, err := EncryptToken(token, username)
	if err != nil {
		return err
	}
	_ = os.MkdirAll(filepath.Dir(path), 0o755)

	existing, err := os.ReadFile(path)
	if err != nil {
		return os.WriteFile(path, []byte(freshLocalVDF(key, encrypted)), 0o644)
	}
	content := string(existing)
	if updated, ok := upsertConnectCacheEntry(content, key, encrypted); ok {
		return os.WriteFile(path, []byte(updated), 0o644)
	}
	if updated, ok := insertConnectCacheBlock(content, key, encrypted); ok {
		return os.WriteFile(path, []byte(updated), 0o644)
	}
	return fmt.Errorf("local.vdf has unexpected layout; not overwriting it")
}

func upsertConnectCacheEntry(content, key, encrypted string) (string, bool) {
	var out strings.Builder
	inside := false
	depth := 0
	written := false
	keyPrefix := `"` + key + `"`

	for _, line := range strings.Split(content, "\n") {
		trim := strings.TrimSpace(line)
		if trim == `"ConnectCache"` {
			inside = true
			depth = 0
			out.WriteString(line)
			out.WriteByte('\n')
			continue
		}
		if !inside {
			out.WriteString(line)
			out.WriteByte('\n')
			continue
		}
		if strings.HasPrefix(trim, "{") {
			depth++
			out.WriteString(line)
			out.WriteByte('\n')
			continue
		}
		if strings.HasPrefix(trim, "}") {
			depth--
			if depth == 0 && !written {
				out.WriteString("\t\t\t\t\t\"" + key + "\"\t\t\"" + encrypted + "\"\n")
				written = true
			}
			out.WriteString(line)
			out.WriteByte('\n')
			if depth == 0 {
				inside = false
			}
			continue
		}
		if strings.HasPrefix(trim, keyPrefix) {
			out.WriteString("\t\t\t\t\t\"" + key + "\"\t\t\"" + encrypted + "\"\n")
			written = true
			continue
		}
		out.WriteString(line)
		out.WriteByte('\n')
	}
	if !written {
		return "", false
	}
	return out.String(), true
}

func insertConnectCacheBlock(content, key, encrypted string) (string, bool) {
	var out strings.Builder
	afterSteam := false
	inserted := false
	for _, line := range strings.Split(content, "\n") {
		out.WriteString(line)
		out.WriteByte('\n')
		trim := strings.TrimSpace(line)
		if trim == `"Steam"` {
			afterSteam = true
			continue
		}
		if !afterSteam {
			continue
		}
		afterSteam = false
		if inserted || !strings.HasPrefix(trim, "{") {
			continue
		}
		indent := leadingWS(line) + "\t"
		out.WriteString(indent + "\"ConnectCache\"\n")
		out.WriteString(indent + "{\n")
		out.WriteString(indent + "\t\"" + key + "\"\t\t\"" + encrypted + "\"\n")
		out.WriteString(indent + "}\n")
		inserted = true
	}
	return out.String(), inserted
}

func freshLocalVDF(key, encrypted string) string {
	return "\"MachineUserConfigStore\"\n" +
		"{\n" +
		"\t\"Software\"\n" +
		"\t{\n" +
		"\t\t\"Valve\"\n" +
		"\t\t{\n" +
		"\t\t\t\"Steam\"\n" +
		"\t\t\t{\n" +
		"\t\t\t\t\"ConnectCache\"\n" +
		"\t\t\t\t{\n" +
		"\t\t\t\t\t\"" + key + "\"\t\t\"" + encrypted + "\"\n" +
		"\t\t\t\t}\n" +
		"\t\t\t}\n" +
		"\t\t}\n" +
		"\t}\n" +
		"}\n"
}

func injectConfigAccount(path, username, steamID string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)
	if strings.Contains(content, `"SteamID"		"`+steamID+`"`) ||
		strings.Contains(content, "\"SteamID\"\t\t\""+steamID+"\"") {
		return nil
	}
	block := "\n\t\t\t\t\t\"" + username + "\"\n\t\t\t\t\t{\n\t\t\t\t\t\t\"SteamID\"\t\t\"" + steamID + "\"\n\t\t\t\t\t}\n"
	idx := strings.Index(content, `"Accounts"`)
	if idx < 0 {
		return fmt.Errorf("no Accounts block")
	}
	rest := content[idx:]
	brace := strings.Index(rest, "{")
	if brace < 0 {
		return fmt.Errorf("Accounts malformed")
	}
	pos := idx + brace + 1
	content = content[:pos] + block + content[pos:]
	return os.WriteFile(path, []byte(content), 0o644)
}

func quotedFields(line string) []string {
	var fields []string
	s := line
	for {
		i := strings.Index(s, `"`)
		if i < 0 {
			break
		}
		s = s[i+1:]
		j := strings.Index(s, `"`)
		if j < 0 {
			break
		}
		fields = append(fields, s[:j])
		s = s[j+1:]
	}
	return fields
}

func leadingWS(line string) string {
	i := 0
	for i < len(line) && (line[i] == '\t' || line[i] == ' ') {
		i++
	}
	return line[:i]
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func fixACL(path string) {
	if path == "" || !fileExists(path) {
		return
	}
	_ = os.Chmod(path, 0o666)
	runHidden("icacls", path, "/grant", "*S-1-5-32-545:(F)", "/Q")
	runHidden("icacls", path, "/grant", "*S-1-5-11:(F)", "/Q")
	runHidden("icacls", path, "/setintegritylevel", "M", "/Q")
}
