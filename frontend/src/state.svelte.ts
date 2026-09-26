import { Browser } from '@wailsio/runtime';
import { AppService } from '../bindings/nfa-tool';
import {
  type Lang,
  loadLang,
  saveLang,
  t,
  translateBackendMessage,
} from './i18n';
import { renderMarkdown } from './markdown';
import { type Theme, applyTheme, loadTheme, saveTheme } from './theme';
import { calmUpdate } from './actions';

export const GUIDE_URL = 'https://teletype.in/@hackerdlc/CS2NFA';

export type Mode = 'simple' | 'advanced' | 'logs';
const MODE_KEY = 'nfa-tool-mode';
const MODE_ORDER: Mode[] = ['simple', 'advanced', 'logs'];

function loadMode(): Mode {
  try {
    const m = localStorage.getItem(MODE_KEY);
    if (m === 'simple' || m === 'advanced' || m === 'logs') return m;
  } catch {
  }
  return 'simple';
}

export type LogEntry = { ts: number; kind: 'ok' | 'err' | 'info'; text: string };
const LOG_KEY = 'nfa-tool-log';
const LOG_CAP = 200;

function loadLogs(): LogEntry[] {
  try {
    const raw = localStorage.getItem(LOG_KEY);
    if (raw) {
      const arr = JSON.parse(raw);
      if (Array.isArray(arr)) {
        return arr
          .slice()
          .sort((a, b) => (a?.ts ?? 0) - (b?.ts ?? 0))
          .slice(-LOG_CAP);
      }
    }
  } catch {
  }
  return [];
}

export type Account = {
  name: string;
  steamId?: string;
  persona?: string;
  expiresIn: string;
  valid: boolean;
  avatar?: string;
};

export type Result = {
  ok: boolean;
  message: string;
};

export type BulkItem = {
  account: string;
  steamId?: string;
  status: string;
  expiresAt?: string;
  detail?: string;
};

export type InfoLogin = { kind: 'key'; key: string } | { kind: 'saved'; name: string };

export type InfoData = {
  steamId: string;
  personaName: string;
  realName?: string;
  avatarFull: string;
  profileUrl: string;
  visibility: string;
  onlineState: string;
  inGame?: string;
  location?: string;
  summary?: string;
  friendsCount: number;
  topGame?: string;
  topGameHours?: string;
  level?: string;
  gamesCount?: string;
  banInfo?: string;
  banDays?: string;
  cs2Items: number;
  cs2Rarity?: string;
  cs2Medals?: string;
  inventoryValue?: string;
  inventoryPartial?: boolean;
  cs2Hours?: string;
  cs2Prime?: string;
  licensesCount: number;
  email?: string;
  wallet?: string;
  country?: string;
  vacBanned: boolean;
  tradeBan: string;
  limited: boolean;
  memberSince: string;
  tokenAlive: boolean;
  profileErr?: string;
  tokenIssued?: string;
  tokenExpires?: string;
  tokenDaysLeft?: number;
  tokenAudiences?: string;
  tokenSteamId?: string;
  tokenClaimsJson?: string;
  tokenChecks?: { id: string; status: string }[];
};

export type UpdateInfo = {
  updateAvailable: boolean;
  currentVersion: string;
  latestVersion: string;
  releaseUrl: string;
  downloadUrl: string;
  releaseNotes: string;
  error?: string;
};

export type DriveStatus = {
  hasCredentials: boolean;
  connected: boolean;
  clientIdHint: string;
};

export type SysStatus = {
  version: string;
  steamRunning: boolean;
  steamPath: string;
  accountsTotal: number;
  accountsValid: number;
  driveConnected: boolean;
  steamApiOnline: boolean;
};

export const LOGIN_STAGES = [
  'find_steam',
  'stop_steam',
  'loginusers',
  'connectcache',
  'config_vdf',
  'registry',
  'acl',
  'launch',
  'wait_window',
] as const;

const HELP_HINT_KEY = 'nfa-tool-help-hint';

function loadHelpHint(): boolean {
  try {
    return localStorage.getItem(HELP_HINT_KEY) !== '1';
  } catch {
    return true;
  }
}

const KEEP_EXISTING_KEY = 'nfa-tool-keep-existing';

function loadKeepExisting(): boolean {
  try {
    return localStorage.getItem(KEEP_EXISTING_KEY) === '1';
  } catch {
    return false;
  }
}

const MANAGEMENT_KEY = 'nfa-tool-management';

function loadManagement(): boolean {
  try {
    return localStorage.getItem(MANAGEMENT_KEY) !== '0';
  } catch {
    return true;
  }
}

const SHOW_PERSONA_KEY = 'nfa-tool-show-persona';

function loadShowPersona(): boolean {
  try {
    return localStorage.getItem(SHOW_PERSONA_KEY) === '1';
  } catch {
    return false;
  }
}

const initialManagement = loadManagement();

