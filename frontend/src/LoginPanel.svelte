<script lang="ts">
  import { t } from './i18n';
  import HelpFab from './HelpFab.svelte';
  import {
    app,
    onKey,
    doLogin,
    importFromField,
    importFromFile,
  } from './state.svelte.js';
</script>

<section class="panel login-panel">
  <div class="panel-label">{t(app.lang, 'accountManagement')}</div>
  <label class="field">
    <span class="field-label">{t(app.lang, 'accountKeyPlaceholder')}</span>
    <input
      type="text"
      spellcheck="false"
      autocomplete="off"
      placeholder="login----token"
      bind:value={app.accountKey}
      onkeydown={onKey}
    />
  </label>
  {#if app.mode === 'advanced'}
    <label class="check">
      <input type="checkbox" bind:checked={app.keepExisting} />
      <span class="box"></span>
      <span>{t(app.lang, 'keepExisting')}</span>
    </label>
    <div class="action-row">
      <button class="btn primary" type="button" disabled={app.loading} onclick={doLogin}>
        {app.loading ? t(app.lang, 'working') : t(app.lang, 'login')}
      </button>
      <button class="btn ghost block" type="button" disabled={app.loading} onclick={importFromField}>
        {t(app.lang, 'importBtn')}
      </button>
    </div>
    <div class="import-file-row">
      <button class="btn ghost block" type="button" disabled={app.loading} onclick={importFromFile}>
        {t(app.lang, 'importFile')}
      </button>
      <HelpFab />
    </div>
  {:else}
    <label class="check">
      <input type="checkbox" bind:checked={app.keepExisting} />
      <span class="box"></span>
      <span>{t(app.lang, 'keepExisting')}</span>
    </label>
    <div class="import-file-row">
      <button class="btn primary" type="button" disabled={app.loading} onclick={doLogin}>
        {app.loading ? t(app.lang, 'working') : t(app.lang, 'login')}
      </button>
      <HelpFab />
    </div>
  {/if}
  {#key app.accounts.length === 0}
    <p class="hint">
      {app.accounts.length === 0 ? t(app.lang, 'hintEmpty') : t(app.lang, 'hintHasAccounts')}
    </p>
  {/key}
</section>
