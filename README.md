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
git clone https://github.com/NeonFast/NFA-Tool.git
cd NFA-Tool
wails3 build
```

Output: **`bin\NFA-Tool-Recode-v2.exe`**

Dev:

```powershell
wails3 dev
```

## Auto releases (GitHub Actions)

After the repo is on GitHub, **you don’t upload the exe by hand**.

1. Bump version in code if you want (`appservice.go` → `AppVersion`, optional).
2. Commit & push `main`.
3. Create and push a tag:

```powershell
git tag v2.0.1
git push origin v2.0.1
```

GitHub Actions (`.github/workflows/release.yml`) will:

- build Windows exe on `windows-latest`
- create a **Release** named like `NFA Tool Recode v2 2.0.1`
- attach `NFA-Tool-Recode-v2-windows-amd64.exe` + `SHA256SUMS.txt`

Also: every push/PR to `main` runs `.github/workflows/windows-build.yml` and uploads a build **artifact** (not a full Release).

Manual run: Actions → **Release** → Run workflow (optional tag input).

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
.github/workflows/
```

## Security

- Treat `accounts.db` like a password file.
- Do not commit tokens, DB files, or logs.
- Unofficial tool — use only with accounts/tokens you are allowed to use.

## License

MIT — see [LICENSE](LICENSE).

**Not affiliated with Valve / Steam.**
