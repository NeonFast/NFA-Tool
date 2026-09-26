package steam

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"nfa-tool/internal/token"
)

// TokenCheck is one named verification of the refresh token.
// Status is "ok", "fail" or "skip" (could not be evaluated).
type TokenCheck struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// AccountInfo aggregates everything learnable about an account from its
// refresh token and public Steam endpoints (no proxy, no API key).
type AccountInfo struct {
	SteamID       string `json:"steamId"`
	PersonaName   string `json:"personaName"`
	RealName      string `json:"realName,omitempty"`
	AvatarFull    string `json:"avatarFull"`
	ProfileURL    string `json:"profileUrl"`
	Visibility    string `json:"visibility"`
	OnlineState   string `json:"onlineState"`
	InGame        string `json:"inGame,omitempty"`
	Location      string `json:"location,omitempty"`
	Summary       string `json:"summary,omitempty"`
	FriendsCount  int    `json:"friendsCount"`
	TopGame       string `json:"topGame,omitempty"`
	TopGameHours  string `json:"topGameHours,omitempty"`
	Level         string `json:"level,omitempty"`
	GamesCount    string `json:"gamesCount,omitempty"`
	BanInfo       string `json:"banInfo,omitempty"`
	BanDays       string `json:"banDays,omitempty"`
	CS2Items      int    `json:"cs2Items"`
	CS2Rarity     string `json:"cs2Rarity,omitempty"`
	CS2Medals     string `json:"cs2Medals,omitempty"`
	CS2Hours      string `json:"cs2Hours,omitempty"`
	CS2Prime      string `json:"cs2Prime,omitempty"`
	InventoryValue   string `json:"inventoryValue,omitempty"`
	InventoryPartial bool   `json:"inventoryPartial,omitempty"`
	LicensesCount int    `json:"licensesCount"`
	Email         string `json:"email,omitempty"`
	Wallet        string `json:"wallet,omitempty"`
	Country       string `json:"country,omitempty"`
	VACBanned     bool   `json:"vacBanned"`
	TradeBan      string `json:"tradeBan"`
	Limited       bool   `json:"limited"`
	MemberSince   string `json:"memberSince"`
	TokenAlive    bool   `json:"tokenAlive"`
	ProfileErr    string `json:"profileErr,omitempty"`

	TokenIssued     string       `json:"tokenIssued,omitempty"`
	TokenExpires    string       `json:"tokenExpires,omitempty"`
	TokenDaysLeft   int          `json:"tokenDaysLeft"`
	TokenAudiences  string       `json:"tokenAudiences,omitempty"`
	TokenSteamID    string       `json:"tokenSteamId,omitempty"`
	TokenClaimsJSON string       `json:"tokenClaimsJson,omitempty"`
	TokenChecks     []TokenCheck `json:"tokenChecks"`

	profileID string
}
var accHTTP = &http.Client{Timeout: 12 * time.Second}

const dateLayout = "2006-01-02 15:04 UTC"

var reSteamID64 = regexp.MustCompile(`^\d{17}$`)

func check(id string, ok bool) TokenCheck {
	if ok {
		return TokenCheck{ID: id, Status: "ok"}
	}
	return TokenCheck{ID: id, Status: "fail"}
}

func checkOrSkip(id string, known, ok bool) TokenCheck {
	if !known {
		return TokenCheck{ID: id, Status: "skip"}
	}
	return check(id, ok)
}

