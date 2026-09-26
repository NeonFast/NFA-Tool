<script lang="ts">
  import { t } from '../i18n';
  import {
    app,
    closeExportPick,
    exportToClipboard,
    exportToFile,
    exportToDrive,
  } from '../state.svelte.js';
</script>

<div class="modal" role="dialog" aria-modal="true" tabindex="-1">
  <div class="modal-card">
    <div class="modal-head">
      <h3>{t(app.lang, 'exportWhere')}</h3>
      <button class="modal-x" type="button" onclick={closeExportPick} aria-label={t(app.lang, 'close')}>✕</button>
    </div>
    <div class="modal-actions">
      <button class="btn primary" type="button" disabled={app.exportBusy || app.driveBusy} onclick={exportToClipboard}>
        {t(app.lang, 'exportToClipboard')}
      </button>
      <button class="btn ghost block" type="button" disabled={app.exportBusy || app.driveBusy} onclick={exportToFile}>
        {t(app.lang, 'exportToFile')}
      </button>
      <button class="btn ghost block" type="button" disabled={app.exportBusy || app.driveBusy} onclick={exportToDrive}>
        {t(app.lang, 'exportDrive')}
      </button>
      <button class="btn ghost block" type="button" disabled={app.exportBusy || app.driveBusy} onclick={closeExportPick}>
        {t(app.lang, 'cancel')}
      </button>
    </div>
  </div>
</div>
