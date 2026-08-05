export type Lang = 'en' | 'ru';

export type Dict = {
  resetSteam: string;
  showInstructions: string;
  accountManagement: string;
  accountKeyPlaceholder: string;
  keepExisting: string;
  login: string;
  working: string;
  hintEmpty: string;
  hintHasAccounts: string;
  savedAccounts: string;
  noSavedAccounts: string;
  delete: string;
  ready: string;
  enterKey: string;
  loggingIn: string;
  loggingInAs: string;
  resettingSteam: string;
  instructions: string;
  help1: string;
  help2: string;
  help3: string;
  help4: string;
  help5: string;
  fullGuide: string;
  gotIt: string;
  lang: string;
  expiredInvalid: string;
  unknown: string;
  // backend / status message maps
  accountDeleted: string;
  cancelled: string;
  steamReset: string;
  accountNotFound: string;
  accountNameRequired: string;
  loggedInAs: string; // {name}
  loggedInToken: string; // {name} {until}
  validUntil: string; // {until}
  successTitle: string;
  errorTitle: string;
  steamNotDetected: string;
  checkUpdate: string;
  updateTitle: string;
  updateAvailable: string; // {current} {latest}
  updateNone: string;
  updateNow: string;
  updateLater: string;
  updateOpenPage: string;
  updateInstalling: string;
  updateFailed: string;
  updateChecking: string;
  export: string;
  exportAll: string;
  exportSelected: string;
  selectAll: string;
  deselectAll: string;
  exportNone: string;
  exportCopied: string;
  exportOk: string;
  exportDrive: string;
  driveTitle: string;
  driveConnect: string;
  driveDisconnect: string;
  driveConnected: string;
  driveNotConnected: string;
  driveSetup: string;
  driveSetupHint: string;
  driveClientId: string;
  driveClientSecret: string;
  driveSave: string;
  driveImport: string;
  driveExport: string;
  driveUploading: string;
  driveHelp1: string;
  driveHelp2: string;
  driveHelp3: string;
  driveHelp4: string;
  driveWaiting: string;
  cancel: string;
  close: string;
  driveTutorialTitle: string;
  driveStep1: string;
  driveStep2: string;
  driveStep3: string;
  driveStep4: string;
  driveStep5: string;
  driveStep6: string;
  driveStep7: string;
  driveStep8: string;
  driveStep9: string;
  driveStep10: string;
  driveStep11: string;
  driveTipTestUsers: string;
  driveTipOwnProject: string;
  driveTipDesktop: string;
  driveTipApisServices: string;
  driveOpenConsole: string;
  driveShowTutorial: string;
  driveHideTutorial: string;
  driveGuideNext: string;
  driveGuideBack: string;
  driveGuideDone: string;
  driveGuideStepOf: string; // {n} {total}
  driveGuideT1: string;
  driveGuideT2: string;
  driveGuideT3: string;
  driveGuideT4: string;
  driveGuideT5: string;
  driveGuideT6: string;
  driveGuideT7: string;
  driveGuideT8: string;
  driveOpenGuide: string;
  help6: string;
};

