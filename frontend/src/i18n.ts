/** Lang is a supported UI language. */
export type Lang = 'en' | 'ru';

/** Dict is the full set of translatable UI strings for one language. */
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
  deleteSelected: string;
  deletedN: string;
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
  accountDeleted: string;
  cancelled: string;
  steamReset: string;
  accountNotFound: string;
  accountNameRequired: string;
  loggedInAs: string;
  loggedInToken: string;
  validUntil: string;
  successTitle: string;
  errorTitle: string;
  creditPrefix: string;
  steamNotDetected: string;
  checkUpdate: string;
  updateTitle: string;
  updateAvailable: string;
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
  exportWhere: string;
  exportToFile: string;
  exportToClipboard: string;
  importBtn: string;
  importFile: string;
  importNone: string;
  importOk: string;
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
  driveGuideStepOf: string;
  driveGuideT1: string;
  driveGuideT2: string;
  driveGuideT3: string;
  driveGuideT4: string;
  driveGuideT5: string;
  driveGuideT6: string;
  driveGuideT7: string;
  driveGuideT8: string;
  driveOpenGuide: string;
  settings: string;
  settingsApp: string;
  settingsDrive: string;
  settingsLang: string;
  settingsTheme: string;
  themeAuto: string;
  themeDark: string;
  themeLight: string;
  help6: string;
  modeSimple: string;
  modeAdvanced: string;
  helpHint: string;
  hintConfirmTitle: string;
  hintConfirmText: string;
  hintConfirmYes: string;
  hintConfirmNo: string;
  info: string;
  accInfoTitle: string;
  accInfoLoading: string;
  accInfoOpenProfile: string;
  rowStatus: string;
  rowVisibility: string;
  rowVac: string;
  rowTrade: string;
  rowLimited: string;
  rowSince: string;
  rowToken: string;
  visPublic: string;
  visPrivate: string;
  visFriends: string;
  stOnline: string;
  stOffline: string;
  stInGame: string;
  yes: string;
  no: string;
  modeLogs: string;
  sysPanel: string;
  logPanel: string;
  logClear: string;
  logEmpty: string;
  logRefresh: string;
  sysVersion: string;
  sysRunning: string;
  sysStopped: string;
  sysPath: string;
  sysAccounts: string;
  apiOk: string;
  apiDown: string;
  keyCheckLabel: string;
  keyCheckBtn: string;
  keyCheckDone: string;
  rowRealName: string;
  rowLocation: string;
  rowInGame: string;
  rowLevel: string;
  rowGames: string;
  rowFriends: string;
  rowTopGame: string;
  rowWallet: string;
  rowCountry: string;
  rowBanInfo: string;
  daysShort: string;
  rowCS2Items: string;
  rowCS2Hours: string;
  rowPrime: string;
  rowLicenses: string;
  saveBtn: string;
  savedAs: string;
  rowCS2Medals: string;
  rowInvValue: string;
  invPartial: string;
  bulkCheckBtn: string;
  bulkTitle: string;
  bulkKeysLabel: string;
  bulkProxyLabel: string;
  bulkStart: string;
  bulkSaved: string;
  bstOk: string;
  bstRejected: string;
  bstExpired: string;
  bstInvalid: string;
  bstError: string;
  bulkProgress: string;
  bulkExportOk: string;
  bulkExportBad: string;
  bulkExportEmpty: string;
  bulkFileCreated: string;
  stageFindSteam: string;
  stageStopSteam: string;
  stageLoginUsers: string;
  stageConnectCache: string;
  stageConfig: string;
  stageRegistry: string;
  stageAcl: string;
  stageLaunch: string;
  stageWaitWindow: string;
  settingsProxy: string;
  proxyHint: string;
  bulkProxyGo: string;
  bulkOpenSettings: string;
  bulkFile: string;
  harvestBtn: string;
  tokSection: string;
  rowTokSteamID: string;
  rowTokIssued: string;
  rowTokExpires: string;
  rowTokAud: string;
  rowTokAccessExp: string;
  rowTokClaims: string;
  checksTitle: string;
  chkJwtStructure: string;
  chkSignature: string;
  chkIssuer: string;
  chkAudience: string;
  chkSteamIDFormat: string;
  chkNotExpired: string;
  chkIatPast: string;
  chkIatBeforeExp: string;
  chkNbfOk: string;
  chkRtExpOk: string;
  chkAccessMinted: string;
  chkAccessSubMatch: string;
  chkAccessExpOk: string;
  chkProfileMatch: string;
  dJwtStructure: string;
  dSignature: string;
  dIssuer: string;
  dAudience: string;
  dSteamIDFormat: string;
  dNotExpired: string;
  dIatPast: string;
  dIatBeforeExp: string;
  dNbfOk: string;
  dRtExpOk: string;
  dProfileMatch: string;
  dAccessMinted: string;
  alphaWarn: string;
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
  deleteSelected: 'Delete selected',
  deletedN: 'Deleted {n} account(s)',
  ready: 'Ready',
  enterKey: 'Enter an account key',
  loggingIn: 'Logging in...',
  loggingInAs: 'Logging in as {name}...',
  resettingSteam: 'Resetting Steam...',
  instructions: 'Instructions',
  help1: 'Paste a key as login----token, or import many from a .txt file',
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
  creditPrefix: 'Created with support of',
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
  exportWhere: 'Export to',
  exportToFile: 'Save as file…',
  exportToClipboard: 'Copy to clipboard',
  importBtn: 'Import',
  importFile: 'Import from file…',
  importNone: 'No valid accounts to import',
  importOk: 'Import completed',
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
  driveWaiting: 'Waiting for Google sign-in in the browser… You may press Cancel at any time.',
  cancel: 'Cancel',
  close: 'Close',
  driveTutorialTitle: 'Google Drive setup (one-time)',
  driveStep1: 'Open https://console.cloud.google.com and sign in with your Google account (the one whose Drive you intend to use).',
  driveStep2: 'Create a project: use the project picker at the top (folder icon / project name) → New Project → enter a name, e.g. NFA-Tool → Create. Wait until creation finishes, then select that project in the same picker.',
  driveStep3: 'If you already have a project: open the same picker and select it. Complete all following steps in this project only.',
  driveStep4: 'Enable the Drive API: use the “Drive API” button below → Enable. Alternatively: ☰ → APIs & Services → Library → search “Google Drive API” → Enable. (APIs & Services remains available.)',
  driveStep5: 'OAuth setup (current UI): ☰ → Google Auth platform → Branding. If you see “not configured yet”, click Get Started. App name: NFA Tool. Support and contact email: your Gmail. Audience: External → accept the policy → Create.',
  driveStep6: 'Test users (required while in Testing): ☰ → Google Auth platform → Audience → Test users → Add users → add your Gmail → Save. Without this step Google returns access_denied / “not completed verification”.',
  driveStep7: 'Scopes (optional): Google Auth platform → Data Access → Add or remove scopes → “…/auth/drive.file” → Save. The application requests drive.file regardless.',
  driveStep8: 'Desktop client: ☰ → Google Auth platform → Clients → Create client → Application type: Desktop app → enter a name → Create. (Alternative path: ☰ → APIs & Services → Credentials → Create credentials → OAuth client ID → Desktop app.)',
  driveStep9: 'Open the created client → copy the Client ID (and Client secret if shown) or Download JSON. In this window, paste the values or use Import JSON → Save.',
  driveStep10: 'In NFA Tool: Connect Google → a browser window opens → choose the same Gmail as in Test users → Allow.',
  driveStep11: 'Upload to Drive. The file appears in the root of your Drive. If sign-in stalls, press Cancel or ✕ at any time.',
  driveTipTestUsers: 'While status is Testing, only emails listed under Google Auth platform → Audience → Test users may sign in. Add yourself first.',
  driveTipOwnProject: 'Each user should create their own Cloud project, Desktop client, and list their email as a Test user. Google verification is not required for personal use.',
  driveTipDesktop: 'Client type must be Desktop app only. Web application is not suitable.',
  driveTipApisServices: 'Both areas are valid: “Google Auth platform” (Branding / Audience / Clients) and classic “APIs & Services” (Library, Credentials). Always select the correct project first.',
  driveOpenConsole: 'OAuth Clients',
  driveShowTutorial: 'Show full guide',
  driveHideTutorial: 'Hide guide',
  driveGuideNext: 'Next',
  driveGuideBack: 'Back',
  driveGuideDone: 'Done',
  driveGuideStepOf: 'Step {n} of {total}',
  driveGuideT1: 'Sign in to Google Cloud',
  driveGuideT2: 'Create a project',
  driveGuideT3: 'Enable Google Drive API',
  driveGuideT4: 'Configure OAuth (Branding)',
  driveGuideT5: 'Add yourself as a Test user',
  driveGuideT6: 'Create a Desktop OAuth client',
  driveGuideT7: 'Enter credentials in NFA Tool',
  driveGuideT8: 'Connect and upload',
  driveOpenGuide: 'Open setup guide',
  settings: 'Settings',
  settingsApp: 'Application',
  settingsDrive: 'Google Drive',
  settingsLang: 'Interface language',
  settingsTheme: 'Theme',
  themeAuto: 'Auto',
  themeDark: 'Dark',
  themeLight: 'Light',
  help6: 'Google Drive: Settings → configure once, then choose Drive when exporting',
  modeSimple: 'Simple',
  modeAdvanced: 'Advanced',
  helpHint: 'INSTRUCTIONS HERE',
  hintConfirmTitle: 'Hide the hint?',
  hintConfirmText: 'Really close this hint? It will not appear again. The instructions are always available via the "?" button.',
  hintConfirmYes: 'Yes, hide it',
  hintConfirmNo: 'Keep it',
  info: 'Info',
  accInfoTitle: 'Account info',
  accInfoLoading: 'Fetching account data…',
  accInfoOpenProfile: 'Open Steam profile',
  rowStatus: 'Status',
  rowVisibility: 'Profile',
  rowVac: 'VAC ban',
  rowTrade: 'Trade ban',
  rowLimited: 'Limited',
  rowSince: 'Registered',
  rowToken: 'Token accepted by Steam',
  visPublic: 'public',
  visPrivate: 'private',
  visFriends: 'friends only',
  stOnline: 'online',
  stOffline: 'offline',
  stInGame: 'in-game',
  yes: 'Yes',
  no: 'No',
  modeLogs: 'Checker',
  sysPanel: 'Status & checks',
  logPanel: 'Log',
  logClear: 'Clear',
  logEmpty: 'No entries yet',
  logRefresh: 'Refresh status',
  sysVersion: 'Version',
  sysRunning: 'running',
  sysStopped: 'not running',
  sysPath: 'Steam path',
  sysAccounts: 'Accounts (valid/total)',
  apiOk: 'reachable',
  apiDown: 'unreachable',
  keyCheckLabel: 'Key check',
  keyCheckBtn: 'Check',
  keyCheckDone: 'Key check finished',
  rowRealName: 'Real name',
  rowLocation: 'Location',
  rowInGame: 'Playing',
  rowLevel: 'Level',
  rowGames: 'Games',
  rowFriends: 'Friends',
  rowTopGame: 'Top game',
  rowWallet: 'Wallet',
  rowCountry: 'Country',
  rowBanInfo: 'Ban details',
  daysShort: 'd.',
  rowCS2Items: 'CS2 items',
  rowCS2Hours: 'CS2 hours',
  rowPrime: 'Prime status',
  rowLicenses: 'Licenses',
  saveBtn: 'Save',
  savedAs: 'Saved: {name} · token valid until {until}',
  rowCS2Medals: 'Medals',
  rowInvValue: 'Inventory value',
  invPartial: '(partial)',
  bulkCheckBtn: 'Bulk check',
  bulkTitle: 'Bulk account check',
  bulkKeysLabel: 'Keys — login----token, one per line',
  bulkProxyLabel: 'Proxies (optional, one per line: http/socks5, user:pass@host:port)',
  bulkStart: 'Check pasted keys',
  bulkSaved: 'Check saved accounts',
  bstOk: 'can log in',
  bstRejected: 'token rejected by Steam',
  bstExpired: 'expired',
  bstInvalid: 'invalid',
  bstError: 'check error',
  bulkProgress: 'Checked {done} of {total}…',
  bulkExportOk: 'Export working',
  bulkExportBad: 'Export failed',
  bulkExportEmpty: 'Nothing to export',
  bulkFileCreated: 'File created',
  stageFindSteam: 'Locating Steam…',
  stageStopSteam: 'Stopping Steam…',
  stageLoginUsers: 'Updating loginusers.vdf…',
  stageConnectCache: 'Writing session token…',
  stageConfig: 'Updating config.vdf…',
  stageRegistry: 'Setting auto-login…',
  stageAcl: 'Fixing file permissions…',
  stageLaunch: 'Launching Steam…',
  stageWaitWindow: 'Waiting for the Steam window…',
  settingsProxy: 'Proxies',
  proxyHint: 'One per line: http/socks5, user:pass@host:port. Used by the bulk checker.',
  bulkProxyGo: 'Proxies are configured in Settings',
  bulkOpenSettings: 'Open settings',
  bulkFile: 'Load .txt…',
  harvestBtn: 'Pull accounts from Steam files',
  tokSection: 'Token',
  rowTokSteamID: 'SteamID from token',
  rowTokIssued: 'Issued',
  rowTokExpires: 'Expires',
  rowTokAud: 'Audiences',
  rowTokAccessExp: 'Access token until',
  rowTokClaims: 'Raw JWT claims',
  checksTitle: 'Token checks',
  chkJwtStructure: 'JWT structure (3 segments)',
  chkSignature: 'Signature segment present',
  chkIssuer: 'Issuer is Steam',
  chkAudience: 'Audience allows client',
  chkSteamIDFormat: 'SteamID64 format',
  chkNotExpired: 'Not expired',
  chkIatPast: 'Issued in the past',
  chkIatBeforeExp: 'Issued before expiry',
  chkNbfOk: 'Not-before (nbf) valid',
  chkRtExpOk: 'Refresh window (rt_exp) valid',
  chkAccessMinted: 'Steam accepted the token (live sign-in)',
  chkAccessSubMatch: 'Access token SteamID matches',
  chkAccessExpOk: 'Access token not expired',
  chkProfileMatch: 'Profile matches token SteamID',
  dJwtStructure: 'The token has the correct shape — three parts separated by dots. If not, it is just a broken piece of text.',
  dSignature: 'The token has its security seal. Without it, Steam treats it as fake.',
  dIssuer: 'The token was really made by Steam, not copied or generated elsewhere.',
  dAudience: 'The token is allowed to log into the Steam app. Without this, login is impossible.',
  dSteamIDFormat: 'The token correctly names which account it belongs to (SteamID).',
  dNotExpired: 'The token has not expired yet.',
  dIatPast: 'The token was created in the past, as it should be. “From the future” would mean a fake.',
  dIatBeforeExp: 'The creation date is earlier than the expiry date. If reversed, the token was tampered with.',
  dNbfOk: 'If the token says “not valid before a date”, that date has already come.',
  dRtExpOk: 'The token can still be renewed — its refresh window has not closed.',
  dProfileMatch: 'The account inside the token matches the public Steam profile. Mismatch = the token belongs to a different account.',
  dAccessMinted: 'Steam actually accepted the token: we signed in to the Steam network with it (harmlessly — the token is not spent). If this fails, the token is revoked and login is impossible.',
  alphaWarn: 'ALPHA: the live checker signs in to the Steam network (CM). Check results may be unstable. Re-verification of rejected accounts is recommended before deletion.',
};

