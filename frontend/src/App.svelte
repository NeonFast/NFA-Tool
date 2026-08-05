<script lang="ts">
  import { onMount } from 'svelte';
  import { Browser } from '@wailsio/runtime';
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

  async function notify(ok: boolean, message: string) {
    const title = ok ? t(lang, 'successTitle') : t(lang, 'errorTitle');
    const text = translateBackendMessage(lang, message);
    try {
      await AppService.Notify(ok, title, text);
    } catch {
      // fallback if bindings lag behind
      window.alert(`${title}\n\n${text}`);
    }
  }

  const GUIDE_URL = 'https://teletype.in/@hackerdlc/CS2NFA';

  type Account = {
    name: string;
    expiresIn: string;
    valid: boolean;
  };

  type Result = {
    ok: boolean;
    message: string;
  };

  let appName = $state('NFA Tool Recode v2');
  let version = $state('2.0.3');
  let lang = $state<Lang>(loadLang());
  let accountKey = $state('');
  let keepExisting = $state(false);
  let accounts = $state<Account[]>([]);
  let selected = $state<Record<string, boolean>>({});
  let dragSelect = $state(false);
  let dragMode = $state(true); // true = check, false = uncheck
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
    try {
      const anySvc = AppService as typeof AppService & { GetAppName?: () => Promise<string> };
      if (typeof anySvc.GetAppName === 'function') {
        appName = await anySvc.GetAppName();
      }
      version = await AppService.GetVersion();
    } catch {
      /* ignore */
    }
    await refreshAccounts();
    void checkUpdates(true);
    return () => {
      window.removeEventListener('pointerup', endDragSelect);
      window.removeEventListener('pointercancel', endDragSelect);
      window.removeEventListener('blur', endDragSelect);
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
      // app should quit shortly
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
    // only primary button; ignore if started on action buttons
    if (e.button !== 0) return;
    e.preventDefault();
    e.stopPropagation();
    // mode = opposite of current cell (paint ON if was off, OFF if was on)
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
      /* ignore */
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
      /* ignore */
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
    status = alreadyTranslated ? msg : translateBackendMessage(lang, msg);
    statusKind = kind;
  }

  function singleLineKey(): string {
    // only first line — multi-import is file-only
    return accountKey.split(/\r?\n/)[0]?.trim() ?? '';
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
    try {
      const res = (await AppService.LoginFromKey(key, keepExisting)) as Result;
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
    try {
      const res = (await AppService.LoginSaved(name, keepExisting)) as Result;
      setStatus(res.message, res.ok ? 'ok' : 'err');
      await notify(res.ok, res.message);
      await refreshAccounts();
    } catch (e) {
      const msg = String(e);
      setStatus(msg, 'err');
      await notify(false, msg);
    }
  }

  async function deleteAccount(name: string) {
    try {
      const res = (await AppService.DeleteAccount(name)) as Result;
      setStatus(res.message, res.ok ? 'ok' : 'err');
      await notify(res.ok, res.message);
      await refreshAccounts();
    } catch (e) {
      const msg = String(e);
      setStatus(msg, 'err');
      await notify(false, msg);
    }
  }

  async function resetSteam() {
    setStatus(t(lang, 'resettingSteam'), '', true);
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
    }
  }

  function onKey(e: KeyboardEvent) {
    if (e.key === 'Enter') doLogin();
  }
</script>