const en: Dict = {
  resetSteam: 'Reset Steam',
  showInstructions: 'Show Instructions',
  accountManagement: 'Account Management',
  accountKeyPlaceholder: 'Enter your account key...',
  keepExisting: 'Keep other Steam accounts',
  login: 'Login',
  working: 'Working…',
  hintEmpty: 'No accounts yet. Paste a key above to add one.',
  hintHasAccounts: 'Pick an account on the right or paste a new key.',
  savedAccounts: 'Saved Accounts',
  noSavedAccounts: 'No saved accounts',
  delete: 'Delete',
  ready: 'Ready',
  enterKey: 'Enter an account key',
  loggingIn: 'Logging in...',
  loggingInAs: 'Logging in as {name}...',
  resettingSteam: 'Resetting Steam...',
  instructions: 'Instructions',
  help1: 'Paste your key as login----token',
  help2: 'If needed, turn on “Keep other Steam accounts”',
  help3: 'Press Login — Steam will restart and sign you in',
  help4: 'Saved accounts on the right can be used again anytime',
  help5: 'Use Reset Steam only if something goes wrong',
  fullGuide: 'Full guide:',
  gotIt: 'Got it',
  lang: 'Language',
  expiredInvalid: 'expired/invalid',
  unknown: 'unknown',
  accountDeleted: 'Account deleted',
  cancelled: 'Cancelled',
  steamReset: 'Steam has been reset',
  accountNotFound: 'account not found',
  accountNameRequired: 'account name required (use login----token)',
  loggedInAs: 'Logged in as {name}',
  loggedInToken: 'Logged in as {name} · token valid until {until}',
  validUntil: 'valid until {until}',
  successTitle: 'Success',
  errorTitle: 'Error',
  steamNotDetected: 'Steam did not start. Check the tray or Task Manager.',
  checkUpdate: 'Check update',
  updateTitle: 'Update available',
  updateAvailable: 'New version {latest} is available (you have {current}).',
  updateNone: 'You have the latest version.',
  updateNow: 'Update now',
  updateLater: 'Later',
  updateOpenPage: 'Open download page',
  updateInstalling: 'Downloading update… The app will restart.',
  updateFailed: 'Update failed',
  updateChecking: 'Checking for updates…',
  export: 'Export',
  exportAll: 'Export all',
  exportSelected: 'Export selected',
  selectAll: 'Select all',
  deselectAll: 'Clear selection',
  exportNone: 'Select accounts to export, or use Export all',
  exportCopied: 'Copied to clipboard',
  exportOk: 'Export done',
  exportDrive: 'Google Drive',
  driveTitle: 'Google Drive export',
  driveConnect: 'Connect Google',
  driveDisconnect: 'Disconnect',
  driveConnected: 'Connected',
  driveNotConnected: 'Not connected',
  driveSetup: 'Setup OAuth',
  driveSetupHint: 'One-time: create a Desktop OAuth client in Google Cloud and paste Client ID (and secret if shown).',
  driveClientId: 'Client ID',
  driveClientSecret: 'Client secret (optional)',
  driveSave: 'Save',
  driveImport: 'Import JSON…',
  driveExport: 'Upload to Drive',
  driveUploading: 'Uploading to Google Drive…',
  driveHelp1: 'Enable Google Drive API, then Google Auth platform → Branding (Get Started if needed)',
  driveHelp2: 'Google Auth platform → Audience → Test users → add your Gmail',
  driveHelp3: 'Google Auth platform → Clients → Create client → Desktop app → copy ID / JSON here',
  driveHelp4: 'Save → Connect Google → allow in browser → Upload to Drive',
  driveWaiting: 'Waiting for Google login in the browser… You can Cancel anytime.',
  cancel: 'Cancel',
  close: 'Close',
  driveTutorialTitle: 'How to set up Google Drive (one-time)',
  driveStep1: 'Open https://console.cloud.google.com and sign in with YOUR Google account (the one that owns the Drive).',
  driveStep2: 'Create a project: top bar project picker (folder icon / project name) → New Project → name e.g. NFA-Tool → Create. Wait until it finishes, then SELECT that project in the same picker (important!).',
  driveStep3: 'If you already have a project: open the same picker → select it. All next steps must be in THIS project.',
  driveStep4: 'Enable Drive API: button “Drive API” below → Enable. Same via menu: ☰ → APIs & Services → Library → search “Google Drive API” → Enable. (APIs & Services still exists and works.)',
  driveStep5: 'OAuth setup (new UI): ☰ → Google Auth platform → Branding. If “not configured yet” → Get Started. App name: NFA Tool. Support email + contact email: your Gmail. Audience: External → agree → Create.',
  driveStep6: 'Test users (required while Testing): ☰ → Google Auth platform → Audience → Test users → Add users → your Gmail → Save. Without this Google shows access_denied / “not completed verification”.',
  driveStep7: 'Optional scopes: Google Auth platform → Data Access → Add or remove scopes → “…/auth/drive.file” if you want it listed → Save. The app requests drive.file anyway.',
  driveStep8: 'Create Desktop client: ☰ → Google Auth platform → Clients → Create client → Application type: Desktop app → name → Create. (Alt path still works: ☰ → APIs & Services → Credentials → Create credentials → OAuth client ID → Desktop app.)',
  driveStep9: 'Open the client → copy Client ID (+ Client secret if shown) or Download JSON. In this window: paste fields or Import JSON → Save.',
  driveStep10: 'In NFA Tool: Connect Google → browser opens → choose the SAME Gmail as Test user → Allow.',
  driveStep11: 'Upload to Drive. File appears in your Drive root. Stuck on login screen → Cancel / ✕ anytime.',
  driveTipTestUsers: 'Status Testing = only emails under Google Auth platform → Audience → Test users. Add yourself first.',
  driveTipOwnProject: 'Each person: own Cloud project + own Desktop client + own email as Test user. No Google verification needed for personal use.',
  driveTipDesktop: 'Client type = Desktop app only. Not Web application.',
  driveTipApisServices: 'Both menus are real: “Google Auth platform” (Branding / Audience / Clients) and classic “APIs & Services” (Library, Credentials). Use either; pick the project first.',
  driveOpenConsole: 'OAuth Clients',
  driveShowTutorial: 'Show full tutorial',
  driveHideTutorial: 'Hide tutorial',
  driveGuideNext: 'Next',
  driveGuideBack: 'Back',
  driveGuideDone: 'Done',
  driveGuideStepOf: 'Step {n} of {total}',
  driveGuideT1: 'Sign in to Google Cloud',
  driveGuideT2: 'Create a project',
  driveGuideT3: 'Enable Google Drive API',
  driveGuideT4: 'Configure OAuth (Branding)',
  driveGuideT5: 'Add yourself as Test user',
  driveGuideT6: 'Create Desktop OAuth client',
  driveGuideT7: 'Paste credentials into NFA Tool',
  driveGuideT8: 'Connect and upload',
  driveOpenGuide: 'Open setup guide',
  help6: 'Google Drive: button Google Drive → Open setup guide (separate window)',
};