const PROXY_KEY = 'nfa-tool-proxies';
function loadProxies(): string {
  try {
    return localStorage.getItem(PROXY_KEY) ?? '';
  } catch {
    return '';
  }
}

export const app = $state({
  appName: 'NFA Tool',
  version: '',
  lang: loadLang() as Lang,
  theme: loadTheme() as Theme,
  mode: (initialManagement ? loadMode() : 'simple') as Mode,
  management: initialManagement,
  showPersona: loadShowPersona(),
  showHelpHint: loadHelpHint(),
  showHintConfirm: false,
  slideFrom: 28,
  swapped: false,
  busyOpen: false,
  busyTitle: '',
  busyMsg: '',
  busySteps: false,
  loginStage: '',
  successOpen: false,
  accountKey: '',
  keepExisting: loadKeepExisting(),
  accounts: [] as Account[],
  selected: {} as Record<string, boolean>,
  dragSelect: false,
  dragMode: true,
  status: '',
  statusKind: 'ok' as 'ok' | 'err' | '',
  loading: false,
  exportBusy: false,
  showExportPick: false,
  exportPendingNames: null as string[] | null,
  showHelp: false,
  showSettings: false,
  driveBusy: false,
  driveAuthWait: false,
  driveClientId: '',
  driveClientSecret: '',
  driveStatus: { hasCredentials: false, connected: false, clientIdHint: '' } as DriveStatus,
  showUpdate: false,
  showAccInfo: false,
  logs: loadLogs() as LogEntry[],
  sysStatus: null as SysStatus | null,
  sysBusy: false,
  infoBusy: false,
  infoData: null as InfoData | null,
  updateBusy: false,
  updateInfo: null as UpdateInfo | null,
  checkKey: '',
  infoLogin: null as InfoLogin | null,
  bulkView: false,
  bulkKeys: '',
  bulkBusy: false,
  bulkResult: null as { total: number; items: BulkItem[] } | null,
  bulkLive: [] as BulkItem[],
  bulkTotal: 0,
  exportToast: '',
  settingsProxies: loadProxies(),
  deleteConfirm: null as string[] | null,
  avatarSync: null as { done: number; total: number } | null,
});

$effect.root(() => {
  $effect(() => {
    try {
      localStorage.setItem(KEEP_EXISTING_KEY, app.keepExisting ? '1' : '0');
    } catch {
    }
  });
  $effect(() => {
    try {
      localStorage.setItem(MANAGEMENT_KEY, app.management ? '1' : '0');
    } catch {
    }
  });
  $effect(() => {
    try {
      localStorage.setItem(SHOW_PERSONA_KEY, app.showPersona ? '1' : '0');
    } catch {
    }
  });
  $effect(() => {
    if (!app.management && app.mode !== 'simple') setMode('simple');
  });
});

let successTimer: ReturnType<typeof setTimeout> | undefined;
let exportTimer: ReturnType<typeof setTimeout> | undefined;

const stageIdx = $derived(
  LOGIN_STAGES.indexOf(app.loginStage as (typeof LOGIN_STAGES)[number]),
);
const busyProgressVal = $derived(
  stageIdx < 0 ? 6 : ((stageIdx + 1) / LOGIN_STAGES.length) * 100,
);
export function busyProgress(): number {
  return busyProgressVal;
}

const releaseNotesHtmlVal = $derived(
  app.updateInfo?.releaseNotes ? renderMarkdown(app.updateInfo.releaseNotes) : '',
);
export function releaseNotesHtml(): string {
  return releaseNotesHtmlVal;
}

const selectedNamesVal = $derived(
  app.accounts.filter((a) => app.selected[a.name]).map((a) => a.name),
);
export function selectedNames(): string[] {
  return selectedNamesVal;
}

const allSelectedVal = $derived(
  app.accounts.length > 0 && selectedNamesVal.length === app.accounts.length,
);
export function allSelected(): boolean {
  return allSelectedVal;
}

const bulkItemsVal = $derived(
  app.bulkBusy || !app.bulkResult ? app.bulkLive : app.bulkResult.items,
);
export function bulkItems(): BulkItem[] {
  return bulkItemsVal;
}

const bulkOkCountVal = $derived(bulkItemsVal.filter((i) => i.status === 'ok').length);
export function bulkOkCount(): number {
  return bulkOkCountVal;
}

const bulkBadCountVal = $derived(bulkItemsVal.length - bulkOkCountVal);
export function bulkBadCount(): number {
  return bulkBadCountVal;
}

export async function notify(ok: boolean, message: string) {
  if (ok) return;
  const title = t(app.lang, 'errorTitle');
  const text = translateBackendMessage(app.lang, message);
  try {
    await AppService.Notify(false, title, text);
  } catch {
    window.alert(`${title}\n\n${text}`);
  }
}