<div class="app">
  <header class="titlebar">
    <div class="brand">
      <span class="logo-mark" aria-hidden="true">N</span>
      <div class="brand-text">
        <span class="brand-name">{appName}</span>
        <span class="brand-ver">v{version}</span>
      </div>
    </div>
    <div class="title-actions">
      <button class="btn ghost" type="button" onclick={openSettings}>{t(lang, 'settings')}</button>
      <button class="btn ghost" type="button" onclick={() => checkUpdates(false)}>{t(lang, 'checkUpdate')}</button>
      <button class="btn ghost" type="button" onclick={() => (showHelp = true)}>{t(lang, 'showInstructions')}</button>
      <div class="win-btns">
        <button class="win" type="button" onclick={() => AppService.WindowMinimise()} aria-label="Minimise">─</button>
        <button
          class="win close"
          type="button"
          onclick={async () => {
            try {
              await (AppService as any).CancelGoogleAuth?.();
            } catch {
              /* ignore */
            }
            await AppService.WindowClose();
          }}
          aria-label="Close"
        >✕</button>
      </div>
    </div>
  </header>

  <main class="content">
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
      <button class="btn ghost block" type="button" disabled={loading} onclick={importFromFile}>
        {t(lang, 'importFile')}
      </button>
      <p class="hint">
        {accounts.length === 0 ? t(lang, 'hintEmpty') : t(lang, 'hintHasAccounts')}
      </p>
    </section>

    <section class="panel accounts-panel">
      <div class="accounts-head">
        <div>
          <div class="panel-label">{t(lang, 'savedAccounts')}</div>
          {#if accounts.length > 0}
            <div class="count-chip">{accounts.length}</div>
          {/if}
        </div>
        {#if accounts.length > 0}
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
              {#if selectedNames.length > 0}({selectedNames.length}){/if}
            </button>
            <button class="btn ghost sm" type="button" disabled={exportBusy} onclick={() => exportAccounts(true)}>
              {t(lang, 'exportAll')}
            </button>
          </div>
        {/if}
      </div>
      {#if accounts.length === 0}
        <div class="empty">
          <div class="empty-icon" aria-hidden="true">∅</div>
          <p>{t(lang, 'noSavedAccounts')}</p>
        </div>
      {:else}
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div class="accounts-list" class:painting={dragSelect}>
          {#each accounts as acc (acc.name)}
            <!-- svelte-ignore a11y_no_static_element_interactions -->
            <div
              class="account-row"
              class:picked={!!selected[acc.name]}
              data-acc={acc.name}
              onpointerenter={() => paintSelect(acc.name)}
              onpointerdown={(e) => {
                const t = e.target as HTMLElement;
                if (t.closest('.row-actions')) return;
                startDragSelect(e, acc.name);
              }}
            >
              <label class="pick">
                <!-- selection is driven only by row pointer handlers (avoids double-toggle) -->
                <input type="checkbox" checked={!!selected[acc.name]} tabindex="-1" />
                <span class="box"></span>
              </label>
              <div class="meta">
                <div class="name">{acc.name}</div>
                <div class="exp" class:ok={acc.valid} class:bad={!acc.valid}>
                  {localizeExpiry(lang, acc.expiresIn)}
                </div>
              </div>
              <!-- svelte-ignore a11y_no_static_element_interactions -->
              <div class="row-actions" onpointerdown={(e) => e.stopPropagation()}>
                <button
                  class="btn mini export"
                  type="button"
                  disabled={exportBusy}
                  onclick={() => exportOne(acc.name)}
                >{t(lang, 'export')}</button>
                <button class="btn mini login" type="button" onclick={() => loginSaved(acc.name)}>{t(lang, 'login')}</button>
                <button class="btn mini del" type="button" onclick={() => deleteAccount(acc.name)}>{t(lang, 'delete')}</button>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </section>
  </main>

  <footer class="statusbar" class:ok={statusKind === 'ok'} class:err={statusKind === 'err'}>
    <span>{status}</span>
  </footer>
</div>

{#if showExportPick}
  <!-- svelte-ignore a11y_click_events_have_key_events a11y_interactive_supports_focus -->
  <div class="modal" role="dialog" aria-modal="true" tabindex="-1" onclick={(e) => e.currentTarget === e.target && closeExportPick()}>
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

{#if showHelp}
  <!-- svelte-ignore a11y_click_events_have_key_events a11y_interactive_supports_focus -->
  <div class="modal" role="dialog" aria-modal="true" tabindex="-1" onclick={(e) => e.currentTarget === e.target && (showHelp = false)}>
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
  <!-- svelte-ignore a11y_click_events_have_key_events a11y_interactive_supports_focus -->
  <div class="modal" role="dialog" aria-modal="true" tabindex="-1" onclick={(e) => e.currentTarget === e.target && closeSettings()}>
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
            <div class="lang-switch" title={t(lang, 'lang')}>
              <button type="button" class="lang-btn" class:active={lang === 'ru'} onclick={() => setLang('ru')}>RU</button>
              <button type="button" class="lang-btn" class:active={lang === 'en'} onclick={() => setLang('en')}>EN</button>
            </div>
          </div>
          <label class="check settings-check">
            <input type="checkbox" bind:checked={keepExisting} />
            <span class="box"></span>
            <span>{t(lang, 'keepExisting')}</span>
          </label>
          <div class="settings-actions">
            <button class="btn ghost block" type="button" onclick={() => checkUpdates(false)}>{t(lang, 'checkUpdate')}</button>
            <button class="btn ghost block" type="button" onclick={resetSteam}>{t(lang, 'resetSteam')}</button>
          </div>
        </section>

        <section class="settings-block">
          <div class="panel-label">{t(lang, 'settingsDrive')}</div>
          <p class="update-msg">
            {driveStatus.connected
              ? `${t(lang, 'driveConnected')}${driveStatus.clientIdHint ? ' · ' + driveStatus.clientIdHint : ''}`
              : driveStatus.hasCredentials
                ? t(lang, 'driveNotConnected')
                : t(lang, 'driveSetupHint')}
          </p>
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
    </div>
  </div>
{/if}

{#if showUpdate && updateInfo}
  <!-- svelte-ignore a11y_click_events_have_key_events a11y_interactive_supports_focus -->
  <div class="modal" role="dialog" aria-modal="true" tabindex="-1" onclick={(e) => !updateBusy && e.currentTarget === e.target && (showUpdate = false)}>
    <div class="modal-card update">
      <h3>{t(lang, 'updateTitle')}</h3>
      <p class="update-msg">
        {t(lang, 'updateAvailable', {
          current: updateInfo.currentVersion,
          latest: updateInfo.latestVersion,
        })}
      </p>
      {#if releaseNotesHtml}
        <!-- eslint-disable-next-line svelte/no-at-html-tags -->
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
    --border: rgba(255, 255, 255, 0.08);
    --border-strong: rgba(255, 255, 255, 0.16);
    --text: #fafafa;
    --text-secondary: #a1a1aa;
    --muted: #71717a;
    --accent: #ffffff;
    --accent-hover: #f4f4f5;
    --accent-fg: #09090b;
    --accent-soft: rgba(255, 255, 255, 0.1);
    --ok: #d4d4d8;
    --danger: #f87171;
    --info: #e4e4e7;
    --radius: 14px;
    --radius-sm: 10px;
    --shadow: 0 1px 0 rgba(255, 255, 255, 0.04) inset, 0 16px 40px rgba(0, 0, 0, 0.4);
    --ease: cubic-bezier(0.22, 1, 0.36, 1);
    --font: "Segoe UI Variable", "Segoe UI", system-ui, -apple-system, sans-serif;
    font-family: var(--font);
    color-scheme: dark;
  }

  :global(*) {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
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

  /* —— Titlebar —— */
  .titlebar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 14px 14px 12px 18px;
    border-bottom: 1px solid var(--border);
    background: rgba(9, 9, 11, 0.72);
    backdrop-filter: blur(12px);
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
    background: #ffffff;
    box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.08), 0 8px 20px rgba(0, 0, 0, 0.35);
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

  .lang-btn {
    border: none;
    background: transparent;
    color: var(--muted);
    padding: 5px 11px;
    font-size: 11px;
    font-weight: 700;
    border-radius: 999px;
    cursor: pointer;
    transition: color 0.2s var(--ease), background 0.2s var(--ease), transform 0.2s var(--ease);
  }

  .lang-btn.active {
    background: #ffffff;
    color: #09090b;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.25);
  }

  .lang-btn:active {
    transform: scale(0.96);
  }

  .win-btns {
    display: flex;
    gap: 4px;
    margin-left: 4px;
  }

  .win {
    width: 36px;
    height: 30px;
    border: none;
    border-radius: 8px;
    background: transparent;
    color: var(--muted);
    cursor: pointer;
    font-size: 12px;
  }

  .win:hover {
    background: rgba(255, 255, 255, 0.06);
    color: var(--text);
  }

  .win.close:hover {
    background: #e11d48;
    color: white;
  }

  /* —— Buttons —— */
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
    background: rgba(255, 255, 255, 0.04);
    border-color: var(--border-strong);
    color: var(--text);
  }

  .btn.ghost.sm {
    padding: 6px 10px;
    font-size: 11.5px;
    border-radius: 8px;
  }

  .btn.ghost.accent {
    color: #ffffff;
    border-color: rgba(255, 255, 255, 0.28);
    background: rgba(255, 255, 255, 0.08);
  }

  .btn.ghost.accent:hover:not(:disabled) {
    border-color: rgba(255, 255, 255, 0.45);
    background: rgba(255, 255, 255, 0.14);
    color: #ffffff;
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
    background: #ffffff;
    color: var(--accent-fg);
    font-weight: 700;
    font-size: 14.5px;
    box-shadow: 0 1px 0 rgba(255, 255, 255, 0.35) inset, 0 8px 22px rgba(0, 0, 0, 0.28);
  }

  .btn.primary:hover:not(:disabled) {
    background: #f4f4f5;
  }

  .btn.primary:active:not(:disabled) {
    background: #e4e4e7;
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
    background: rgba(255, 255, 255, 0.1);
    border-color: rgba(255, 255, 255, 0.28);
    color: #ffffff;
  }

  .btn.mini.login {
    background: rgba(255, 255, 255, 0.1);
    color: #ffffff;
    border-color: transparent;
  }

  .btn.mini.login:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.18);
  }

  .btn.mini.del {
    background: rgba(251, 113, 133, 0.1);
    color: var(--danger);
  }

  .btn.mini.del:hover:not(:disabled) {
    background: rgba(251, 113, 133, 0.2);
  }

  .danger-btn {
    background: #fb7185 !important;
    color: #1c0508 !important;
    box-shadow: none !important;
  }

  /* —— Layout —— */
  .content {
    display: grid;
    grid-template-columns: 340px 1fr;
    gap: 16px;
    padding: 16px 18px;
    min-height: 0;
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
    background: #0c0c0e;
    color: var(--text);
    padding: 0 14px;
    font-size: 13.5px;
    outline: none;
    user-select: text;
    -webkit-user-select: text;
    transition: border-color 0.2s var(--ease), box-shadow 0.2s var(--ease), background 0.2s var(--ease);
  }

  .field input::placeholder {
    color: #52525b;
  }

  .field input:focus {
    border-color: rgba(255, 255, 255, 0.45);
    box-shadow: 0 0 0 3px rgba(255, 255, 255, 0.08);
    background: #101012;
  }

  .action-row {
    display: grid;
    grid-template-columns: 1.4fr 1fr;
    gap: 8px;
  }

  .action-row .btn.primary {
    width: 100%;
  }

  .action-row .btn.ghost.block {
    height: 46px;
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
    background: #0c0c0e;
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
    background: #ffffff;
    border-color: #ffffff;
    box-shadow: 0 0 0 3px rgba(255, 255, 255, 0.08);
  }

  .check input:checked + .box::after,
  .pick input:checked + .box::after {
    content: '';
    width: 13px;
    height: 13px;
    background-color: #09090b;
    /* proper checkmark, not a rotated border stub */
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
  }

  .export-bar {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
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
    grid-template-columns: auto 1fr auto;
    gap: 12px;
    align-items: center;
    background: #121214;
    border: 1px solid var(--border);
    border-radius: 12px;
    padding: 12px 14px;
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
    background: #17171a;
    transform: translateY(-1px);
    box-shadow: 0 8px 20px rgba(0, 0, 0, 0.22);
  }

  .account-row.picked {
    border-color: rgba(255, 255, 255, 0.35);
    background: rgba(255, 255, 255, 0.06);
  }

  .pick {
    display: flex;
    align-items: center;
    cursor: pointer;
    pointer-events: none; /* row handles drag-select */
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
  }

  .statusbar {
    min-height: 30px;
    display: flex;
    align-items: center;
    padding: 0 18px;
    font-size: 12px;
    color: var(--muted);
    border-top: 1px solid var(--border);
    background: rgba(9, 9, 11, 0.9);
  }

  .statusbar.ok {
    color: var(--ok);
  }

  .statusbar.err {
    color: var(--danger);
  }

  /* —— Modal —— */
  .modal {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.55);
    backdrop-filter: blur(8px);
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
    box-shadow: 0 24px 64px rgba(0, 0, 0, 0.5);
    animation: scale-in 0.28s var(--ease) both;
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

  .settings-block {
    margin-bottom: 18px;
    padding-bottom: 14px;
    border-bottom: 1px solid var(--border);
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

  .tut-toggle {
    margin-bottom: 10px;
  }

  .notes {
    flex: 1;
    min-height: 180px;
    max-height: 360px;
    overflow: auto;
    word-break: break-word;
    background: #0c0c0e;
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
  .notes.md :global(h4) { font-size: 1em; color: #ffffff; }

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
    color: #ffffff;
    text-decoration: underline;
    text-underline-offset: 2px;
  }

  .notes.md :global(a:hover) { color: #e4e4e7; }

  .notes.md :global(strong),
  .notes.md :global(b) {
    color: var(--text);
    font-weight: 700;
  }

  .notes.md :global(code) {
    background: rgba(255, 255, 255, 0.08);
    padding: 1px 6px;
    border-radius: 6px;
    color: #ffffff;
    font-size: 12px;
    font-family: ui-monospace, Consolas, monospace;
  }

  .notes.md :global(pre) {
    background: rgba(0, 0, 0, 0.35);
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
    border-left: 3px solid rgba(255, 255, 255, 0.35);
    color: var(--muted);
    background: rgba(255, 255, 255, 0.04);
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
    background: rgba(255, 255, 255, 0.03);
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
    color: #ffffff;
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

  @media (max-width: 820px) {
    .content {
      grid-template-columns: 1fr;
    }
  }
</style>