const ru: Dict = {
  resetSteam: 'Сброс Steam',
  showInstructions: 'Инструкция',
  accountManagement: 'Управление аккаунтом',
  accountKeyPlaceholder: 'Вставьте ключ аккаунта...',
  keepExisting: 'Сохранить другие аккаунты Steam',
  login: 'Войти',
  working: 'Подождите…',
  hintEmpty: 'Аккаунтов пока нет. Вставьте ключ выше.',
  hintHasAccounts: 'Выберите аккаунт справа или вставьте новый ключ.',
  savedAccounts: 'Сохранённые аккаунты',
  noSavedAccounts: 'Нет сохранённых аккаунтов',
  delete: 'Удалить',
  ready: 'Готово',
  enterKey: 'Введите ключ аккаунта',
  loggingIn: 'Выполняется вход...',
  loggingInAs: 'Вход в {name}...',
  resettingSteam: 'Сброс Steam...',
  instructions: 'Инструкция',
  help1: 'Вставьте ключ в формате login----token',
  help2: 'При необходимости включите «Сохранить другие аккаунты Steam»',
  help3: 'Нажмите «Войти» — Steam перезапустится и войдёт в аккаунт',
  help4: 'Сохранённые аккаунты справа можно использовать снова',
  help5: '«Сброс Steam» — только если что-то сломалось',
  fullGuide: 'Полный гайд:',
  gotIt: 'Понятно',
  lang: 'Язык',
  expiredInvalid: 'истёк/невалиден',
  unknown: 'неизвестно',
  accountDeleted: 'Аккаунт удалён',
  cancelled: 'Отменено',
  steamReset: 'Steam сброшен',
  accountNotFound: 'аккаунт не найден',
  accountNameRequired: 'нужен логин (формат login----token)',
  loggedInAs: 'Вход выполнен: {name}',
  loggedInToken: 'Вход выполнен: {name} · токен валиден до {until}',
  validUntil: 'валиден до {until}',
  successTitle: 'Успех',
  errorTitle: 'Ошибка',
  steamNotDetected: 'Steam не запустился. Проверьте трей или Диспетчер задач.',
  checkUpdate: 'Обновления',
  updateTitle: 'Доступно обновление',
  updateAvailable: 'Доступна версия {latest} (у вас {current}).',
  updateNone: 'У вас последняя версия.',
  updateNow: 'Обновить',
  updateLater: 'Позже',
  updateOpenPage: 'Страница загрузки',
  updateInstalling: 'Скачиваем обновление… Приложение перезапустится.',
  updateFailed: 'Не удалось обновить',
  updateChecking: 'Проверка обновлений…',
  export: 'Экспорт',
  exportAll: 'Экспорт всех',
  exportSelected: 'Экспорт выбранных',
  selectAll: 'Выбрать все',
  deselectAll: 'Снять выбор',
  exportNone: 'Выберите аккаунты или нажмите «Экспорт всех»',
  exportCopied: 'Скопировано в буфер',
  exportOk: 'Экспорт готов',
  exportDrive: 'Google Drive',
  driveTitle: 'Экспорт в Google Drive',
  driveConnect: 'Подключить Google',
  driveDisconnect: 'Отключить',
  driveConnected: 'Подключено',
  driveNotConnected: 'Не подключено',
  driveSetup: 'Настройка OAuth',
  driveSetupHint: 'Один раз: создайте OAuth-клиент Desktop в Google Cloud и вставьте Client ID (и secret, если есть).',
  driveClientId: 'Client ID',
  driveClientSecret: 'Client secret (необязательно)',
  driveSave: 'Сохранить',
  driveImport: 'Импорт JSON…',
  driveExport: 'Загрузить на Drive',
  driveUploading: 'Загрузка на Google Drive…',
  driveHelp1: 'Включи Google Drive API, затем Google Auth platform → Branding (при необходимости Get Started)',
  driveHelp2: 'Google Auth platform → Audience → Test users → добавь свой Gmail',
  driveHelp3: 'Google Auth platform → Clients → Create client → Desktop app → ID / JSON сюда',
  driveHelp4: 'Сохранить → Подключить Google → разреши в браузере → Загрузить на Drive',
  driveWaiting: 'Ждём вход Google в браузере… Можно нажать «Отмена» в любой момент.',
  cancel: 'Отмена',
  close: 'Закрыть',
  driveTutorialTitle: 'Как настроить Google Drive (один раз)',
  driveStep1: 'Открой https://console.cloud.google.com и войди СВОИМ Google (тот аккаунт, чей Drive нужен).',
  driveStep2: 'Создать проект: сверху селектор проекта (иконка папки / имя проекта) → New Project → имя, напр. NFA-Tool → Create. Дождись создания и ОБЯЗАТЕЛЬНО выбери этот проект в том же селекторе.',
  driveStep3: 'Если проект уже есть: тот же селектор сверху → выбери его. Все шаги ниже только в ЭТОМ проекте.',
  driveStep4: 'Включить Drive API: кнопка «Drive API» ниже → Enable. Или через меню: ☰ → APIs & Services → Library → «Google Drive API» → Enable. (Раздел APIs & Services по-прежнему есть и работает.)',
  driveStep5: 'OAuth (новый UI): ☰ → Google Auth platform → Branding. Если «not configured yet» → Get Started. App name: NFA Tool. Support + contact email: свой Gmail. Audience: External → согласие → Create.',
  driveStep6: 'Test users (обязательно в Testing): ☰ → Google Auth platform → Audience → Test users → Add users → свой Gmail → Save. Без этого: access_denied / «not completed verification».',
  driveStep7: 'Scopes по желанию: Google Auth platform → Data Access → Add or remove scopes → “…/auth/drive.file” → Save. Программа и так просит drive.file.',
  driveStep8: 'Клиент Desktop: ☰ → Google Auth platform → Clients → Create client → Application type: Desktop app → имя → Create. (Старый путь тоже ок: ☰ → APIs & Services → Credentials → Create credentials → OAuth client ID → Desktop app.)',
  driveStep9: 'Открой клиента → скопируй Client ID (+ secret) или Download JSON. Здесь: вставь в поля или «Импорт JSON» → Сохранить.',
  driveStep10: 'В NFA Tool: Подключить Google → браузер → тот же Gmail, что в Test users → Allow.',
  driveStep11: 'Загрузить на Drive. Файл в корне Drive. Зависло на входе — Отмена / ✕.',
  driveTipTestUsers: 'Статус Testing = только email из Google Auth platform → Audience → Test users. Сначала добавь себя.',
  driveTipOwnProject: 'Каждый: свой Cloud-проект + свой Desktop-клиент + свой email в Test users. Верификация Google для личного использования не нужна.',
  driveTipDesktop: 'Тип клиента только Desktop app. Не Web application.',
  driveTipApisServices: 'Оба меню живые: «Google Auth platform» (Branding / Audience / Clients) и классика «APIs & Services» (Library, Credentials). Главное — сначала выбрать нужный проект сверху.',
  driveOpenConsole: 'OAuth Clients',
  driveShowTutorial: 'Показать полный туториал',
  driveHideTutorial: 'Скрыть туториал',
  driveGuideNext: 'Далее',
  driveGuideBack: 'Назад',
  driveGuideDone: 'Готово',
  driveGuideStepOf: 'Шаг {n} из {total}',
  driveGuideT1: 'Вход в Google Cloud',
  driveGuideT2: 'Создать проект',
  driveGuideT3: 'Включить Google Drive API',
  driveGuideT4: 'Настроить OAuth (Branding)',
  driveGuideT5: 'Добавить себя в Test users',
  driveGuideT6: 'Создать Desktop OAuth-клиент',
  driveGuideT7: 'Вставить данные в NFA Tool',
  driveGuideT8: 'Подключить и загрузить',
  driveOpenGuide: 'Открыть гайд',
  help6: 'Google Drive: кнопка Google Drive → «Открыть гайд» (отдельное окно)',
};