// FetchAccountInfo never fails hard: sources are independent and fetched in
// parallel; each goroutine writes its own fields, partial data is fine.
func FetchAccountInfo(refreshToken, steamID string) *AccountInfo {
	info := &AccountInfo{SteamID: steamID}

	var static []TokenCheck
	claims, claimsErr := token.DecodeClaims(refreshToken)
	if claimsErr != nil {
		static = append(static, TokenCheck{ID: "jwt_structure", Status: "fail"})
	} else {
		now := time.Now()
		parts := strings.Split(refreshToken, ".")
		static = append(static, check("signature", len(parts) == 3 && parts[2] != ""))

		iss, _ := claims["iss"].(string)
		static = append(static, check("issuer", strings.EqualFold(iss, "steam")))

		aud := token.ClaimStrings(claims, "aud")
		hasClient := false
		for _, a := range aud {
			if strings.Contains(a, "client") {
				hasClient = true
			}
		}
		static = append(static, check("audience", hasClient))
		info.TokenAudiences = strings.Join(aud, ", ")

		sub, _ := claims["sub"].(string)
		info.TokenSteamID = sub
		static = append(static, check("steamid_format", reSteamID64.MatchString(sub)))

		expT, hasExp := token.ClaimTime(claims, "exp")
		iatT, hasIat := token.ClaimTime(claims, "iat")
		if hasExp {
			info.TokenExpires = expT.Format(dateLayout)
			if days := int(time.Until(expT).Hours() / 24); days > 0 {
				info.TokenDaysLeft = days
			}
		}
		if hasIat {
			info.TokenIssued = iatT.Format(dateLayout)
		}
		static = append(static, checkOrSkip("not_expired", hasExp, expT.After(now)))
		static = append(static, checkOrSkip("iat_past", hasIat, !iatT.After(now.Add(5*time.Minute))))
		static = append(static, checkOrSkip("iat_before_exp", hasIat && hasExp, iatT.Before(expT)))
		nbfT, hasNbf := token.ClaimTime(claims, "nbf")
		static = append(static, checkOrSkip("nbf_ok", hasNbf, !nbfT.After(now.Add(5*time.Minute))))
		rtExpT, hasRtExp := token.ClaimTime(claims, "rt_exp")
		static = append(static, checkOrSkip("rt_exp_ok", hasRtExp, rtExpT.After(now)))

		if pretty, err := json.MarshalIndent(claims, "", "  "); err == nil {
			info.TokenClaimsJSON = string(pretty)
		}
	}

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		if err := fillCommunityProfile(info, steamID); err != nil {
			info.ProfileErr = err.Error()
		} else {
			fillProfilePage(info, steamID)
		}
	}()
	go func() {
		defer wg.Done()
		fillCS2(info, steamID)
	}()
	go func() {
		defer wg.Done()
		if eres, err := CheckTokenCM(refreshToken, 15*time.Second); err == nil && eres == 1 {
			info.TokenAlive = true
		}
	}()
	wg.Wait()

	static = append(static,
		check("access_minted", info.TokenAlive),
		checkOrSkip("profile_match",
			info.profileID != "" && info.TokenSteamID != "",
			info.profileID == info.TokenSteamID),
	)
	info.TokenChecks = static
	return info
}

type xmlPlayedGame struct {
	Name  string `xml:"gameName"`
	Hours string `xml:"hoursPlayed"`
}

type xmlProfile struct {
	SteamID64    string          `xml:"steamID64"`
	PersonaName  string          `xml:"steamID"`
	RealName     string          `xml:"realname"`
	OnlineState  string          `xml:"onlineState"`
	PrivacyState string          `xml:"privacyState"`
	AvatarFull   string          `xml:"avatarFull"`
	CustomURL    string          `xml:"customURL"`
	MemberSince  string          `xml:"memberSince"`
	Location     string          `xml:"location"`
	Summary      string          `xml:"summary"`
	InGameName   string          `xml:"inGameInfo>gameName"`
	Friends      []string        `xml:"friends>friend"`
	MostPlayed   []xmlPlayedGame `xml:"mostPlayedGames>mostPlayedGame"`
	VACBanned    int             `xml:"vacBanned"`
	TradeBan     string          `xml:"tradeBanState"`
	Limited      int             `xml:"isLimitedAccount"`
}

// FetchProfileBasics fetches just the avatar URL and persona name from the
// public XML profile. Cheap single-request probe for background prefetching.
func FetchProfileBasics(steamID string) (avatarURL, persona string) {
	u := fmt.Sprintf("https://steamcommunity.com/profiles/%s?xml=1", url.PathEscape(steamID))
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return "", ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := accHTTP.Do(req)
	if err != nil {
		return "", ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", ""
	}
	var p xmlProfile
	if err := xml.NewDecoder(resp.Body).Decode(&p); err != nil {
		return "", ""
	}
	if p.SteamID64 == "" {
		return "", ""
	}
	return p.AvatarFull, p.PersonaName
}