export function stageLabel(id: string): string {
  const m: Record<string, keyof import('./i18n').Dict> = {
    find_steam: 'stageFindSteam',
    stop_steam: 'stageStopSteam',
    loginusers: 'stageLoginUsers',
    connectcache: 'stageConnectCache',
    config_vdf: 'stageConfig',
    registry: 'stageRegistry',
    acl: 'stageAcl',
    launch: 'stageLaunch',
    wait_window: 'stageWaitWindow',
  };
  const key = m[id];
  return key ? t(app.lang, key) : t(app.lang, 'working');
}

export function openBusy(title: string, msg = '', steps = false) {
  app.busyTitle = title;
  app.busyMsg = msg;
  app.busySteps = steps;
  app.loginStage = '';
  app.busyOpen = true;
}

export function closeBusy() {
  app.busyOpen = false;
  app.loginStage = '';
}

export function showSuccess() {
  app.successOpen = true;
  clearTimeout(successTimer);
  successTimer = setTimeout(() => (app.successOpen = false), 2600);
}

export function openHelp() {
  app.showHelp = true;
}

export function confirmHideHint() {
  app.showHintConfirm = false;
  app.showHelpHint = false;
  try {
    localStorage.setItem(HELP_HINT_KEY, '1');
  } catch {
  }
  app.showHelp = true;
}

export async function checkUpdates(silent = false) {
  if (!silent) {
    setStatus(t(app.lang, 'updateChecking'), '', true);
  }
  try {
    const info = await (AppService as any).CheckForUpdates();
    app.updateInfo = info;
    if (info?.error && !silent) {
      setStatus(info.error, 'err', true);
      await notify(false, info.error);
      return;
    }
    if (info?.updateAvailable) {
      app.showUpdate = true;
      if (!silent) {
        setStatus(
          t(app.lang, 'updateAvailable', { current: info.currentVersion, latest: info.latestVersion }),
          'ok',
          true,
        );
      }
    } else if (!silent) {
      const msg = t(app.lang, 'updateNone');
      setStatus(msg, 'ok', true);
      await notify(true, msg);
    }
  } catch (e) {
    if (!silent) {
      setStatus(String(e), 'err');
      await notify(false, String(e));
    }
  }
}

export async function installUpdate() {
  if (!app.updateInfo?.downloadUrl) {
    if (app.updateInfo?.releaseUrl) {
      try {
        await (AppService as any).OpenURL(app.updateInfo.releaseUrl);
      } catch {
        await Browser.OpenURL(app.updateInfo.releaseUrl);
      }
    }
    return;
  }
  app.updateBusy = true;
  setStatus(t(app.lang, 'updateInstalling'), '', true);
  try {
    const res = (await (AppService as any).InstallUpdate(app.updateInfo.downloadUrl)) as Result;
    if (!res.ok) {
      setStatus(res.message, 'err');
      await notify(false, res.message || t(app.lang, 'updateFailed'));
      app.updateBusy = false;
      return;
    }
  } catch (e) {
    setStatus(String(e), 'err');
    await notify(false, String(e));
    app.updateBusy = false;
  }
}

export async function openReleasePage() {
  const url = app.updateInfo?.releaseUrl || app.updateInfo?.downloadUrl;
  if (!url) return;
  try {
    await (AppService as any).OpenURL(url);
  } catch {
    await Browser.OpenURL(url);
  }
}

export function setLang(next: Lang) {
  if (next === app.lang) return;
  calmUpdate(() => {
    app.lang = next;
    saveLang(next);
    if (app.statusKind === 'ok' && (app.status === t('en', 'ready') || app.status === t('ru', 'ready') || app.status === 'Ready' || app.status === 'Готово')) {
      app.status = t(next, 'ready');
    }
  });
}

export function setTheme(next: Theme) {
  app.theme = next;
  saveTheme(next);
  applyTheme(next);
}

export function addLog(kind: LogEntry['kind'], text: string) {
  if (!text) return;
  const last = app.logs[app.logs.length - 1];
  if (last && last.text === text) return;
  app.logs = [...app.logs, { ts: Date.now(), kind, text }].slice(-LOG_CAP);
  try {
    localStorage.setItem(LOG_KEY, JSON.stringify(app.logs));
  } catch {
  }
}

export function clearLogs() {
  app.logs = [];
  try {
    localStorage.removeItem(LOG_KEY);
  } catch {
  }
}

export function fmtTime(ts: number): string {
  return new Date(ts).toLocaleTimeString(app.lang === 'ru' ? 'ru-RU' : 'en-US', { hour12: false });
}

export async function refreshSysStatus() {
  app.sysBusy = true;
  try {
    app.sysStatus = await (AppService as any).GetSystemStatus();
  } catch (e) {
    setStatus(String(e), 'err');
  } finally {
    app.sysBusy = false;
  }
}

export function setMode(next: Mode) {
  if (next !== app.mode) {
    app.slideFrom = (MODE_ORDER.indexOf(next) >= MODE_ORDER.indexOf(app.mode) ? 1 : -1) * 28;
    app.swapped = true;
    app.bulkView = false;
  }
  app.mode = next;
  try {
    localStorage.setItem(MODE_KEY, next);
  } catch {
  }
  if (next === 'simple') {
    app.selected = {};
    closeExportPick();
  }
  if (next === 'logs') {
    void refreshSysStatus();
  }
}

