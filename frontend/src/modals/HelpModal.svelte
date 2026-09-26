<script lang="ts">
  import { Browser } from '@wailsio/runtime';
  import { t } from '../i18n';
  import { app, GUIDE_URL } from '../state.svelte.js';
</script>

<div class="modal" role="dialog" aria-modal="true" tabindex="-1">
  <div class="modal-card">
    <h3>{t(app.lang, 'instructions')}</h3>
    <ol>
      <li>{t(app.lang, 'help1')}</li>
      <li>{t(app.lang, 'help2')}</li>
      <li>{t(app.lang, 'help3')}</li>
      <li>{t(app.lang, 'help4')}</li>
      <li>{t(app.lang, 'help5')}</li>
      <li>{t(app.lang, 'help6')}</li>
    </ol>
    <p class="guide">
      {t(app.lang, 'fullGuide')}
      <button class="link" type="button" onclick={() => Browser.OpenURL(GUIDE_URL)}>
        teletype.in/@hackerdlc/CS2NFA
      </button>
    </p>
    <button class="btn primary" type="button" onclick={() => (app.showHelp = false)}>{t(app.lang, 'gotIt')}</button>
  </div>
</div>
