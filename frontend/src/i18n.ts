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
};

const en: Dict = {
  resetSteam: 'Reset Steam',
  showInstructions: 'Show Instructions',
  accountManagement: 'Account Management',
  accountKeyPlaceholder: 'Enter your account key...',
  keepExisting: 'Keep other Steam logins (multi ConnectCache)',
  login: 'Login',
  working: 'Working…',
  hintEmpty: 'No accounts added yet. Use the field above to add one. Read the instructions before use.',
  hintHasAccounts: 'Select a saved account on the right or paste a new key above.',
  savedAccounts: 'Saved Accounts',
  noSavedAccounts: 'No saved accounts',
  delete: 'Delete',
  ready: 'Ready',
  enterKey: 'Enter an account key',
  loggingIn: 'Logging in...',
  loggingInAs: 'Logging in as {name}...',
  resettingSteam: 'Resetting Steam...',
  instructions: 'Instructions',
  help1: 'Paste account key as login----eya_token',
  help2: 'Keep existing = merge ConnectCache + loginusers (AutoLogin). Off = clean single-account write.',
  help3: 'Press Login — Steam restarts with the injected session',
  help4: 'Saved accounts appear on the right for one-click login',
  help5: 'Reset Steam wipes config/userdata if something breaks',
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
  steamNotDetected: 'Warning: steam.exe was not detected after launch. Check the tray or Task Manager.',
};

const ru: Dict = {
  resetSteam: 'Сброс Steam',
  showInstructions: 'Инструкция',
  accountManagement: 'Управление аккаунтом',
  accountKeyPlaceholder: 'Вставьте ключ аккаунта...',
  keepExisting: 'Сохранить другие логины Steam (multi ConnectCache)',
  login: 'Войти',
  working: 'Подождите…',
  hintEmpty: 'Аккаунтов пока нет. Вставьте ключ выше. Перед использованием прочитайте инструкцию.',
  hintHasAccounts: 'Выберите сохранённый аккаунт справа или вставьте новый ключ выше.',
  savedAccounts: 'Сохранённые аккаунты',
  noSavedAccounts: 'Нет сохранённых аккаунтов',
  delete: 'Удалить',
  ready: 'Готово',
  enterKey: 'Введите ключ аккаунта',
  loggingIn: 'Выполняется вход...',
  loggingInAs: 'Вход в {name}...',
  resettingSteam: 'Сброс Steam...',
  instructions: 'Инструкция',
  help1: 'Вставьте ключ в формате login----eya_token',
  help2: '«Сохранить» = merge ConnectCache + loginusers (AutoLogin). Выкл = чистый вход одного аккаунта.',
  help3: 'Нажмите «Войти» — Steam перезапустится с новой сессией',
  help4: 'Сохранённые аккаунты справа — вход в один клик',
  help5: '«Сброс Steam» удаляет config/userdata, если что-то сломалось',
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
  steamNotDetected: 'Внимание: steam.exe не найден после запуска. Проверьте трей или Диспетчер задач.',
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
  if (m.startsWith('account name required')) return t(lang, 'accountNameRequired');

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

  // legacy duration form
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