func fillCommunityProfile(info *AccountInfo, steamID string) error {
	u := fmt.Sprintf("https://steamcommunity.com/profiles/%s?xml=1", url.PathEscape(steamID))
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := accHTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("profile http %d", resp.StatusCode)
	}
	var p xmlProfile
	if err := xml.NewDecoder(resp.Body).Decode(&p); err != nil {
		return err
	}
	if p.SteamID64 == "" {
		return fmt.Errorf("profile not found")
	}
	info.profileID = p.SteamID64
	info.PersonaName = p.PersonaName
	info.RealName = p.RealName
	info.AvatarFull = p.AvatarFull
	info.OnlineState = p.OnlineState
	info.Visibility = p.PrivacyState
	info.VACBanned = p.VACBanned == 1
	info.TradeBan = p.TradeBan
	info.Limited = p.Limited == 1
	info.MemberSince = p.MemberSince
	info.Location = p.Location
	info.Summary = strings.TrimSpace(p.Summary)
	info.InGame = p.InGameName
	info.FriendsCount = len(p.Friends)
	if len(p.MostPlayed) > 0 {
		info.TopGame = p.MostPlayed[0].Name
		info.TopGameHours = p.MostPlayed[0].Hours
	}
	if p.CustomURL != "" {
		info.ProfileURL = "https://steamcommunity.com/id/" + p.CustomURL
	} else {
		info.ProfileURL = "https://steamcommunity.com/profiles/" + steamID
	}
	return nil
}

var (
	reLevel       = regexp.MustCompile(`friendPlayerLevelNum">\s*(\d+)`)
	reGamesCount  = regexp.MustCompile(`(?is)count_link_label">Games</span>.*?profile_count_link_total">\s*([\d,]+)`)
	reBanInfo     = regexp.MustCompile(`(?i)(\d+)\s+((?:VAC|game)\s+bans?)\s+on record`)
	reBanDays     = regexp.MustCompile(`(?is)(\d[\d,]*)\s+day\(s\) since last ban`)
	reCS2Hours    = regexp.MustCompile(`(?s)"appid"\s*:\s*730\s*,.*?"hours_forever"\s*:\s*"([\d,\.]+)"`)
)

func fillProfilePage(info *AccountInfo, steamID string) {
	body := fetchBody(fmt.Sprintf("https://steamcommunity.com/profiles/%s/", url.PathEscape(steamID)), nil)
	if body == "" {
		return
	}
	if m := reLevel.FindStringSubmatch(body); m != nil {
		info.Level = m[1]
	}
	if m := reGamesCount.FindStringSubmatch(body); m != nil {
		info.GamesCount = m[1]
	}
	if m := reBanInfo.FindStringSubmatch(body); m != nil {
		info.BanInfo = m[1] + " " + m[2]
	}
	if m := reBanDays.FindStringSubmatch(body); m != nil {
		info.BanDays = m[1]
	}
}

type invTag struct {
	Category string `json:"category"`
	Name     string `json:"localized_tag_name"`
}

type invDescription struct {
	Name           string   `json:"name"`
	MarketHashName string   `json:"market_hash_name"`
	Marketable     int      `json:"marketable"`
	Tags           []invTag `json:"tags"`
}

type invResponse struct {
	Total        int              `json:"total_inventory_count"`
	Descriptions []invDescription `json:"descriptions"`
}

var rarityOrder = []string{"Extraordinary", "Covert", "Classified", "Restricted", "Mil-Spec Grade"}

func fillCS2(info *AccountInfo, steamID string) {
	body := fetchBody(fmt.Sprintf(
		"https://steamcommunity.com/inventory/%s/730/2?l=english&count=500", url.PathEscape(steamID)), nil)
	if body != "" {
		var inv invResponse
		if json.Unmarshal([]byte(body), &inv) == nil {
			if inv.Total > 0 {
				info.CS2Items = inv.Total
				info.CS2Rarity = summarizeRarities(inv.Descriptions)
			}
			info.CS2Medals = collectMedals(inv.Descriptions)
			fillInventoryValue(info, inv.Descriptions)
		}
	}
	games := fetchBody(fmt.Sprintf("https://steamcommunity.com/profiles/%s/games/?tab=all", url.PathEscape(steamID)), nil)
	if m := reCS2Hours.FindStringSubmatch(games); m != nil {
		info.CS2Hours = m[1]
	}
}

