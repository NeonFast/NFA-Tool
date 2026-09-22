<script lang="ts">
  import { onMount } from 'svelte';
  import { fly } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import { Browser, Events } from '@wailsio/runtime';
  import { AppService } from '../bindings/nfa-tool';
  import {
    type Lang,
    loadLang,
    saveLang,
    t,
    translateBackendMessage,
    localizeExpiry,
  } from './i18n';
  import { renderMarkdown } from './markdown';
  import { type Theme, applyTheme, loadTheme, saveTheme } from './theme';

  async function notify(ok: boolean, message: string) {
    if (ok) return;
    const title = t(lang, 'errorTitle');
    const text = translateBackendMessage(lang, message);
    try {
      await AppService.Notify(false, title, text);
    } catch {
      window.alert(`${title}\n\n${text}`);
    }
  }

  const GUIDE_URL = 'https://teletype.in/@hackerdlc/CS2NFA';

  type Mode = 'simple' | 'advanced' | 'logs';
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

  type LogEntry = { ts: number; kind: 'ok' | 'err' | 'info'; text: string };
  const LOG_KEY = 'nfa-tool-log';
  const LOG_CAP = 200;

  function loadLogs(): LogEntry[] {
    try {
      const raw = localStorage.getItem(LOG_KEY);
      if (raw) {
        const arr = JSON.parse(raw);
        if (Array.isArray(arr)) return arr.slice(0, LOG_CAP);
      }
    } catch {
    }
    return [];
  }

  function aimHint(node: HTMLElement) {
    const update = () => {
      try {
        const fab = node.querySelector<HTMLElement>('.help-fab');
        if (!fab) return;
        const centerX = fab.offsetLeft + fab.offsetWidth / 2;
        node.style.setProperty('--aim-r', `${node.offsetWidth - centerX}px`);
      } catch {
      }
    };
    update();
    const ro = new ResizeObserver(update);
    ro.observe(node);
    return {
      destroy() {
        ro.disconnect();
      },
    };
  }

  function slidingPill(node: HTMLElement) {
    let raf = 0;
    const update = () => {
      cancelAnimationFrame(raf);
      raf = requestAnimationFrame(() => {
        try {
          const active =
            node.querySelector<HTMLElement>('.lang-btn.active') ??
            node.querySelector<HTMLElement>('.lang-btn');
          if (!active) return;
          node.style.setProperty('--pill-x', `${active.offsetLeft - node.clientLeft}px`);
          node.style.setProperty('--pill-w', `${active.offsetWidth}px`);
        } catch {
        }
      });
    };
    update();
    const ro = new ResizeObserver(update);
    ro.observe(node);
    const mo = new MutationObserver(update);
    mo.observe(node, { attributes: true, subtree: true, attributeFilter: ['class'] });
    return {
      destroy() {
        cancelAnimationFrame(raf);
        ro.disconnect();
        mo.disconnect();
      },
    };
  }

  type Account = {
    name: string;
    expiresIn: string;
    valid: boolean;
    avatar?: string;
  };

  type Result = {
    ok: boolean;
    message: string;
  };

  let appName = $state('NFA Tool');
  let version = $state('');
  let lang = $state<Lang>(loadLang());
  let theme = $state<Theme>(loadTheme());
  let mode = $state<Mode>(loadMode());
  let showHelpHint = $state(loadHelpHint());
  let showHintConfirm = $state(false);
  let slideFrom = $state(28);
  let swapped = $state(false);

  const LOGIN_STAGES = [
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
  let busyOpen = $state(false);
  let busyTitle = $state('');
  let busyMsg = $state('');
  let busySteps = $state(false);
  let loginStage = $state('');

  const stageIdx = $derived(LOGIN_STAGES.indexOf(loginStage as (typeof LOGIN_STAGES)[number]));
  const busyProgress = $derived(stageIdx < 0 ? 6 : ((stageIdx + 1) / LOGIN_STAGES.length) * 100);

  function stageLabel(id: string): string {
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
    return key ? t(lang, key) : t(lang, 'working');
  }

  function openBusy(title: string, msg = '', steps = false) {
    busyTitle = title;
    busyMsg = msg;
    busySteps = steps;
    loginStage = '';
    busyOpen = true;
  }

  function closeBusy() {
    busyOpen = false;
    loginStage = '';
  }

  let successOpen = $state(false);
  let successTimer: ReturnType<typeof setTimeout> | undefined;
  function showSuccess() {
    successOpen = true;
    clearTimeout(successTimer);
    successTimer = setTimeout(() => (successOpen = false), 2600);
  }

  const HELP_HINT_KEY = 'nfa-tool-help-hint';

  function loadHelpHint(): boolean {
    try {
      return localStorage.getItem(HELP_HINT_KEY) !== '1';
    } catch {
      return true;
    }
  }

  function openHelp() {
    showHelp = true;
  }

  function confirmHideHint() {
    showHintConfirm = false;
    showHelpHint = false;
    try {
      localStorage.setItem(HELP_HINT_KEY, '1');
    } catch {
    }
    showHelp = true;
  }
  let accountKey = $state('');
  let keepExisting = $state(false);
  let accounts = $state<Account[]>([]);
  let selected = $state<Record<string, boolean>>({});
  let dragSelect = $state(false);
  let dragMode = $state(true);
  let status = $state('');
  let statusKind = $state<'ok' | 'err' | ''>('ok');
  let loading = $state(false);
  let exportBusy = $state(false);
  let showExportPick = $state(false);
  let exportPendingNames = $state<string[] | null>(null);
  let showHelp = $state(false);
  let showSettings = $state(false);
  let driveBusy = $state(false);
  let driveAuthWait = $state(false);
  let driveClientId = $state('');
  let driveClientSecret = $state('');
  let driveStatus = $state<{
    hasCredentials: boolean;
    connected: boolean;
    clientIdHint: string;
  }>({ hasCredentials: false, connected: false, clientIdHint: '' });
  let showUpdate = $state(false);
  let showAccInfo = $state(false);
  let logs = $state<LogEntry[]>(loadLogs());
  let sysStatus = $state<{
    version: string;
    steamRunning: boolean;
    steamPath: string;
    accountsTotal: number;
    accountsValid: number;
    driveConnected: boolean;
    steamApiOnline: boolean;
  } | null>(null);
  let sysBusy = $state(false);
  let infoBusy = $state(false);
  let infoData = $state<{
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
  } | null>(null);
  let updateBusy = $state(false);
  let updateInfo = $state<{
    updateAvailable: boolean;
    currentVersion: string;
    latestVersion: string;
    releaseUrl: string;
    downloadUrl: string;
    releaseNotes: string;
    error?: string;
  } | null>(null);

  const releaseNotesHtml = $derived(
    updateInfo?.releaseNotes ? renderMarkdown(updateInfo.releaseNotes) : '',
  );

  onMount(async () => {
    status = t(lang, 'ready');
    window.addEventListener('pointerup', endDragSelect);
    window.addEventListener('pointercancel', endDragSelect);
    window.addEventListener('blur', endDragSelect);
    const offStage = Events.On('login:stage', (ev) => {
      const st = (ev as { data?: { stage?: string } })?.data?.stage;
      if (typeof st === 'string') loginStage = st;
    });
    const offBulk = Events.On('bulk:item', (ev) => {
      if (!bulkBusy) return;
      const it = (ev as { data?: BulkItem })?.data;
      if (it) bulkLive = [...bulkLive, it];
    });
    try {
      const anySvc = AppService as typeof AppService & { GetAppName?: () => Promise<string> };
      if (typeof anySvc.GetAppName === 'function') {
        appName = await anySvc.GetAppName();
      }
      version = await AppService.GetVersion();
    } catch {
    }
    await refreshAccounts();
    void checkUpdates(true);
    return () => {
      window.removeEventListener('pointerup', endDragSelect);
      window.removeEventListener('pointercancel', endDragSelect);
      window.removeEventListener('blur', endDragSelect);
      offStage();
      offBulk();
    };
  });

  async function checkUpdates(silent = false) {
    if (!silent) {
      setStatus(t(lang, 'updateChecking'), '', true);
    }
    try {
      const info = await (AppService as any).CheckForUpdates();
      updateInfo = info;
      if (info?.error && !silent) {
        setStatus(info.error, 'err', true);
        await notify(false, info.error);
        return;
      }
      if (info?.updateAvailable) {
        showUpdate = true;
        if (!silent) {
          setStatus(
            t(lang, 'updateAvailable', { current: info.currentVersion, latest: info.latestVersion }),
            'ok',
            true,
          );
        }
      } else if (!silent) {
        const msg = t(lang, 'updateNone');
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

  async function installUpdate() {
    if (!updateInfo?.downloadUrl) {
      if (updateInfo?.releaseUrl) {
        try {
          await (AppService as any).OpenURL(updateInfo.releaseUrl);
        } catch {
          await Browser.OpenURL(updateInfo.releaseUrl);
        }
      }
      return;
    }
    updateBusy = true;
    setStatus(t(lang, 'updateInstalling'), '', true);
    try {
      const res = (await (AppService as any).InstallUpdate(updateInfo.downloadUrl)) as Result;
      if (!res.ok) {
        setStatus(res.message, 'err');
        await notify(false, res.message || t(lang, 'updateFailed'));
        updateBusy = false;
        return;
      }
    } catch (e) {
      setStatus(String(e), 'err');
      await notify(false, String(e));
      updateBusy = false;
    }
  }

  async function openReleasePage() {
    const url = updateInfo?.releaseUrl || updateInfo?.downloadUrl;
    if (!url) return;
    try {
      await (AppService as any).OpenURL(url);
    } catch {
      await Browser.OpenURL(url);
    }
  }

  function setLang(next: Lang) {
    lang = next;
    saveLang(next);
    if (statusKind === 'ok' && (status === t('en', 'ready') || status === t('ru', 'ready') || status === 'Ready' || status === 'Готово')) {
      status = t(next, 'ready');
    }
  }

  function setTheme(next: Theme) {
    theme = next;
    saveTheme(next);
    applyTheme(next);
  }

  function addLog(kind: LogEntry['kind'], text: string) {
    if (!text) return;
    if (logs[0] && logs[0].text === text) return;
    logs = [{ ts: Date.now(), kind, text }, ...logs].slice(0, LOG_CAP);
    try {
      localStorage.setItem(LOG_KEY, JSON.stringify(logs));
    } catch {
    }
  }

  function clearLogs() {
    logs = [];
    try {
      localStorage.removeItem(LOG_KEY);
    } catch {
    }
  }

  function fmtTime(ts: number): string {
    return new Date(ts).toLocaleTimeString(lang === 'ru' ? 'ru-RU' : 'en-US', { hour12: false });
  }

  async function refreshSysStatus() {
    sysBusy = true;
    try {
      sysStatus = await (AppService as any).GetSystemStatus();
    } catch (e) {
      setStatus(String(e), 'err');
    } finally {
      sysBusy = false;
    }
  }

  function setMode(next: Mode) {
    if (next !== mode) {
      slideFrom = (MODE_ORDER.indexOf(next) >= MODE_ORDER.indexOf(mode) ? 1 : -1) * 64;
      swapped = true;
    }
    mode = next;
    try {
      localStorage.setItem(MODE_KEY, next);
    } catch {
    }
    if (next === 'simple') {
      selected = {};
      closeExportPick();
    }
    if (next === 'logs') {
      void refreshSysStatus();
    }
  }

  async function refreshAccounts() {
    try {
      const list = await AppService.ListAccounts();
      accounts = (list ?? []).slice().sort((a, b) => a.name.localeCompare(b.name));
      const next: Record<string, boolean> = {};
      for (const a of accounts) {
        if (selected[a.name]) next[a.name] = true;
      }
      selected = next;
    } catch (e) {
      setStatus(String(e), 'err');
    }
  }

  const selectedNames = $derived(
    accounts.filter((a) => selected[a.name]).map((a) => a.name),
  );
  const allSelected = $derived(
    accounts.length > 0 && selectedNames.length === accounts.length,
  );

  function setSelect(name: string, on: boolean) {
    if (!!selected[name] === on) return;
    selected = { ...selected, [name]: on };
  }

  function toggleSelect(name: string) {
    setSelect(name, !selected[name]);
  }

  function startDragSelect(e: PointerEvent, name: string) {
    if (mode !== 'advanced') return;
    if (e.button !== 0) return;
    e.preventDefault();
    e.stopPropagation();
    const next = !selected[name];
    dragSelect = true;
    dragMode = next;
    setSelect(name, next);
  }

  function paintSelect(name: string) {
    if (!dragSelect) return;
    setSelect(name, dragMode);
  }

  function endDragSelect() {
    dragSelect = false;
  }

  function toggleSelectAll() {
    if (allSelected) {
      selected = {};
      return;
    }
    const next: Record<string, boolean> = {};
    for (const a of accounts) next[a.name] = true;
    selected = next;
  }

  function askExport(names: string[]) {
    exportPendingNames = names;
    showExportPick = true;
  }

  function closeExportPick() {
    showExportPick = false;
    exportPendingNames = null;
  }

  async function exportAccounts(all: boolean) {
    const names = all ? [] : selectedNames;
    if (!all && names.length === 0) {
      const msg = t(lang, 'exportNone');
      setStatus(msg, 'err', true);
      await notify(false, msg);
      return;
    }
    askExport(names);
  }

  async function exportOne(name: string) {
    askExport([name]);
  }

  async function exportToClipboard() {
    const names = exportPendingNames;
    if (names === null) return;
    closeExportPick();
    exportBusy = true;
    try {
      const text = await (AppService as any).ExportTokens(names);
      await navigator.clipboard.writeText(text);
      const msg = t(lang, 'exportCopied');
      setStatus(msg, 'ok', true);
      await notify(true, msg);
    } catch (e) {
      const msg = String(e);
      setStatus(msg, 'err');
      await notify(false, msg);
    } finally {
      exportBusy = false;
    }
  }

  async function exportToFile() {
    const names = exportPendingNames;
    if (names === null) return;
    closeExportPick();
    exportBusy = true;
    try {
      const res = (await (AppService as any).ExportTokensToFile(names)) as Result;
      if (res.message === 'Cancelled' || res.message === 'cancelled') {
        setStatus(t(lang, 'cancelled'), 'ok', true);
        return;
      }
      setStatus(res.message, res.ok ? 'ok' : 'err');
      await notify(res.ok, res.message);
    } catch (e) {
      const msg = String(e);
      setStatus(msg, 'err');
      await notify(false, msg);
    } finally {
      exportBusy = false;
    }
  }

  async function refreshDriveStatus() {
    try {
      const st = await (AppService as any).GoogleDriveStatus();
      driveStatus = {
        hasCredentials: !!st?.hasCredentials,
        connected: !!st?.connected,
        clientIdHint: st?.clientIdHint || '',
      };
    } catch {
    }
  }

  async function openSettings() {
    await refreshDriveStatus();
    showSettings = true;
  }

  async function openDriveGuide() {
    try {
      await (AppService as any).OpenDriveGuide();
    } catch (e) {
      setStatus(String(e), 'err');
    }
  }

  async function saveDriveCreds() {
    driveBusy = true;
    try {
      const res = (await (AppService as any).SaveGoogleCredentials(
        driveClientId.trim(),
        driveClientSecret.trim(),
      )) as Result;
      setStatus(res.message, res.ok ? 'ok' : 'err');
      await notify(res.ok, res.message);
      if (res.ok) {
        driveClientId = '';
        driveClientSecret = '';
        await refreshDriveStatus();
      }
    } catch (e) {
      const msg = String(e);
      setStatus(msg, 'err');
      await notify(false, msg);
    } finally {
      driveBusy = false;
    }
  }

  async function importDriveCreds() {
    driveBusy = true;
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
      driveBusy = false;
    }
  }

  async function cancelDriveAuth() {
    try {
      await (AppService as any).CancelGoogleAuth();
    } catch {
    }
    driveAuthWait = false;
    driveBusy = false;
    setStatus(t(lang, 'cancelled'), 'ok', true);
  }

  async function closeSettings() {
    if (driveAuthWait) {
      await cancelDriveAuth();
    }
    showSettings = false;
  }

  async function connectDrive() {
    driveBusy = true;
    driveAuthWait = true;
    setStatus(t(lang, 'driveWaiting'), '', true);
    try {
      const res = (await (AppService as any).ConnectGoogleDrive()) as Result;
      const cancelled =
        res.message === 'Cancelled' ||
        res.message === 'cancelled' ||
        (res.message || '').toLowerCase().includes('cancelled');
      if (cancelled) {
        setStatus(t(lang, 'cancelled'), 'ok', true);
      } else {
        setStatus(res.message, res.ok ? 'ok' : 'err');
        await notify(res.ok, res.message);
      }
      await refreshDriveStatus();
    } catch (e) {
      const msg = String(e);
      if (msg.toLowerCase().includes('cancel')) {
        setStatus(t(lang, 'cancelled'), 'ok', true);
      } else {
        setStatus(msg, 'err');
        await notify(false, msg);
      }
    } finally {
      driveAuthWait = false;
      driveBusy = false;
    }
  }

  async function disconnectDrive() {
    driveBusy = true;
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
      driveBusy = false;
    }
  }

  async function exportToDrive() {
    const names = exportPendingNames;
    if (names === null) return;
    closeExportPick();
    await refreshDriveStatus();
    if (!driveStatus.hasCredentials) {
      const msg =
        lang === 'ru'
          ? 'Сначала настройте Google Drive в Настройках'
          : 'Please configure Google Drive in Settings first';
      setStatus(msg, 'err', true);
      await notify(false, msg);
      await openSettings();
      return;
    }
    exportBusy = true;
    driveBusy = true;
    driveAuthWait = !driveStatus.connected;
    setStatus(driveAuthWait ? t(lang, 'driveWaiting') : t(lang, 'driveUploading'), '', true);
    try {
      const res = (await (AppService as any).ExportTokensToGoogleDrive(names)) as Result;
      const cancelled =
        res.message === 'Cancelled' ||
        res.message === 'cancelled' ||
        (res.message || '').toLowerCase().includes('cancelled');
      if (cancelled) {
        setStatus(t(lang, 'cancelled'), 'ok', true);
      } else {
        setStatus(res.message, res.ok ? 'ok' : 'err');
        await notify(res.ok, res.message);
      }
      await refreshDriveStatus();
    } catch (e) {
      const msg = String(e);
      if (msg.toLowerCase().includes('cancel')) {
        setStatus(t(lang, 'cancelled'), 'ok', true);
      } else {
        setStatus(msg, 'err');
        await notify(false, msg);
      }
    } finally {
      driveAuthWait = false;
      driveBusy = false;
      exportBusy = false;
    }
  }

  function setStatus(msg: string, kind: 'ok' | 'err' | '' = '', alreadyTranslated = false) {
    const text = alreadyTranslated ? msg : translateBackendMessage(lang, msg);
    status = text;
    statusKind = kind;
    addLog(kind === 'err' ? 'err' : kind === 'ok' ? 'ok' : 'info', text);
  }

  let checkKey = $state('');
  let infoLogin = $state<{ kind: 'key'; key: string } | { kind: 'saved'; name: string } | null>(null);

  type BulkItem = {
    account: string;
    steamId?: string;
    status: string;
    expiresAt?: string;
    detail?: string;
  };
  let showBulk = $state(false);
  let bulkKeys = $state('');
  let bulkBusy = $state(false);
  let bulkResult = $state<{ total: number; items: BulkItem[] } | null>(null);
  let bulkLive = $state<BulkItem[]>([]);
  let bulkTotal = $state(0);
  let exportToast = $state('');
  let exportTimer: ReturnType<typeof setTimeout> | undefined;

  const bulkItems = $derived(bulkBusy || !bulkResult ? bulkLive : bulkResult.items);
  const bulkOkCount = $derived(bulkItems.filter((i) => i.status === 'ok').length);
  const bulkBadCount = $derived(bulkItems.length - bulkOkCount);

  const PROXY_KEY = 'nfa-tool-proxies';
  function loadProxies(): string {
    try {
      return localStorage.getItem(PROXY_KEY) ?? '';
    } catch {
      return '';
    }
  }
  let settingsProxies = $state(loadProxies());
  function saveProxies() {
    try {
      localStorage.setItem(PROXY_KEY, settingsProxies);
    } catch {
    }
  }

  async function bulkPickFile() {
    try {
      const text = (await (AppService as any).PickTextFile()) as string;
      if (text?.trim()) {
        bulkKeys = bulkKeys.trim() ? bulkKeys.trimEnd() + '\n' + text.trim() : text.trim();
      }
    } catch (e) {
      setStatus(String(e), 'err');
    }
  }

  function bulkInfo(it: BulkItem) {
    const name = it.account.split(':')[0].trim();
    if (!name) return;
    const saved = accounts.find((a) => a.name.toLowerCase() === name.toLowerCase());
    if (saved) {
      openAccInfo(saved.name);
      return;
    }
    const line = bulkKeys
      .split(/\r?\n/)
      .map((l) => l.trim())
      .find((l) => l.toLowerCase().startsWith(name.toLowerCase() + '----'));
    if (line) {
      infoLogin = { kind: 'key', key: line };
      void showInfoFor((AppService as any).CheckAccountKey(line));
    }
  }

  function bulkStatusText(status: string): string {
    const m: Record<string, keyof import('./i18n').Dict> = {
      ok: 'bstOk',
      rejected: 'bstRejected',
      expired: 'bstExpired',
      invalid: 'bstInvalid',
      error: 'bstError',
    };
    const key = m[status];
    return key ? t(lang, key) : status;
  }

  async function runBulkCheck(saved: boolean) {
    if (!saved && !bulkKeys.trim()) return;
    bulkBusy = true;
    bulkResult = null;
    bulkLive = [];
    bulkTotal = saved
      ? accounts.length
      : bulkKeys.split(/\r?\n/).filter((l) => l.trim()).length;
    try {
      bulkResult = saved
        ? await (AppService as any).CheckSavedAccounts(settingsProxies)
        : await (AppService as any).CheckAccountKeys(bulkKeys, settingsProxies);
    } catch (e) {
      const msg = String(e);
      setStatus(msg, 'err');
      await notify(false, msg);
    } finally {
      bulkBusy = false;
    }
  }

  async function exportBulk(which: 'ok' | 'bad') {
    try {
      const path = (await (AppService as any).ExportBulkResults(which)) as string;
      if (!path) return;
      clearTimeout(exportTimer);
      exportToast = path;
      exportTimer = setTimeout(() => (exportToast = ''), 2600);
    } catch (e) {
      const msg = String(e);
      setStatus(msg, 'err');
      await notify(false, msg);
    }
  }

  async function showInfoFor(promise: Promise<any>) {
    showAccInfo = true;
    infoData = null;
    infoBusy = true;
    try {
      infoData = await promise;
    } catch (e) {
      showAccInfo = false;
      infoLogin = null;
      const msg = String(e);
      setStatus(msg, 'err');
      await notify(false, msg);
    } finally {
      infoBusy = false;
    }
  }

  function openAccInfo(name: string) {
    infoLogin = { kind: 'saved', name };
    void showInfoFor((AppService as any).GetAccountInfo(name));
  }

  function runKeyCheck() {
    const key = checkKey.trim();
    if (!key) return;
    infoLogin = { kind: 'key', key };
    void showInfoFor((AppService as any).CheckAccountKey(key));
    setStatus(t(lang, 'keyCheckDone'), 'ok', true);
  }

  async function loginFromInfo() {
    if (!infoLogin) return;
    const src = infoLogin;
    loading = true;
    openBusy(
      src.kind === 'saved' ? t(lang, 'loggingInAs', { name: src.name }) : t(lang, 'loggingIn'),
      '',
      true,
    );
    try {
      const res = (
        src.kind === 'key'
          ? await AppService.LoginFromKey(src.key, keepExisting)
          : await AppService.LoginSaved(src.name, keepExisting)
      ) as Result;
      setStatus(res.message, res.ok ? 'ok' : 'err');
      await notify(res.ok, res.message);
      if (res.ok) {
        showAccInfo = false;
        infoLogin = null;
        if (src.kind === 'key') accountKey = '';
        showSuccess();
        await refreshAccounts();
      }
    } catch (e) {
      const msg = String(e);
      setStatus(msg, 'err');
      await notify(false, msg);
    } finally {
      closeBusy();
      loading = false;
    }
  }

  function trVisibility(v: string): string {
    const m: Record<string, string> = {
      public: t(lang, 'visPublic'),
      private: t(lang, 'visPrivate'),
      friendsonly: t(lang, 'visFriends'),
    };
    return m[(v || '').toLowerCase()] || v || '—';
  }

  function trOnline(v: string): string {
    const m: Record<string, string> = {
      online: t(lang, 'stOnline'),
      offline: t(lang, 'stOffline'),
      'in-game': t(lang, 'stInGame'),
    };
    return m[(v || '').toLowerCase()] || v || '—';
  }

  function yn(b: boolean): string {
    return b ? t(lang, 'yes') : t(lang, 'no');
  }

  function checkDesc(id: string): string {
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
    return key ? t(lang, key) : '';
  }

  function checkLabel(id: string): string {
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
    return key ? t(lang, key) : id;
  }

  function singleLineKey(): string {
    return accountKey.split(/\r?\n/)[0]?.trim() ?? '';
  }

  async function saveFromInfo() {
    if (!infoLogin || infoLogin.kind !== 'key') return;
    loading = true;
    try {
      const res = (await (AppService as any).SaveAccountKey(infoLogin.key)) as Result;
      setStatus(res.message, res.ok ? 'ok' : 'err');
      await notify(res.ok, res.message);
      if (res.ok) {
        accountKey = '';
        await refreshAccounts();
      }
    } catch (e) {
      const msg = String(e);
      setStatus(msg, 'err');
      await notify(false, msg);
    } finally {
      loading = false;
    }
  }

  async function doLogin() {
    const key = singleLineKey();
    if (!key) {
      const msg = t(lang, 'enterKey');
      setStatus(msg, 'err', true);
      await notify(false, msg);
      return;
    }
    loading = true;
    setStatus(t(lang, 'loggingIn'), '', true);
    openBusy(t(lang, 'loggingIn'), '', true);
    try {
      const res = (await AppService.LoginFromKey(key, keepExisting)) as Result;
      setStatus(res.message, res.ok ? 'ok' : 'err');
      await notify(res.ok, res.message);
      if (res.ok) {
        accountKey = '';
        showSuccess();
        await refreshAccounts();
      }
    } catch (e) {
      const msg = String(e);
      setStatus(msg, 'err');
      await notify(false, msg);
    } finally {
      closeBusy();
      loading = false;
    }
  }

  async function importFromField() {
    const key = singleLineKey();
    if (!key) {
      const msg = t(lang, 'enterKey');
      setStatus(msg, 'err', true);
      await notify(false, msg);
      return;
    }
    loading = true;
    try {
      const res = (await (AppService as any).ImportTokens(key)) as Result;
      setStatus(res.message, res.ok ? 'ok' : 'err');
      await notify(res.ok, res.message);
      if (res.ok) {
        accountKey = '';
        await refreshAccounts();
      }
    } catch (e) {
      const msg = String(e);
      setStatus(msg, 'err');
      await notify(false, msg);
    } finally {
      loading = false;
    }
  }

  async function importFromFile() {
    loading = true;
    try {
      const res = (await (AppService as any).ImportTokensFromFile()) as Result;
      if (res.message === 'Cancelled' || res.message === 'cancelled') {
        setStatus(t(lang, 'cancelled'), 'ok', true);
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
      loading = false;
    }
  }

  async function loginSaved(name: string) {
    setStatus(t(lang, 'loggingInAs', { name }), '', true);
    openBusy(t(lang, 'loggingInAs', { name }), '', true);
    try {
      const res = (await AppService.LoginSaved(name, keepExisting)) as Result;
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

  function clearSelectionNames(names: string[]) {
    const drop = new Set(names.map((n) => n.toLowerCase()));
    const next: Record<string, boolean> = {};
    for (const [k, v] of Object.entries(selected)) {
      if (!drop.has(k.toLowerCase())) next[k] = v;
    }
    selected = next;
  }

  async function deleteAccount(name: string, e?: Event) {
    e?.preventDefault();
    e?.stopPropagation();
    endDragSelect();
    const acc = name.trim();
    if (!acc) return;
    try {
      const res = (await AppService.DeleteAccount(acc)) as Result;
      setStatus(res.message, res.ok ? 'ok' : 'err');
      await notify(res.ok, res.message);
      if (res.ok) {
        clearSelectionNames([acc]);
        accounts = accounts.filter((a) => a.name.toLowerCase() !== acc.toLowerCase());
      }
      await refreshAccounts();
    } catch (err) {
      const msg = String(err);
      setStatus(msg, 'err');
      await notify(false, msg);
    }
  }

  async function deleteSelected() {
    endDragSelect();
    const names = selectedNames.slice();
    if (names.length === 0) {
      const msg = t(lang, 'exportNone');
      setStatus(msg, 'err', true);
      return;
    }
    try {
      const res = (await (AppService as any).DeleteAccounts(names)) as Result;
      setStatus(res.message, res.ok ? 'ok' : 'err');
      await notify(res.ok, res.message);
      if (res.ok) {
        const drop = new Set(names.map((n) => n.toLowerCase()));
        accounts = accounts.filter((a) => !drop.has(a.name.toLowerCase()));
        selected = {};
      }
      await refreshAccounts();
    } catch (err) {
      const msg = String(err);
      setStatus(msg, 'err');
      await notify(false, msg);
    }
  }

  async function harvestSteam() {
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

  async function resetSteam() {
    setStatus(t(lang, 'resettingSteam'), '', true);
    openBusy(t(lang, 'resettingSteam'));
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

  function onKey(e: KeyboardEvent) {
    if (e.key !== 'Enter') return;
    doLogin();
  }
</script>

{#snippet helpButton()}
  <div class="help-wrap" use:aimHint>
    {#if showHelpHint}
      <button class="help-hint" type="button" onclick={() => (showHintConfirm = true)} aria-label={t(lang, 'showInstructions')}>
        <svg class="hh-arrow" viewBox="0 0 56 40" aria-hidden="true">
          <path d="M10 35 C 22 34, 34 26, 40 14" />
          <path d="M32 14 L 40 5 L 48 14" />
        </svg>
        <span class="hh-text">{t(lang, 'helpHint')}</span>
      </button>
    {/if}
    <button
      class="help-fab"
      type="button"
      title={t(lang, 'showInstructions')}
      aria-label={t(lang, 'showInstructions')}
      onclick={openHelp}
    >
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <path d="M10.5 8.67709C10.8665 8.26188 11.4027 8 12 8C13.1046 8 14 8.89543 14 10C14 10.9337 13.3601 11.718 12.4949 11.9383C12.2273 12.0064 12 12.2239 12 12.5V12.5V13" />
        <path d="M12 16H12.01" />
      </svg>
    </button>
  </div>
{/snippet}

<div class="app">
  <header class="titlebar">
    <div class="brand">
      <span class="logo-mark" aria-hidden="true">N</span>
      <div class="brand-text">
        <span class="brand-name">{appName}</span>
        {#if version}
          <span class="brand-ver">v{version}</span>
        {/if}
      </div>
    </div>
    <div class="title-actions">
      <div class="lang-switch" use:slidingPill>
        <button type="button" class="lang-btn" class:active={mode === 'simple'} onclick={() => setMode('simple')}>
          <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path fill-rule="evenodd" clip-rule="evenodd" d="M12.3327 3.63004C12.9116 2.74599 14.2858 3.15586 14.2858 4.21256V9.39999H17.7635C18.6601 9.39999 19.1983 10.3956 18.7071 11.1457L12.7695 20.214C12.1817 21.1118 10.7144 20.7524 10.7144 19.6027V14.6H7.18878C6.31275 14.6 5.78696 13.6272 6.26683 12.8943L12.3327 3.63004Z" /></svg>
          {t(lang, 'modeSimple')}
        </button>
        <button type="button" class="lang-btn" class:active={mode === 'advanced'} onclick={() => setMode('advanced')}>
          <svg viewBox="0 0 24 24" fill="currentColor" style="transform: translateY(-1.5px)" aria-hidden="true"><path fill-rule="evenodd" clip-rule="evenodd" d="M12 6C8.68629 6 6 8.68629 6 12C6 13.6332 6.65387 15.1157 7.71186 16.1966C7.97971 16.4703 8.1241 16.7217 8.16867 16.9444L8.69776 19.5886C8.97833 20.9908 10.2095 22 11.6395 22H12.3605C13.7905 22 15.0217 20.9908 15.3022 19.5886L15.8313 16.9444C15.8759 16.7217 16.0203 16.4703 16.2881 16.1966C17.3461 15.1157 18 13.6332 18 12C18 8.68629 15.3137 6 12 6ZM11 16C10.4477 16 10 16.4477 10 17C10 17.5523 10.4477 18 11 18H13C13.5523 18 14 17.5523 14 17C14 16.4477 13.5523 16 13 16H11Z" /></svg>
          {t(lang, 'modeAdvanced')}
        </button>
        <button type="button" class="lang-btn" class:active={mode === 'logs'} onclick={() => setMode('logs')}>
          <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path fill-rule="evenodd" clip-rule="evenodd" d="M7.29291 14.2929C6.90238 14.6834 6.90238 15.3166 7.29291 15.7071C7.68343 16.0976 8.31659 16.0976 8.70712 15.7071L11.2071 13.2071C11.8738 12.5404 11.8738 11.4596 11.2071 10.7929L8.70712 8.29289C8.3166 7.90237 7.68343 7.90237 7.29291 8.29289C6.90238 8.68342 6.90238 9.31658 7.29291 9.70711L9.5858 12L7.29291 14.2929ZM13 14C12.4477 14 12 14.4477 12 15C12 15.5523 12.4477 16 13 16H16C16.5523 16 17 15.5523 17 15C17 14.4477 16.5523 14 16 14H13ZM22 7.93418C22 7.95604 22 7.97799 22 8.00001L22 16.0658C22.0001 16.9523 22.0001 17.7161 21.9179 18.3278C21.8297 18.9833 21.631 19.6117 21.1213 20.1213C20.6117 20.631 19.9833 20.8297 19.3278 20.9179C18.7161 21.0001 17.9523 21.0001 17.0658 21L6.9342 21C6.0477 21.0001 5.28388 21.0001 4.67222 20.9179C4.0167 20.8297 3.38835 20.631 2.87869 20.1213C2.36902 19.6117 2.17028 18.9833 2.08215 18.3278C1.99991 17.7161 1.99995 16.9523 2 16.0658L2 7.9342C1.99995 7.0477 1.99991 6.28388 2.08215 5.67221C2.17028 5.0167 2.36902 4.38835 2.87869 3.87868C3.38835 3.36902 4.0167 3.17028 4.67222 3.08215C5.28388 2.99991 6.04769 2.99995 6.93418 3L17 3.00001C17.022 3.00001 17.044 3 17.0658 3C17.9523 2.99995 18.7161 2.99991 19.3278 3.08215C19.9833 3.17028 20.6117 3.36902 21.1213 3.87869C21.631 4.38835 21.8297 5.0167 21.9179 5.67221C22.0001 6.28387 22.0001 7.04769 22 7.93418Z" /></svg>
          {t(lang, 'modeLogs')}<span class="beta-chip" title={t(lang, 'alphaWarn')}>ALPHA</span>
        </button>
      </div>
      <button
        class="icon-btn"
        type="button"
        title={t(lang, 'settings')}
        aria-label={t(lang, 'settings')}
        onclick={openSettings}
      >
        <svg viewBox="0 0 72 72" fill="currentColor" aria-hidden="true">
          <path d="M57.531,30.556C58.96,30.813,60,32.057,60,33.509v4.983c0,1.452-1.04,2.696-2.469,2.953l-2.974,0.535c-0.325,1.009-0.737,1.977-1.214,2.907l1.73,2.49c0.829,1.192,0.685,2.807-0.342,3.834l-3.523,3.523c-1.027,1.027-2.642,1.171-3.834,0.342l-2.49-1.731c-0.93,0.477-1.898,0.889-2.906,1.214l-0.535,2.974C41.187,58.96,39.943,60,38.491,60h-4.983c-1.452,0-2.696-1.04-2.953-2.469l-0.535-2.974c-1.009-0.325-1.977-0.736-2.906-1.214l-2.49,1.731c-1.192,0.829-2.807,0.685-3.834-0.342l-3.523-3.523c-1.027-1.027-1.171-2.641-0.342-3.834l1.73-2.49c-0.477-0.93-0.889-1.898-1.214-2.907l-2.974-0.535C13.04,41.187,12,39.943,12,38.491v-4.983c0-1.452,1.04-2.696,2.469-2.953l2.974-0.535c0.325-1.009,0.737-1.977,1.214-2.907l-1.73-2.49c-0.829-1.192-0.685-2.807,0.342-3.834l3.523-3.523c1.027-1.027,2.642-1.171,3.834-0.342l2.49,1.731c0.93-0.477,1.898-0.889,2.906-1.214l0.535-2.974C30.813,13.04,32.057,12,33.509,12h4.983c1.452,0,2.696,1.04,2.953,2.469l0.535,2.974c1.009,0.325,1.977,0.736,2.906,1.214l2.49-1.731c1.192-0.829,2.807-0.685,3.834,0.342l3.523,3.523c1.027,1.027,1.171,2.641,0.342,3.834l-1.73,2.49c0.477,0.93,0.889,1.898,1.214,2.907L57.531,30.556z M36,45c4.97,0,9-4.029,9-9c0-4.971-4.03-9-9-9s-9,4.029-9,9C27,40.971,31.03,45,36,45z" />
        </svg>
      </button>
      <button
        class="icon-btn"
        type="button"
        title={t(lang, 'checkUpdate')}
        aria-label={t(lang, 'checkUpdate')}
        onclick={() => checkUpdates(false)}
      >
        <svg viewBox="0 0 21 21" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <g transform="translate(2 2)">
            <path d="m4.5 1.5c-2.4138473 1.37729434-4 4.02194088-4 7 0 4.418278 3.581722 8 8 8s8-3.581722 8-8-3.581722-8-8-8" />
            <path d="m4.5 5.5v-4h-4" />
          </g>
        </svg>
      </button>
      <div class="win-btns">
        <button class="win" type="button" onclick={() => AppService.WindowMinimise()} aria-label="Minimise">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M9 12H15" />
          </svg>
        </button>
        <button
          class="win close"
          type="button"
          onclick={async () => {
            try {
              await (AppService as any).CancelGoogleAuth?.();
            } catch {
            }
            await AppService.WindowClose();
          }}
          aria-label="Close"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M9 9L15 15" />
            <path d="M15 9L9 15" />
          </svg>
        </button>
      </div>
    </div>
  </header>

  <div class="swap-wrap">
  {#key mode}
  {#if mode === 'logs'}
    <main
      class="content"
      class:mode-swap={swapped}
      in:fly={{ x: swapped ? slideFrom : 0, duration: 380, opacity: 1, easing: cubicOut }}
      out:fly={{ x: swapped ? -slideFrom : 0, duration: 380, opacity: 1, easing: cubicOut }}
    >
      <section class="panel sys-panel">
        <div class="panel-label">{t(lang, 'sysPanel')}</div>
        <p class="alpha-warn">{t(lang, 'alphaWarn')}</p>
        {#if sysStatus}
          <div class="accinfo-grid">
            <span class="ai-label">{t(lang, 'sysVersion')}</span>
            <span class="ai-value">v{sysStatus.version}</span>
            <span class="ai-label">Steam</span>
            <span class="ai-value" class:good={sysStatus.steamRunning} class:bad={!sysStatus.steamRunning}>
              {sysStatus.steamRunning ? t(lang, 'sysRunning') : t(lang, 'sysStopped')}
            </span>
            <span class="ai-label">{t(lang, 'sysPath')}</span>
            <span class="ai-value path" title={sysStatus.steamPath}>{sysStatus.steamPath || '—'}</span>
            <span class="ai-label">{t(lang, 'sysAccounts')}</span>
            <span class="ai-value">{sysStatus.accountsValid} / {sysStatus.accountsTotal}</span>
            <span class="ai-label">Google Drive</span>
            <span class="ai-value" class:good={sysStatus.driveConnected} class:bad={!sysStatus.driveConnected}>
              {sysStatus.driveConnected ? t(lang, 'driveConnected') : t(lang, 'driveNotConnected')}
            </span>
            <span class="ai-label">Steam API</span>
            <span class="ai-value" class:good={sysStatus.steamApiOnline} class:bad={!sysStatus.steamApiOnline}>
              {sysStatus.steamApiOnline ? t(lang, 'apiOk') : t(lang, 'apiDown')}
            </span>
          </div>
        {:else}
          <p class="update-msg wait">{t(lang, 'accInfoLoading')}</p>
        {/if}
        <button class="btn ghost block" type="button" disabled={sysBusy} onclick={refreshSysStatus}>
          {t(lang, 'logRefresh')}
        </button>
        <button class="btn ghost block" type="button" onclick={() => checkUpdates(false)}>
          {t(lang, 'checkUpdate')}
        </button>
        <div class="panel-label keycheck-label">{t(lang, 'keyCheckLabel')}</div>
        <label class="field">
          <input
            type="text"
            spellcheck="false"
            autocomplete="off"
            placeholder="login----token"
            bind:value={checkKey}
            onkeydown={(e) => e.key === 'Enter' && runKeyCheck()}
          />
        </label>
        <button class="btn primary" type="button" disabled={infoBusy || !checkKey.trim()} onclick={runKeyCheck}>
          {infoBusy ? t(lang, 'working') : t(lang, 'keyCheckBtn')}
        </button>
        <button class="btn ghost block" type="button" onclick={() => (showBulk = true)}>
          {t(lang, 'bulkCheckBtn')}
        </button>
      </section>

      <section class="panel logs-panel">
        <div class="accounts-head">
          <div>
            <div class="panel-label">
              {t(lang, 'logPanel')}
              <span class="beta-chip" title={t(lang, 'alphaWarn')}>ALPHA</span>
            </div>
            {#if logs.length > 0}
              <div class="count-chip">{logs.length}</div>
            {/if}
          </div>
          {#if logs.length > 0}
            <button class="btn ghost sm" type="button" onclick={clearLogs}>{t(lang, 'logClear')}</button>
          {/if}
        </div>
        {#if logs.length === 0}
          <div class="empty">
            <div class="empty-icon" aria-hidden="true">
              <svg viewBox="0 0 24 24" fill="currentColor">
                <path d="M6.99486 7.00636C6.60433 7.39689 6.60433 8.03005 6.99486 8.42058L10.58 12.0057L6.99486 15.5909C6.60433 15.9814 6.60433 16.6146 6.99486 17.0051C7.38538 17.3956 8.01855 17.3956 8.40907 17.0051L11.9942 13.4199L15.5794 17.0051C15.9699 17.3956 16.6031 17.3956 16.9936 17.0051C17.3841 16.6146 17.3841 15.9814 16.9936 15.5909L13.4084 12.0057L16.9936 8.42059C17.3841 8.03007 17.3841 7.3969 16.9936 7.00638C16.603 6.61585 15.9699 6.61585 15.5794 7.00638L11.9942 10.5915L8.40907 7.00636C8.01855 6.61584 7.38538 6.61584 6.99486 7.00636Z" />
              </svg>
            </div>
            <p>{t(lang, 'logEmpty')}</p>
          </div>
        {:else}
          <div class="logs-list">
            {#each logs as entry, i (entry.ts + '-' + i)}
              <div class="log-row {entry.kind}">
                <span class="log-ts">{fmtTime(entry.ts)}</span>
                <span class="log-text">{entry.text}</span>
              </div>
            {/each}
          </div>
        {/if}
      </section>
    </main>
  {:else}
  <main
    class="content"
    class:mode-swap={swapped}
    in:fly={{ x: swapped ? slideFrom : 0, duration: 380, opacity: 1, easing: cubicOut }}
    out:fly={{ x: swapped ? -slideFrom : 0, duration: 380, opacity: 1, easing: cubicOut }}
  >
    <section class="panel login-panel">
      <div class="panel-label">{t(lang, 'accountManagement')}</div>
      <label class="field">
        <span class="field-label">{t(lang, 'accountKeyPlaceholder')}</span>
        <input
          type="text"
          spellcheck="false"
          autocomplete="off"
          placeholder="login----token"
          bind:value={accountKey}
          onkeydown={onKey}
        />
      </label>
      {#if mode === 'advanced'}
        <label class="check">
          <input type="checkbox" bind:checked={keepExisting} />
          <span class="box"></span>
          <span>{t(lang, 'keepExisting')}</span>
        </label>
        <div class="action-row">
          <button class="btn primary" type="button" disabled={loading} onclick={doLogin}>
            {loading ? t(lang, 'working') : t(lang, 'login')}
          </button>
          <button class="btn ghost block" type="button" disabled={loading} onclick={importFromField}>
            {t(lang, 'importBtn')}
          </button>
        </div>
        <div class="import-file-row">
          <button class="btn ghost block" type="button" disabled={loading} onclick={importFromFile}>
            {t(lang, 'importFile')}
          </button>
          {@render helpButton()}
        </div>
      {:else}
        <label class="check">
          <input type="checkbox" bind:checked={keepExisting} />
          <span class="box"></span>
          <span>{t(lang, 'keepExisting')}</span>
        </label>
        <div class="import-file-row">
          <button class="btn primary" type="button" disabled={loading} onclick={doLogin}>
            {loading ? t(lang, 'working') : t(lang, 'login')}
          </button>
          {@render helpButton()}
        </div>
      {/if}
      {#key accounts.length === 0}
        <p class="hint">
          {accounts.length === 0 ? t(lang, 'hintEmpty') : t(lang, 'hintHasAccounts')}
        </p>
      {/key}
    </section>

    <section class="panel accounts-panel">
      <div class="accounts-head">
        <div>
          <div class="panel-label">{t(lang, 'savedAccounts')}</div>
          {#if accounts.length > 0}
            {#key accounts.length}
              <div class="count-chip">{accounts.length}</div>
            {/key}
          {/if}
        </div>
        <div class="head-actions">
          {#if mode === 'advanced'}
            <button class="btn ghost sm" type="button" onclick={() => (showBulk = true)}>
              {t(lang, 'bulkCheckBtn')}
            </button>
          {/if}
          {#if accounts.length > 0 && mode === 'advanced'}
          <div class="export-bar">
            <button class="btn ghost sm" type="button" disabled={exportBusy} onclick={toggleSelectAll}>
              {allSelected ? t(lang, 'deselectAll') : t(lang, 'selectAll')}
            </button>
            <button
              class="btn ghost sm"
              type="button"
              disabled={exportBusy || selectedNames.length === 0}
              onclick={() => exportAccounts(false)}
            >
              {t(lang, 'exportSelected')}
              {#if selectedNames.length > 0}
                {#key selectedNames.length}
                  <span class="num-pop">({selectedNames.length})</span>
                {/key}
              {/if}
            </button>
            <button class="btn ghost sm" type="button" disabled={exportBusy} onclick={() => exportAccounts(true)}>
              {t(lang, 'exportAll')}
            </button>
            <button
              class="btn ghost sm danger-sm"
              type="button"
              disabled={selectedNames.length === 0}
              onclick={deleteSelected}
            >
              {t(lang, 'deleteSelected')}
              {#if selectedNames.length > 0}
                {#key selectedNames.length}
                  <span class="num-pop">({selectedNames.length})</span>
                {/key}
              {/if}
            </button>
          </div>
          {/if}
        </div>
      </div>
      {#if accounts.length === 0}
        <div class="empty">
          <div class="empty-icon" aria-hidden="true">∅</div>
          <p>{t(lang, 'noSavedAccounts')}</p>
        </div>
      {:else}
        <div class="accounts-list" class:painting={dragSelect}>
          {#each accounts as acc, i (acc.name)}
            <div
              class="account-row"
              class:simple={mode === 'simple'}
              class:picked={mode === 'advanced' && !!selected[acc.name]}
              data-acc={acc.name}
              style="animation-delay: {Math.min(i * 30, 300)}ms"
              onpointerenter={() => paintSelect(acc.name)}
              onpointerdown={(e) => {
                const t = e.target as HTMLElement;
                if (t.closest('.row-actions')) return;
                startDragSelect(e, acc.name);
              }}
            >
              {#if mode === 'advanced'}
                <label class="pick">
                  <input type="checkbox" checked={!!selected[acc.name]} tabindex="-1" />
                  <span class="box"></span>
                </label>
              {/if}
              {#if acc.avatar}
                <img class="acc-ava" src={acc.avatar} alt="" />
              {:else}
                <div class="acc-ava placeholder" aria-hidden="true">{acc.name.slice(0, 1).toUpperCase()}</div>
              {/if}
              <div class="meta">
                <div class="name">{acc.name}</div>
                <div class="exp" class:ok={acc.valid} class:bad={!acc.valid}>
                  {localizeExpiry(lang, acc.expiresIn)}
                </div>
              </div>
              <div
                class="row-actions"
                onpointerdown={(e) => {
                  e.stopPropagation();
                  endDragSelect();
                }}
              >
                {#if mode === 'advanced'}
                  <button
                    class="btn mini export"
                    type="button"
                    disabled={exportBusy}
                    onclick={(e) => {
                      e.stopPropagation();
                      exportOne(acc.name);
                    }}
                  >{t(lang, 'export')}</button>
                {/if}
                <button
                  class="btn mini info"
                  type="button"
                  onclick={(e) => {
                    e.stopPropagation();
                    openAccInfo(acc.name);
                  }}
                >{t(lang, 'info')}</button>
                <button
                  class="btn mini login"
                  type="button"
                  onclick={(e) => {
                    e.stopPropagation();
                    loginSaved(acc.name);
                  }}
                >{t(lang, 'login')}</button>
                <button
                  class="btn mini del"
                  type="button"
                  onclick={(e) => deleteAccount(acc.name, e)}
                >{t(lang, 'delete')}</button>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </section>
  </main>
  {/if}
  {/key}
  </div>

  <footer class="statusbar" class:ok={statusKind === 'ok'} class:err={statusKind === 'err'}>
    <span class="status-dot" aria-hidden="true"></span>
    {#key status}
      <span class="status-text">{status}</span>
    {/key}
    <span class="credit">
      {t(lang, 'creditPrefix')}
      <span class="credit-brand">HACKERSHOP</span>
    </span>
  </footer>
</div>

{#if showExportPick}
  <div class="modal" role="dialog" aria-modal="true" tabindex="-1">
    <div class="modal-card">
      <div class="modal-head">
        <h3>{t(lang, 'exportWhere')}</h3>
        <button class="modal-x" type="button" onclick={closeExportPick} aria-label={t(lang, 'close')}>✕</button>
      </div>
      <div class="modal-actions">
        <button class="btn primary" type="button" disabled={exportBusy || driveBusy} onclick={exportToClipboard}>
          {t(lang, 'exportToClipboard')}
        </button>
        <button class="btn ghost block" type="button" disabled={exportBusy || driveBusy} onclick={exportToFile}>
          {t(lang, 'exportToFile')}
        </button>
        <button class="btn ghost block" type="button" disabled={exportBusy || driveBusy} onclick={exportToDrive}>
          {t(lang, 'exportDrive')}
        </button>
        <button class="btn ghost block" type="button" disabled={exportBusy || driveBusy} onclick={closeExportPick}>
          {t(lang, 'cancel')}
        </button>
      </div>
    </div>
  </div>
{/if}

{#if showHintConfirm}
  <div class="modal" role="dialog" aria-modal="true" tabindex="-1">
    <div class="modal-card">
      <h3>{t(lang, 'hintConfirmTitle')}</h3>
      <p class="update-msg">{t(lang, 'hintConfirmText')}</p>
      <div class="modal-actions">
        <button class="btn primary" type="button" onclick={confirmHideHint}>{t(lang, 'hintConfirmYes')}</button>
        <button class="btn ghost block" type="button" onclick={() => (showHintConfirm = false)}>{t(lang, 'hintConfirmNo')}</button>
      </div>
    </div>
  </div>
{/if}

{#if showAccInfo}
  <div class="modal" role="dialog" aria-modal="true" tabindex="-1">
    <div class="modal-card update">
      <div class="modal-head">
        <h3>{t(lang, 'accInfoTitle')}</h3>
        <button class="modal-x" type="button" onclick={() => (showAccInfo = false)} aria-label={t(lang, 'close')}>✕</button>
      </div>
      {#if infoBusy}
        <p class="update-msg wait">{t(lang, 'accInfoLoading')}</p>
      {:else if infoData}
        <div class="accinfo-head">
          {#if infoData.avatarFull}
            <img class="accinfo-ava" src={infoData.avatarFull} alt="" />
          {/if}
          <div class="accinfo-id-block">
            <div class="accinfo-nick">{infoData.personaName || '—'}</div>
            <div class="accinfo-id">{infoData.steamId}</div>
          </div>
        </div>
        <div class="accinfo-grid">
          <span class="ai-label">{t(lang, 'rowStatus')}</span>
          <span class="ai-value">{infoData.profileErr ? '—' : trOnline(infoData.onlineState)}</span>
          <span class="ai-label">{t(lang, 'rowVisibility')}</span>
          <span class="ai-value">{infoData.profileErr ? '—' : trVisibility(infoData.visibility)}</span>
          {#if infoData.realName}
            <span class="ai-label">{t(lang, 'rowRealName')}</span>
            <span class="ai-value">{infoData.realName}</span>
          {/if}
          {#if infoData.location}
            <span class="ai-label">{t(lang, 'rowLocation')}</span>
            <span class="ai-value">{infoData.location}</span>
          {/if}
          {#if infoData.inGame}
            <span class="ai-label">{t(lang, 'rowInGame')}</span>
            <span class="ai-value good">{infoData.inGame}</span>
          {/if}
          {#if infoData.level}
            <span class="ai-label">{t(lang, 'rowLevel')}</span>
            <span class="ai-value">{infoData.level}</span>
          {/if}
          {#if infoData.gamesCount}
            <span class="ai-label">{t(lang, 'rowGames')}</span>
            <span class="ai-value">{infoData.gamesCount}</span>
          {/if}
          {#if infoData.friendsCount > 0}
            <span class="ai-label">{t(lang, 'rowFriends')}</span>
            <span class="ai-value">{infoData.friendsCount}</span>
          {/if}
          {#if infoData.topGame}
            <span class="ai-label">{t(lang, 'rowTopGame')}</span>
            <span class="ai-value">{infoData.topGame}{infoData.topGameHours ? ` · ${infoData.topGameHours}h` : ''}</span>
          {/if}
          <span class="ai-label">{t(lang, 'rowVac')}</span>
          <span class="ai-value" class:bad={infoData.vacBanned}>{infoData.profileErr ? '—' : yn(infoData.vacBanned)}</span>
          {#if infoData.banInfo}
            <span class="ai-label">{t(lang, 'rowBanInfo')}</span>
            <span class="ai-value bad">{infoData.banInfo}{infoData.banDays ? ` · ${infoData.banDays} ${t(lang, 'daysShort')}` : ''}</span>
          {/if}
          {#if infoData.cs2Items > 0}
            <span class="ai-label">{t(lang, 'rowCS2Items')}</span>
            <span class="ai-value">{infoData.cs2Items}{infoData.cs2Rarity ? ` · ${infoData.cs2Rarity}` : ''}</span>
          {/if}
          {#if infoData.cs2Hours}
            <span class="ai-label">{t(lang, 'rowCS2Hours')}</span>
            <span class="ai-value">{infoData.cs2Hours}h</span>
          {/if}
          {#if infoData.cs2Medals}
            <span class="ai-label">{t(lang, 'rowCS2Medals')}</span>
            <span class="ai-value" title={infoData.cs2Medals}>{infoData.cs2Medals}</span>
          {/if}
          {#if infoData.inventoryValue}
            <span class="ai-label">{t(lang, 'rowInvValue')}</span>
            <span class="ai-value good">
              {infoData.inventoryValue}{infoData.inventoryPartial ? ` ${t(lang, 'invPartial')}` : ''}
            </span>
          {/if}
          {#if infoData.cs2Prime}
            <span class="ai-label">{t(lang, 'rowPrime')}</span>
            <span class="ai-value" class:good={infoData.cs2Prime === 'yes'}>{yn(infoData.cs2Prime === 'yes')}</span>
          {/if}
          {#if infoData.licensesCount > 0}
            <span class="ai-label">{t(lang, 'rowLicenses')}</span>
            <span class="ai-value">{infoData.licensesCount}</span>
          {/if}
          <span class="ai-label">{t(lang, 'rowTrade')}</span>
          <span class="ai-value" class:bad={!!infoData.tradeBan && infoData.tradeBan.toLowerCase() !== 'none'}>
            {infoData.profileErr ? '—' : !infoData.tradeBan || infoData.tradeBan.toLowerCase() === 'none' ? t(lang, 'no') : infoData.tradeBan}
          </span>
          <span class="ai-label">{t(lang, 'rowLimited')}</span>
          <span class="ai-value" class:bad={infoData.limited}>{infoData.profileErr ? '—' : yn(infoData.limited)}</span>
          <span class="ai-label">{t(lang, 'rowSince')}</span>
          <span class="ai-value">{infoData.memberSince || '—'}</span>
          {#if infoData.email}
            <span class="ai-label">Email</span>
            <span class="ai-value">{infoData.email}</span>
          {/if}
          {#if infoData.wallet}
            <span class="ai-label">{t(lang, 'rowWallet')}</span>
            <span class="ai-value">{infoData.wallet}</span>
          {/if}
          {#if infoData.country}
            <span class="ai-label">{t(lang, 'rowCountry')}</span>
            <span class="ai-value">{infoData.country}</span>
          {/if}
          <span class="ai-label">{t(lang, 'rowToken')}</span>
          <span class="ai-value" class:good={infoData.tokenAlive} class:bad={!infoData.tokenAlive}>{yn(infoData.tokenAlive)}</span>
        </div>
        {#if infoData.tokenSteamId || infoData.tokenIssued || infoData.tokenExpires}
          <div class="panel-label tok-label">{t(lang, 'tokSection')}</div>
          <div class="accinfo-grid">
            {#if infoData.tokenSteamId}
              <span class="ai-label">{t(lang, 'rowTokSteamID')}</span>
              <span class="ai-value">{infoData.tokenSteamId}</span>
            {/if}
            {#if infoData.tokenIssued}
              <span class="ai-label">{t(lang, 'rowTokIssued')}</span>
              <span class="ai-value">{infoData.tokenIssued}</span>
            {/if}
            {#if infoData.tokenExpires}
              <span class="ai-label">{t(lang, 'rowTokExpires')}</span>
              <span class="ai-value" class:bad={!infoData.tokenDaysLeft}>
                {infoData.tokenExpires}{infoData.tokenDaysLeft ? ` · ${infoData.tokenDaysLeft} ${t(lang, 'daysShort')}` : ''}
              </span>
            {/if}
            {#if infoData.tokenAudiences}
              <span class="ai-label">{t(lang, 'rowTokAud')}</span>
              <span class="ai-value" title={infoData.tokenAudiences}>{infoData.tokenAudiences}</span>
            {/if}
          </div>
        {/if}
        {#if infoData.tokenChecks?.length}
          <div class="panel-label tok-label">{t(lang, 'checksTitle')}</div>
          <div class="accinfo-grid checks-grid">
            {#each infoData.tokenChecks as c (c.id)}
              <span class="ai-label" title={checkDesc(c.id)}>{checkLabel(c.id)}</span>
              <span
                class="ai-value"
                title={checkDesc(c.id)}
                class:good={c.status === 'ok'}
                class:bad={c.status === 'fail'}
              >{c.status === 'ok' ? '✓' : c.status === 'fail' ? '✗' : '—'}</span>
            {/each}
          </div>
        {/if}
        {#if infoData.tokenClaimsJson}
          <details class="claims">
            <summary>{t(lang, 'rowTokClaims')}</summary>
            <pre class="claims-pre">{infoData.tokenClaimsJson}</pre>
          </details>
        {/if}
        {#if infoData.summary}
          <div class="accinfo-summary">{infoData.summary}</div>
        {/if}
        {#if infoLogin}
          <button class="btn primary" type="button" disabled={loading} onclick={loginFromInfo}>
            {loading ? t(lang, 'working') : t(lang, 'login')}
          </button>
        {/if}
        {#if infoLogin?.kind === 'key'}
          <button class="btn ghost block" type="button" disabled={loading} onclick={saveFromInfo}>
            {t(lang, 'saveBtn')}
          </button>
        {/if}
        {#if infoData.profileUrl}
          <button
            class="btn ghost block"
            type="button"
            onclick={() => (AppService as any).OpenURL(infoData!.profileUrl)}
          >{t(lang, 'accInfoOpenProfile')}</button>
        {/if}
      {/if}
    </div>
  </div>
{/if}

{#if showHelp}
  <div class="modal" role="dialog" aria-modal="true" tabindex="-1">
    <div class="modal-card">
      <h3>{t(lang, 'instructions')}</h3>
      <ol>
        <li>{t(lang, 'help1')}</li>
        <li>{t(lang, 'help2')}</li>
        <li>{t(lang, 'help3')}</li>
        <li>{t(lang, 'help4')}</li>
        <li>{t(lang, 'help5')}</li>
        <li>{t(lang, 'help6')}</li>
      </ol>
      <p class="guide">
        {t(lang, 'fullGuide')}
        <button class="link" type="button" onclick={() => Browser.OpenURL(GUIDE_URL)}>
          teletype.in/@hackerdlc/CS2NFA
        </button>
      </p>
      <button class="btn primary" type="button" onclick={() => (showHelp = false)}>{t(lang, 'gotIt')}</button>
    </div>
  </div>
{/if}

{#if showSettings}
  <div class="modal" role="dialog" aria-modal="true" tabindex="-1">
    <div class="modal-card update settings-modal">
      <div class="modal-head">
        <h3>{t(lang, 'settings')}</h3>
        <button class="modal-x" type="button" onclick={closeSettings} aria-label={t(lang, 'close')}>✕</button>
      </div>

      {#if driveAuthWait}
        <p class="update-msg wait">{t(lang, 'driveWaiting')}</p>
        <div class="modal-actions">
          <button class="btn primary danger-btn" type="button" onclick={cancelDriveAuth}>{t(lang, 'cancel')}</button>
          <button class="btn ghost block" type="button" onclick={closeSettings}>{t(lang, 'close')}</button>
        </div>
      {:else}
        <section class="settings-block">
          <div class="panel-label">{t(lang, 'settingsApp')}</div>
          <div class="settings-row">
            <span>{t(lang, 'settingsLang')}</span>
            <div class="lang-switch" use:slidingPill title={t(lang, 'lang')}>
              <button type="button" class="lang-btn" class:active={lang === 'ru'} onclick={() => setLang('ru')}>RU</button>
              <button type="button" class="lang-btn" class:active={lang === 'en'} onclick={() => setLang('en')}>EN</button>
            </div>
          </div>
          <div class="settings-row">
            <span>{t(lang, 'settingsTheme')}</span>
            <div class="lang-switch" use:slidingPill>
              <button type="button" class="lang-btn" class:active={theme === 'auto'} onclick={() => setTheme('auto')}>{t(lang, 'themeAuto')}</button>
              <button type="button" class="lang-btn" class:active={theme === 'dark'} onclick={() => setTheme('dark')}>{t(lang, 'themeDark')}</button>
              <button type="button" class="lang-btn" class:active={theme === 'light'} onclick={() => setTheme('light')}>{t(lang, 'themeLight')}</button>
            </div>
          </div>
          {#if mode === 'advanced'}
            <label class="check settings-check">
              <input type="checkbox" bind:checked={keepExisting} />
              <span class="box"></span>
              <span>{t(lang, 'keepExisting')}</span>
            </label>
          {/if}
          <div class="settings-actions">
            <button class="btn ghost block" type="button" onclick={harvestSteam}>{t(lang, 'harvestBtn')}</button>
            <button class="btn ghost block" type="button" onclick={() => checkUpdates(false)}>{t(lang, 'checkUpdate')}</button>
            <button class="btn ghost block" type="button" onclick={resetSteam}>{t(lang, 'resetSteam')}</button>
          </div>
        </section>

        <section class="settings-block">
          <div class="panel-label">{t(lang, 'settingsProxy')}</div>
          <label class="field">
            <textarea
              class="bulk-ta sm"
              spellcheck="false"
              autocomplete="off"
              placeholder="http://user:pass@host:port"
              bind:value={settingsProxies}
              oninput={saveProxies}
            ></textarea>
          </label>
          <p class="proxy-hint">{t(lang, 'proxyHint')}</p>
        </section>

        {#if mode === 'advanced'}
          <section class="settings-block">
            <div class="panel-label">{t(lang, 'settingsDrive')}</div>
          {#key driveStatus.connected}
            <p class="update-msg drive-status">
              {driveStatus.connected
                ? `${t(lang, 'driveConnected')}${driveStatus.clientIdHint ? ' · ' + driveStatus.clientIdHint : ''}`
                : driveStatus.hasCredentials
                  ? t(lang, 'driveNotConnected')
                  : t(lang, 'driveSetupHint')}
            </p>
          {/key}
          <button class="btn ghost block tut-toggle" type="button" onclick={openDriveGuide}>
            {t(lang, 'driveOpenGuide')}
          </button>
          {#if !driveStatus.hasCredentials}
            <label class="field">
              <input type="text" spellcheck="false" autocomplete="off" placeholder={t(lang, 'driveClientId')} bind:value={driveClientId} />
            </label>
            <label class="field" style="margin-top:8px">
              <input type="text" spellcheck="false" autocomplete="off" placeholder={t(lang, 'driveClientSecret')} bind:value={driveClientSecret} />
            </label>
            <div class="settings-actions" style="margin-top:10px">
              <button class="btn primary" type="button" disabled={driveBusy || !driveClientId.trim()} onclick={saveDriveCreds}>
                {t(lang, 'driveSave')}
              </button>
              <button class="btn ghost block" type="button" disabled={driveBusy} onclick={importDriveCreds}>
                {t(lang, 'driveImport')}
              </button>
            </div>
          {:else}
            <div class="settings-actions">
              {#if !driveStatus.connected}
                <button class="btn primary" type="button" disabled={driveBusy} onclick={connectDrive}>
                  {t(lang, 'driveConnect')}
                </button>
              {:else}
                <button class="btn ghost block" type="button" disabled={driveBusy} onclick={disconnectDrive}>
                  {t(lang, 'driveDisconnect')}
                </button>
              {/if}
              <button class="btn ghost block" type="button" disabled={driveBusy} onclick={importDriveCreds}>
                {t(lang, 'driveImport')}
              </button>
            </div>
          {/if}
          </section>
        {/if}
      {/if}
    </div>
  </div>
{/if}

{#if showBulk}
  <div class="modal" role="dialog" aria-modal="true" tabindex="-1">
    <div class="modal-card update settings-modal bulk-modal">
      <div class="modal-head">
        <h3>{t(lang, 'bulkTitle')}</h3>
        <button class="modal-x" type="button" disabled={bulkBusy} onclick={() => (showBulk = false)} aria-label={t(lang, 'close')}>✕</button>
      </div>
      <div class="bulk-keys-head">
        <span class="field-label">{t(lang, 'bulkKeysLabel')}</span>
        <button class="btn ghost sm" type="button" disabled={bulkBusy} onclick={bulkPickFile}>
          {t(lang, 'bulkFile')}
        </button>
      </div>
      <textarea
        class="bulk-ta"
        spellcheck="false"
        autocomplete="off"
        placeholder="login----token"
        bind:value={bulkKeys}
      ></textarea>
      <div class="bulk-proxy-hint">
        <span>{t(lang, 'bulkProxyGo')}</span>
        <button
          class="btn ghost sm"
          type="button"
          onclick={() => {
            showBulk = false;
            void openSettings();
          }}
        >{t(lang, 'bulkOpenSettings')}</button>
      </div>
      <div class="settings-actions">
        <button class="btn primary" type="button" disabled={bulkBusy || !bulkKeys.trim()} onclick={() => runBulkCheck(false)}>
          {bulkBusy ? t(lang, 'working') : t(lang, 'bulkStart')}
        </button>
        <button class="btn ghost block" type="button" disabled={bulkBusy} onclick={() => runBulkCheck(true)}>
          {t(lang, 'bulkSaved')}
        </button>
      </div>
      {#if bulkBusy}
        <div class="bulk-progress">
          <div class="bulk-progress-bar">
            <span
              class:indeterminate={bulkTotal === 0}
              style="width: {bulkTotal > 0 ? Math.round((bulkLive.length / bulkTotal) * 100) : 100}%"
            ></span>
          </div>
          <div class="bulk-progress-text">
            {bulkTotal > 0
              ? t(lang, 'bulkProgress', { done: bulkLive.length, total: bulkTotal })
              : t(lang, 'working')}
          </div>
        </div>
      {/if}
      {#if bulkItems.length > 0}
        <div class="logs-list bulk-results">
          {#each bulkItems as it, i (it.account + '-' + i)}
            <div
              class="log-row"
              class:ok={it.status === 'ok'}
              class:err={it.status === 'rejected' || it.status === 'expired' || it.status === 'invalid'}
              title={it.detail || it.steamId || ''}
            >
              <span class="log-text bulk-acc">{it.account || '—'}</span>
              <span class="bulk-status">{bulkStatusText(it.status)}</span>
              <span class="bulk-date">
                {#if it.expiresAt}
                  <span class="log-ts">{it.expiresAt}</span>
                {/if}
                {#if it.steamId}
                  <button class="btn mini info bulk-info" type="button" onclick={() => bulkInfo(it)}>
                    {t(lang, 'info')}
                  </button>
                {/if}
              </span>
            </div>
          {/each}
        </div>
        {#if !bulkBusy && bulkResult}
          <div class="bulk-export-row">
            <button class="btn ghost sm" type="button" disabled={bulkOkCount === 0} onclick={() => exportBulk('ok')}>
              {t(lang, 'bulkExportOk')} ({bulkOkCount})
            </button>
            <button class="btn ghost sm" type="button" disabled={bulkBadCount === 0} onclick={() => exportBulk('bad')}>
              {t(lang, 'bulkExportBad')} ({bulkBadCount})
            </button>
          </div>
        {/if}
      {:else if bulkResult && !bulkBusy}
        <p class="update-msg">{t(lang, 'logEmpty')}</p>
      {/if}
    </div>
  </div>
{/if}

{#if exportToast}
  <div class="modal success-modal export-toast">
    <div class="success-card" role="status">
      <svg class="success-check" viewBox="0 0 52 52" aria-hidden="true">
        <circle cx="26" cy="26" r="24" />
        <path d="M15 27l7 7 15-15" />
      </svg>
      <div class="busy-title">{t(lang, 'bulkFileCreated')}</div>
      <div class="export-path" title={exportToast}>{exportToast}</div>
    </div>
  </div>
{/if}

{#if busyOpen}
  <div class="modal busy-modal">
    <div class="modal-card busy-card" role="status">
      <div class="busy-ring" aria-hidden="true"></div>
      <div class="busy-title">{busyTitle}</div>
      {#if busySteps}
        <div class="busy-bar"><span style="width: {busyProgress}%"></span></div>
        {#key loginStage}
          <div class="busy-stage">{stageLabel(loginStage)}</div>
        {/key}
      {:else if busyMsg}
        <div class="busy-stage">{busyMsg}</div>
      {/if}
    </div>
  </div>
{/if}

{#if successOpen}
  <div class="modal success-modal">
    <div class="success-card" role="status">
      <svg class="success-check" viewBox="0 0 52 52" aria-hidden="true">
        <circle cx="26" cy="26" r="24" />
        <path d="M15 27l7 7 15-15" />
      </svg>
      <div class="busy-title">{t(lang, 'successTitle')}</div>
    </div>
  </div>
{/if}

{#if showUpdate && updateInfo}
  <div class="modal" role="dialog" aria-modal="true" tabindex="-1">
    <div class="modal-card update">
      <h3>{t(lang, 'updateTitle')}</h3>
      <p class="update-msg">
        {t(lang, 'updateAvailable', {
          current: updateInfo.currentVersion,
          latest: updateInfo.latestVersion,
        })}
      </p>
      {#if releaseNotesHtml}
        <div class="notes md">{@html releaseNotesHtml}</div>
      {/if}
      <div class="modal-actions">
        <button class="btn primary" type="button" disabled={updateBusy} onclick={installUpdate}>
          {updateBusy ? t(lang, 'updateInstalling') : t(lang, 'updateNow')}
        </button>
        <button class="btn ghost block" type="button" disabled={updateBusy} onclick={openReleasePage}>
          {t(lang, 'updateOpenPage')}
        </button>
        <button class="btn ghost block" type="button" disabled={updateBusy} onclick={() => (showUpdate = false)}>
          {t(lang, 'updateLater')}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  :global(:root) {
    --bg: #09090b;
    --bg-elevated: #141416;
    --bg-soft: #1f1f23;
    --bg-input: #0c0c0e;
    --bg-input-focus: #101012;
    --bg-row: #121214;
    --bg-row-hover: #17171a;
    --bg-titlebar: rgba(14, 14, 17, 0.96);
    --bg-statusbar: rgba(9, 9, 11, 0.9);
    --border: rgba(255, 255, 255, 0.08);
    --border-strong: rgba(255, 255, 255, 0.16);
    --text: #fafafa;
    --text-secondary: #b4b4bc;
    --muted: #8b8b94;
    --accent: #ffffff;
    --pill-bg: rgba(255, 255, 255, 0.16);
    --accent-hover: #f4f4f5;
    --accent-active: #e4e4e7;
    --accent-fg: #09090b;
    --accent-soft: rgba(255, 255, 255, 0.1);
    --ok: #d4d4d8;
    --good: #4ade80;
    --danger: #f87171;
    --info: #e4e4e7;
    --hover: rgba(255, 255, 255, 0.04);
    --hover-strong: rgba(255, 255, 255, 0.1);
    --overlay: rgba(0, 0, 0, 0.72);
    --placeholder: #52525b;
    --focus-border: rgba(255, 255, 255, 0.45);
    --focus-ring: rgba(255, 255, 255, 0.08);
    --picked-bg: rgba(255, 255, 255, 0.06);
    --picked-border: rgba(255, 255, 255, 0.35);
    --summary-bg: rgba(255, 255, 255, 0.03);
    --mini-login-bg: rgba(255, 255, 255, 0.1);
    --mini-login-bg-hover: rgba(255, 255, 255, 0.18);
    --mini-hover-border: rgba(255, 255, 255, 0.28);
    --md-code-bg: rgba(255, 255, 255, 0.08);
    --md-pre-bg: rgba(0, 0, 0, 0.35);
    --md-quote-border: rgba(255, 255, 255, 0.35);
    --md-quote-bg: rgba(255, 255, 255, 0.04);
    --md-th-bg: rgba(255, 255, 255, 0.03);
    --logo-ring: rgba(255, 255, 255, 0.08);
    --beta-fg: #fbbf24;
    --row-hover-shadow: 0 4px 16px rgba(0, 0, 0, 0.18);
    --logo-shadow: 0 0 0 1px var(--logo-ring), 0 8px 20px rgba(0, 0, 0, 0.35);
    --primary-shadow: 0 1px 0 rgba(255, 255, 255, 0.35) inset, 0 8px 22px rgba(0, 0, 0, 0.28);
    --primary-shadow-hover: 0 1px 0 rgba(255, 255, 255, 0.35) inset, 0 12px 28px rgba(0, 0, 0, 0.34);
    --modal-shadow: 0 24px 64px rgba(0, 0, 0, 0.5);
    --pill-shadow: 0 1px 2px rgba(0, 0, 0, 0.25);
    --radius: 14px;
    --radius-sm: 10px;
    --shadow: 0 1px 0 rgba(255, 255, 255, 0.04) inset, 0 16px 40px rgba(0, 0, 0, 0.4);
    --ease: cubic-bezier(0.22, 1, 0.36, 1);
    --font: "Segoe UI Variable", "Segoe UI", system-ui, -apple-system, sans-serif;
    font-family: var(--font);
    color-scheme: dark;
  }

  :global(:root[data-theme='light']) {
    --bg: #f4f4f5;
    --bg-elevated: #ffffff;
    --bg-soft: #e4e4e7;
    --bg-input: #f4f4f5;
    --bg-input-focus: #ffffff;
    --bg-row: #f4f4f5;
    --bg-row-hover: #ececee;
    --bg-titlebar: rgba(250, 250, 250, 0.96);
    --bg-statusbar: rgba(244, 244, 245, 0.9);
    --border: rgba(0, 0, 0, 0.09);
    --border-strong: rgba(0, 0, 0, 0.22);
    --text: #18181b;
    --text-secondary: #45454e;
    --muted: #61616b;
    --accent: #18181b;
    --pill-bg: rgba(0, 0, 0, 0.1);
    --accent-hover: #27272a;
    --accent-active: #3f3f46;
    --accent-fg: #fafafa;
    --accent-soft: rgba(0, 0, 0, 0.07);
    --ok: #16a34a;
    --good: #16a34a;
    --danger: #dc2626;
    --info: #3f3f46;
    --hover: rgba(0, 0, 0, 0.04);
    --hover-strong: rgba(0, 0, 0, 0.08);
    --overlay: rgba(0, 0, 0, 0.4);
    --placeholder: #a1a1aa;
    --focus-border: rgba(0, 0, 0, 0.45);
    --focus-ring: rgba(0, 0, 0, 0.08);
    --picked-bg: rgba(0, 0, 0, 0.05);
    --picked-border: rgba(0, 0, 0, 0.35);
    --summary-bg: rgba(0, 0, 0, 0.03);
    --mini-login-bg: rgba(0, 0, 0, 0.07);
    --mini-login-bg-hover: rgba(0, 0, 0, 0.12);
    --mini-hover-border: rgba(0, 0, 0, 0.3);
    --md-code-bg: rgba(0, 0, 0, 0.06);
    --md-pre-bg: rgba(0, 0, 0, 0.04);
    --md-quote-border: rgba(0, 0, 0, 0.3);
    --md-quote-bg: rgba(0, 0, 0, 0.03);
    --md-th-bg: rgba(0, 0, 0, 0.03);
    --logo-ring: rgba(0, 0, 0, 0.1);
    --beta-fg: #b45309;
    --row-hover-shadow: 0 4px 14px rgba(0, 0, 0, 0.07);
    --logo-shadow: 0 0 0 1px var(--logo-ring), 0 3px 10px rgba(0, 0, 0, 0.14);
    --primary-shadow: 0 1px 0 rgba(255, 255, 255, 0.1) inset, 0 3px 10px rgba(0, 0, 0, 0.14);
    --primary-shadow-hover: 0 1px 0 rgba(255, 255, 255, 0.1) inset, 0 5px 14px rgba(0, 0, 0, 0.17);
    --modal-shadow: 0 12px 32px rgba(0, 0, 0, 0.14);
    --pill-shadow: 0 1px 2px rgba(0, 0, 0, 0.16);
    --shadow: 0 1px 2px rgba(0, 0, 0, 0.04), 0 6px 20px rgba(0, 0, 0, 0.06);
    color-scheme: light;
  }

  :global(*) {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
    scrollbar-width: none;
  }

  :global(::-webkit-scrollbar) {
    width: 0;
    height: 0;
    display: none;
  }

  :global(html),
  :global(body),
  :global(#app) {
    height: 100%;
    overflow: hidden;
    background: var(--bg);
    color: var(--text);
    user-select: none;
  }

  :global(body) {
    background: var(--bg);
    --wails-draggable: drag;
  }

  @keyframes fade-up {
    from { opacity: 0; transform: translateY(8px); }
    to { opacity: 1; transform: translateY(0); }
  }

  @keyframes fade-in {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  @keyframes scale-in {
    from { opacity: 0; transform: scale(0.96); }
    to { opacity: 1; transform: scale(1); }
  }

  @keyframes slide-x {
    from { opacity: 0; transform: translateX(10px); }
    to { opacity: 1; transform: translateX(0); }
  }

  .app {
    height: 100%;
    display: grid;
    grid-template-rows: auto 1fr auto;
  }

  .titlebar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 14px 14px 12px 18px;
    border-bottom: 1px solid var(--border);
    background: var(--bg-titlebar);
  }

  .brand {
    display: flex;
    align-items: center;
    gap: 12px;
    min-width: 0;
  }

  .logo-mark {
    width: 32px;
    height: 32px;
    border-radius: 10px;
    display: grid;
    place-items: center;
    font-weight: 800;
    font-size: 14px;
    color: var(--accent-fg);
    background: var(--accent);
    box-shadow: var(--logo-shadow);
    flex-shrink: 0;
  }

  .brand-text {
    display: flex;
    flex-direction: column;
    gap: 1px;
    min-width: 0;
  }

  .brand-name {
    font-size: 14px;
    font-weight: 650;
    letter-spacing: -0.01em;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .brand-ver {
    font-size: 11px;
    color: var(--muted);
    font-weight: 600;
  }

  .title-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    --wails-draggable: no-drag;
  }

  .lang-switch {
    display: inline-flex;
    padding: 3px;
    background: var(--bg-elevated);
    border: 1px solid var(--border);
    border-radius: 999px;
    position: relative;
  }

  .lang-switch::before {
    content: '';
    position: absolute;
    top: 3px;
    bottom: 3px;
    left: 0;
    width: var(--pill-w, calc(50% - 3px));
    border-radius: 999px;
    background: var(--pill-bg);
    box-shadow: var(--pill-shadow);
    transform: translateX(var(--pill-x, 3px));
    transition:
      transform 0.3s var(--ease),
      width 0.3s var(--ease);
  }

  .lang-btn {
    white-space: nowrap;
    border: none;
    background: transparent;
    color: var(--muted);
    padding: 5px 11px;
    font-size: 11px;
    font-weight: 700;
    line-height: 1;
    border-radius: 999px;
    cursor: pointer;
    position: relative;
    z-index: 1;
    display: inline-flex;
    align-items: center;
    gap: 5px;
    transition: color 0.2s var(--ease), transform 0.2s var(--ease);
  }

  .lang-btn.active {
    color: var(--text);
  }

  .lang-btn:active {
    transform: scale(0.96);
  }

  .win-btns {
    display: flex;
    gap: 4px;
    margin-left: 4px;
  }

  .icon-btn {
    width: 32px;
    height: 32px;
    display: inline-grid;
    place-items: center;
    border-radius: 999px;
    border: 1px solid var(--border);
    background: var(--bg-elevated);
    color: var(--muted);
    cursor: pointer;
    transition:
      background 0.2s var(--ease),
      border-color 0.2s var(--ease),
      color 0.2s var(--ease),
      transform 0.2s var(--ease);
  }

  .icon-btn svg {
    width: 16px;
    height: 16px;
  }

  .icon-btn:hover {
    background: var(--hover);
    border-color: var(--border-strong);
    color: var(--text);
  }

  .icon-btn:active {
    transform: scale(0.92);
  }

  .win svg {
    width: 26px;
    height: 26px;
    display: block;
  }

  .win {
    width: 38px;
    height: 32px;
    border: none;
    border-radius: 8px;
    background: transparent;
    color: var(--muted);
    cursor: pointer;
    display: inline-grid;
    place-items: center;
    transition: background 0.2s var(--ease), color 0.2s var(--ease);
  }

  .win:hover {
    background: var(--picked-bg);
    color: var(--text);
  }

  .win.close:hover {
    background: #e11d48;
    color: white;
  }

  .btn {
    border: 1px solid transparent;
    border-radius: var(--radius-sm);
    font: inherit;
    font-weight: 600;
    cursor: pointer;
    transition:
      background 0.2s var(--ease),
      border-color 0.2s var(--ease),
      color 0.2s var(--ease),
      opacity 0.2s var(--ease),
      transform 0.2s var(--ease),
      box-shadow 0.2s var(--ease);
  }

  .btn:active:not(:disabled) {
    transform: scale(0.98);
  }

  .btn.ghost {
    background: transparent;
    border-color: var(--border);
    color: var(--text-secondary);
    padding: 7px 12px;
    font-size: 12.5px;
  }

  .btn.ghost:hover:not(:disabled) {
    background: var(--hover);
    border-color: var(--border-strong);
    color: var(--text);
    transform: translateY(-1px);
  }

  .btn.ghost:active:not(:disabled) {
    transform: translateY(0) scale(0.98);
  }

  .btn.ghost.sm {
    padding: 6px 10px;
    font-size: 11.5px;
    border-radius: 8px;
  }

  .btn.ghost.accent {
    color: var(--text);
    border-color: var(--mini-hover-border);
    background: var(--accent-soft);
  }

  .btn.ghost.accent:hover:not(:disabled) {
    border-color: var(--focus-border);
    background: var(--hover-strong);
    color: var(--text);
  }

  .btn.ghost.block,
  .btn.block {
    width: 100%;
    height: 42px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }

  .btn.primary {
    height: 46px;
    width: 100%;
    border: none;
    border-radius: 12px;
    background: var(--accent);
    color: var(--accent-fg);
    font-weight: 700;
    font-size: 14.5px;
    box-shadow: var(--primary-shadow);
  }

  .btn.primary:hover:not(:disabled) {
    background: var(--accent-hover);
    transform: translateY(-1px);
    box-shadow: var(--primary-shadow-hover);
  }

  .btn.primary:active:not(:disabled) {
    background: var(--accent-active);
    transform: translateY(0) scale(0.98);
  }

  .btn:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }

  .btn.mini {
    padding: 7px 12px;
    font-size: 12px;
    border-radius: 8px;
  }

  .btn.mini.export {
    background: transparent;
    color: var(--text-secondary);
    border-color: var(--border);
    opacity: 0;
    pointer-events: none;
    transform: translateX(6px);
    transition:
      opacity 0.18s var(--ease),
      transform 0.18s var(--ease),
      background 0.2s var(--ease),
      border-color 0.2s var(--ease),
      color 0.2s var(--ease);
  }

  .account-row:hover .btn.mini.export,
  .account-row:focus-within .btn.mini.export {
    opacity: 1;
    pointer-events: auto;
    transform: translateX(0);
  }

  .btn.mini.export:hover:not(:disabled) {
    background: var(--hover-strong);
    border-color: var(--mini-hover-border);
    color: var(--text);
  }

  .btn.mini.login {
    background: var(--mini-login-bg);
    color: var(--text);
    border-color: transparent;
  }

  .btn.mini.login:hover:not(:disabled) {
    background: var(--mini-login-bg-hover);
  }

  .btn.mini.del {
    background: rgba(251, 113, 133, 0.1);
    color: var(--danger);
  }
  .btn.mini.del:hover:not(:disabled) {
    background: rgba(251, 113, 133, 0.2);
  }

  .btn.mini.info {
    background: transparent;
    color: var(--text-secondary);
    border-color: var(--border);
  }

  .btn.mini.info:hover:not(:disabled) {
    background: var(--hover-strong);
    border-color: var(--mini-hover-border);
    color: var(--text);
  }

  .accinfo-head {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 14px;
  }

  .accinfo-ava {
    width: 52px;
    height: 52px;
    border-radius: 12px;
    border: 1px solid var(--border);
    flex-shrink: 0;
  }

  .accinfo-id-block {
    min-width: 0;
  }

  .accinfo-nick {
    font-size: 16px;
    font-weight: 700;
    color: var(--text);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .accinfo-id {
    font-size: 12px;
    color: var(--muted);
    margin-top: 2px;
    user-select: text;
    -webkit-user-select: text;
  }

  .accinfo-grid {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 8px 16px;
    background: var(--bg-input);
    border: 1px solid var(--border);
    border-radius: 12px;
    padding: 14px 16px;
    margin-bottom: 14px;
    font-size: 13px;
  }

  .ai-label {
    color: var(--muted);
  }

  .ai-value {
    color: var(--text-secondary);
    text-align: right;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .ai-value.good {
    color: var(--good);
  }

  .ai-value.bad {
    color: var(--danger);
  }

  .accinfo-summary {
    font-size: 12.5px;
    line-height: 1.5;
    color: var(--muted);
    background: var(--summary-bg);
    border-left: 3px solid var(--border-strong);
    border-radius: 0 8px 8px 0;
    padding: 8px 12px;
    margin: -6px 0 14px;
    max-height: 90px;
    overflow: auto;
    user-select: text;
    -webkit-user-select: text;
    white-space: pre-wrap;
    word-break: break-word;
  }

  .ai-value.path {
    direction: rtl;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .tok-label {
    margin: 2px 0 10px;
  }

  .checks-grid {
    font-size: 12.5px;
  }

  .claims {
    margin: -4px 0 14px;
  }

  .claims summary {
    cursor: pointer;
    color: var(--text-secondary);
    font-size: 12.5px;
    font-weight: 600;
    user-select: none;
  }

  .claims-pre {
    margin-top: 8px;
    background: var(--bg-input);
    border: 1px solid var(--border);
    border-radius: 10px;
    padding: 10px 12px;
    font-size: 11.5px;
    line-height: 1.5;
    font-family: ui-monospace, Consolas, monospace;
    color: var(--text-secondary);
    max-height: 180px;
    overflow: auto;
    user-select: text;
    -webkit-user-select: text;
    white-space: pre-wrap;
    word-break: break-word;
  }

  .sys-panel {
    padding: 16px 18px;
    display: flex;
    flex-direction: column;
    gap: 10px;
    min-height: 0;
    overflow: hidden;
  }

  .sys-panel .btn,
  .logs-panel .btn {
    flex-shrink: 0;
  }

  .sys-panel .btn.ghost.block {
    height: 38px;
  }

  .sys-panel .accinfo-grid {
    margin-bottom: 2px;
    padding: 12px 14px;
  }

  .sys-panel .alpha-warn {
    padding: 7px 10px;
  }

  .keycheck-label {
    margin-top: 2px;
    padding-top: 10px;
    border-top: 1px solid var(--border);
  }

  .beta-chip {
    display: inline-block;    padding: 2px 6px;
    border-radius: 999px;
    font-size: 9px;
    font-weight: 800;
    letter-spacing: 0.08em;
    color: var(--danger);
    background: rgba(239, 68, 68, 0.12);
    border: 1px solid rgba(239, 68, 68, 0.35);
    line-height: 1.3;
    flex-shrink: 0;
  }

  .accounts-head .panel-label {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }

  .logs-panel .accounts-head {
    align-items: center;
  }

  .logs-panel .accounts-head .beta-chip {
    line-height: 1;
  }

  .lang-btn .beta-chip {
    margin-left: 0;
    padding: 2px 5px;
    font-size: 8px;
    line-height: 1;
    display: inline-flex;
    align-items: center;
    transform: translateY(-0.5px);
  }

  .alpha-warn {
    font-size: 12px;
    line-height: 1.45;
    color: var(--danger);
    background: rgba(239, 68, 68, 0.1);
    border: 1px solid rgba(239, 68, 68, 0.3);
    border-radius: 10px;
    padding: 8px 12px;
  }

  .logs-panel {
    padding: 16px 16px 12px;
    display: flex;
    flex-direction: column;
    min-height: 0;
  }

  .logs-list {
    overflow: auto;
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-height: 0;
    flex: 1;
    padding-right: 2px;
    user-select: text;
    -webkit-user-select: text;
  }

  .log-row {
    display: flex;
    gap: 10px;
    align-items: baseline;
    font-size: 12.5px;
    padding: 8px 12px;
    background: var(--bg-row);
    border: 1px solid var(--border);
    border-radius: 10px;
    animation: fade-up 0.25s var(--ease) both;
  }

  .log-ts {
    color: var(--muted);
    font-variant-numeric: tabular-nums;
    flex-shrink: 0;
  }

  .log-text {
    color: var(--text-secondary);
    word-break: break-word;
  }

  .log-row.ok .log-text {
    color: var(--ok);
  }

  .log-row.err .log-text {
    color: var(--danger);
  }

  .btn.ghost.sm.danger-sm {
    color: var(--danger);
    border-color: rgba(251, 113, 133, 0.35);
  }

  .btn.ghost.sm.danger-sm:hover:not(:disabled) {
    background: rgba(251, 113, 133, 0.12);
    border-color: rgba(251, 113, 133, 0.5);
    color: var(--danger);
  }

  .danger-btn {
    background: #fb7185 !important;
    color: #1c0508 !important;
    box-shadow: none !important;
  }

  .swap-wrap {
    display: grid;
    min-height: 0;
    overflow: hidden;
  }

  .swap-wrap > .content {
    grid-area: 1 / 1;
  }

  .content {
    display: grid;
    grid-template-columns: 340px 1fr;
    grid-template-rows: minmax(0, 1fr);
    gap: 16px;
    padding: 16px 18px;
    min-height: 0;
  }

  .mode-swap .panel,
  .mode-swap .account-row,
  .mode-swap .log-row,
  .mode-swap .count-chip,
  .mode-swap .action-row,
  .mode-swap .import-file-row,
  .mode-swap .login-panel .check,
  .mode-swap .export-bar,
  .mode-swap .hint,
  .mode-swap .pick {
    animation: none;
  }

  .panel {
    background: var(--bg-elevated);
    border: 1px solid var(--border);
    border-radius: calc(var(--radius) + 2px);
    box-shadow: var(--shadow);
    --wails-draggable: no-drag;
    animation: fade-up 0.35s var(--ease) both;
  }

  .accounts-panel {
    animation-delay: 0.05s;
  }

  .login-panel {
    padding: 20px 18px;
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .accounts-panel {
    padding: 16px 16px 12px;
    display: flex;
    flex-direction: column;
    min-height: 0;
  }

  .panel-label {
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--muted);
  }

  .login-panel .panel-label {
    margin-bottom: 2px;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .field-label {
    font-size: 12px;
    color: var(--text-secondary);
    font-weight: 500;
  }

  .field input {
    width: 100%;
    height: 44px;
    border-radius: 11px;
    border: 1px solid var(--border);
    background: var(--bg-input);
    color: var(--text);
    padding: 0 14px;
    font-size: 13.5px;
    outline: none;
    user-select: text;
    -webkit-user-select: text;
    transition: border-color 0.2s var(--ease), box-shadow 0.2s var(--ease), background 0.2s var(--ease);
  }

  .field input::placeholder {
    color: var(--placeholder);
  }

  .field input:focus {
    border-color: var(--focus-border);
    box-shadow: 0 0 0 3px var(--focus-ring);
    background: var(--bg-input-focus);
  }

  .action-row {
    display: grid;
    grid-template-columns: 1.4fr 1fr;
    gap: 8px;
    animation: fade-up 0.3s var(--ease) both;
  }

  .login-panel .check {
    animation: fade-up 0.3s var(--ease) both;
  }

  .action-row .btn.primary {
    width: 100%;
  }

  .action-row .btn.ghost.block {
    height: 46px;
  }

  .import-file-row {
    display: flex;
    gap: 8px;
    align-items: stretch;
    animation: fade-up 0.3s var(--ease) 0.05s both;
  }

  .import-file-row .btn {
    flex: 1;
    width: auto;
    height: 46px;
  }

  .lang-btn svg {
    width: 12px;
    height: 12px;
    flex-shrink: 0;
  }

  .help-fab {
    width: 46px;
    flex-shrink: 0;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: transparent;
    color: var(--muted);
    font: inherit;
    cursor: pointer;
    display: inline-grid;
    place-items: center;
    transition:
      background 0.2s var(--ease),
      border-color 0.2s var(--ease),
      color 0.2s var(--ease),
      transform 0.2s var(--ease);
  }

  .help-fab svg {
    width: 24px;
    height: 24px;
  }

  .help-fab:hover {
    background: var(--hover);
    border-color: var(--border-strong);
    color: var(--text);
    transform: scale(1.06);
  }

  .help-fab:active {
    transform: scale(0.94);
  }

  .help-wrap {
    position: relative;
    display: flex;
    flex-shrink: 0;
  }

  .help-hint {
    position: absolute;
    top: calc(100% + 2px);
    right: calc(var(--aim-r, 21px) - 14px);
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 0;
    border: none;
    background: none;
    padding: 0;
    font: inherit;
    cursor: pointer;
    white-space: nowrap;
    z-index: 6;
    --wails-draggable: no-drag;
    animation: hint-bob 1.6s ease-in-out infinite;
  }

  .hh-arrow {
    width: 52px;
    height: 38px;
    margin-right: -1px;
    overflow: visible;
    filter: drop-shadow(0 0 5px rgba(255, 51, 85, 0.9)) drop-shadow(0 0 14px rgba(255, 51, 85, 0.45));
  }

  .hh-arrow path {
    fill: none;
    stroke: #ff3355;
    stroke-width: 3;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  .hh-text {
    font-size: 12px;
    font-weight: 800;
    letter-spacing: 0.08em;
    color: #ff5c73;
    text-shadow:
      0 0 6px rgba(255, 51, 85, 0.85),
      0 0 16px rgba(255, 51, 85, 0.5);
    animation: neon-pulse 1.4s ease-in-out infinite;
  }

  @keyframes hint-bob {
    0%, 100% { transform: translateY(0); }
    50% { transform: translateY(4px); }
  }

  @keyframes neon-pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.7; }
  }

  .check {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 13px;
    color: var(--text-secondary);
    cursor: pointer;
  }

  .check input,
  .pick input {
    display: none;
  }

  .check .box,
  .pick .box {
    width: 18px;
    height: 18px;
    border-radius: 5px;
    border: 1.5px solid var(--border-strong);
    background: var(--bg-input);
    flex-shrink: 0;
    position: relative;
    display: inline-grid;
    place-items: center;
    transition: background 0.2s var(--ease), border-color 0.2s var(--ease), transform 0.2s var(--ease), box-shadow 0.2s var(--ease);
  }

  .check:active .box,
  .pick:active .box {
    transform: scale(0.92);
  }

  .check input:checked + .box,
  .pick input:checked + .box {
    background: var(--accent);
    border-color: var(--accent);
    box-shadow: 0 0 0 3px var(--focus-ring);
  }

  .check input:checked + .box::after,
  .pick input:checked + .box::after {
    content: '';
    width: 13px;
    height: 13px;
    background-color: var(--accent-fg);
    -webkit-mask: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 16 16'%3E%3Cpath fill='black' d='M6.2 11.6 2.6 8l1.2-1.2 2.4 2.4 5.9-5.9L13.3 4.5z'/%3E%3C/svg%3E") center / contain no-repeat;
    mask: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 16 16'%3E%3Cpath fill='black' d='M6.2 11.6 2.6 8l1.2-1.2 2.4 2.4 5.9-5.9L13.3 4.5z'/%3E%3C/svg%3E") center / contain no-repeat;
    animation: scale-in 0.16s var(--ease) both;
  }

  .hint {
    margin-top: auto;
    font-size: 12.5px;
    line-height: 1.45;
    color: var(--text-secondary);
    padding-top: 4px;
    animation: fade-in 0.3s var(--ease) both;
  }

  .accounts-head {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 14px;
  }

  .accounts-head > div:first-child {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .head-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .bulk-ta {
    width: 100%;
    min-height: 96px;
    border-radius: 11px;
    border: 1px solid var(--border);
    background: var(--bg-input);
    color: var(--text);
    padding: 10px 14px;
    font-size: 12.5px;
    font-family: ui-monospace, Consolas, monospace;
    line-height: 1.5;
    outline: none;
    resize: vertical;
    user-select: text;
    -webkit-user-select: text;
    transition: border-color 0.2s var(--ease), box-shadow 0.2s var(--ease), background 0.2s var(--ease);
  }

  .bulk-ta.sm {
    min-height: 56px;
  }

  .bulk-ta::placeholder {
    color: var(--placeholder);
  }

  .bulk-ta:focus {
    border-color: var(--focus-border);
    box-shadow: 0 0 0 3px var(--focus-ring);
    background: var(--bg-input-focus);
  }

  .bulk-proxy-hint {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    font-size: 12px;
    color: var(--muted);
    margin: 10px 0 12px;
  }

  .bulk-keys-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    margin-bottom: 6px;
  }

  .proxy-hint {
    margin-top: 8px;
    font-size: 12px;
    color: var(--muted);
    line-height: 1.4;
  }

  .bulk-results {
    margin-top: 12px;
    max-height: 300px;
  }

  .bulk-acc {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .bulk-date {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .bulk-info {
    opacity: 0;
    pointer-events: none;
    transform: translateX(6px);
    transition:
      opacity 0.18s var(--ease),
      transform 0.18s var(--ease),
      background 0.2s var(--ease),
      border-color 0.2s var(--ease),
      color 0.2s var(--ease);
  }

  .log-row:hover .bulk-info,
  .log-row:focus-within .bulk-info {
    opacity: 1;
    pointer-events: auto;
    transform: translateX(0);
  }

  .bulk-progress {
    margin-top: 12px;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .bulk-progress-bar {
    height: 6px;
    border-radius: 999px;
    background: var(--bg-row);
    border: 1px solid var(--border);
    overflow: hidden;
  }

  .bulk-progress-bar span {
    display: block;
    height: 100%;
    border-radius: 999px;
    background: linear-gradient(90deg, var(--accent), var(--accent-hover, var(--accent)));
    transition: width 0.3s var(--ease);
    position: relative;
    overflow: hidden;
  }

  .bulk-progress-bar span::after {
    content: '';
    position: absolute;
    inset: 0;
    background: linear-gradient(90deg, transparent, rgba(0, 0, 0, 0.22), transparent);
    animation: bulk-shimmer 1.2s linear infinite;
  }

  :root[data-theme='light'] .bulk-progress-bar span::after {
    background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.35), transparent);
  }

  @keyframes bulk-shimmer {
    from { transform: translateX(-100%); }
    to { transform: translateX(100%); }
  }

  .bulk-progress-text {
    font-size: 12px;
    color: var(--muted);
    text-align: center;
  }

  .bulk-export-row {
    margin-top: 10px;
    display: flex;
    gap: 8px;
    justify-content: flex-end;
    animation: fade-up 0.3s var(--ease) both;
  }

  .export-toast .success-card {
    padding: 20px 26px;
  }

  .export-path {
    font-size: 11px;
    color: var(--muted);
    max-width: 340px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    direction: rtl;
  }

  .bulk-status {
    font-weight: 700;
    flex-shrink: 0;
  }

  .log-row.ok .bulk-status {
    color: var(--good);
  }

  .log-row.err .bulk-status {
    color: var(--danger);
  }

  .count-chip {
    min-width: 22px;
    height: 22px;
    padding: 0 7px;
    border-radius: 999px;
    background: var(--bg-soft);
    border: 1px solid var(--border);
    color: var(--text-secondary);
    font-size: 11px;
    font-weight: 700;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    animation: scale-in 0.25s var(--ease) both;
  }

  .num-pop {
    display: inline-block;
    margin-left: 2px;
    animation: num-pop 0.28s var(--ease) both;
  }

  @keyframes num-pop {
    from { opacity: 0; transform: scale(0.5) translateY(2px); }
    60% { opacity: 1; transform: scale(1.2) translateY(0); }
    to { opacity: 1; transform: scale(1); }
  }

  .export-bar {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    animation: slide-x 0.28s var(--ease) both;
  }

  .pick {
    animation: check-pop 0.25s var(--ease) both;
    animation-delay: inherit;
  }

  @keyframes check-pop {
    from { opacity: 0; transform: scale(0.4); }
    60% { opacity: 1; transform: scale(1.15); }
    to { opacity: 1; transform: scale(1); }
  }

  .accounts-list {
    overflow: auto;
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-height: 0;
    flex: 1;
    padding-right: 2px;
    user-select: none;
  }

  .accounts-list.painting {
    cursor: crosshair;
  }

  .accounts-list.painting .account-row {
    transform: none;
  }

  .account-row {
    display: grid;
    grid-template-columns: auto auto 1fr auto;
    gap: 12px;
    align-items: center;
    background: var(--bg-row);
    border: 1px solid var(--border);
    border-radius: 12px;
    padding: 15px 14px;
    cursor: default;
    transition:
      border-color 0.22s var(--ease),
      background 0.22s var(--ease),
      transform 0.22s var(--ease),
      box-shadow 0.22s var(--ease);
    animation: fade-up 0.3s var(--ease) both;
  }

  .account-row:hover {
    border-color: var(--border-strong);
    background: var(--bg-row-hover);
    box-shadow: var(--row-hover-shadow);
  }

  .account-row.picked {
    border-color: var(--picked-border);
    background: var(--picked-bg);
  }

  .account-row.simple {
    grid-template-columns: auto 1fr auto;
  }

  .acc-ava {
    width: 34px;
    height: 34px;
    border-radius: 9px;
    border: 1px solid var(--border);
    object-fit: cover;
    flex-shrink: 0;
  }

  .acc-ava.placeholder {
    display: grid;
    place-items: center;
    background: var(--bg-soft);
    color: var(--muted);
    font-size: 14px;
    font-weight: 700;
    text-transform: uppercase;
  }

  .pick {
    display: flex;
    align-items: center;
    cursor: pointer;
    pointer-events: none;
  }

  .meta .name {
    font-weight: 600;
    font-size: 14px;
    letter-spacing: -0.01em;
  }

  .meta .exp {
    margin-top: 3px;
    font-size: 12px;
    color: var(--muted);
    transition: color 0.35s var(--ease);
  }

  .meta .exp.ok {
    color: var(--ok);
  }

  .meta .exp.bad {
    color: var(--danger);
  }

  .row-actions {
    display: flex;
    gap: 6px;
  }

  .empty {
    flex: 1;
    display: grid;
    place-content: center;
    gap: 8px;
    text-align: center;
    color: var(--muted);
    font-size: 13.5px;
    padding: 40px 12px;
  }

  .empty-icon {
    font-size: 22px;
    opacity: 0.45;
    animation: float 3.2s ease-in-out infinite;
    display: flex;
    justify-content: center;
  }

  .empty-icon svg {
    width: 22px;
    height: 22px;
    display: block;
  }

  @keyframes float {
    0%, 100% { transform: translateY(0); }
    50% { transform: translateY(-4px); }
  }

  .statusbar {
    min-height: 30px;
    display: flex;
    align-items: center;
    padding: 0 18px;
    font-size: 12px;
    line-height: 1;
    color: var(--muted);
    border-top: 1px solid var(--border);
    background: var(--bg-statusbar);
    transition: color 0.3s var(--ease);
  }

  .statusbar.ok {
    color: var(--ok);
  }

  .statusbar.err {
    color: var(--danger);
  }

  .modal {
    position: fixed;
    inset: 0;
    background: var(--overlay);
    display: grid;
    place-items: center;
    z-index: 50;
    --wails-draggable: no-drag;
    animation: fade-in 0.2s var(--ease) both;
  }

  .modal-card {
    width: min(440px, 90vw);
    background: var(--bg-elevated);
    border: 1px solid var(--border);
    border-radius: 18px;
    padding: 22px;
    box-shadow: var(--modal-shadow);
    animation: scale-in 0.28s var(--ease) both;
  }

  .modal-card > * {
    animation: fade-up 0.3s var(--ease) both;
    flex-shrink: 0;
  }

  .modal-card > *:nth-child(2) { animation-delay: 0.04s; }
  .modal-card > *:nth-child(3) { animation-delay: 0.08s; }
  .modal-card > *:nth-child(4) { animation-delay: 0.12s; }
  .modal-card > *:nth-child(n + 5) { animation-delay: 0.16s; }

  .busy-modal {
    z-index: 60;
    cursor: wait;
  }

  .busy-card {
    width: min(340px, 84vw);
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    padding: 26px 24px;
    text-align: center;
  }

  @keyframes busy-spin {
    to { transform: rotate(360deg); }
  }

  .busy-ring {
    width: 44px;
    height: 44px;
    border-radius: 50%;
    border: 3px solid var(--border-strong);
    border-top-color: var(--text);
    animation: busy-spin 0.8s linear infinite;
  }

  .busy-title {
    font-size: 14px;
    font-weight: 700;
    color: var(--text);
  }

  .busy-bar {
    width: 100%;
    height: 6px;
    border-radius: 999px;
    background: var(--bg-soft);
    overflow: hidden;
  }

  .busy-bar span {
    display: block;
    height: 100%;
    border-radius: 999px;
    background: var(--accent);
    transition: width 0.35s var(--ease);
  }

  .busy-stage {
    font-size: 12.5px;
    color: var(--text-secondary);
    min-height: 18px;
    animation: fade-up 0.25s var(--ease) both;
  }

  @keyframes fade-out {
    from { opacity: 1; }
    to { opacity: 0; }
  }

  @keyframes draw-on {
    to { stroke-dashoffset: 0; }
  }

  .success-modal {
    z-index: 61;
    background: var(--overlay);
    animation: fade-in 0.25s var(--ease) both, fade-out 0.6s var(--ease) 1.9s both;
  }

  .success-card {
    background: var(--bg-elevated);
    border: 1px solid var(--border);
    border-radius: 18px;
    box-shadow: var(--modal-shadow);
    padding: 26px 34px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    animation: scale-in 0.3s var(--ease) both;
  }

  .success-check {
    width: 64px;
    height: 64px;
  }

  .success-check circle,
  .success-check path {
    fill: none;
    stroke: var(--good);
    stroke-width: 3;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  .success-check circle {
    stroke-dasharray: 151;
    stroke-dashoffset: 151;
    animation: draw-on 0.55s var(--ease) 0.1s forwards;
  }

  .success-check path {
    stroke-dasharray: 40;
    stroke-dashoffset: 40;
    animation: draw-on 0.35s var(--ease) 0.55s forwards;
  }

  .modal-card.update {
    width: min(560px, 94vw);
    max-height: min(720px, 88vh);
    display: flex;
    flex-direction: column;
    overflow: auto;
  }

  .modal-card.update.settings-modal {
    width: min(480px, 92vw);
    max-height: min(680px, 90vh);
  }

  .modal-card.update.settings-modal.bulk-modal {
    width: min(660px, 94vw);
    max-height: min(760px, 92vh);
  }

  .settings-block {
    margin-bottom: 18px;
    padding-bottom: 14px;
    border-bottom: 1px solid var(--border);
    animation: fade-up 0.3s var(--ease) both;
  }

  .settings-block:nth-of-type(2) {
    animation-delay: 0.06s;
  }

  .settings-block:last-of-type {
    border-bottom: none;
    margin-bottom: 8px;
    padding-bottom: 0;
  }

  .settings-block .panel-label {
    margin-bottom: 12px;
  }

  .settings-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 12px;
    font-size: 13.5px;
    color: var(--text-secondary);
  }

  .settings-check {
    margin-bottom: 12px;
  }

  .settings-actions {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .settings-actions .btn.primary {
    width: 100%;
  }

  .modal-card h3 {
    color: var(--text);
    font-size: 17px;
    font-weight: 700;
    letter-spacing: -0.02em;
    margin-bottom: 12px;
  }

  .modal-card ol {
    padding-left: 18px;
    font-size: 13.5px;
    line-height: 1.55;
    margin-bottom: 18px;
    color: var(--text-secondary);
  }

  .modal-card .btn.primary {
    width: 100%;
  }

  .modal-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 4px;
  }

  .modal-head h3 {
    margin-bottom: 0;
  }

  .modal-x {
    width: 32px;
    height: 32px;
    border: none;
    border-radius: 8px;
    background: transparent;
    color: var(--muted);
    cursor: pointer;
    font-size: 14px;
    flex-shrink: 0;
  }

  .modal-x:hover {
    background: rgba(251, 113, 133, 0.15);
    color: var(--danger);
  }

  .update-msg {
    font-size: 14px;
    line-height: 1.5;
    margin-bottom: 12px;
    color: var(--text-secondary);
  }

  .update-msg.wait {
    color: var(--info);
    font-weight: 600;
  }

  .drive-status {
    animation: fade-in 0.3s var(--ease) both;
  }

  .tut-toggle {
    margin-bottom: 10px;
  }

  .notes {
    flex: 1;
    min-height: 180px;
    max-height: 360px;
    overflow: auto;
    word-break: break-word;
    background: var(--bg-input);
    border: 1px solid var(--border);
    border-radius: 12px;
    padding: 14px 16px;
    font-size: 13px;
    line-height: 1.5;
    color: var(--text-secondary);
    margin-bottom: 14px;
    font-family: inherit;
    user-select: text;
    -webkit-user-select: text;
  }

  .notes.md :global(h1),
  .notes.md :global(h2),
  .notes.md :global(h3),
  .notes.md :global(h4) {
    color: var(--text);
    font-weight: 700;
    line-height: 1.3;
    margin: 0.85em 0 0.4em;
  }

  .notes.md :global(h1) { font-size: 1.15em; }
  .notes.md :global(h2) { font-size: 1.08em; }
  .notes.md :global(h3),
  .notes.md :global(h4) { font-size: 1em; color: var(--text); }

  .notes.md :global(h1:first-child),
  .notes.md :global(h2:first-child),
  .notes.md :global(h3:first-child),
  .notes.md :global(p:first-child) {
    margin-top: 0;
  }

  .notes.md :global(p) { margin: 0.45em 0; }

  .notes.md :global(ul),
  .notes.md :global(ol) {
    margin: 0.4em 0 0.6em;
    padding-left: 1.35em;
  }

  .notes.md :global(li) { margin: 0.2em 0; }

  .notes.md :global(a) {
    color: var(--text);
    text-decoration: underline;
    text-underline-offset: 2px;
  }

  .notes.md :global(a:hover) { color: var(--text-secondary); }

  .notes.md :global(strong),
  .notes.md :global(b) {
    color: var(--text);
    font-weight: 700;
  }

  .notes.md :global(code) {
    background: var(--md-code-bg);
    padding: 1px 6px;
    border-radius: 6px;
    color: var(--text);
    font-size: 12px;
    font-family: ui-monospace, Consolas, monospace;
  }

  .notes.md :global(pre) {
    background: var(--md-pre-bg);
    border-radius: 10px;
    padding: 10px 12px;
    overflow: auto;
    margin: 0.6em 0;
    border: 1px solid var(--border);
  }

  .notes.md :global(pre code) {
    background: none;
    padding: 0;
    color: var(--text);
  }

  .notes.md :global(hr) {
    border: none;
    border-top: 1px solid var(--border);
    margin: 0.9em 0;
  }

  .notes.md :global(blockquote) {
    margin: 0.5em 0;
    padding: 6px 12px;
    border-left: 3px solid var(--md-quote-border);
    color: var(--muted);
    background: var(--md-quote-bg);
    border-radius: 0 8px 8px 0;
  }

  .notes.md :global(table) {
    width: 100%;
    border-collapse: collapse;
    margin: 0.6em 0;
    font-size: 12.5px;
  }

  .notes.md :global(th),
  .notes.md :global(td) {
    border: 1px solid var(--border);
    padding: 6px 8px;
    text-align: left;
  }

  .notes.md :global(th) {
    color: var(--text);
    background: var(--md-th-bg);
  }

  .modal-actions {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .guide {
    font-size: 13px;
    color: var(--muted);
    margin: -6px 0 14px;
  }

  .link {
    border: none;
    background: none;
    color: var(--text);
    cursor: pointer;
    font: inherit;
    text-decoration: underline;
    text-underline-offset: 2px;
    padding: 0;
    transition: opacity 0.2s var(--ease);
  }

  .link:hover {
    opacity: 0.75;
  }

  .statusbar span {
    animation: fade-in 0.25s var(--ease) both;
  }

  .statusbar span.status-dot {
    animation: none;
  }

  .status-text {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: calc(100% - 170px);
    line-height: 1.35;
  }

  .status-dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--muted);
    margin-right: 8px;
    flex-shrink: 0;
    align-self: center;
    position: relative;
    top: -0.5px;
    transition: background 0.3s var(--ease), box-shadow 0.3s var(--ease);
  }

  .statusbar.ok .status-dot {
    background: var(--good);
    box-shadow: 0 0 6px var(--good);
  }

  .statusbar.err .status-dot {
    background: var(--danger);
    box-shadow: 0 0 6px var(--danger);
  }

  .statusbar .credit {
    margin-left: auto;
    font-size: 11px;
    color: var(--muted);
    letter-spacing: 0.02em;
    white-space: nowrap;
    flex-shrink: 0;
  }

  .credit-brand {
    color: var(--text-secondary);
    font-weight: 700;
    letter-spacing: 0.08em;
  }

  @media (max-width: 820px) {
    .content {
      grid-template-columns: 1fr;
    }
  }
</style>
