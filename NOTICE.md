# NOTICE

## AI-assisted development

Parts of this codebase were **polished, refactored, and ported with assistance from AI** (code review, UI/i18n, project packaging, and Go/Wails scaffolding).

- Core login behavior is based on the original Steam cache / ConnectCache approach.
- AI was used to improve structure, UI, localization, build setup, and documentation — not as a substitute for testing on a real Windows + Steam environment.
- Always review security-sensitive areas (DPAPI, token storage, admin elevation) before shipping your own builds.

## Security

- Refresh tokens may be stored in `user_backup.json` next to the executable.
- Do **not** commit `user_backup.json`, logs, or tokens to Git.
- The app requests **Administrator** rights on Windows (UAC).