func collectMedals(descs []invDescription) string {
	var out []string
	for _, d := range descs {
		for _, tg := range d.Tags {
			if tg.Category != "Type" {
				continue
			}
			l := strings.ToLower(tg.Name)
			for _, m := range []string{"medal", "coin", "trophy", "collectible", "pin"} {
				if strings.Contains(l, m) {
					out = append(out, d.Name)
					break
				}
			}
		}
	}
	return strings.Join(out, " · ")
}

const marketPriceCap = 60

var rePriceNum = regexp.MustCompile(`[\d.,]+`)

func parseMarketPrice(s string) float64 {
	num := rePriceNum.FindString(s)
	num = strings.ReplaceAll(num, ",", "")
	f, _ := strconv.ParseFloat(num, 64)
	return f
}

func fetchMarketPrice(hashName string) (float64, bool) {
	u := "https://steamcommunity.com/market/priceoverview/?appid=730&currency=1&market_hash_name=" +
		url.QueryEscape(hashName)
	body := fetchBody(u, nil)
	if body == "" {
		return 0, false
	}
	var out struct {
		Success     bool   `json:"success"`
		LowestPrice string `json:"lowest_price"`
	}
	if json.Unmarshal([]byte(body), &out) != nil || !out.Success || out.LowestPrice == "" {
		return 0, false
	}
	return parseMarketPrice(out.LowestPrice), true
}

func fillInventoryValue(info *AccountInfo, descs []invDescription) {
	seen := map[string]bool{}
	var names []string
	for _, d := range descs {
		if d.Marketable != 1 || d.MarketHashName == "" || seen[d.MarketHashName] {
			continue
		}
		seen[d.MarketHashName] = true
		names = append(names, d.MarketHashName)
	}
	if len(names) == 0 {
		return
	}
	if len(names) > marketPriceCap {
		names = names[:marketPriceCap]
		info.InventoryPartial = true
	}

	var mu sync.Mutex
	var sum float64
	var priced int
	var stop atomic.Bool
	var wg sync.WaitGroup
	sem := make(chan struct{}, 4)
	for _, n := range names {
		if stop.Load() {
			break
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(name string) {
			defer wg.Done()
			defer func() { <-sem }()
			price, ok := fetchMarketPrice(name)
			if !ok {
				stop.Store(true)
				return
			}
			mu.Lock()
			sum += price
			priced++
			mu.Unlock()
		}(n)
	}
	wg.Wait()

	if priced > 0 {
		info.InventoryValue = fmt.Sprintf("$%.2f", sum)
		if priced < len(names) {
			info.InventoryPartial = true
		}
	}
}

func summarizeRarities(descs []invDescription) string {
	rarity := map[string]int{}
	knives, gloves := 0, 0
	for _, d := range descs {
		for _, tag := range d.Tags {
			switch tag.Category {
			case "Rarity":
				rarity[tag.Name]++
			case "Type":
				if strings.Contains(tag.Name, "Knife") {
					knives++
				}
				if strings.Contains(tag.Name, "Gloves") {
					gloves++
				}
			}
		}
	}
	var parts []string
	for _, name := range rarityOrder {
		if n := rarity[name]; n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", n, name))
		}
	}
	if knives > 0 {
		parts = append(parts, fmt.Sprintf("%d Knife", knives))
	}
	if gloves > 0 {
		parts = append(parts, fmt.Sprintf("%d Gloves", gloves))
	}
	return strings.Join(parts, " · ")
}

func fetchBody(u string, cookies []*http.Cookie) string {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	resp, err := accHTTP.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return ""
	}
	return string(raw)
}

// APIReachable reports whether the Steam Web API answers (quick connectivity check).
func APIReachable() bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.steampowered.com/ISteamWebAPIUtil/GetSupportedAPIList/v1/")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