export function toggleShowPersona() {
  app.showPersona = !app.showPersona;
}

// openBulkView shows the batch check screen; it lives under the Checker tab,
// so the tab pill moves there too and the back arrow returns to the Checker.
export function openBulkView() {
  if (app.bulkView) return;
  app.slideFrom = 28;
  app.swapped = true;
  app.bulkView = true;
  if (app.mode !== 'logs') {
    app.mode = 'logs';
    try {
      localStorage.setItem(MODE_KEY, 'logs');
    } catch {
    }
    void refreshSysStatus();
  }
}

export function closeBulkView() {
  if (!app.bulkView) return;
  app.slideFrom = -28;
  app.swapped = true;
  app.bulkView = false;
}

export async function refreshAccounts() {
  try {
    const list = await AppService.ListAccounts();
    app.accounts = (list ?? []).slice().sort((a, b) => a.name.localeCompare(b.name));
    void (AppService as any).PrefetchAvatars?.();
    const next: Record<string, boolean> = {};
    for (const a of app.accounts) {
      if (app.selected[a.name]) next[a.name] = true;
    }
    app.selected = next;
  } catch (e) {
    setStatus(String(e), 'err');
  }
}

export function setSelect(name: string, on: boolean) {
  if (!!app.selected[name] === on) return;
  app.selected = { ...app.selected, [name]: on };
}

export function toggleSelect(name: string) {
  setSelect(name, !app.selected[name]);
}

export function startDragSelect(e: PointerEvent, name: string) {
  if (app.mode !== 'advanced') return;
  if (e.button !== 0) return;
  e.preventDefault();
  e.stopPropagation();
  const next = !app.selected[name];
  app.dragSelect = true;
  app.dragMode = next;
  setSelect(name, next);
}

export function paintSelect(name: string) {
  if (!app.dragSelect) return;
  setSelect(name, app.dragMode);
}

export function endDragSelect() {
  app.dragSelect = false;
}

export function toggleSelectAll() {
  if (allSelectedVal) {
    app.selected = {};
    return;
  }
  const next: Record<string, boolean> = {};
  for (const a of app.accounts) next[a.name] = true;
  app.selected = next;
}

export function askExport(names: string[]) {
  app.exportPendingNames = names;
  app.showExportPick = true;
}

export function closeExportPick() {
  app.showExportPick = false;
  app.exportPendingNames = null;
}

export async function exportAccounts(all: boolean) {
  const names = all ? [] : selectedNamesVal;
  if (!all && names.length === 0) {
    const msg = t(app.lang, 'exportNone');
    setStatus(msg, 'err', true);
    await notify(false, msg);
    return;
  }
  askExport(names);
}

export async function exportOne(name: string) {
  askExport([name]);
}

export async function exportToClipboard() {
  const names = app.exportPendingNames;
  if (names === null) return;
  closeExportPick();
  app.exportBusy = true;
  try {
    const text = await (AppService as any).ExportTokens(names);
    await navigator.clipboard.writeText(text);
    const msg = t(app.lang, 'exportCopied');
    setStatus(msg, 'ok', true);
    await notify(true, msg);
  } catch (e) {
    const msg = String(e);
    setStatus(msg, 'err');
    await notify(false, msg);
  } finally {
    app.exportBusy = false;
  }
}

export async function exportToFile() {
  const names = app.exportPendingNames;
  if (names === null) return;
  closeExportPick();
  app.exportBusy = true;
  try {
    const res = (await (AppService as any).ExportTokensToFile(names)) as Result;
    if (res.message === 'Cancelled' || res.message === 'cancelled') {
      setStatus(t(app.lang, 'cancelled'), 'ok', true);
      return;
    }
    setStatus(res.message, res.ok ? 'ok' : 'err');
    await notify(res.ok, res.message);
  } catch (e) {
    const msg = String(e);
    setStatus(msg, 'err');
    await notify(false, msg);
  } finally {
    app.exportBusy = false;
  }
}

export async function refreshDriveStatus() {
  try {
    const st = await (AppService as any).GoogleDriveStatus();
    app.driveStatus = {
      hasCredentials: !!st?.hasCredentials,
      connected: !!st?.connected,
      clientIdHint: st?.clientIdHint || '',
    };
  } catch {
  }
}

export async function openSettings() {
  await refreshDriveStatus();
  app.showSettings = true;
}

export async function openDriveGuide() {
  try {
    await (AppService as any).OpenDriveGuide();
  } catch (e) {
    setStatus(String(e), 'err');
  }
}

