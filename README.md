# NFA Tool

Windows desktop utility for Steam **ConnectCache** token login.

Stack: **Go** + **Wails v3** + **Svelte 5** (TypeScript).

> **AI notice:** code was polished / ported with AI assistance. See [NOTICE.md](NOTICE.md).

## Features

- Login via `login----eya_token`
- Saved accounts list + token TTL
- Optional “keep existing Steam accounts”
- RU / EN UI (auto-detect system language + manual switch)
- Reset Steam (config / userdata)
- Requires Administrator (UAC)
- Single native `.exe` (no Python runtime)

## Requirements

- Windows 10/11 x64
- [Go](https://go.dev/dl/) 1.24+
- Node.js 20+ (frontend build)
- [Wails v3 CLI](https://v3.wails.io/getting-started/installation/):

```powershell
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
```

- WebView2 (usually preinstalled on Windows 11)

## Build

```powershell
git clone <your-repo-url>
cd nfa-tool   # or repo root
wails3 build
```

Output: `bin\nfa-tool-v3.exe` (name from `Taskfile.yml` → `APP_NAME`).

Dev mode:

```powershell
wails3 dev
```

## Usage

1. Run the exe **as Administrator** (UAC prompt).
2. Paste key: `username----eya_jwt_token`
3. Press **Login** — Steam restarts with the injected session.
4. Guide: https://teletype.in/@hackerdlc/CS2NFA

## Project layout

```
appservice.go          Wails service API
main.go                App entry + window
elevate_*.go           Admin elevation (Windows)
internal/steam/        DPAPI, VDF, path, login (core logic)
internal/token/        JWT parse/validate
internal/storage/      user_backup.json
frontend/src/          Svelte UI + i18n
build/                 Wails packaging assets
```

## Configuration

| File | Purpose |
|------|---------|
| `build/config.yml` | Product metadata / dev mode |
| `build/windows/info.json` | Windows version resource |
| `build/windows/wails.exe.manifest` | UAC `requireAdministrator` |
| `appservice.go` → `AppVersion` | In-app version string |

## Security notes

- Tokens can be saved in `user_backup.json` beside the exe — **gitignored**.
- JWT signature is not verified (same class of tool as the original).
- Use only with accounts/tokens you are allowed to use.

## License

MIT — see [LICENSE](LICENSE).

## Disclaimer

Unofficial tool, not affiliated with Valve / Steam. Use at your own risk.
