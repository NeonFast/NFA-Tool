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

1. Run **as Administrator** (UAC prompt).
2. If Steam was never opened on this PC — launch Steam once, then close it.
3. Paste key: `login----token`
4. Press **Login** — Steam restarts with that session.
5. Guide: https://teletype.in/@hackerdlc/CS2NFA

### Keep existing

- **On** — merge into existing ConnectCache / loginusers (other cached sessions kept).
- **Off** — still uses surgical writes; does not wipe the whole Steam install.

Accounts in the app list are always stored in `accounts.db` regardless of the checkbox.

## Project layout

```
main.go / appservice.go     Wails app + API
internal/steam/             Steam path, DPAPI, VDF surgery, login, launch
internal/token/             JWT parse / TTL
internal/storage/           SQLite accounts.db
frontend/                   Svelte UI + i18n
build/                      Wails packaging / Windows metadata
legacy/                     Old Python / Wails v2 (archive only)
.github/workflows/          Windows CI build
```

## How login works (high level)

1. Stop `steam.exe` / `steamwebhelper.exe`
2. Encrypt refresh token with DPAPI (`BObfuscateBuffer` + account entropy)
3. Upsert hex blob into `%LOCALAPPDATA%\Steam\local.vdf` → `ConnectCache`
4. Mark user active in `Steam\config\loginusers.vdf`
5. Ensure `Accounts` entry in `config.vdf` (no full rewrite)
6. Set `HKCU\SOFTWARE\Valve\Steam\AutoLoginUser`
7. Start `steam.exe` detached / unelevated

## Production checklist

- [x] Release build: `wails3 build` (`production`, stripped, GUI subsystem)
- [x] UAC manifest: `requireAdministrator`
- [x] Version metadata: `build/config.yml`, `build/windows/info.json`, `AppVersion`
- [x] Secrets gitignored: `accounts.db`, `*.db-*`, tokens, logs
- [x] CI: `.github/workflows/windows-build.yml`
- [x] License: MIT
- [x] Attribution: [NOTICE.md](NOTICE.md)

## Security

- Treat `accounts.db` like a password file.
- Do not commit tokens, DB files, or logs.
- Unofficial tool — use only with accounts/tokens you are allowed to use.

## License

MIT — see [LICENSE](LICENSE).

**Not affiliated with Valve / Steam.**
