# NFA Tool Recode v2

Windows desktop app for Steam **session login via refresh token** (ConnectCache).

**Stack:** Go · [Wails v3](https://v3.wails.io/) · Svelte 5 · SQLite  

**Version:** 2.0.0

> See [NOTICE.md](NOTICE.md) for **inspiration credits**, AI note, and security.

## Features

- Login with `username----eya_jwt` (or marketplace-style multi-segment keys)
- Saved accounts in local **SQLite** (`accounts.db`) + token expiry display
- Optional **keep other Steam logins** (surgical ConnectCache / loginusers merge)
- RU / EN UI (system language + manual switch)
- Native success/error dialogs
- Reset Steam (config / userdata / local cache)
- Runs elevated for config writes; starts Steam **unelevated**
- Single native `NFA-Tool-Recode-v2.exe`

## Requirements

| | |
|---|---|
| OS | Windows 10/11 x64 |
| Steam | Installed; open Steam **once** so `config.vdf` + `loginusers.vdf` exist |
| Build | Go 1.24+, Node 20+, [Wails v3 CLI](https://v3.wails.io/getting-started/installation/) |

```powershell
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
```

## Build (production)

```powershell
git clone <your-repo-url>
cd <repo>
wails3 build
```

Output: **`bin\NFA-Tool-Recode-v2.exe`**

Dev:

```powershell
wails3 dev
```

## Usage

1. Run as **Administrator**.
2. Open Steam once on this PC if you never did, then close it.
3. Paste key: `login----token`
4. Press **Login**.
5. Guide: https://teletype.in/@hackerdlc/CS2NFA

Optional: **Keep other Steam accounts** — leaves your previous Steam logins available.

## Project layout

```
main.go / appservice.go
internal/steam/
internal/token/
internal/storage/
frontend/
build/
legacy/                 archive only
.github/workflows/
```

## Security

- Treat `accounts.db` like a password file.
- Do not commit tokens, DB files, or logs.
- Unofficial tool — use only with accounts/tokens you are allowed to use.

## License

MIT — see [LICENSE](LICENSE).

**Not affiliated with Valve / Steam.**
