<script lang="ts">
  import { t } from './i18n';
  import { app } from './state.svelte.js';
</script>

<footer class="statusbar" class:ok={app.statusKind === 'ok'} class:err={app.statusKind === 'err'}>
  <span class="status-dot" aria-hidden="true"></span>
  {#key app.status}
    <span class="status-text">{app.status}</span>
  {/key}
  {#if app.avatarSync}
    <span class="avatar-sync" aria-hidden="true">
      {t(app.lang, 'avatarSync', { done: app.avatarSync.done, total: app.avatarSync.total })}
    </span>
  {/if}
  <span class="credit">
    {t(app.lang, 'creditPrefix')}
    <span class="credit-brand">HACKERSHOP</span>
  </span>
</footer>
