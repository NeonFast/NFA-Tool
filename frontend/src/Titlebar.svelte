<script lang="ts">
  import { AppService } from '../bindings/nfa-tool';
  import { t } from './i18n';
  import { pillTransition } from './actions';
  import { app, setMode, openSettings, checkUpdates } from './state.svelte.js';
</script>

<header class="titlebar">
  <div class="brand">
    <span class="logo-mark" aria-hidden="true">N</span>
    <div class="brand-text">
      <span class="brand-name">{app.appName}</span>
      {#if app.version}
        <span class="brand-ver">v{app.version}</span>
      {/if}
    </div>
  </div>
  <div class="title-actions">
    <div class="lang-switch">
      <button type="button" class="lang-btn" class:active={app.mode === 'simple'} onclick={(e) => pillTransition(e.currentTarget, () => setMode('simple'))}>
        <span class="pill" aria-hidden="true"></span>
        <svg viewBox="5 1.9 15 20" fill="currentColor" aria-hidden="true"><path fill-rule="evenodd" clip-rule="evenodd" d="M12.3327 3.63004C12.9116 2.74599 14.2858 3.15586 14.2858 4.21256V9.39999H17.7635C18.6601 9.39999 19.1983 10.3956 18.7071 11.1457L12.7695 20.214C12.1817 21.1118 10.7144 20.7524 10.7144 19.6027V14.6H7.18878C6.31275 14.6 5.78696 13.6272 6.26683 12.8943L12.3327 3.63004Z" /></svg>
        <span class="lbl">{t(app.lang, 'modeSimple')}</span>
      </button>
      {#if app.management}
      <button type="button" class="lang-btn" class:active={app.mode === 'advanced'} onclick={(e) => pillTransition(e.currentTarget, () => setMode('advanced'))}>
        <span class="pill" aria-hidden="true"></span>
        <svg viewBox="5.2 5.2 13.6 17.6" fill="currentColor" aria-hidden="true"><path fill-rule="evenodd" clip-rule="evenodd" d="M12 6C8.68629 6 6 8.68629 6 12C6 13.6332 6.65387 15.1157 7.71186 16.1966C7.97971 16.4703 8.1241 16.7217 8.16867 16.9444L8.69776 19.5886C8.97833 20.9908 10.2095 22 11.6395 22H12.3605C13.7905 22 15.0217 20.9908 15.3022 19.5886L15.8313 16.9444C15.8759 16.7217 16.0203 16.4703 16.2881 16.1966C17.3461 15.1157 18 13.6332 18 12C18 8.68629 15.3137 6 12 6ZM11 16C10.4477 16 10 16.4477 10 17C10 17.5523 10.4477 18 11 18H13C13.5523 18 14 17.5523 14 17C14 16.4477 13.5523 16 13 16H11Z" /></svg>
        <span class="lbl">{t(app.lang, 'modeAdvanced')}</span>
      </button>
      <button type="button" class="lang-btn" class:active={app.mode === 'logs'} onclick={(e) => pillTransition(e.currentTarget, () => setMode('logs'))}>
        <span class="pill" aria-hidden="true"></span>
        <svg viewBox="1.2 2.4 21.6 18.2" fill="currentColor" aria-hidden="true"><path fill-rule="evenodd" clip-rule="evenodd" d="M7.29291 14.2929C6.90238 14.6834 6.90238 15.3166 7.29291 15.7071C7.68343 16.0976 8.31659 16.0976 8.70712 15.7071L11.2071 13.2071C11.8738 12.5404 11.8738 11.4596 11.2071 10.7929L8.70712 8.29289C8.3166 7.90237 7.68343 7.90237 7.29291 8.29289C6.90238 8.68342 6.90238 9.31658 7.29291 9.70711L9.5858 12L7.29291 14.2929ZM13 14C12.4477 14 12 14.4477 12 15C12 15.5523 12.4477 16 13 16H16C16.5523 16 17 15.5523 17 15C17 14.4477 16.5523 14 16 14H13ZM22 7.93418C22 7.95604 22 7.97799 22 8.00001L22 16.0658C22.0001 16.9523 22.0001 17.7161 21.9179 18.3278C21.8297 18.9833 21.631 19.6117 21.1213 20.1213C20.6117 20.631 19.9833 20.8297 19.3278 20.9179C18.7161 21.0001 17.9523 21.0001 17.0658 21L6.9342 21C6.0477 21.0001 5.28388 21.0001 4.67222 20.9179C4.0167 20.8297 3.38835 20.631 2.87869 20.1213C2.36902 19.6117 2.17028 18.9833 2.08215 18.3278C1.99991 17.7161 1.99995 16.9523 2 16.0658L2 7.9342C1.99995 7.0477 1.99991 6.28388 2.08215 5.67221C2.17028 5.0167 2.36902 4.38835 2.87869 3.87868C3.38835 3.36902 4.0167 3.17028 4.67222 3.08215C5.28388 2.99991 6.04769 2.99995 6.93418 3L17 3.00001C17.022 3.00001 17.044 3 17.0658 3C17.9523 2.99995 18.7161 2.99991 19.3278 3.08215C19.9833 3.17028 20.6117 3.36902 21.1213 3.87869C21.631 4.38835 21.8297 5.0167 21.9179 5.67221C22.0001 6.28387 22.0001 7.04769 22 7.93418Z" /></svg>
        <span class="lbl">{t(app.lang, 'modeLogs')}</span><span class="beta-chip" title={t(app.lang, 'alphaWarn')}>ALPHA</span>
      </button>
      {/if}
    </div>
    <button
      class="icon-btn"
      type="button"
      title={t(app.lang, 'settings')}
      aria-label={t(app.lang, 'settings')}
      onclick={openSettings}
    >
      <svg viewBox="0 0 72 72" fill="currentColor" aria-hidden="true">
        <path d="M57.531,30.556C58.96,30.813,60,32.057,60,33.509v4.983c0,1.452-1.04,2.696-2.469,2.953l-2.974,0.535c-0.325,1.009-0.737,1.977-1.214,2.907l1.73,2.49c0.829,1.192,0.685,2.807-0.342,3.834l-3.523,3.523c-1.027,1.027-2.642,1.171-3.834,0.342l-2.49-1.731c-0.93,0.477-1.898,0.889-2.906,1.214l-0.535,2.974C41.187,58.96,39.943,60,38.491,60h-4.983c-1.452,0-2.696-1.04-2.953-2.469l-0.535-2.974c-1.009-0.325-1.977-0.736-2.906-1.214l-2.49,1.731c-1.192,0.829-2.807,0.685-3.834-0.342l-3.523-3.523c-1.027-1.027-1.171-2.641-0.342-3.834l1.73-2.49c-0.477-0.93-0.889-1.898-1.214-2.907l-2.974-0.535C13.04,41.187,12,39.943,12,38.491v-4.983c0-1.452,1.04-2.696,2.469-2.953l2.974-0.535c0.325-1.009,0.737-1.977,1.214-2.907l-1.73-2.49c-0.829-1.192-0.685-2.807,0.342-3.834l3.523-3.523c1.027-1.027,2.642-1.171,3.834-0.342l2.49,1.731c0.93-0.477,1.898-0.889,2.906-1.214l0.535-2.974C30.813,13.04,32.057,12,33.509,12h4.983c1.452,0,2.696,1.04,2.953,2.469l0.535,2.974c1.009,0.325,1.977,0.736,2.906,1.214l2.49-1.731c1.192-0.829,2.807-0.685,3.834,0.342l3.523,3.523c1.027,1.027,1.171,2.641,0.342,3.834l-1.73,2.49c0.477,0.93,0.889,1.898,1.214,2.907L57.531,30.556z M36,45c4.97,0,9-4.029,9-9c0-4.971-4.03-9-9-9s-9,4.029-9,9C27,40.971,31.03,45,36,45z" />
      </svg>
    </button>
    <button
      class="icon-btn"
      type="button"
      title={t(app.lang, 'checkUpdate')}
      aria-label={t(app.lang, 'checkUpdate')}
      onclick={() => checkUpdates(false)}
    >
      <svg viewBox="0 0 21 21" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <g transform="translate(2 2)">
          <path d="m4.5 1.5c-2.4138473 1.37729434-4 4.02194088-4 7 0 4.418278 3.581722 8 8 8s8-3.581722 8-8-3.581722-8-8-8" />
          <path d="m4.5 5.5v-4h-4" />
        </g>
      </svg>
    </button>
    <div class="win-btns">
      <button class="win" type="button" onclick={() => AppService.WindowMinimise()} aria-label="Minimise">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <path d="M9 12H15" />
        </svg>
      </button>
      <button
        class="win close"
        type="button"
        onclick={async () => {
          try {
            await (AppService as any).CancelGoogleAuth?.();
          } catch {
          }
          await AppService.WindowClose();
        }}
        aria-label="Close"
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <path d="M9 9L15 15" />
          <path d="M15 9L9 15" />
        </svg>
      </button>
    </div>
  </div>
</header>
