<script lang="ts">
  import { t } from '../i18n';
  import { app, confirmHideHint } from '../state.svelte.js';
</script>

<div class="modal" role="dialog" aria-modal="true" tabindex="-1">
  <div class="modal-card">
    <h3>{t(app.lang, 'hintConfirmTitle')}</h3>
    <p class="update-msg">{t(app.lang, 'hintConfirmText')}</p>
    <div class="modal-actions">
      <button class="btn primary" type="button" onclick={confirmHideHint}>{t(app.lang, 'hintConfirmYes')}</button>
      <button class="btn ghost block" type="button" onclick={() => (app.showHintConfirm = false)}>{t(app.lang, 'hintConfirmNo')}</button>
    </div>
  </div>
</div>
