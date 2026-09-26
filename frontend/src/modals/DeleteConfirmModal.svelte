<script lang="ts">
  import { t } from '../i18n';
  import { app, confirmDelete, cancelDelete } from '../state.svelte.js';

  const many = $derived((app.deleteConfirm?.length ?? 0) > 1);
</script>

<div class="modal" role="dialog" aria-modal="true" tabindex="-1">
  <div class="modal-card">
    <h3>{t(app.lang, many ? 'delConfirmTitleMany' : 'delConfirmTitleOne')}</h3>
    <p class="update-msg">
      {many
        ? t(app.lang, 'delConfirmTextMany', { n: app.deleteConfirm?.length ?? 0 })
        : t(app.lang, 'delConfirmTextOne', { name: app.deleteConfirm?.[0] ?? '' })}
    </p>
    <div class="modal-actions">
      <button class="btn ghost danger-sm" type="button" onclick={() => void confirmDelete()}>
        {t(app.lang, 'delete')}
      </button>
      <button class="btn ghost block" type="button" onclick={cancelDelete}>{t(app.lang, 'cancel')}</button>
    </div>
  </div>
</div>