const ru: Dict = {
  resetSteam: 'Сброс Steam',
  showInstructions: 'Инструкция',
  accountManagement: 'Управление аккаунтом',
  accountKeyPlaceholder: 'Вставьте ключ аккаунта…',
  keepExisting: 'Сохранить другие аккаунты Steam',
  login: 'Войти',
  working: 'Пожалуйста, подождите…',
  hintEmpty: 'Сохранённых аккаунтов пока нет. Вставьте ключ в поле выше.',
  hintHasAccounts: 'Выберите аккаунт справа или вставьте новый ключ.',
  savedAccounts: 'Сохранённые аккаунты',
  noSavedAccounts: 'Нет сохранённых аккаунтов',
  delete: 'Удалить',
  deleteSelected: 'Удалить выбранные',
  deletedN: 'Удалено аккаунтов: {n}',
  ready: 'Готово',
  enterKey: 'Введите ключ аккаунта',
  loggingIn: 'Выполняется вход…',
  loggingInAs: 'Выполняется вход: {name}…',
  resettingSteam: 'Выполняется сброс Steam…',
  instructions: 'Инструкция',
  help1: 'Вставьте ключ в формате login----token или импортируйте несколько из .txt файла',
  help2: 'При необходимости включите параметр «Сохранить другие аккаунты Steam»',
  help3: 'Нажмите «Войти» — Steam перезапустится и выполнит вход в аккаунт',
  help4: 'Сохранённые аккаунты справа можно использовать повторно в любое время',
  help5: 'Функцию «Сброс Steam» используйте только при возникновении неполадок',
  fullGuide: 'Полное руководство:',
  gotIt: 'Понятно',
  lang: 'Язык',
  expiredInvalid: 'истёк / недействителен',
  unknown: 'неизвестно',
  accountDeleted: 'Аккаунт удалён',
  cancelled: 'Отменено',
  steamReset: 'Steam сброшен',
  accountNotFound: 'аккаунт не найден',
  accountNameRequired: 'укажите логин (формат login----token)',
  loggedInAs: 'Вход выполнен: {name}',
  loggedInToken: 'Вход выполнен: {name} · токен действителен до {until}',
  validUntil: 'действителен до {until}',
  successTitle: 'Успешно',
  errorTitle: 'Ошибка',
  creditPrefix: 'Создано при поддержке',
  steamNotDetected: 'Steam не запустился. Проверьте область уведомлений или Диспетчер задач.',
  checkUpdate: 'Обновления',
  updateTitle: 'Доступно обновление',
  updateAvailable: 'Доступна версия {latest} (у вас установлена {current}).',
  updateNone: 'У вас установлена актуальная версия.',
  updateNow: 'Обновить',
  updateLater: 'Позже',
  updateOpenPage: 'Страница загрузки',
  updateInstalling: 'Выполняется загрузка обновления… Приложение будет перезапущено.',
  updateFailed: 'Не удалось выполнить обновление',
  updateChecking: 'Проверка обновлений…',
  export: 'Экспорт',
  exportAll: 'Экспорт всех',
  exportSelected: 'Экспорт выбранных',
  selectAll: 'Выбрать все',
  deselectAll: 'Снять выбор',
  exportNone: 'Выберите аккаунты или воспользуйтесь пунктом «Экспорт всех»',
  exportCopied: 'Скопировано в буфер обмена',
  exportOk: 'Экспорт выполнен',
  exportWhere: 'Куда экспортировать',
  exportToFile: 'Сохранить в файл…',
  exportToClipboard: 'Копировать в буфер',
  importBtn: 'Импорт',
  importFile: 'Импорт из файла…',
  importNone: 'Нет корректных аккаунтов для импорта',
  importOk: 'Импорт выполнен',
  exportDrive: 'Google Drive',
  driveTitle: 'Экспорт в Google Drive',
  driveConnect: 'Подключить Google',
  driveDisconnect: 'Отключить',
  driveConnected: 'Подключено',
  driveNotConnected: 'Не подключено',
  driveSetup: 'Настройка OAuth',
  driveSetupHint: 'Один раз настройте OAuth-клиент Desktop в Google Cloud и вставьте Client ID (и Client secret, если он указан).',
  driveClientId: 'Client ID',
  driveClientSecret: 'Client secret (необязательно)',
  driveSave: 'Сохранить',
  driveImport: 'Импорт JSON…',
  driveExport: 'Загрузить на Drive',
  driveUploading: 'Выполняется загрузка на Google Drive…',
  driveHelp1: 'Включите Google Drive API, затем откройте Google Auth platform → Branding (при необходимости Get Started)',
  driveHelp2: 'Google Auth platform → Audience → Test users → добавьте ваш Gmail',
  driveHelp3: 'Google Auth platform → Clients → Create client → Desktop app → скопируйте ID / JSON сюда',
  driveHelp4: 'Сохраните → Подключите Google → подтвердите доступ в браузере → Загрузите на Drive',
  driveWaiting: 'Ожидание входа Google в браузере… Вы можете нажать «Отмена» в любой момент.',
  cancel: 'Отмена',
  close: 'Закрыть',
  driveTutorialTitle: 'Настройка Google Drive (один раз)',
  driveStep1: 'Откройте https://console.cloud.google.com и войдите в ваш аккаунт Google (тот, чей Drive требуется использовать).',
  driveStep2: 'Создайте проект: в верхней панели селектор проекта (иконка папки / имя проекта) → New Project → укажите имя, например NFA-Tool → Create. Дождитесь завершения и обязательно выберите созданный проект в том же селекторе.',
  driveStep3: 'Если проект уже создан: откройте тот же селектор сверху и выберите его. Все дальнейшие шаги выполняйте только в этом проекте.',
  driveStep4: 'Включите Drive API: кнопка «Drive API» ниже → Enable. Либо через меню: ☰ → APIs & Services → Library → «Google Drive API» → Enable. (Раздел APIs & Services по-прежнему доступен.)',
  driveStep5: 'Настройка OAuth (новый интерфейс): ☰ → Google Auth platform → Branding. Если отображается «not configured yet» — нажмите Get Started. App name: NFA Tool. Support и contact email: ваш Gmail. Audience: External → примите условия → Create.',
  driveStep6: 'Test users (обязательно в режиме Testing): ☰ → Google Auth platform → Audience → Test users → Add users → укажите ваш Gmail → Save. Без этого Google возвращает access_denied / «not completed verification».',
  driveStep7: 'Scopes (по желанию): Google Auth platform → Data Access → Add or remove scopes → «…/auth/drive.file» → Save. Приложение и так запрашивает scope drive.file.',
  driveStep8: 'Клиент Desktop: ☰ → Google Auth platform → Clients → Create client → Application type: Desktop app → укажите имя → Create. (Альтернативный путь: ☰ → APIs & Services → Credentials → Create credentials → OAuth client ID → Desktop app.)',
  driveStep9: 'Откройте созданного клиента → скопируйте Client ID (и Client secret, если отображается) либо Download JSON. В этом окне вставьте данные в поля или выполните «Импорт JSON» → Сохранить.',
  driveStep10: 'В NFA Tool: «Подключить Google» → откроется браузер → выберите тот же Gmail, что указан в Test users → Allow.',
  driveStep11: 'Загрузите файл на Drive. Он появится в корне вашего Drive. Если вход завис — нажмите «Отмена» или ✕.',
  driveTipTestUsers: 'При статусе Testing вход доступен только адресам из Google Auth platform → Audience → Test users. Сначала добавьте себя.',
  driveTipOwnProject: 'Каждому пользователю рекомендуется создать собственный проект Cloud, Desktop-клиент и указать свой email в Test users. Для личного использования верификация Google не требуется.',
  driveTipDesktop: 'Тип клиента — только Desktop app. Тип Web application не подходит.',
  driveTipApisServices: 'Доступны оба раздела: «Google Auth platform» (Branding / Audience / Clients) и классический «APIs & Services» (Library, Credentials). Важно сначала выбрать нужный проект в верхней панели.',
  driveOpenConsole: 'OAuth Clients',
  driveShowTutorial: 'Показать полное руководство',
  driveHideTutorial: 'Скрыть руководство',
  driveGuideNext: 'Далее',
  driveGuideBack: 'Назад',
  driveGuideDone: 'Готово',
  driveGuideStepOf: 'Шаг {n} из {total}',
  driveGuideT1: 'Вход в Google Cloud',
  driveGuideT2: 'Создание проекта',
  driveGuideT3: 'Включение Google Drive API',
  driveGuideT4: 'Настройка OAuth (Branding)',
  driveGuideT5: 'Добавление себя в Test users',
  driveGuideT6: 'Создание Desktop OAuth-клиента',
  driveGuideT7: 'Ввод данных в NFA Tool',
  driveGuideT8: 'Подключение и загрузка',
  driveOpenGuide: 'Открыть руководство',
  settings: 'Настройки',
  settingsApp: 'Приложение',
  settingsDrive: 'Google Drive',
  settingsLang: 'Язык интерфейса',
  settingsTheme: 'Тема',
  themeAuto: 'Авто',
  themeDark: 'Тёмная',
  themeLight: 'Светлая',
  help6: 'Google Drive: Настройки → настройте один раз, затем выбирайте Drive при экспорте',
  modeSimple: 'Простой',
  modeAdvanced: 'Расширенный',
  helpHint: 'ТУТ ИНСТРУКЦИЯ',
  hintConfirmTitle: 'Скрыть подсказку?',
  hintConfirmText: 'Точно закрыть подсказку? Она больше не появится. Инструкция всегда доступна по кнопке «?».',
  hintConfirmYes: 'Да, скрыть',
  hintConfirmNo: 'Оставить',
  info: 'Инфо',
  accInfoTitle: 'Информация об аккаунте',
  accInfoLoading: 'Получаем данные аккаунта…',
  accInfoOpenProfile: 'Открыть профиль Steam',
  rowStatus: 'Статус',
  rowVisibility: 'Профиль',
  rowVac: 'VAC-бан',
  rowTrade: 'Трейд-бан',
  rowLimited: 'Лимит',
  rowSince: 'Регистрация',
  rowToken: 'Токен принят Steam',
  visPublic: 'открытый',
  visPrivate: 'закрытый',
  visFriends: 'только друзья',
  stOnline: 'в сети',
  stOffline: 'не в сети',
  stInGame: 'в игре',
  yes: 'Да',
  no: 'Нет',
  modeLogs: 'Чекер',
  sysPanel: 'Статус и проверки',
  logPanel: 'Журнал',
  logClear: 'Очистить',
  logEmpty: 'Записей пока нет',
  logRefresh: 'Обновить статус',
  sysVersion: 'Версия',
  sysRunning: 'запущен',
  sysStopped: 'не запущен',
  sysPath: 'Путь Steam',
  sysAccounts: 'Аккаунты (валидные/всего)',
  apiOk: 'доступен',
  apiDown: 'недоступен',
  keyCheckLabel: 'Проверка ключа',
  keyCheckBtn: 'Проверить',
  keyCheckDone: 'Проверка ключа выполнена',
  rowRealName: 'Имя',
  rowLocation: 'Локация',
  rowInGame: 'Играет в',
  rowLevel: 'Уровень',
  rowGames: 'Игр',
  rowFriends: 'Друзей',
  rowTopGame: 'Топ-игра',
  rowWallet: 'Кошелёк',
  rowCountry: 'Страна',
  rowBanInfo: 'Детали банов',
  daysShort: 'дн.',
  rowCS2Items: 'Предметы CS2',
  rowCS2Hours: 'Часы в CS2',
  rowPrime: 'Prime-статус',
  rowLicenses: 'Лицензий',
  saveBtn: 'Сохранить',
  savedAs: 'Сохранено: {name} · токен действителен до {until}',
  rowCS2Medals: 'Медали',
  rowInvValue: 'Цена инвентаря',
  invPartial: '(не все)',
  bulkCheckBtn: 'Масс. проверка',
  bulkTitle: 'Массовая проверка аккаунтов',
  bulkKeysLabel: 'Ключи — login----token, по одному на строку',
  bulkProxyLabel: 'Прокси (необязательно, по одному на строку: http/socks5, user:pass@host:port)',
  bulkStart: 'Проверить вставленные',
  bulkSaved: 'Проверить сохранённые',
  bstOk: 'вход возможен',
  bstRejected: 'токен отклонён Steam',
  bstExpired: 'истёк',
  bstInvalid: 'невалиден',
  bstError: 'ошибка проверки',
  bulkProgress: 'Проверено {done} из {total}…',
  bulkExportOk: 'Экспорт рабочих',
  bulkExportBad: 'Экспорт нерабочих',
  bulkExportEmpty: 'Нечего экспортировать',
  bulkFileCreated: 'Файл создан',
  stageFindSteam: 'Поиск Steam…',
  stageStopSteam: 'Остановка Steam…',
  stageLoginUsers: 'Обновление loginusers.vdf…',
  stageConnectCache: 'Запись токена сессии…',
  stageConfig: 'Обновление config.vdf…',
  stageRegistry: 'Настройка автологина…',
  stageAcl: 'Настройка прав доступа…',
  stageLaunch: 'Запуск Steam…',
  stageWaitWindow: 'Ожидание окна Steam…',
  settingsProxy: 'Прокси',
  proxyHint: 'По одному на строку: http/socks5, user:pass@host:port. Используются массовой проверкой.',
  bulkProxyGo: 'Прокси задаются в настройках',
  bulkOpenSettings: 'Открыть настройки',
  bulkFile: 'Загрузить .txt…',
  harvestBtn: 'Забрать аккаунты из файлов Steam',
  tokSection: 'Токен',
  rowTokSteamID: 'SteamID из токена',
  rowTokIssued: 'Выдан',
  rowTokExpires: 'Истекает',
  rowTokAud: 'Audience',
  rowTokAccessExp: 'Access-токен до',
  rowTokClaims: 'Сырые JWT claims',
  checksTitle: 'Проверки токена',
  chkJwtStructure: 'Структура JWT (3 сегмента)',
  chkSignature: 'Сегмент подписи на месте',
  chkIssuer: 'Issuer — Steam',
  chkAudience: 'Audience допускает client',
  chkSteamIDFormat: 'Формат SteamID64',
  chkNotExpired: 'Не истёк',
  chkIatPast: 'Выдан в прошлом',
  chkIatBeforeExp: 'Выдан раньше истечения',
  chkNbfOk: 'Not-before (nbf) корректен',
  chkRtExpOk: 'Окно refresh (rt_exp) действует',
  chkAccessMinted: 'Steam принял токен (живой вход)',
  chkAccessSubMatch: 'SteamID access-токена совпадает',
  chkAccessExpOk: 'Access-токен не истёк',
  chkProfileMatch: 'Профиль совпадает со SteamID токена',
  dJwtStructure: 'Токен имеет правильную форму — три части через точки. Если нет, это просто битый кусок текста.',
  dSignature: 'У токена есть защитная печать. Без неё Steam считает его подделкой.',
  dIssuer: 'Токен действительно сделан Steam, а не скопирован или сгенерирован где-то ещё.',
  dAudience: 'Токену разрешено входить в приложение Steam. Без этого вход невозможен.',
  dSteamIDFormat: 'В токене правильно указано, к какому аккаунту он относится (SteamID).',
  dNotExpired: 'Срок действия токена ещё не вышел.',
  dIatPast: 'Токен создан в прошлом, как и положено. «Из будущего» — значит подделка.',
  dIatBeforeExp: 'Дата создания раньше даты окончания. Если наоборот — токен подправили.',
  dNbfOk: 'Если в токене написано «действует с такой-то даты» — эта дата уже наступила.',
  dRtExpOk: 'Токен ещё можно продлить — его окно обновления не закрылось.',
  dProfileMatch: 'Аккаунт внутри токена совпадает с публичным профилем Steam. Не совпадает — токен от другого аккаунта.',
  dAccessMinted: 'Steam реально принял токен: мы вошли с ним в сеть Steam (безопасно — токен при этом не тратится). Если тут крестик — токен отозван и войти нельзя.',
  alphaWarn: 'ALPHA: живой чекер выполняет вход в сеть Steam (CM). Результаты проверки могут быть нестабильными. Перед удалением отклонённых аккаунтов рекомендуется выполнить повторную проверку.',
};