export async function saveDriveCreds() {
  app.driveBusy = true;
  try {
    const res = (await (AppService as any).SaveGoogleCredentials(
      app.driveClientId.trim(),
      app.driveClientSecret.trim(),
    )) as Result;
    setStatus(res.message, res.ok ? 'ok' : 'err');
    await notify(res.ok, res.message);
    if (res.ok) {
      app.driveClientId = '';
      app.driveClientSecret = '';
      await refreshDriveStatus();
    }
  } catch (e) {
    const msg = String(e);
    setStatus(msg, 'err');
    await notify(false, msg);
  } finally {
    app.driveBusy = false;
  }
}

export async function importDriveCreds() {
  app.driveBusy = true;
  try {
    const res = (await (AppService as any).ImportGoogleCredentials()) as Result;
    if (res.message !== 'Cancelled' && res.message !== 'cancelled') {
      setStatus(res.message, res.ok ? 'ok' : 'err');
      await notify(res.ok, res.message);
    }
    await refreshDriveStatus();
  } catch (e) {
    const msg = String(e);
    setStatus(msg, 'err');
    await notify(false, msg);
  } finally {
    app.driveBusy = false;
  }
}

export async function cancelDriveAuth() {
  try {
    await (AppService as any).CancelGoogleAuth();
  } catch {
  }
  app.driveAuthWait = false;
  app.driveBusy = false;
  setStatus(t(app.lang, 'cancelled'), 'ok', true);
}

export async function closeSettings() {
  if (app.driveAuthWait) {
    await cancelDriveAuth();
  }
  app.showSettings = false;
}

export async function connectDrive() {
  app.driveBusy = true;
  app.driveAuthWait = true;
  setStatus(t(app.lang, 'driveWaiting'), '', true);
  try {
    const res = (await (AppService as any).ConnectGoogleDrive()) as Result;
    const cancelled =
      res.message === 'Cancelled' ||
      res.message === 'cancelled' ||
      (res.message || '').toLowerCase().includes('cancelled');
    if (cancelled) {
      setStatus(t(app.lang, 'cancelled'), 'ok', true);
    } else {
      setStatus(res.message, res.ok ? 'ok' : 'err');
      await notify(res.ok, res.message);
    }
    await refreshDriveStatus();
  } catch (e) {
    const msg = String(e);
    if (msg.toLowerCase().includes('cancel')) {
      setStatus(t(app.lang, 'cancelled'), 'ok', true);
    } else {
      setStatus(msg, 'err');
      await notify(false, msg);
    }
  } finally {
    app.driveAuthWait = false;
    app.driveBusy = false;
  }
}

export async function disconnectDrive() {
  app.driveBusy = true;
  try {
    const res = (await (AppService as any).DisconnectGoogleDrive()) as Result;
    setStatus(res.message, res.ok ? 'ok' : 'err');
    await notify(res.ok, res.message);
    await refreshDriveStatus();
  } catch (e) {
    const msg = String(e);
    setStatus(msg, 'err');
    await notify(false, msg);
  } finally {
    app.driveBusy = false;
  }
}

export async function exportToDrive() {
  const names = app.exportPendingNames;
  if (names === null) return;
  closeExportPick();
  await refreshDriveStatus();
  if (!app.driveStatus.hasCredentials) {
    const msg =
      app.lang === 'ru'
        ? 'Сначала настройте Google Drive в Настройках'
        : 'Please configure Google Drive in Settings first';
    setStatus(msg, 'err', true);
    await notify(false, msg);
    await openSettings();
    return;
  }
  app.exportBusy = true;
  app.driveBusy = true;
  app.driveAuthWait = !app.driveStatus.connected;
  setStatus(app.driveAuthWait ? t(app.lang, 'driveWaiting') : t(app.lang, 'driveUploading'), '', true);
  try {
    const res = (await (AppService as any).ExportTokensToGoogleDrive(names)) as Result;
    const cancelled =
      res.message === 'Cancelled' ||
      res.message === 'cancelled' ||
      (res.message || '').toLowerCase().includes('cancelled');
    if (cancelled) {
      setStatus(t(app.lang, 'cancelled'), 'ok', true);
    } else {
      setStatus(res.message, res.ok ? 'ok' : 'err');
      await notify(res.ok, res.message);
    }
    await refreshDriveStatus();
  } catch (e) {
    const msg = String(e);
    if (msg.toLowerCase().includes('cancel')) {
      setStatus(t(app.lang, 'cancelled'), 'ok', true);
    } else {
      setStatus(msg, 'err');
      await notify(false, msg);
    }
  } finally {
    app.driveAuthWait = false;
    app.driveBusy = false;
    app.exportBusy = false;
  }
}

export function setStatus(msg: string, kind: 'ok' | 'err' | '' = '', alreadyTranslated = false) {
  const text = alreadyTranslated ? msg : translateBackendMessage(app.lang, msg);
  app.status = text;
  app.statusKind = kind;
  addLog(kind === 'err' ? 'err' : kind === 'ok' ? 'ok' : 'info', text);
}

export function saveProxies() {
  try {
    localStorage.setItem(PROXY_KEY, app.settingsProxies);
  } catch {
  }
}

