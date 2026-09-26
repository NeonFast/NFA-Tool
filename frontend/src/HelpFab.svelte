<script lang="ts">
  import { t } from './i18n';
  import { aimHint } from './actions';
  import { app, openHelp } from './state.svelte.js';
</script>

<div class="help-wrap" use:aimHint>
  {#if app.showHelpHint}
    <button class="help-hint" type="button" onclick={() => (app.showHintConfirm = true)} aria-label={t(app.lang, 'showInstructions')}>
      <svg class="hh-arrow" viewBox="0 0 56 40" aria-hidden="true">
        <path d="M10 35 C 22 34, 34 26, 40 14" />
        <path d="M32 14 L 40 5 L 48 14" />
      </svg>
      <span class="hh-text">{t(app.lang, 'helpHint')}</span>
    </button>
  {/if}
  <button
    class="help-fab"
    type="button"
    title={t(app.lang, 'showInstructions')}
    aria-label={t(app.lang, 'showInstructions')}
    onclick={openHelp}
  >
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
      <path d="M10.5 8.67709C10.8665 8.26188 11.4027 8 12 8C13.1046 8 14 8.89543 14 10C14 10.9337 13.3601 11.718 12.4949 11.9383C12.2273 12.0064 12 12.2239 12 12.5V12.5V13" />
      <path d="M12 16H12.01" />
    </svg>
  </button>
</div>
