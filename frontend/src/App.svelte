<script lang="ts">
  import { onMount } from 'svelte';
  import { Browser } from '@wailsio/runtime';
  import { AppService } from '../bindings/nfa-tool';
  import {
    type Lang,
    loadLang,
    saveLang,
    t,
    translateBackendMessage,
    localizeExpiry,
  } from './i18n';

  async function notify(ok: boolean, message: string) {
    const title = ok ? t(lang, 'successTitle') : t(lang, 'errorTitle');
    const text = translateBackendMessage(lang, message);
    try {
      await AppService.Notify(ok, title, text);
    } catch {
      // fallback if bindings lag behind
      window.alert(`${title}\n\n${text}`);
    }
  }

  const GUIDE_URL = 'https://teletype.in/@hackerdlc/CS2NFA';

  type Account = {
    name: string;
    expiresIn: string;
    valid: boolean;
  };

  type Result = {
    ok: boolean;
    message: string;
  };

  let appName = $state('NFA Tool Recode v2');
  let version = $state('2.0.0');
  let lang = $state<Lang>(loadLang());
  let accountKey = $state('');
  let keepExisting = $state(false);
  let accounts = $state<Account[]>([]);
  let status = $state('');
  let statusKind = $state<'ok' | 'err' | ''>('ok');
  let loading = $state(false);
  let showHelp = $state(false);

  const title = $derived(`${appName} · ${version}`);

  onMount(async () => {
    status = t(lang, 'ready');
    try {
      // GetAppName may be missing until bindings regenerate
      const anySvc = AppService as typeof AppService & { GetAppName?: () => Promise<string> };
      if (typeof anySvc.GetAppName === 'function') {
        appName = await anySvc.GetAppName();
      }
      version = await AppService.GetVersion();
    } catch {
      /* ignore */
    }
    await refreshAccounts();
  });

  function setLang(next: Lang) {
    lang = next;
    saveLang(next);
    if (statusKind === 'ok' && (status === t('en', 'ready') || status === t('ru', 'ready') || status === 'Ready' || status === 'Готово')) {
      status = t(next, 'ready');
    }
  }

  async function refreshAccounts() {
    try {
      const list = await AppService.ListAccounts();
      accounts = (list ?? []).slice().sort((a, b) => a.name.localeCompare(b.name));
    } catch (e) {
      setStatus(String(e), 'err');
    }
  }

  function setStatus(msg: string, kind: 'ok' | 'err' | '' = '', alreadyTranslated = false) {
    status = alreadyTranslated ? msg : translateBackendMessage(lang, msg);
    statusKind = kind;
  }

  async function doLogin() {
    const key = accountKey.trim();
    if (!key) {
      const msg = t(lang, 'enterKey');
      setStatus(msg, 'err', true);
      await notify(false, msg);
      return;
    }
    loading = true;
    setStatus(t(lang, 'loggingIn'), '', true);
    try {
      const res = (await AppService.LoginFromKey(key, keepExisting)) as Result;
      setStatus(res.message, res.ok ? 'ok' : 'err');
      await notify(res.ok, res.message);
      if (res.ok) {
        accountKey = '';
        await refreshAccounts();
      }
    } catch (e) {
      const msg = String(e);
      setStatus(msg, 'err');
      await notify(false, msg);
    } finally {
      loading = false;
    }
  }

  async function loginSaved(name: string) {
    setStatus(t(lang, 'loggingInAs', { name }), '', true);
    try {
      const res = (await AppService.LoginSaved(name, keepExisting)) as Result;
      setStatus(res.message, res.ok ? 'ok' : 'err');
      await notify(res.ok, res.message);
      await refreshAccounts();
    } catch (e) {
      const msg = String(e);
      setStatus(msg, 'err');
      await notify(false, msg);
    }
  }

  async function deleteAccount(name: string) {
    try {
      const res = (await AppService.DeleteAccount(name)) as Result;
      setStatus(res.message, res.ok ? 'ok' : 'err');
      await notify(res.ok, res.message);
      await refreshAccounts();
    } catch (e) {
      const msg = String(e);
      setStatus(msg, 'err');
      await notify(false, msg);
    }
  }

  async function resetSteam() {
    setStatus(t(lang, 'resettingSteam'), '', true);
    try {
      const res = (await AppService.ResetSteam()) as Result;
      setStatus(res.message, res.ok ? 'ok' : 'err');
      if (res.message !== 'Cancelled' && res.message !== 'cancelled') {
        await notify(res.ok, res.message);
      }
    } catch (e) {
      const msg = String(e);
      setStatus(msg, 'err');
      await notify(false, msg);
    }
  }

  function onKey(e: KeyboardEvent) {
    if (e.key === 'Enter') doLogin();
  }