export async function bulkPickFile() {
  try {
    const text = (await (AppService as any).PickTextFile()) as string;
    if (text?.trim()) {
      app.bulkKeys = app.bulkKeys.trim() ? app.bulkKeys.trimEnd() + '\n' + text.trim() : text.trim();
    }
  } catch (e) {
    setStatus(String(e), 'err');
  }
}

export function bulkInfo(it: BulkItem) {
  const name = it.account.split(':')[0].trim();
  if (!name) return;
  const saved = app.accounts.find((a) => a.name.toLowerCase() === name.toLowerCase());
  if (saved) {
    openAccInfo(saved.name);
    return;
  }
  const line = app.bulkKeys
    .split(/\r?\n/)
    .map((l) => l.trim())
    .find((l) => l.toLowerCase().startsWith(name.toLowerCase() + '----'));
  if (line) {
    app.infoLogin = { kind: 'key', key: line };
    void showInfoFor((AppService as any).CheckAccountKey(line));
  }
}

export function bulkStatusText(status: string): string {
  const m: Record<string, keyof import('./i18n').Dict> = {
    ok: 'bstOk',
    rejected: 'bstRejected',
    expired: 'bstExpired',
    invalid: 'bstInvalid',
    error: 'bstError',
  };
  const key = m[status];
  return key ? t(app.lang, key) : status;
}

export async function runBulkCheck(saved: boolean) {
  if (!saved && !app.bulkKeys.trim()) return;
  app.bulkBusy = true;
  app.bulkResult = null;
  app.bulkLive = [];
  app.bulkTotal = saved
    ? app.accounts.length
    : app.bulkKeys.split(/\r?\n/).filter((l) => l.trim()).length;
  try {
    app.bulkResult = saved
      ? await (AppService as any).CheckSavedAccounts(app.settingsProxies)
      : await (AppService as any).CheckAccountKeys(app.bulkKeys, app.settingsProxies);
  } catch (e) {
    const msg = String(e);
    setStatus(msg, 'err');
    await notify(false, msg);
  } finally {
    app.bulkBusy = false;
  }
}

export async function exportBulk(which: 'ok' | 'bad') {
  try {
    const path = (await (AppService as any).ExportBulkResults(which)) as string;
    if (!path) return;
    clearTimeout(exportTimer);
    app.exportToast = path;
    exportTimer = setTimeout(() => (app.exportToast = ''), 2600);
  } catch (e) {
    const msg = String(e);
    if (msg.toLowerCase().includes('cancel')) {
      setStatus(t(app.lang, 'cancelled'), 'ok', true);
      return;
    }
    setStatus(msg, 'err');
    await notify(false, msg);
  }
}

export async function exportBulkToDrive(which: 'ok' | 'bad') {
  await refreshDriveStatus();
  if (!app.driveStatus.hasCredentials) {
    const msg =
      app.lang === 'ru'
        ? 'Сначала настройте Google Drive в Настройках'
        : 'Please configure Google Drive in Settings first';
    setStatus(msg, 'err', true);
    await notify(false, msg);
    await openSettings();
    return;
  }
  app.driveBusy = true;
  try {
    const res = (await (AppService as any).ExportBulkResultsToGoogleDrive(which)) as Result;
    setStatus(res.message, res.ok ? 'ok' : 'err');
    await notify(res.ok, res.message);
  } catch (e) {
    const msg = String(e);
    setStatus(msg, 'err');
    await notify(false, msg);
  } finally {
    app.driveBusy = false;
  }
}

export async function showInfoFor(promise: Promise<any>) {
  app.showAccInfo = true;
  app.infoData = null;
  app.infoBusy = true;
  try {
    app.infoData = await promise;
  } catch (e) {
    app.showAccInfo = false;
    app.infoLogin = null;
    const msg = String(e);
    setStatus(msg, 'err');
    await notify(false, msg);
  } finally {
    app.infoBusy = false;
  }
}

export function openAccInfo(name: string) {
  app.infoLogin = { kind: 'saved', name };
  void showInfoFor((AppService as any).GetAccountInfo(name));
}

export function runKeyCheck() {
  const key = app.checkKey.trim();
  if (!key) return;
  app.infoLogin = { kind: 'key', key };
  void showInfoFor((AppService as any).CheckAccountKey(key));
  setStatus(t(app.lang, 'keyCheckDone'), 'ok', true);
}

export async function loginFromInfo() {
  if (!app.infoLogin) return;
  const src = app.infoLogin;
  app.loading = true;
  openBusy(
    src.kind === 'saved' ? t(app.lang, 'loggingInAs', { name: src.name }) : t(app.lang, 'loggingIn'),
    '',
    true,
  );
  try {
    const res = (
      src.kind === 'key'
        ? await AppService.LoginFromKey(src.key, app.keepExisting)
        : await AppService.LoginSaved(src.name, app.keepExisting)
    ) as Result;
    setStatus(res.message, res.ok ? 'ok' : 'err');
    await notify(res.ok, res.message);
    if (res.ok) {
      app.showAccInfo = false;
      app.infoLogin = null;
      if (src.kind === 'key') app.accountKey = '';
      showSuccess();
      await refreshAccounts();
    }
  } catch (e) {
    const msg = String(e);
    setStatus(msg, 'err');
    await notify(false, msg);
  } finally {
    closeBusy();
    app.loading = false;
  }
}

