<script lang="ts">
  import { onMount } from 'svelte';
  import { t, translateBackendMessage } from './i18n';
  import {
    app,
    bulkItems,
    bulkOkCount,
    bulkBadCount,
    bulkPickFile,
    openSettings,
    runBulkCheck,
    bulkStatusText,
    bulkInfo,
    exportBulk,
    exportBulkToDrive,
    closeBulkView,
    type BulkItem,
  } from './state.svelte.js';

  let exportMenu = $state<'ok' | 'bad' | null>(null);
  let menuX = $state(0);
  let menuY = $state(0);
  let ta: HTMLTextAreaElement;

  function toggleMenu(which: 'ok' | 'bad', e: MouseEvent) {
    if (exportMenu === which) {
      exportMenu = null;
      return;
    }
    const r = (e.currentTarget as HTMLElement).getBoundingClientRect();
    menuX = Math.max(8, Math.min(r.right - 190, window.innerWidth - 198));
    menuY = r.bottom + 6;
    exportMenu = which;
  }

  function pickExport(which: 'ok' | 'bad', dest: 'file' | 'drive') {
    exportMenu = null;
    if (dest === 'drive') void exportBulkToDrive(which);
    else void exportBulk(which);
  }

  function bulkAccountText(it: BulkItem): string {
    const a = it.account || '—';
    return it.status === 'invalid' ? translateBackendMessage(app.lang, a) : a;
  }

  onMount(() => {
    ta.focus();
    const onClick = (e: MouseEvent) => {
      if (!(e.target as HTMLElement).closest('.export-menu-wrap, .export-menu')) exportMenu = null;
    };
    window.addEventListener('click', onClick);
    return () => window.removeEventListener('click', onClick);
  });
</script>

<section class="panel bulk-side">
  <div class="bulk-head">
    <button class="icon-btn bulk-back" type="button" onclick={closeBulkView} aria-label={t(app.lang, 'close')}>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M19 12H5" /><path d="M12 19L5 12L12 5" /></svg>
    </button>
    <div class="panel-label">{t(app.lang, 'bulkTitle')}</div>
  </div>

  <div class="bulk-keys-head">
    <span class="field-label">{t(app.lang, 'bulkKeysLabel')}</span>
    <button class="btn ghost sm" type="button" disabled={app.bulkBusy} onclick={bulkPickFile}>
      {t(app.lang, 'bulkFile')}
    </button>
  </div>
  <textarea
    class="bulk-ta"
    spellcheck="false"
    autocomplete="off"
    placeholder="login----token"
    bind:this={ta}
    bind:value={app.bulkKeys}
  ></textarea>
  <div class="bulk-proxy-hint">
    <span>{t(app.lang, 'bulkProxyGo')}</span>
    <button class="btn ghost sm" type="button" onclick={() => void openSettings()}>
      {t(app.lang, 'bulkOpenSettings')}
    </button>
  </div>
  <div class="settings-actions">
    <button class="btn primary" type="button" disabled={app.bulkBusy || !app.bulkKeys.trim()} onclick={() => runBulkCheck(false)}>
      {app.bulkBusy ? t(app.lang, 'working') : t(app.lang, 'bulkStart')}
    </button>
    <button class="btn ghost block" type="button" disabled={app.bulkBusy} onclick={() => runBulkCheck(true)}>
      {t(app.lang, 'bulkSaved')}
    </button>
  </div>
  {#if app.bulkBusy}
    <div class="bulk-progress">
      <div class="bulk-progress-bar">
        <span
          class:indeterminate={app.bulkTotal === 0}
          style="width: {app.bulkTotal > 0 ? Math.round((app.bulkLive.length / app.bulkTotal) * 100) : 100}%"
        ></span>
      </div>
      <div class="bulk-progress-text">
        {app.bulkTotal > 0
          ? t(app.lang, 'bulkProgress', { done: app.bulkLive.length, total: app.bulkTotal })
          : t(app.lang, 'working')}
      </div>
    </div>
  {/if}
</section>

<section class="panel bulk-results-panel">
  <div class="accounts-head">
    <div class="panel-label">
      <span class="lbl-t">{t(app.lang, 'bulkResults')}</span>
      {#if bulkItems().length > 0}
        <span class="count-chip">{bulkItems().length}</span>
      {/if}
    </div>
    {#if !app.bulkBusy && app.bulkResult}
      <div class="bulk-export-row">
        <div class="export-menu-wrap">
          <button
            class="btn ghost sm"
            type="button"
            disabled={bulkOkCount() === 0 || app.driveBusy}
            onclick={(e) => toggleMenu('ok', e)}
          >
            {t(app.lang, 'bulkExportOk')} ({bulkOkCount()})
          </button>
        </div>
        <div class="export-menu-wrap">
          <button
            class="btn ghost sm"
            type="button"
            disabled={bulkBadCount() === 0 || app.driveBusy}
            onclick={(e) => toggleMenu('bad', e)}
          >
            {t(app.lang, 'bulkExportBad')} ({bulkBadCount()})
          </button>
        </div>
      </div>
    {/if}
  </div>
  {#if bulkItems().length > 0}
    <div class="logs-list bulk-results">
      {#each bulkItems() as it, i (it.account + '-' + i)}
        <div
          class="log-row"
          class:ok={it.status === 'ok'}
          class:err={it.status === 'rejected' || it.status === 'expired' || it.status === 'invalid'}
          title={it.detail || it.steamId || ''}
        >
          <span class="log-text bulk-acc">{bulkAccountText(it)}</span>
          <span class="bulk-status">{bulkStatusText(it.status)}</span>
          <span class="bulk-date">
            {#if it.expiresAt}
              <span class="log-ts">{it.expiresAt}</span>
            {/if}
            {#if it.steamId}
              <button class="btn mini info bulk-info" type="button" onclick={() => bulkInfo(it)}>
                {t(app.lang, 'info')}
              </button>
            {/if}
          </span>
        </div>
      {/each}
    </div>
  {:else}
    <div class="empty">
      <p>{t(app.lang, 'logEmpty')}</p>
    </div>
  {/if}
</section>

{#if exportMenu}
  <div class="export-menu" style="left: {menuX}px; top: {menuY}px" role="menu">
    <button type="button" onclick={() => pickExport(exportMenu ?? 'ok', 'file')}>{t(app.lang, 'exportToFile')}</button>
    <button type="button" onclick={() => pickExport(exportMenu ?? 'ok', 'drive')}>{t(app.lang, 'exportDrive')}</button>
  </div>
{/if}
