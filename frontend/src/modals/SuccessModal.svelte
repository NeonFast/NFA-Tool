<script lang="ts">
  import { t } from '../i18n';
  import { app } from '../state.svelte.js';
</script>

<div class="modal success-modal">
  <div class="success-card" role="status">
    <svg class="success-check" viewBox="0 0 52 52" aria-hidden="true">
      <circle cx="26" cy="26" r="24" />
      <path d="M15 27l7 7 15-15" />
    </svg>
    <div class="busy-title">{t(app.lang, 'successTitle')}</div>
  </div>
</div>