export function trVisibility(v: string): string {
  const m: Record<string, string> = {
    public: t(app.lang, 'visPublic'),
    private: t(app.lang, 'visPrivate'),
    friendsonly: t(app.lang, 'visFriends'),
  };
  return m[(v || '').toLowerCase()] || v || '—';
}

export function trOnline(v: string): string {
  const m: Record<string, string> = {
    online: t(app.lang, 'stOnline'),
    offline: t(app.lang, 'stOffline'),
    'in-game': t(app.lang, 'stInGame'),
  };
  return m[(v || '').toLowerCase()] || v || '—';
}

export function yn(b: boolean): string {
  return b ? t(app.lang, 'yes') : t(app.lang, 'no');
}

export function checkDesc(id: string): string {
  const m: Record<string, keyof import('./i18n').Dict> = {
    jwt_structure: 'dJwtStructure',
    signature: 'dSignature',
    issuer: 'dIssuer',
    audience: 'dAudience',
    steamid_format: 'dSteamIDFormat',
    not_expired: 'dNotExpired',
    iat_past: 'dIatPast',
    iat_before_exp: 'dIatBeforeExp',
    nbf_ok: 'dNbfOk',
    rt_exp_ok: 'dRtExpOk',
    access_minted: 'dAccessMinted',
    profile_match: 'dProfileMatch',
  };
  const key = m[id];
  return key ? t(app.lang, key) : '';
}

export function checkLabel(id: string): string {
  const m: Record<string, keyof import('./i18n').Dict> = {
    jwt_structure: 'chkJwtStructure',
    signature: 'chkSignature',
    issuer: 'chkIssuer',
    audience: 'chkAudience',
    steamid_format: 'chkSteamIDFormat',
    not_expired: 'chkNotExpired',
    iat_past: 'chkIatPast',
    iat_before_exp: 'chkIatBeforeExp',
    nbf_ok: 'chkNbfOk',
    rt_exp_ok: 'chkRtExpOk',
    access_minted: 'chkAccessMinted',
    profile_match: 'chkProfileMatch',
  };
  const key = m[id];
  return key ? t(app.lang, key) : id;
}

export function singleLineKey(): string {
  return app.accountKey.split(/\r?\n/)[0]?.trim() ?? '';
}

export async function saveFromInfo() {
  if (!app.infoLogin || app.infoLogin.kind !== 'key') return;
  app.loading = true;
  try {
    const res = (await (AppService as any).SaveAccountKey(app.infoLogin.key)) as Result;
    setStatus(res.message, res.ok ? 'ok' : 'err');
    await notify(res.ok, res.message);
    if (res.ok) {
      app.accountKey = '';
      await refreshAccounts();
    }
  } catch (e) {
    const msg = String(e);
    setStatus(msg, 'err');
    await notify(false, msg);
  } finally {
    app.loading = false;
  }
}

export async function doLogin() {
  const key = singleLineKey();
  if (!key) {
    const msg = t(app.lang, 'enterKey');
    setStatus(msg, 'err', true);
    await notify(false, msg);
    return;
  }
  app.loading = true;
  setStatus(t(app.lang, 'loggingIn'), '', true);
  openBusy(t(app.lang, 'loggingIn'), '', true);
  try {
    const res = (await AppService.LoginFromKey(key, app.keepExisting)) as Result;
    setStatus(res.message, res.ok ? 'ok' : 'err');
    await notify(res.ok, res.message);
    if (res.ok) {
      app.accountKey = '';
      showSuccess();
      await refreshAccounts();
    }
  } catch (e) {
    const msg = String(e);
    setStatus(msg, 'err');
    await notify(false, msg);
  } finally {
    closeBusy();
    app.loading = false;
  }
}

export async function importFromField() {
  const key = singleLineKey();
  if (!key) {
    const msg = t(app.lang, 'enterKey');
    setStatus(msg, 'err', true);
    await notify(false, msg);
    return;
  }
  app.loading = true;
  try {
    const res = (await (AppService as any).ImportTokens(key)) as Result;
    setStatus(res.message, res.ok ? 'ok' : 'err');
    await notify(res.ok, res.message);
    if (res.ok) {
      app.accountKey = '';
      await refreshAccounts();
    }
  } catch (e) {
    const msg = String(e);
    setStatus(msg, 'err');
    await notify(false, msg);
  } finally {
    app.loading = false;
  }
}