const catalogs: Record<Lang, Dict> = { en, ru };

const STORAGE_KEY = 'nfa-tool-lang';

export function detectSystemLang(): Lang {
  const list = [
    ...(typeof navigator !== 'undefined' ? navigator.languages ?? [] : []),
    typeof navigator !== 'undefined' ? navigator.language : '',
  ]
    .filter(Boolean)
    .map((x) => x.toLowerCase());

  for (const l of list) {
    if (l.startsWith('ru')) return 'ru';
  }
  return 'en';
}

export function loadLang(): Lang {
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved === 'en' || saved === 'ru') return saved;
  } catch {
    /* ignore */
  }
  return detectSystemLang();
}

export function saveLang(lang: Lang) {
  try {
    localStorage.setItem(STORAGE_KEY, lang);
  } catch {
    /* ignore */
  }
}

export function t(lang: Lang, key: keyof Dict, vars?: Record<string, string | number>): string {
  let s = catalogs[lang][key] ?? catalogs.en[key] ?? String(key);
  if (vars) {
    for (const [k, v] of Object.entries(vars)) {
      s = s.replaceAll(`{${k}}`, String(v));
    }
  }
  return s;
}

/** Translate backend English status/result messages when possible. */
export function translateBackendMessage(lang: Lang, msg: string): string {
  if (!msg) return msg;
  const m = msg.trim();

  if (m === 'Account deleted' || m === 'account deleted') return t(lang, 'accountDeleted');
  if (m === 'Cancelled' || m === 'cancelled') return t(lang, 'cancelled');
  if (m === 'Steam has been reset') return t(lang, 'steamReset');
  if (m === 'account not found') return t(lang, 'accountNotFound');
  if (m === 'no accounts to export') return t(lang, 'exportNone');
  if (m === 'Google OAuth not configured') {
    return lang === 'ru'
      ? 'Сначала настройте Google OAuth (Client ID)'
      : 'Configure Google OAuth first (Client ID)';
  }
  if (m === 'Google OAuth saved') {
    return lang === 'ru' ? 'OAuth сохранён' : 'OAuth saved';
  }
  if (m === 'Google OAuth imported') {
    return lang === 'ru' ? 'OAuth импортирован' : 'OAuth imported';
  }
  if (m === 'Google Drive connected') {
    return lang === 'ru' ? 'Google Drive подключён' : 'Google Drive connected';
  }
  if (m === 'Google Drive disconnected') {
    return lang === 'ru' ? 'Google Drive отключён' : 'Google Drive disconnected';
  }
  if (m.startsWith('account name required')) return t(lang, 'accountNameRequired');
  if (m.startsWith('Exported ')) {
    return m.replace(/^Exported /, lang === 'ru' ? 'Экспортировано: ' : 'Exported ');
  }
  if (m.startsWith('Uploaded ')) {
    return m.replace(/^Uploaded /, lang === 'ru' ? 'Загружено: ' : 'Uploaded ');
  }

  const steamWarn = m.includes('warning: steam.exe not detected after launch');
  const core = m.replace(/\s*·\s*warning: steam\.exe not detected after launch\s*/i, '').trim();

  // Logged in as NAME · token valid until DATE
  let re = /^Logged in as (.+?) · token valid until (.+)$/;
  let match = core.match(re);
  if (match) {
    let out = t(lang, 'loggedInToken', { name: match[1], until: match[2] });
    if (steamWarn) out += '\n\n' + t(lang, 'steamNotDetected');
    return out;
  }

  // older duration form
  re = /^Logged in as (.+?) · token valid (.+)$/;
  match = m.match(re);
  if (match) return t(lang, 'loggedInToken', { name: match[1], until: match[2] });

  re = /^Logged in as (.+)$/;
  match = core.match(re);
  if (match) {
    let out = t(lang, 'loggedInAs', { name: match[1] });
    if (steamWarn) out += '\n\n' + t(lang, 'steamNotDetected');
    return out;
  }

  // common token errors stay readable; light-touch RU hints
  if (lang === 'ru') {
    if (m === 'token expired') return 'токен истёк';
    if (m === 'invalid token format') return 'неверный формат токена';
    if (m === 'empty input') return 'пустой ввод';
    if (m.startsWith('no JWT found')) return 'JWT не найден в ключе';
    if (m.startsWith('token issuer')) return 'issuer токена не steam';
    if (m.startsWith('token audience')) return 'в токене нет audience client';
    if (m.startsWith('token missing')) return 'в токене нет steam id';
    if (m.includes('Directory not recognized') || m.includes('directory not recognized')) {
      return 'папка Steam не найдена — сначала запустите Steam';
    }
    if (m.includes('encrypt token')) return 'ошибка шифрования токена (DPAPI)';
    if (m.includes('set AutoLoginUser')) return 'не удалось записать AutoLoginUser в реестр';
    if (m.includes('steam not found')) return 'Steam не найден';
  }

  return m;
}

export function localizeExpiry(lang: Lang, exp: string): string {
  if (exp === 'expired/invalid') return t(lang, 'expiredInvalid');
  if (exp === 'unknown') return t(lang, 'unknown');
  // Backend sends absolute date: "2026-09-15 14:30 UTC"
  if (/^\d{4}-\d{2}-\d{2}/.test(exp)) {
    return t(lang, 'validUntil', { until: exp });
  }
  return exp;
}
