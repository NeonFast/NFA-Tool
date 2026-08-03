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

type LoginOptions struct {
	AccountName  string
	Token        string
	SteamID      string
	KeepExisting bool // always merge-safe now (surgical edits)
}

// Login injects a refresh token into the Steam client session files.
// Approach: surgical VDF edits (inspired by public switchers; reimplemented here).
//
//  1. Require Steam config already exists (open Steam once first)
//  2. Kill steam.exe / steamwebhelper
//  3. Patch loginusers.vdf (active user)
//  4. Upsert ConnectCache in local.vdf (keep other tokens)
//  5. Ensure Accounts entry in config.vdf (no full rewrite)
//  6. Set AutoLoginUser
//  7. Launch steam.exe detached / unelevated
func Login(opt LoginOptions) error {
	account := strings.ToLower(strings.TrimSpace(opt.AccountName))
	if i := strings.Index(account, "@"); i >= 0 {
		account = account[:i]
	}
	token := strings.TrimSpace(opt.Token)
	steamID := strings.TrimSpace(opt.SteamID)
	if account == "" || token == "" || steamID == "" {
		return fmt.Errorf("account, token and steam id are required")
	}

	install, err := GetSteamInstallPathNoKill()
	if err != nil {
		return err
	}
	configDir := filepath.Join(install, "config")
	configPath := filepath.Join(configDir, "config.vdf")
	usersPath := filepath.Join(configDir, "loginusers.vdf")
	localDir := filepath.Join(os.Getenv("LOCALAPPDATA"), "Steam")
	if os.Getenv("LOCALAPPDATA") == "" {
		localDir = filepath.Join(os.Getenv("localappdata"), "Steam")
	}
	localPath := filepath.Join(localDir, "local.vdf")

	// Roster: fail if Steam never created configs
	if !fileExists(configPath) || !fileExists(usersPath) {
		return fmt.Errorf("open Steam and sign in once so it can create its config files, then try again")
	}

	// stop client
	_ = KillSteam()

	// 1) loginusers — line surgery
	if err := setLoginUsersActive(usersPath, account, steamID); err != nil {
		return fmt.Errorf("loginusers.vdf: %w", err)
	}

	// 2) local.vdf ConnectCache — upsert only
	if err := storeConnectCacheToken(localPath, account, token); err != nil {
		return fmt.Errorf("local.vdf: %w", err)
	}

	// 3) config.vdf Accounts inject if needed
	_ = injectConfigAccount(configPath, account, steamID)

	// 4) registry
	if err := SetAutoLoginUser(account); err != nil {
		return fmt.Errorf("AutoLoginUser: %w", err)
	}

	// ACL so medium-IL Steam can read what elevated process wrote
	fixACL(usersPath)
	fixACL(localPath)
	fixACL(configPath)

	time.Sleep(400 * time.Millisecond)
	return LaunchSteam(install)
}

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
	// map store_key -> encrypted
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

func AccountCRCKey(account string) string {
	crc := crc32.ChecksumIEEE([]byte(account))
	hex := fmt.Sprintf("%08x", crc)
	trimmed := strings.TrimLeft(hex, "0")
	if trimmed == "" {
		return "01"
	}
	return trimmed + "1"
}

// --- surgical loginusers (roster set_active) ---

func setLoginUsersActive(path, username, steamID string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	s := string(content)
	// demote all MostRecent
	s = strings.ReplaceAll(s, `"MostRecent"		"1"`, `"MostRecent"		"0"`)
	s = strings.ReplaceAll(s, "\"MostRecent\"\t\t\"1\"", "\"MostRecent\"\t\t\"0\"")
	// demote AutoLogin too (live Steam)
	s = strings.ReplaceAll(s, `"AutoLogin"		"1"`, `"AutoLogin"		"0"`)
	s = strings.ReplaceAll(s, "\"AutoLogin\"\t\t\"1\"", "\"AutoLogin\"\t\t\"0\"")
	s = strings.ReplaceAll(s, `"AllowAutoLogin"		"1"`, `"AllowAutoLogin"		"0"`)
	s = strings.ReplaceAll(s, "\"AllowAutoLogin\"\t\t\"1\"", "\"AllowAutoLogin\"\t\t\"0\"")

	if strings.Contains(s, `"`+steamID+`"`) {
		s = refreshUserBlock(s, username, steamID)
	} else {
		block := newUserBlock(username, steamID)
		// insert before last closing brace of file
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
			// rewrite body until closing }
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
			// ensure required fields before last }
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
		// keep existing persona if present — only set if empty-ish; roster keeps/overwrites account name style
		return line // preserve persona from Steam
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

// --- ConnectCache surgical upsert (roster store_token) ---

func storeConnectCacheToken(path, username, token string) error {
	key := AccountCRCKey(username)
	encrypted, err := EncryptToken(token, username)
	if err != nil {
		return err
	}
	_ = os.MkdirAll(filepath.Dir(path), 0o755)

	existing, err := os.ReadFile(path)
	if err != nil {
		// fresh file only if missing
		return os.WriteFile(path, []byte(freshLocalVDF(key, encrypted)), 0o644)
	}
	content := string(existing)
	if updated, ok := upsertConnectCacheEntry(content, key, encrypted); ok {
		return os.WriteFile(path, []byte(updated), 0o644)
	}
	if updated, ok := insertConnectCacheBlock(content, key, encrypted); ok {
		return os.WriteFile(path, []byte(updated), 0o644)
	}
	// last resort: do NOT wipe other tokens — refuse
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
		// indent one more tab than Steam's brace line
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

// injectConfigAccount — shefu style: only add Accounts entry if SteamID missing
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