export async function importFromFile() {
  app.loading = true;
  try {
    const res = (await (AppService as any).ImportTokensFromFile()) as Result;
    if (res.message === 'Cancelled' || res.message === 'cancelled') {
      setStatus(t(app.lang, 'cancelled'), 'ok', true);
      return;
    }
    setStatus(res.message, res.ok ? 'ok' : 'err');
    await notify(res.ok, res.message);
    if (res.ok) await refreshAccounts();
  } catch (e) {
    const msg = String(e);
    setStatus(msg, 'err');
    await notify(false, msg);
  } finally {
    app.loading = false;
  }
}

export async function loginSaved(name: string) {
  setStatus(t(app.lang, 'loggingInAs', { name }), '', true);
  openBusy(t(app.lang, 'loggingInAs', { name }), '', true);
  try {
    const res = (await AppService.LoginSaved(name, app.keepExisting)) as Result;
    setStatus(res.message, res.ok ? 'ok' : 'err');
    await notify(res.ok, res.message);
    if (res.ok) showSuccess();
    await refreshAccounts();
  } catch (e) {
    const msg = String(e);
    setStatus(msg, 'err');
    await notify(false, msg);
  } finally {
    closeBusy();
  }
}

export function clearSelectionNames(names: string[]) {
  const drop = new Set(names.map((n) => n.toLowerCase()));
  const next: Record<string, boolean> = {};
  for (const [k, v] of Object.entries(app.selected)) {
    if (!drop.has(k.toLowerCase())) next[k] = v;
  }
  app.selected = next;
}

export function deleteAccount(name: string, e?: Event) {
  e?.preventDefault();
  e?.stopPropagation();
  endDragSelect();
  const acc = name.trim();
  if (!acc) return;
  app.deleteConfirm = [acc];
}

export function deleteSelected() {
  endDragSelect();
  const names = selectedNamesVal.slice();
  if (names.length === 0) {
    const msg = t(app.lang, 'exportNone');
    setStatus(msg, 'err', true);
    return;
  }
  app.deleteConfirm = names;
}

export function cancelDelete() {
  app.deleteConfirm = null;
}

export async function confirmDelete() {
  const names = app.deleteConfirm ?? [];
  app.deleteConfirm = null;
  if (names.length === 1) {
    await doDeleteAccount(names[0]);
  } else if (names.length > 1) {
    await doDeleteAccounts(names);
  }
}

async function doDeleteAccount(acc: string) {
  try {
    const res = (await AppService.DeleteAccount(acc)) as Result;
    setStatus(res.message, res.ok ? 'ok' : 'err');
    await notify(res.ok, res.message);
    if (res.ok) {
      clearSelectionNames([acc]);
      app.accounts = app.accounts.filter((a) => a.name.toLowerCase() !== acc.toLowerCase());
    }
    await refreshAccounts();
  } catch (err) {
    const msg = String(err);
    setStatus(msg, 'err');
    await notify(false, msg);
  }
}

async function doDeleteAccounts(names: string[]) {
  try {
    const res = (await (AppService as any).DeleteAccounts(names)) as Result;
    setStatus(res.message, res.ok ? 'ok' : 'err');
    await notify(res.ok, res.message);
    if (res.ok) {
      const drop = new Set(names.map((n) => n.toLowerCase()));
      app.accounts = app.accounts.filter((a) => !drop.has(a.name.toLowerCase()));
      app.selected = {};
    }
    await refreshAccounts();
  } catch (err) {
    const msg = String(err);
    setStatus(msg, 'err');
    await notify(false, msg);
  }
}

export async function dumpDiagnostics() {
  try {
    const journal = app.logs
      .slice(-200)
      .map((l) => `${new Date(l.ts).toLocaleString()} [${l.kind}] ${l.text}`)
      .join('\n');
    const res = (await (AppService as any).SaveDiagnosticsDump(journal)) as Result;
    setStatus(res.message, res.ok ? 'ok' : 'err');
    await notify(res.ok, res.message);
  } catch (e) {
    setStatus(String(e), 'err');
  }
}

export async function harvestSteam() {
  try {
    const res = (await (AppService as any).HarvestSteamAccounts()) as Result;
    setStatus(res.message, res.ok ? 'ok' : 'err');
    await notify(res.ok, res.message);
    if (res.ok) await refreshAccounts();
  } catch (e) {
    const msg = String(e);
    setStatus(msg, 'err');
    await notify(false, msg);
  }
}

export async function resetSteam() {
  setStatus(t(app.lang, 'resettingSteam'), '', true);
  openBusy(t(app.lang, 'resettingSteam'));
  try {
    const res = (await AppService.ResetSteam()) as Result;
    setStatus(res.message, res.ok ? 'ok' : 'err');
    if (res.message !== 'Cancelled' && res.message !== 'cancelled') {
      await notify(res.ok, res.message);
    }
  } catch (e) {
    const msg = String(e);
    setStatus(msg, 'err');
    await notify(false, msg);
  } finally {
    closeBusy();
  }
}

export function onKey(e: KeyboardEvent) {
  if (e.key !== 'Enter') return;
  doLogin();
}
