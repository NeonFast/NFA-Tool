<script lang="ts">
  import { t } from './i18n';
  import {
    app,
    refreshSysStatus,
    checkUpdates,
    runKeyCheck,
    openBulkView,
  } from './state.svelte.js';
</script>

<section class="panel sys-panel">
  <div class="panel-label">{t(app.lang, 'sysPanel')}</div>
  <p class="alpha-warn">{t(app.lang, 'alphaWarn')}</p>
  {#if app.sysStatus}
    <div class="accinfo-grid">
      <span class="ai-label">{t(app.lang, 'sysVersion')}</span>
      <span class="ai-value">v{app.sysStatus.version}</span>
      <span class="ai-label">Steam</span>
      <span class="ai-value" class:good={app.sysStatus.steamRunning} class:bad={!app.sysStatus.steamRunning}>
        {app.sysStatus.steamRunning ? t(app.lang, 'sysRunning') : t(app.lang, 'sysStopped')}
      </span>
      <span class="ai-label">{t(app.lang, 'sysPath')}</span>
      <span class="ai-value path" title={app.sysStatus.steamPath}>{app.sysStatus.steamPath || '—'}</span>
      <span class="ai-label">{t(app.lang, 'sysAccounts')}</span>
      <span class="ai-value">{app.sysStatus.accountsValid} / {app.sysStatus.accountsTotal}</span>
      <span class="ai-label">Google Drive</span>
      <span class="ai-value" class:good={app.sysStatus.driveConnected} class:bad={!app.sysStatus.driveConnected}>
        {app.sysStatus.driveConnected ? t(app.lang, 'driveConnected') : t(app.lang, 'driveNotConnected')}
      </span>
      <span class="ai-label">Steam API</span>
      <span class="ai-value" class:good={app.sysStatus.steamApiOnline} class:bad={!app.sysStatus.steamApiOnline}>
        {app.sysStatus.steamApiOnline ? t(app.lang, 'apiOk') : t(app.lang, 'apiDown')}
      </span>
    </div>
  {:else}
    <p class="update-msg wait">{t(app.lang, 'accInfoLoading')}</p>
  {/if}
  <button class="btn ghost block" type="button" disabled={app.sysBusy} onclick={refreshSysStatus}>
    {t(app.lang, 'logRefresh')}
  </button>
  <button class="btn ghost block" type="button" onclick={() => checkUpdates(false)}>
    {t(app.lang, 'checkUpdate')}
  </button>
  <div class="panel-label keycheck-label">{t(app.lang, 'keyCheckLabel')}</div>
  <label class="field">
    <input
      type="text"
      spellcheck="false"
      autocomplete="off"
      placeholder="login----token"
      bind:value={app.checkKey}
      onkeydown={(e) => e.key === 'Enter' && runKeyCheck()}
    />
  </label>
  <button class="btn primary" type="button" disabled={app.infoBusy || !app.checkKey.trim()} onclick={runKeyCheck}>
    {app.infoBusy ? t(app.lang, 'working') : t(app.lang, 'keyCheckBtn')}
  </button>
  <button class="btn ghost block" type="button" onclick={openBulkView}>
    {t(app.lang, 'bulkCheckBtn')}
  </button>
</section>