</script>

<div class="app">
  <header class="titlebar">
    <div class="brand">
      <span class="logo-dot"></span>
      <span>{title}</span>
    </div>
    <div class="title-actions">
      <div class="lang-switch" title={t(lang, 'lang')}>
        <button
          type="button"
          class="lang-btn"
          class:active={lang === 'ru'}
          onclick={() => setLang('ru')}
        >RU</button>
        <button
          type="button"
          class="lang-btn"
          class:active={lang === 'en'}
          onclick={() => setLang('en')}
        >EN</button>
      </div>
      <button class="ghost" type="button" onclick={resetSteam}>{t(lang, 'resetSteam')}</button>
      <button class="ghost" type="button" onclick={() => (showHelp = true)}>{t(lang, 'showInstructions')}</button>
      <div class="win-btns">
        <button class="win" type="button" onclick={() => AppService.WindowMinimise()} aria-label="Minimise">─</button>
        <button class="win close" type="button" onclick={() => AppService.WindowClose()} aria-label="Close">✕</button>
      </div>
    </div>
  </header>

  <main class="content">
    <section class="card">
      <h2>{t(lang, 'accountManagement')}</h2>
      <label class="field">
        <input
          type="text"
          spellcheck="false"
          autocomplete="off"
          placeholder={t(lang, 'accountKeyPlaceholder')}
          bind:value={accountKey}
          onkeydown={onKey}
        />
      </label>
      <label class="check">
        <input type="checkbox" bind:checked={keepExisting} />
        <span class="box"></span>
        <span>{t(lang, 'keepExisting')}</span>
      </label>
      <button class="primary" type="button" disabled={loading} onclick={doLogin}>
        {loading ? t(lang, 'working') : t(lang, 'login')}
      </button>
      <p class="hint">
        {accounts.length === 0 ? t(lang, 'hintEmpty') : t(lang, 'hintHasAccounts')}
      </p>
    </section>

    <section class="accounts-pane">
      <h2 class="accounts-title">{t(lang, 'savedAccounts')}</h2>
      {#if accounts.length === 0}
        <p class="empty">{t(lang, 'noSavedAccounts')}</p>
      {:else}
        <div class="accounts-list">
          {#each accounts as acc (acc.name)}
            <div class="account-row">
              <div class="meta">
                <div class="name">{acc.name}</div>
                <div class="exp" class:ok={acc.valid} class:bad={!acc.valid}>
                  {localizeExpiry(lang, acc.expiresIn)}
                </div>
              </div>
              <div class="row-actions">
                <button class="mini login" type="button" onclick={() => loginSaved(acc.name)}>{t(lang, 'login')}</button>
                <button class="mini del" type="button" onclick={() => deleteAccount(acc.name)}>{t(lang, 'delete')}</button>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </section>
  </main>

  <footer class="statusbar" class:ok={statusKind === 'ok'} class:err={statusKind === 'err'}>
    <span>{status}</span>
  </footer>
</div>

{#if showHelp}
  <!-- svelte-ignore a11y_click_events_have_key_events a11y_interactive_supports_focus -->
  <div class="modal" role="dialog" aria-modal="true" tabindex="-1" onclick={(e) => e.currentTarget === e.target && (showHelp = false)}>
    <div class="modal-card">
      <h3>{t(lang, 'instructions')}</h3>
      <ol>
        <li>{t(lang, 'help1')}</li>
        <li>{t(lang, 'help2')}</li>
        <li>{t(lang, 'help3')}</li>
        <li>{t(lang, 'help4')}</li>
        <li>{t(lang, 'help5')}</li>
      </ol>
      <p class="guide">
        {t(lang, 'fullGuide')}
        <button class="link" type="button" onclick={() => Browser.OpenURL(GUIDE_URL)}>
          teletype.in/@hackerdlc/CS2NFA
        </button>
      </p>
      <button class="primary" type="button" onclick={() => (showHelp = false)}>{t(lang, 'gotIt')}</button>
    </div>
  </div>
{/if}

<style>
  :global(:root) {
    --bg: #0f0f1a;
    --card-border: rgba(139, 92, 246, 0.12);
    --text: #e8e6f0;
    --muted: #8b87a0;
    --accent: #c4b5fd;
    --accent-strong: #a78bfa;
    --accent-dim: rgba(167, 139, 250, 0.15);
    --cyan: #2dd4bf;
    --green: #4ade80;
    --danger: #f87171;
    --radius: 16px;
    --shadow: 0 12px 40px rgba(0, 0, 0, 0.45);
    font-family: 'Segoe UI', system-ui, -apple-system, sans-serif;
    color-scheme: dark;
  }

  :global(*) {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
  }

  :global(html),
  :global(body) {
    height: 100%;
    overflow: hidden;
    background: var(--bg);
    color: var(--text);
    user-select: none;
  }

  :global(body) {
    background:
      radial-gradient(1200px 600px at 10% -10%, rgba(139, 92, 246, 0.18), transparent 55%),
      radial-gradient(900px 500px at 100% 0%, rgba(45, 212, 191, 0.08), transparent 50%),
      var(--bg);
    --wails-draggable: drag;
  }

  :global(#app) {
    height: 100%;
  }

  .app {
    height: 100%;
    display: grid;
    grid-template-rows: auto 1fr auto;
  }

  .titlebar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 14px 8px 18px;
    gap: 12px;
  }

  .brand {
    display: flex;
    align-items: center;
    gap: 10px;
    font-weight: 600;
    color: var(--accent);
    font-size: 15px;
  }

  .logo-dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: linear-gradient(135deg, #a78bfa, #2dd4bf);
    box-shadow: 0 0 12px rgba(167, 139, 250, 0.8);
  }

  .title-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    --wails-draggable: no-drag;
  }

  .lang-switch {
    display: inline-flex;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 10px;
    overflow: hidden;
    margin-right: 2px;
  }

  .lang-btn {
    border: none;
    background: transparent;
    color: var(--muted);
    padding: 6px 10px;
    font-size: 11.5px;
    font-weight: 700;
    letter-spacing: 0.4px;
    cursor: pointer;
  }

  .lang-btn.active {
    background: rgba(167, 139, 250, 0.22);
    color: var(--accent);
  }

  .lang-btn:hover:not(.active) {
    color: var(--text);
    background: rgba(255, 255, 255, 0.04);
  }

  .ghost {
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.08);
    color: var(--text);
    border-radius: 10px;
    padding: 7px 14px;
    font-size: 12.5px;
    cursor: pointer;
    transition: 0.15s ease;
  }

  .ghost:hover {
    background: rgba(167, 139, 250, 0.12);
    border-color: rgba(167, 139, 250, 0.35);
    color: var(--accent);
  }

  .win-btns {
    display: flex;
    margin-left: 6px;
    gap: 4px;
  }

  .win {
    width: 36px;
    height: 28px;
    border: none;
    border-radius: 8px;
    background: transparent;
    color: var(--muted);
    cursor: pointer;
    font-size: 12px;
  }

  .win:hover {
    background: rgba(255, 255, 255, 0.08);
    color: var(--text);
  }

  .win.close:hover {
    background: #e11d48;
    color: white;
  }

  .content {
    display: grid;
    grid-template-columns: 360px 1fr;
    gap: 22px;
    padding: 10px 22px 16px;
    min-height: 0;
  }

  .card {
    background: linear-gradient(180deg, rgba(36, 36, 58, 0.95), rgba(22, 22, 40, 0.98));
    border: 1px solid var(--card-border);
    border-radius: var(--radius);
    padding: 22px 20px;
    box-shadow: var(--shadow);
    display: flex;
    flex-direction: column;
    gap: 14px;
    --wails-draggable: no-drag;
  }

  .card h2 {
    text-align: center;
    font-size: 20px;
    font-weight: 650;
    color: var(--accent);
  }

  .field input {
    width: 100%;
    height: 46px;
    border-radius: 12px;
    border: 1px solid rgba(255, 255, 255, 0.06);
    background: rgba(0, 0, 0, 0.28);
    color: var(--text);
    padding: 0 14px;
    font-size: 13.5px;
    outline: none;
    user-select: text;
    -webkit-user-select: text;
  }

  .field input::placeholder {
    color: #6b6780;
  }

  .field input:focus {
    border-color: rgba(167, 139, 250, 0.55);
    box-shadow: 0 0 0 3px rgba(167, 139, 250, 0.12);
  }

  .check {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 13px;
    cursor: pointer;
  }

  .check input {
    display: none;
  }

  .check .box {
    width: 18px;
    height: 18px;
    border-radius: 5px;
    background: var(--accent-strong);
    flex-shrink: 0;
    opacity: 0.35;
    position: relative;
    transition: 0.15s ease;
  }

  .check input:checked + .box {
    opacity: 1;
  }

  .check input:checked + .box::after {
    content: '';
    position: absolute;
    left: 5px;
    top: 2px;
    width: 5px;
    height: 9px;
    border: solid #1a1030;
    border-width: 0 2px 2px 0;
    transform: rotate(45deg);
  }

  .primary {
    height: 46px;
    border: none;
    border-radius: 12px;
    background: linear-gradient(90deg, #c4b5fd 0%, #a78bfa 100%);
    color: #1a1030;
    font-weight: 700;
    font-size: 15px;
    cursor: pointer;
    transition: 0.15s ease;
  }

  .primary:hover:not(:disabled) {
    filter: brightness(1.06);
    transform: translateY(-1px);
  }

  .primary:disabled {
    opacity: 0.55;
    cursor: wait;
  }

  .hint {
    margin-top: auto;
    font-size: 12.5px;
    line-height: 1.45;
    color: var(--green);
  }

  .accounts-pane {
    min-height: 0;
    display: flex;
    flex-direction: column;
    padding: 6px 4px 0 8px;
    --wails-draggable: no-drag;
  }

  .accounts-title {
    color: var(--cyan);
    font-size: 22px;
    font-weight: 650;
    margin-bottom: 14px;
  }

  .accounts-list {
    overflow: auto;
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding-right: 6px;
    min-height: 0;
    flex: 1;
  }

  .account-row {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 10px;
    align-items: center;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 14px;
    padding: 12px 14px;
  }

  .account-row:hover {
    border-color: rgba(45, 212, 191, 0.25);
    background: rgba(45, 212, 191, 0.04);
  }

  .meta .name {
    font-weight: 600;
    font-size: 14px;
  }

  .meta .exp {
    margin-top: 3px;
    font-size: 12px;
    color: var(--muted);
  }

  .meta .exp.ok {
    color: var(--green);
  }

  .meta .exp.bad {
    color: var(--danger);
  }

  .row-actions {
    display: flex;
    gap: 6px;
  }

  .mini {
    border: none;
    border-radius: 9px;
    padding: 8px 12px;
    font-size: 12px;
    font-weight: 600;
    cursor: pointer;
  }

  .mini.login {
    background: var(--accent-dim);
    color: var(--accent);
  }

  .mini.login:hover {
    background: rgba(167, 139, 250, 0.28);
  }

  .mini.del {
    background: rgba(248, 113, 113, 0.12);
    color: var(--danger);
  }

  .mini.del:hover {
    background: rgba(248, 113, 113, 0.25);
  }

  .empty {
    color: var(--muted);
    font-size: 13px;
    padding: 20px 0;
  }

  .statusbar {
    height: 28px;
    display: flex;
    align-items: center;
    padding: 0 16px;
    font-size: 12px;
    color: var(--muted);
    border-top: 1px solid rgba(255, 255, 255, 0.04);
    background: rgba(0, 0, 0, 0.2);
  }

  .statusbar.ok {
    color: var(--green);
  }

  .statusbar.err {
    color: var(--danger);
  }

  .modal {
    position: fixed;
    inset: 0;
    background: rgba(5, 5, 12, 0.65);
    backdrop-filter: blur(6px);
    display: grid;
    place-items: center;
    z-index: 50;
    --wails-draggable: no-drag;
  }

  .modal-card {
    width: min(440px, 90vw);
    background: #1a1a2e;
    border: 1px solid var(--card-border);
    border-radius: 18px;
    padding: 22px;
    box-shadow: var(--shadow);
  }

  .modal-card h3 {
    color: var(--accent);
    margin-bottom: 12px;
  }

  .modal-card ol {
    padding-left: 18px;
    font-size: 13.5px;
    line-height: 1.55;
    margin-bottom: 18px;
  }

  .modal-card code {
    background: rgba(0, 0, 0, 0.35);
    padding: 1px 6px;
    border-radius: 6px;
    color: var(--cyan);
    font-size: 12px;
  }

  .modal-card .primary {
    width: 100%;
  }

  .guide {
    font-size: 13px;
    color: var(--muted);
    margin: -6px 0 14px;
  }

  .link {
    border: none;
    background: none;
    color: var(--cyan);
    cursor: pointer;
    font: inherit;
    text-decoration: underline;
    padding: 0;
  }

  .link:hover {
    color: var(--accent);
  }

  @media (max-width: 820px) {
    .content {
      grid-template-columns: 1fr;
    }
  }
</style>
