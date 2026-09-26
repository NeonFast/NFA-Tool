<script lang="ts">
  import { t } from '../i18n';
  import { pillTransition } from '../actions';
  import {
    app,
    closeSettings,
    cancelDriveAuth,
    setLang,
    setTheme,
    harvestSteam,
    checkUpdates,
    resetSteam,
    dumpDiagnostics,
    saveProxies,
    openDriveGuide,
    saveDriveCreds,
    importDriveCreds,
    connectDrive,
    disconnectDrive,
  } from '../state.svelte.js';
</script>

<div class="modal" role="dialog" aria-modal="true" tabindex="-1">
  <div class="modal-card update settings-modal">
    <div class="modal-head">
      <h3>{t(app.lang, 'settings')}</h3>
      <button class="modal-x" type="button" onclick={closeSettings} aria-label={t(app.lang, 'close')}>✕</button>
    </div>

    {#if app.driveAuthWait}
      <p class="update-msg wait">{t(app.lang, 'driveWaiting')}</p>
      <div class="modal-actions">
        <button class="btn primary danger-btn" type="button" onclick={cancelDriveAuth}>{t(app.lang, 'cancel')}</button>
        <button class="btn ghost block" type="button" onclick={closeSettings}>{t(app.lang, 'close')}</button>
      </div>
    {:else}
      <section class="settings-block">
        <div class="panel-label">{t(app.lang, 'settingsApp')}</div>
        <div class="settings-row">
          <span>{t(app.lang, 'settingsLang')}</span>
          <div class="lang-switch compact" title={t(app.lang, 'lang')}>
            <button type="button" class="lang-btn" class:active={app.lang === 'ru'} onclick={() => setLang('ru')}><span class="pill" aria-hidden="true"></span><span class="lbl">RU</span></button>
            <button type="button" class="lang-btn" class:active={app.lang === 'en'} onclick={() => setLang('en')}><span class="pill" aria-hidden="true"></span><span class="lbl">EN</span></button>
          </div>
        </div>
        <div class="settings-row">
          <span>{t(app.lang, 'settingsTheme')}</span>
          <div class="lang-switch compact">
            <button type="button" class="lang-btn" class:active={app.theme === 'auto'} onclick={(e) => pillTransition(e.currentTarget, () => setTheme('auto'))}><span class="pill" aria-hidden="true"></span><span class="lbl">{t(app.lang, 'themeAuto')}</span></button>
            <button type="button" class="lang-btn" class:active={app.theme === 'dark'} onclick={(e) => pillTransition(e.currentTarget, () => setTheme('dark'))}><span class="pill" aria-hidden="true"></span><span class="lbl">{t(app.lang, 'themeDark')}</span></button>
            <button type="button" class="lang-btn" class:active={app.theme === 'light'} onclick={(e) => pillTransition(e.currentTarget, () => setTheme('light'))}><span class="pill" aria-hidden="true"></span><span class="lbl">{t(app.lang, 'themeLight')}</span></button>
          </div>
        </div>
        <label class="check settings-check">
          <input type="checkbox" bind:checked={app.keepExisting} />
          <span class="box"></span>
          <span>{t(app.lang, 'keepExisting')}</span>
        </label>
        <div class="settings-actions">
          <button class="btn ghost block" type="button" onclick={harvestSteam}>{t(app.lang, 'harvestBtn')}</button>
          <button class="btn ghost block" type="button" onclick={() => checkUpdates(false)}>{t(app.lang, 'checkUpdate')}</button>
          <button class="btn ghost block" type="button" onclick={resetSteam}>{t(app.lang, 'resetSteam')}</button>
          <button class="btn ghost block" type="button" onclick={() => void dumpDiagnostics()}>{t(app.lang, 'diagDumpBtn')}</button>
        </div>
      </section>

      <section class="settings-block">
        <div class="panel-label">{t(app.lang, 'settingsProxy')}</div>
        <label class="field">
          <textarea
            class="bulk-ta sm"
            spellcheck="false"
            autocomplete="off"
            placeholder="http://user:pass@host:port"
            bind:value={app.settingsProxies}
            oninput={saveProxies}
          ></textarea>
        </label>
        <p class="proxy-hint">{t(app.lang, 'proxyHint')}</p>
      </section>

      <section class="settings-block">
        <div class="panel-label">{t(app.lang, 'settingsDrive')}</div>
        {#key app.driveStatus.connected}
          <p class="update-msg drive-status">
            {app.driveStatus.connected
              ? `${t(app.lang, 'driveConnected')}${app.driveStatus.clientIdHint ? ' · ' + app.driveStatus.clientIdHint : ''}`
              : app.driveStatus.hasCredentials
                ? t(app.lang, 'driveNotConnected')
                : t(app.lang, 'driveSetupHint')}
          </p>
        {/key}
        <button class="btn ghost block tut-toggle" type="button" onclick={openDriveGuide}>
          {t(app.lang, 'driveOpenGuide')}
        </button>
        {#if !app.driveStatus.hasCredentials}
          <label class="field">
            <input type="text" spellcheck="false" autocomplete="off" placeholder={t(app.lang, 'driveClientId')} bind:value={app.driveClientId} />
          </label>
          <label class="field" style="margin-top:8px">
            <input type="text" spellcheck="false" autocomplete="off" placeholder={t(app.lang, 'driveClientSecret')} bind:value={app.driveClientSecret} />
          </label>
          <div class="settings-actions" style="margin-top:10px">
            <button class="btn primary" type="button" disabled={app.driveBusy || !app.driveClientId.trim()} onclick={saveDriveCreds}>
              {t(app.lang, 'driveSave')}
            </button>
            <button class="btn ghost block" type="button" disabled={app.driveBusy} onclick={importDriveCreds}>
              {t(app.lang, 'driveImport')}
            </button>
          </div>
        {:else}
          <div class="settings-actions">
            {#if !app.driveStatus.connected}
              <button class="btn primary" type="button" disabled={app.driveBusy} onclick={connectDrive}>
                {t(app.lang, 'driveConnect')}
              </button>
            {:else}
              <button class="btn ghost block" type="button" disabled={app.driveBusy} onclick={disconnectDrive}>
                {t(app.lang, 'driveDisconnect')}
              </button>
            {/if}
            <button class="btn ghost block" type="button" disabled={app.driveBusy} onclick={importDriveCreds}>
              {t(app.lang, 'driveImport')}
            </button>
          </div>
        {/if}
        </section>

      <section class="settings-block">
        <div class="panel-label">{t(app.lang, 'settingsManagement')}</div>
        <label class="check settings-check">
          <input type="checkbox" bind:checked={app.management} />
          <span class="box"></span>
          <span>{t(app.lang, 'managementMode')}</span>
        </label>
      </section>
    {/if}
  </div>
</div>
