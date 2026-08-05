# NFA Tool Recode v2

Windows desktop app for Steam **session login via refresh token** (ConnectCache).

**Stack:** Go · [Wails v3](https://v3.wails.io/) · Svelte 5 · SQLite  

**Version:** 2.1.0

> See [NOTICE.md](NOTICE.md) for **inspiration credits**, AI note, and security.

## Features

- Login with `login----token` (marketplace multi-segment keys supported)
- Saved accounts in local **SQLite** (`accounts.db`) with expiry display
- **Import** one key from the field, or **bulk import** from a `.txt` file (`login----token` per line)
- **Export** selected / all / per-row → clipboard, file, or **Google Drive**
- Drag-select accounts (paint checkboxes with the mouse)
- Optional **keep other Steam logins** (surgical ConnectCache / loginusers merge)
- RU / EN UI (system language + Settings)
- Settings: language, Steam options, Google Drive OAuth, updates, Reset Steam
- Native success/error dialogs
- **Update checker** (GitHub Releases) + one-click install/restart
- Markdown release notes in the update dialog
- Reset Steam (config / userdata)
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

1. Bump `AppVersion` in `appservice.go` (and `build/config.yml` / `info.json`) to match the tag.
2. Commit & push `main`.
3. Tag and push:

```powershell
git tag v2.1.0
git push origin v2.1.0
```

GitHub Actions (`.github/workflows/release.yml`) builds Windows exe and publishes a Release with  
`NFA-Tool-Recode-v2-windows-amd64.exe` + `SHA256SUMS.txt`.

Pushes/PRs to `main` also run `.github/workflows/windows-build.yml` (artifact only).

## Usage

1. Run as **Administrator**.
2. Open Steam once on this PC if you never did, then close it.
3. Paste a key: `login----token` → **Login** (or **Import** to save only).
4. Bulk: **Import from file…** — same format as export (`login----token` lines).
5. Export: select accounts (click or drag) → **Export** → clipboard / file / Google Drive.
6. Google Drive: **Settings** → configure OAuth once (see in-app guide), then pick Drive on export.
7. Full Steam guide: https://teletype.in/@hackerdlc/CS2NFA

Optional: **Keep other Steam accounts** — previous Steam logins stay available.

## Project layout

```
main.go / appservice.go
internal/steam/
internal/token/
internal/storage/
internal/gdrive/
internal/update/
frontend/
build/
.github/workflows/
```

## Security

- Tokens in `accounts.db` are **DPAPI-encrypted** (Windows user + app entropy). Stealing the file alone is not enough.
- Same-user malware can still decrypt — this is not antivirus.
- Google OAuth tokens are stored sealed next to the app (`gdrive-token.sealed`); client credentials in `google-oauth.json`.
- Hot-update from GitHub Releases (in-app).
- Do not commit DB files, OAuth secrets, or logs.
- Unofficial tool — use only with accounts/tokens you are allowed to use.

## License

MIT — see [LICENSE](LICENSE).

**Not affiliated with Valve / Steam.**
