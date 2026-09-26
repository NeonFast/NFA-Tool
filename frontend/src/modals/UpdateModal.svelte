<script lang="ts">
  import { t } from '../i18n';
  import {
    app,
    releaseNotesHtml,
    installUpdate,
    openReleasePage,
  } from '../state.svelte.js';
</script>

<div class="modal" role="dialog" aria-modal="true" tabindex="-1">
  <div class="modal-card update">
    <h3>{t(app.lang, 'updateTitle')}</h3>
    <p class="update-msg">
      {t(app.lang, 'updateAvailable', {
        current: app.updateInfo!.currentVersion,
        latest: app.updateInfo!.latestVersion,
      })}
    </p>
    {#if releaseNotesHtml()}
      <div class="notes md">{@html releaseNotesHtml()}</div>
    {/if}
    <div class="modal-actions">
      <button class="btn primary" type="button" disabled={app.updateBusy} onclick={installUpdate}>
        {app.updateBusy ? t(app.lang, 'updateInstalling') : t(app.lang, 'updateNow')}
      </button>
      <button class="btn ghost block" type="button" disabled={app.updateBusy} onclick={openReleasePage}>
        {t(app.lang, 'updateOpenPage')}
      </button>
      <button class="btn ghost block" type="button" disabled={app.updateBusy} onclick={() => (app.showUpdate = false)}>
        {t(app.lang, 'updateLater')}
      </button>
    </div>
  </div>
</div>
