# NOTICE

## Credits & inspiration

This project is an **independent implementation** (Go + Wails v3 + Svelte).

The **general idea** of writing a Steam refresh token into the client session
store (ConnectCache / `local.vdf`, `loginusers.vdf`, registry `AutoLoginUser`)
is shared by several public tools and community write-ups. In particular, the
**surgical edit** approach (patch existing VDF instead of rewriting whole files)
was informed by studying open-source account switchers such as:

- [kWAYTV/roster](https://github.com/kWAYTV/roster) — Steam refresh-token switcher (Rust/Tauri)
- [shefu223/nfa-tool](https://github.com/shefu223/nfa-tool) — lightweight token loader (Rust/Tauri)

We did **not** copy their codebases wholesale. Logic was reimplemented for this
stack, verified against a live Steam install, and adapted (admin elevation,
unelevated Steam launch, SQLite storage, RU/EN UI, etc.).

Legacy Python sources under `legacy/` are archived earlier prototypes.

## AI-assisted development

Parts of this codebase were **polished and refined with assistance from AI**
(structure, UI/i18n, packaging, documentation, debugging).

- AI help is not a substitute for testing on a real Windows + Steam setup.
- Review security-sensitive areas (DPAPI, token storage, admin elevation)
  before shipping your own builds.

## Security

- Accounts/tokens are stored in `accounts.db` (SQLite) next to the executable.
- Do **not** commit `accounts.db`, `user_backup.json`, logs, or tokens to Git.
- The app requests **Administrator** rights on Windows (UAC) so it can update
  Steam config under Program Files; Steam itself is launched **without**
  elevation (elevated Steam often fails to show a normal UI).

## Trademark

Steam and Valve are trademarks of Valve Corporation. This project is
**unofficial** and not affiliated with or endorsed by Valve.
