<script lang="ts">
  import { t } from './i18n';
  import { app, clearLogs, fmtTime } from './state.svelte.js';

  let listEl = $state<HTMLDivElement>();
  let firstPaint = true;

  // Autoscroll to the newest entry (bottom), unless the user scrolled up.
  $effect(() => {
    void app.logs.length;
    const el = listEl;
    if (!el) return;
    const nearBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 48;
    if (firstPaint || nearBottom) {
      firstPaint = false;
      el.scrollTop = el.scrollHeight;
    }
  });
</script>

<section class="panel logs-panel">
  <div class="accounts-head">
    <div>
      <div class="panel-label">
        <span class="lbl-t">{t(app.lang, 'logPanel')}</span>
        <span class="beta-chip" title={t(app.lang, 'alphaWarn')}>ALPHA</span>
      </div>
      {#if app.logs.length > 0}
        <div class="count-chip">{app.logs.length}</div>
      {/if}
    </div>
    {#if app.logs.length > 0}
      <button class="btn ghost sm" type="button" onclick={clearLogs}>{t(app.lang, 'logClear')}</button>
    {/if}
  </div>
  {#if app.logs.length === 0}
    <div class="empty">
      <div class="empty-icon" aria-hidden="true">
        <svg viewBox="0 0 24 24" fill="currentColor">
          <path d="M6.99486 7.00636C6.60433 7.39689 6.60433 8.03005 6.99486 8.42058L10.58 12.0057L6.99486 15.5909C6.60433 15.9814 6.60433 16.6146 6.99486 17.0051C7.38538 17.3956 8.01855 17.3956 8.40907 17.0051L11.9942 13.4199L15.5794 17.0051C15.9699 17.3956 16.6031 17.3956 16.9936 17.0051C17.3841 16.6146 17.3841 15.9814 16.9936 15.5909L13.4084 12.0057L16.9936 8.42059C17.3841 8.03007 17.3841 7.3969 16.9936 7.00638C16.603 6.61585 15.9699 6.61585 15.5794 7.00638L11.9942 10.5915L8.40907 7.00636C8.01855 6.61584 7.38538 6.61584 6.99486 7.00636Z" />
        </svg>
      </div>
      <p>{t(app.lang, 'logEmpty')}</p>
    </div>
  {:else}
    <div class="logs-list" bind:this={listEl}>
      {#each app.logs as entry (entry.ts + '-' + entry.kind + '-' + entry.text.length)}
        <div class="log-row {entry.kind}">
          <span class="log-ts">{fmtTime(entry.ts)}</span>
          <span class="log-text">{entry.text}</span>
        </div>
      {/each}
    </div>
  {/if}
</section>
