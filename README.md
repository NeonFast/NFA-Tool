# NFA Tool

Windows desktop utility for Steam **ConnectCache** token login.

**This repository root is the main application** (Go + Wails v3 + Svelte 5).

> **AI notice:** code was polished / ported with AI assistance. See [NOTICE.md](NOTICE.md).

## Features

- Login via `login----eya_token`
- Saved accounts list + token TTL
- Optional “keep existing Steam accounts”
- RU / EN UI (auto-detect + manual switch)
- Reset Steam (config / userdata)
- Requires Administrator (UAC)
- Single native `.exe` (no Python runtime)

## Requirements

- Windows 10/11 x64
- [Go](https://go.dev/dl/) 1.24+
- Node.js 20+
- Wails v3 CLI:

```powershell
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
```

## Build

```powershell
git clone <your-repo-url>
cd <repo>
wails3 build
```

Output: `bin\NFA-Tool.exe`

Dev:

```powershell
wails3 dev
```

## Usage

1. Run **as Administrator** (UAC).
2. Paste: `username----eya_jwt_token`
3. Press **Login**.
4. Guide: https://teletype.in/@hackerdlc/CS2NFA

## Layout

```
.                     ← main app (v3)
appservice.go
main.go
internal/steam/       DPAPI, VDF, login
internal/token/
internal/storage/
frontend/             Svelte UI + i18n
build/                Wails packaging
legacy/               old Python tool + Wails v2 prototype
```

## Legacy

Previous versions live under [`legacy/`](legacy/):

| Path | What |
|------|------|
| `legacy/GUI.py` | Original Python app |
| `legacy/nfa-tool/` | Wails v2 prototype |
| `legacy/binaries/` | Old packaged exe (not in git if large) |

## Security

- Tokens may be stored in `user_backup.json` next to the exe — **gitignored**.
- Do not commit tokens or logs.

## License

MIT — see [LICENSE](LICENSE).

Unofficial tool, not affiliated with Valve / Steam.