const catalogs: Record<Lang, Dict> = { en, ru };

const STORAGE_KEY = 'nfa-tool-lang';

/** detectSystemLang picks 'ru' when the browser prefers Russian, else 'en'. */
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

/** loadLang reads the saved language, defaulting to the system language. */
export function loadLang(): Lang {
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved === 'en' || saved === 'ru') return saved;
  } catch {
  }
  return detectSystemLang();
}

/** saveLang persists the language preference. */
export function saveLang(lang: Lang) {
  try {
    localStorage.setItem(STORAGE_KEY, lang);
  } catch {
  }
}

/** t translates a dictionary key, substituting {var} placeholders. */
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
  if (m.startsWith('Deleted ')) {
    const n = m.match(/Deleted (\d+)/)?.[1] ?? '';
    return n ? t(lang, 'deletedN', { n }) : m.replace(/^Deleted /, lang === 'ru' ? 'Удалено: ' : 'Deleted ');
  }
  if (m === 'no accounts to delete') {
    return lang === 'ru' ? 'Нет аккаунтов для удаления' : 'No accounts to delete';
  }
  if (m === 'Cancelled' || m === 'cancelled') return t(lang, 'cancelled');
  if (m === 'Steam has been reset') return t(lang, 'steamReset');
  if (m === 'account not found') return t(lang, 'accountNotFound');
  if (m === 'no accounts to export') return t(lang, 'exportNone');
  if (m === 'no accounts to import') return t(lang, 'importNone');
  if (m.startsWith('Imported ')) {
    return m.replace(/^Imported /, lang === 'ru' ? 'Импортировано: ' : 'Imported ');
  }
  if (m.startsWith('import failed')) {
    return lang === 'ru' ? m.replace(/^import failed/, 'Импорт не выполнен') : m;
  }
  if (m === 'Google OAuth not configured') {
    return lang === 'ru'
      ? 'Сначала настройте Google OAuth (укажите Client ID)'
      : 'Please configure Google OAuth first (Client ID)';
  }
  if (m === 'Google OAuth saved') {
    return lang === 'ru' ? 'Данные OAuth сохранены' : 'OAuth credentials saved';
  }
  if (m === 'Google OAuth imported') {
    return lang === 'ru' ? 'Данные OAuth импортированы' : 'OAuth credentials imported';
  }
  if (m === 'Google Drive connected') {
    return lang === 'ru' ? 'Google Drive успешно подключён' : 'Google Drive connected';
  }
  if (m === 'Google Drive disconnected') {
    return lang === 'ru' ? 'Google Drive отключён' : 'Google Drive disconnected';
  }
  if (m.startsWith('account name required')) return t(lang, 'accountNameRequired');

  let saveRe = /^Saved as (.+?) · token valid until (.+)$/;
  const saveMatch = m.match(saveRe);
  if (saveMatch) return t(lang, 'savedAs', { name: saveMatch[1], until: saveMatch[2] });
  if (m.startsWith('Exported ')) {
    return m.replace(/^Exported /, lang === 'ru' ? 'Экспортировано: ' : 'Exported ');
  }
  if (m.startsWith('Harvested ')) {
    const n = m.match(/Harvested (\d+)/)?.[1] ?? '';
    if (!n) return m;
    return lang === 'ru' ? `Найдено аккаунтов в файлах Steam: ${n}` : m;
  }
  if (m === 'no accounts found in Steam files') {
    return lang === 'ru' ? 'В файлах Steam аккаунты не найдены' : 'No accounts found in Steam files';
  }
  if (m.startsWith('Uploaded ')) {
    return m.replace(/^Uploaded /, lang === 'ru' ? 'Загружено: ' : 'Uploaded ');
  }

  const steamWarn = m.includes('warning: steam.exe not detected after launch');
  const core = m.replace(/\s*·\s*warning: steam\.exe not detected after launch\s*/i, '').trim();

  let re = /^Logged in as (.+?) · token valid until (.+)$/;
  let match = core.match(re);
  if (match) {
    let out = t(lang, 'loggedInToken', { name: match[1], until: match[2] });
    if (steamWarn) out += '\n\n' + t(lang, 'steamNotDetected');
    return out;
  }

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

  if (lang === 'ru') {
    if (m === 'token expired') return 'Срок действия токена истёк';
    if (m === 'invalid token format') return 'Неверный формат токена';
    if (m === 'empty input') return 'Пустой ввод';
    if (m.startsWith('no JWT found')) return 'JWT не найден в ключе';
    if (m.startsWith('token issuer')) return 'Issuer токена не соответствует Steam';
    if (m.startsWith('token audience')) return 'В токене отсутствует audience client';
    if (m.startsWith('token missing')) return 'В токене отсутствует Steam ID';
    if (m.includes('Directory not recognized') || m.includes('directory not recognized')) {
      return 'Папка Steam не найдена. Сначала запустите Steam на этом компьютере.';
    }
    if (m.includes('encrypt token')) return 'Ошибка шифрования токена (DPAPI)';
    if (m.includes('set AutoLoginUser')) return 'Не удалось записать AutoLoginUser в реестр';
    if (m.includes('steam not found')) return 'Steam не найден';
    if (m.includes('authorization timed out')) {
      return 'Время ожидания авторизации истекло. Нажмите «Отмена» и повторите попытку.';
    }
  }

  return m;
}

/** localizeExpiry renders a backend expiry string in the UI language. */
export function localizeExpiry(lang: Lang, exp: string): string {
  if (exp === 'expired/invalid') return t(lang, 'expiredInvalid');
  if (exp === 'unknown') return t(lang, 'unknown');
  if (/^\d{4}-\d{2}-\d{2}/.test(exp)) {
    return t(lang, 'validUntil', { until: exp });
  }
  return exp;
}
