package steam

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/coder/websocket"

	"nfa-tool/internal/token"
)

const (
	emsgMulti               = 1
	emsgClientLogOnResponse = 751
	emsgClientLoggedOff     = 757
	emsgClientLogon         = 5514

	protoMask            = 0x80000000
	logonProtocolVersion = 65580
	osTypeWindows10      = 16
)

// CheckTokenCM performs a proof logon against a Steam CM server over
// WebSocket using the given refresh token and returns the EResult code from
// CMsgClientLogonResponse. Unlike the GenerateAccessTokenForApp web API, this
// does not consume or rotate the refresh token. EResult 1 (OK) means the
// token is valid; typical failures are 5 (InvalidPassword) for garbage or
// expired tokens and 8 (InvalidParam) for malformed ones.
func CheckTokenCM(refreshToken string, timeout time.Duration) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	addrs, err := pickCMWebSockets(ctx)
	if err != nil {
		return 0, err
	}

	var conn *websocket.Conn
	var lastErr error
	for _, addr := range addrs {
		conn, lastErr = dialCM(ctx, addr)
		if conn != nil {
			break
		}
		if ctx.Err() != nil {
			break
		}
	}
	if conn == nil {
		return 0, lastErr
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")
	conn.SetReadLimit(1 << 20)

	if err := conn.Write(ctx, websocket.MessageBinary, buildClientLogon(refreshToken)); err != nil {
		return 0, fmt.Errorf("cm write logon: %w", err)
	}

	for {
		mt, data, err := conn.Read(ctx)
		if err != nil {
			return 0, fmt.Errorf("cm read: %w", err)
		}
		if mt != websocket.MessageBinary {
			continue
		}
		if os.Getenv("CMPROBE_DEBUG") != "" {
			emsg := uint32(0)
			if len(data) >= 4 {
				emsg = binary.LittleEndian.Uint32(data) &^ protoMask
			}
			fmt.Fprintf(os.Stderr, "cm rx emsg=%d len=%d hex=%x\n", emsg, len(data), data)
		}
		eresult, done, err := handleCMMessage(data)
		if err != nil {
			return 0, err
		}
		if done {
			return eresult, nil
		}
	}
}

type cmListResponse struct {
	Response struct {
		ServerlistWebsockets []string `json:"serverlist_websockets"`
	} `json:"response"`
}

func dialCM(ctx context.Context, addr string) (*websocket.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		host, port = addr, "27018"
	}
	candidates := []string{net.JoinHostPort(host, "443")}
	if port != "443" {
		candidates = append(candidates, net.JoinHostPort(host, port))
	}
	var lastErr error
	for _, c := range candidates {
		attempt, cancel := context.WithTimeout(ctx, 6*time.Second)
		conn, _, err := websocket.Dial(attempt, "wss://"+c+"/cmsocket/", nil)
		cancel()
		if err == nil {
			return conn, nil
		}
		lastErr = fmt.Errorf("cm dial %s: %w", c, err)
		if ctx.Err() != nil {
			break
		}
	}
	return nil, lastErr
}

func pickCMWebSockets(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://api.steampowered.com/ISteamDirectory/GetCMList/v1/?cellid=0&maxcount=5", nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get cm list: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read cm list: %w", err)
	}
	var list cmListResponse
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("parse cm list: %w", err)
	}
	if len(list.Response.ServerlistWebsockets) == 0 {
		return nil, fmt.Errorf("cm list has no websocket servers")
	}
	return list.Response.ServerlistWebsockets, nil
}

func buildClientLogon(refreshToken string) []byte {
	var steamID uint64
	if claims, err := token.DecodeClaims(refreshToken); err == nil {
		if sub, _ := claims["sub"].(string); sub != "" {
			steamID, _ = strconv.ParseUint(sub, 10, 64)
		}
	}

	var header []byte
	header = pbAppendFixed64Field(header, 1, steamID)
	header = pbAppendVarintField(header, 2, 0)

	var body []byte
	body = pbAppendVarintField(body, 1, logonProtocolVersion)
	body = pbAppendStringField(body, 6, "english")
	body = pbAppendVarintField(body, 7, osTypeWindows10)
	body = pbAppendVarintField(body, 8, 1)
	body = pbAppendVarintField(body, 33, 2)
	body = pbAppendStringField(body, 96, "cmprobe")
	body = pbAppendVarintField(body, 102, 1)
	body = pbAppendStringField(body, 108, refreshToken)

	msg := make([]byte, 0, 8+len(header)+len(body))
	var prefix [8]byte
	binary.LittleEndian.PutUint32(prefix[0:4], emsgClientLogon|protoMask)
	binary.LittleEndian.PutUint32(prefix[4:8], uint32(len(header)))
	msg = append(msg, prefix[:]...)
	msg = append(msg, header...)
	return append(msg, body...)
}

func handleCMMessage(data []byte) (eresult int, done bool, err error) {
	if len(data) < 8 {
		return 0, false, fmt.Errorf("cm: short message (%d bytes)", len(data))
	}
	rawEMsg := binary.LittleEndian.Uint32(data)
	emsg := rawEMsg &^ protoMask
	if rawEMsg&protoMask == 0 {
		return 0, false, nil
	}
	headerLen := int(binary.LittleEndian.Uint32(data[4:8]))
	if len(data) < 8+headerLen {
		return 0, false, fmt.Errorf("cm: truncated header in emsg %d", emsg)
	}
	headerFields, err := pbParse(data[8 : 8+headerLen])
	if err != nil {
		return 0, false, fmt.Errorf("cm: header parse: %w", err)
	}
	body := data[8+headerLen:]

	switch emsg {
	case emsgMulti:
		fields, err := pbParse(body)
		if err != nil {
			return 0, false, fmt.Errorf("cm: multi parse: %w", err)
		}
		payload, ok := pbFindBytes(fields, 2)
		if !ok {
			return 0, false, nil
		}
		if v, ok := pbFindVarint(fields, 1); ok && v > 0 {
			zr, err := gzip.NewReader(bytes.NewReader(payload))
			if err != nil {
				return 0, false, fmt.Errorf("cm: multi gunzip: %w", err)
			}
			payload, err = io.ReadAll(zr)
			zr.Close()
			if err != nil {
				return 0, false, fmt.Errorf("cm: multi gunzip: %w", err)
			}
		}
		for len(payload) >= 4 {
			n := int(binary.LittleEndian.Uint32(payload))
			payload = payload[4:]
			if n > len(payload) {
				return 0, false, fmt.Errorf("cm: truncated multi payload")
			}
			eresult, done, err := handleCMMessage(payload[:n])
			if err != nil {
				return 0, false, err
			}
			if done {
				return eresult, true, nil
			}
			payload = payload[n:]
		}
		return 0, false, nil
	case emsgClientLogOnResponse:
		fields, err := pbParse(body)
		if err != nil {
			return 0, false, fmt.Errorf("cm: logon response parse: %w", err)
		}
		if v, ok := pbFindVarint(fields, 1); ok {
			return int(int32(v)), true, nil
		}
		if v, ok := pbFindVarint(headerFields, 13); ok {
			return int(int32(v)), true, nil
		}
		return 2, true, nil
	case emsgClientLoggedOff:
		if v, ok := pbFindVarint(headerFields, 13); ok {
			return 0, false, fmt.Errorf("cm: logged off before logon response, eresult %d", int32(v))
		}
		return 0, false, fmt.Errorf("cm: logged off before logon response")
	default:
		return 0, false, nil
	}
}
