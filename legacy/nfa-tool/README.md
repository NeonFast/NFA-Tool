# NFA Tool (Go)

Production-oriented rewrite of Steam Cache Login in Go + Wails.

## Features

- Dark purple UI (frameless), similar to NFA Tool Rebound
- Login via `account----eya_token`
- Keep existing Steam accounts (merge mode)
- Saved accounts list with token TTL
- Harvest tokens from current Steam ConnectCache
- Reset Steam (config + userdata)
- Single native `.exe` (standalone binary)

## Build

```powershell
# once
go install github.com/wailsapp/wails/v2/cmd/wails@latest

cd nfa-tool
wails build
```

Output: `build\bin\NFA-Tool.exe`

Dev mode:

```powershell
wails dev
```

## Project layout

```
internal/steam/    DPAPI, VDF, path, login, reset
internal/token/    JWT parse/validate
internal/storage/  user_backup.json
app.go             Wails API
frontend/          UI (HTML/CSS/JS)
```

## Notes

- Windows only (DPAPI + registry)
- JWT signature is not verified
- Tokens stored in `user_backup.json` next to the exe
