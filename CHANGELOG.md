# Changelog

## NFA-Tool · 3.1.0

EN

### What's new
- **Batch check is now a full screen**, not a modal: autofocused key field, live results, and result export to a **file or Google Drive** (your choice)
- Result rows show an **Info button on hover** — the date gracefully fades under it; info opens **stacked above** the check, closable without losing results
- **Background avatar & nickname sync** with a live counter in the status bar; avatars fetched via the info dialog appear in the list instantly and count towards the sync
- **Account search** — by login, Steam nickname or SteamID64
- **Login / Nickname display toggle** next to saved accounts (remembered)
- **Management features switch** at the bottom of Settings — hides the Advanced / Checker tabs for a minimal UI
- **Delete confirmation** for single and bulk account deletion
- **Diagnostics dump** button in Settings: one text file with system info (Windows build, RAM, WebView2 runtime, elevation), app state, UI journal and log tail
- Background **error & crash logging** to `logs/app.log` (frontend JS errors included; fatal crashes captured)
- Journal: **newest entries at the bottom** with smart auto-scroll

### Fixes
- Fixed a **WebView2 renderer crash** when switching tabs (access violation under software rendering — pill animation now uses transform-only motion)
- Tab pills are perfectly symmetrical; labels and icons aligned across headers, tabs and buttons
- Language switch is now a calm fade instead of a jumpy repaint
- Mode/batch-check slide transitions no longer show the previous screen behind the new one
- Leaving the batch check when switching tabs; entering it moves the tab pill to Checker
- External links always open in the **system browser** — the app window can no longer be navigated away (drop-navigation and middle-click popups blocked)
- Old Crashpad crash dumps are pruned automatically (they ate ~10 MB each)
- Backend errors (Steam first run, batch parse errors, Drive upload) now translated in the UI
- Modal button spacing standardised

### Notes
- Run as **Administrator**
- Open Steam once on this PC before first login
- Batch check is rate-limited by Steam — use proxies for large lists
- Download: `NFA-Tool-windows-amd64.exe`

---

RU

### Что нового
- **Пакетная проверка теперь полноценный экран**, а не модалка: автофокус поля ключей, живые результаты, экспорт в **файл или на Google Drive** (на выбор)
- В строках результатов кнопка **«Инфо» появляется при наведении** — дата плавно растворяется под ней; инфо открывается **поверх** проверки и закрывается без её потери
- **Фоновая подгрузка аватарок и ников** со счётчиком в статусбаре; аватарка, полученная через «Инфо», появляется в списке мгновенно и засчитывается в счётчик
- **Поиск по аккаунтам** — по логину, нику Steam или SteamID64
- Переключатель **логин / ник Steam** возле сохранённых аккаунтов (запоминается)
- Свитч **«Функции управления»** внизу настроек — скрывает вкладки «Расширенный» и «Чекер» для минимального интерфейса
- **Подтверждение удаления** аккаунта и группы аккаунтов
- Кнопка **дампа диагностики** в настройках: один txt с данными системы (сборка Windows, RAM, WebView2, права), состоянием приложения, журналом и хвостом лога
- Фоновое **логирование ошибок и крашей** в `logs/app.log` (включая ошибки интерфейса и фатальные падения)
- Журнал: **новые записи снизу** с умным автоскроллом

### Исправления
- Исправлен **краш рендерера WebView2** при переключении вкладок (access violation на софтверном рендере — анимация пилюли переведена на transform)
- Пилюли вкладок идеально симметричны; выровнены тексты и иконки в заголовках, вкладках и кнопках
- Смена языка — спокойный фейд вместо дёрганой перерисовки
- Свайп-переходы режимов/пакетной проверки больше не показывают старый экран под новым
- Выход из пакетной проверки при переключении вкладки; при входе пилюля переезжает на «Чекер»
- Внешние ссылки всегда открываются в **системном браузере** — окно приложения больше нельзя увести на чужой сайт (блокированы drag-навигация и попапы средней кнопкой)
- Старые краш-дампы Crashpad чистятся автоматически (жрали по ~10 МБ каждый)
- Ошибки бэкенда (первый запуск Steam, разбор ключей, выгрузка на Drive) переведены в интерфейсе
- Унифицированы отступы кнопок в модальных окнах

### Важно
- Запуск **от администратора**
- Один раз откройте Steam перед первым входом
- Пакетная проверка ограничена рейт-лимитом Steam — для больших списков используйте прокси
- Скачать: `NFA-Tool-windows-amd64.exe`

---

**Full Changelog**: https://github.com/NeonFast/NFA-Tool/compare/v3.0.0...v3.1.0

---

## NFA Tool Recode v2 · 3.0.0

### What's new
- **Account info panel**: avatar, persona/real name, profile visibility, online state & in-game, level, games, friends, wallet, VAC / trade ban / limited status, account creation date
- **Token check without burning it**: proof logon against Steam CM over WebSocket — the refresh token is **not** consumed or rotated
- **Bulk check** for pasted keys or all saved accounts, with **proxy support** (http / https / socks5, one per line) and exportable results
- **Harvest** accounts already logged into Steam on this PC (ConnectCache scan)
- **Light theme** + theme switcher (auto / dark / light)
- **Simple / Advanced** UI modes
- **Logs panel** and **System status** panel (Steam running state, install path, stored accounts, Steam API reachability)
- Live **login stage** progress during sign-in
- Locally cached avatars for saved accounts

### Fixes
- More reliable Steam window detection (matched by PID instead of window class)
- Assorted stability and UI polish across the board

### Notes
- Run as **Administrator**
- Open Steam once on this PC before first login
- Bulk check is rate-limited by Steam — use proxies for large lists
- Download: `NFA-Tool-Recode-v2-windows-amd64.exe`

**Full Changelog**: https://github.com/NeonFast/NFA-Tool/compare/v2.1.2...v3.0.0
