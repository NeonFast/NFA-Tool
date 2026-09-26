<script lang="ts">
  import { Browser } from '@wailsio/runtime';
  import { onMount } from 'svelte';
  import { AppService } from '../bindings/nfa-tool';
  import { type Lang, loadLang, saveLang, t } from './i18n';
  import { calmUpdate, installLinkGuard } from './actions';

  onMount(() => {
    installLinkGuard();
  });

  const steps: {
    titleKey: 'driveGuideT1' | 'driveGuideT2' | 'driveGuideT3' | 'driveGuideT4' | 'driveGuideT5' | 'driveGuideT6' | 'driveGuideT7' | 'driveGuideT8';
    bodyKey: 'driveStep1' | 'driveStep2' | 'driveStep4' | 'driveStep5' | 'driveStep6' | 'driveStep8' | 'driveStep9' | 'driveStep10';
    warn?: boolean;
    links?: { label: string; url: string }[];
  }[] = [
    {
      titleKey: 'driveGuideT1',
      bodyKey: 'driveStep1',
      links: [{ label: 'Cloud Console', url: 'https://console.cloud.google.com/' }],
    },
    {
      titleKey: 'driveGuideT2',
      bodyKey: 'driveStep2',
      warn: true,
      links: [{ label: 'New Project', url: 'https://console.cloud.google.com/projectcreate' }],
    },
    {
      titleKey: 'driveGuideT3',
      bodyKey: 'driveStep4',
      links: [
        { label: 'Drive API', url: 'https://console.cloud.google.com/flows/enableapi?apiid=drive.googleapis.com' },
        { label: 'API Library', url: 'https://console.cloud.google.com/apis/library' },
      ],
    },
    {
      titleKey: 'driveGuideT4',
      bodyKey: 'driveStep5',
      links: [{ label: 'Branding', url: 'https://console.cloud.google.com/auth/branding' }],
    },
    {
      titleKey: 'driveGuideT5',
      bodyKey: 'driveStep6',
      warn: true,
      links: [{ label: 'Audience / Test users', url: 'https://console.cloud.google.com/auth/audience' }],
    },
    {
      titleKey: 'driveGuideT6',
      bodyKey: 'driveStep8',
      links: [
        { label: 'Clients', url: 'https://console.cloud.google.com/auth/clients' },
        { label: 'Credentials', url: 'https://console.cloud.google.com/apis/credentials' },
      ],
    },
    {
      titleKey: 'driveGuideT7',
      bodyKey: 'driveStep9',
    },
    {
      titleKey: 'driveGuideT8',
      bodyKey: 'driveStep10',
    },
  ];

  let lang = $state<Lang>(loadLang());
  let step = $state(0);

  const total = steps.length;
  const cur = $derived(steps[step]);
  const isLast = $derived(step >= total - 1);
  const isFirst = $derived(step <= 0);

  function setLang(next: Lang) {
    if (next === lang) return;
    calmUpdate(() => {
      lang = next;
      saveLang(next);
    });
  }

  async function openURL(url: string) {
    try {
      await (AppService as any).OpenURL(url);
    } catch {
      await Browser.OpenURL(url);
    }
  }

  async function closeWin() {
    try {
      await (AppService as any).CloseDriveGuide();
    } catch {
      window.close();
    }
  }

  function next() {
    if (isLast) {
      void closeWin();
      return;
    }
    step += 1;
  }

  function back() {
    if (!isFirst) step -= 1;
  }
</script>

<div class="guide">
  <header class="titlebar">
    <div class="brand">
      <span class="logo-mark" aria-hidden="true">G</span>
      <div class="brand-text">
        <span class="brand-name">{t(lang, 'driveTutorialTitle')}</span>
        <span class="brand-ver">Google Drive</span>
      </div>
    </div>
    <div class="title-actions">
      <div class="lang-switch">
        <button type="button" class="lang-btn" class:active={lang === 'ru'} onclick={() => setLang('ru')}><span class="pill" aria-hidden="true"></span><span class="lbl">RU</span></button>
        <button type="button" class="lang-btn" class:active={lang === 'en'} onclick={() => setLang('en')}><span class="pill" aria-hidden="true"></span><span class="lbl">EN</span></button>
      </div>
      <div class="win-btns">
        <button class="win close" type="button" onclick={closeWin} aria-label={t(lang, 'close')}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M9 9L15 15" />
            <path d="M15 9L9 15" />
          </svg>
        </button>
      </div>
    </div>
  </header>

  <div class="progress">
    {#each steps as _, i}
      <button
        type="button"
        class="pip"
        class:on={i === step}
        class:done={i < step}
        onclick={() => (step = i)}
        aria-label={`Step ${i + 1}`}
      ></button>
    {/each}
    <span class="count">{step + 1} / {total}</span>
  </div>

  <main class="body panel" style="animation: none">
    {#key step}
      <div class="step-anim">
    <p class="eyebrow">{t(lang, 'driveGuideStepOf', { n: step + 1, total })}</p>
    <h1 class:warn={cur.warn}>{t(lang, cur.titleKey)}</h1>
    <p class="text" class:warn={cur.warn}>{t(lang, cur.bodyKey)}</p>

    {#if step === 4}
      <div class="callout">⚠ {t(lang, 'driveTipTestUsers')}</div>
    {/if}
    {#if step === 1}
      <div class="callout soft">💡 {t(lang, 'driveTipOwnProject')}</div>
    {/if}
    {#if step === 5}
      <div class="callout soft">🖥 {t(lang, 'driveTipDesktop')}</div>
      <div class="callout soft">📂 {t(lang, 'driveTipApisServices')}</div>
    {/if}

    {#if cur.links?.length}
      <div class="links">
        {#each cur.links as link}
          <button class="link-btn" type="button" onclick={() => openURL(link.url)}>{link.label}</button>
        {/each}
      </div>
    {/if}
      </div>
    {/key}
  </main>

  <footer class="foot">
    <button class="ghost" type="button" disabled={isFirst} onclick={back}>{t(lang, 'driveGuideBack')}</button>
    <button class="primary" type="button" onclick={next}>
      {isLast ? t(lang, 'driveGuideDone') : t(lang, 'driveGuideNext')}
    </button>
  </footer>
</div>

<style>
  :global(:root) {
    --bg: #09090b;
    --bg-elevated: #141416;
    --bg-titlebar: rgba(14, 14, 17, 0.96);
    --border: rgba(255, 255, 255, 0.08);
    --border-strong: rgba(255, 255, 255, 0.16);
    --text: #fafafa;
    --text-secondary: #b4b4bc;
    --muted: #8b8b94;
    --accent: #ffffff;
    --pill-bg: rgba(255, 255, 255, 0.16);
    --accent-hover: #f4f4f5;
    --accent-active: #e4e4e7;
    --accent-fg: #09090b;
    --accent-soft: rgba(255, 255, 255, 0.1);
    --hover: rgba(255, 255, 255, 0.04);
    --hover-strong: rgba(255, 255, 255, 0.1);
    --picked-bg: rgba(255, 255, 255, 0.06);
    --mini-hover-border: rgba(255, 255, 255, 0.28);
    --focus-border: rgba(255, 255, 255, 0.45);
    --logo-ring: rgba(255, 255, 255, 0.08);
    --beta-fg: #fbbf24;
    --warn-text: #fde68a;
    --logo-shadow: 0 0 0 1px var(--logo-ring), 0 8px 20px rgba(0, 0, 0, 0.35);
    --primary-shadow: 0 1px 0 rgba(255, 255, 255, 0.35) inset, 0 8px 22px rgba(0, 0, 0, 0.28);
    --primary-shadow-hover: 0 1px 0 rgba(255, 255, 255, 0.35) inset, 0 12px 28px rgba(0, 0, 0, 0.34);
    --pill-shadow: 0 1px 2px rgba(0, 0, 0, 0.25);
    --radius: 14px;
    --radius-sm: 10px;
    --shadow: 0 1px 0 rgba(255, 255, 255, 0.04) inset, 0 16px 40px rgba(0, 0, 0, 0.4);
    --ease: cubic-bezier(0.22, 1, 0.36, 1);
    --font: "Segoe UI Variable", "Segoe UI", system-ui, -apple-system, sans-serif;
    font-family: var(--font);
    color-scheme: dark;
  }

  :global(:root[data-theme='light']) {
    --bg: #f4f4f5;
    --bg-elevated: #ffffff;
    --bg-titlebar: rgba(250, 250, 250, 0.96);
    --border: rgba(0, 0, 0, 0.09);
    --border-strong: rgba(0, 0, 0, 0.22);
    --text: #18181b;
    --text-secondary: #45454e;
    --muted: #61616b;
    --accent: #18181b;
    --pill-bg: rgba(0, 0, 0, 0.1);
    --accent-hover: #27272a;
    --accent-active: #3f3f46;
    --accent-fg: #fafafa;
    --accent-soft: rgba(0, 0, 0, 0.07);
    --hover: rgba(0, 0, 0, 0.04);
    --hover-strong: rgba(0, 0, 0, 0.08);
    --picked-bg: rgba(0, 0, 0, 0.05);
    --mini-hover-border: rgba(0, 0, 0, 0.3);
    --focus-border: rgba(0, 0, 0, 0.45);
    --logo-ring: rgba(0, 0, 0, 0.1);
    --beta-fg: #b45309;
    --warn-text: #92400e;
    --logo-shadow: 0 0 0 1px var(--logo-ring), 0 3px 10px rgba(0, 0, 0, 0.14);
    --primary-shadow: 0 1px 0 rgba(255, 255, 255, 0.1) inset, 0 3px 10px rgba(0, 0, 0, 0.14);
    --primary-shadow-hover: 0 1px 0 rgba(255, 255, 255, 0.1) inset, 0 5px 14px rgba(0, 0, 0, 0.17);
    --pill-shadow: 0 1px 2px rgba(0, 0, 0, 0.16);
    --shadow: 0 1px 2px rgba(0, 0, 0, 0.04), 0 6px 20px rgba(0, 0, 0, 0.06);
    color-scheme: light;
  }

  :global(*) {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
    scrollbar-width: none;
  }

  :global(::-webkit-scrollbar) {
    width: 0;
    height: 0;
    display: none;
  }

  :global(html),
  :global(body),
  :global(#app) {
    height: 100%;
    overflow: hidden;
    background: var(--bg);
    color: var(--text);
    user-select: none;
  }

  :global(body) {
    background: var(--bg);
    --wails-draggable: drag;
  }

  @keyframes slide-step {
    from { opacity: 0; transform: translateX(14px); }
    to { opacity: 1; transform: translateX(0); }
  }

  .guide {
    height: 100%;
    display: grid;
    grid-template-rows: auto auto 1fr auto;
  }

  .titlebar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 14px 14px 12px 18px;
    border-bottom: 1px solid var(--border);
    background: var(--bg-titlebar);
  }

  .brand {
    display: flex;
    align-items: center;
    gap: 12px;
    min-width: 0;
  }

  .logo-mark {
    width: 32px;
    height: 32px;
    border-radius: 10px;
    display: grid;
    place-items: center;
    font-weight: 800;
    font-size: 14px;
    color: var(--accent-fg);
    background: var(--accent);
    box-shadow: var(--logo-shadow);
    flex-shrink: 0;
  }

  .brand-text {
    display: flex;
    flex-direction: column;
    gap: 1px;
    min-width: 0;
  }

  .brand-name {
    font-size: 14px;
    font-weight: 650;
    letter-spacing: -0.01em;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .brand-ver {
    font-size: 11px;
    color: var(--muted);
    font-weight: 600;
  }

  .title-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    --wails-draggable: no-drag;
  }

  :global(#app) {
    transition: opacity 0.16s ease;
  }

  :global(html.calm-fade #app) {
    opacity: 0;
  }

  .lang-switch {
    display: inline-flex;
    padding: 3px;
    background: var(--bg-elevated);
    border: 1px solid var(--border);
    border-radius: 999px;
    position: relative;
  }

  .lang-btn .pill {
    position: absolute;
    inset: 0;
    z-index: -1;
    border-radius: 999px;
    background: var(--pill-bg);
    box-shadow: var(--pill-shadow);
    opacity: 0;
  }

  .lang-btn.active .pill {
    opacity: 1;
  }

  .lang-btn {
    white-space: nowrap;
    border: none;
    background: transparent;
    color: var(--muted);
    height: 22px;
    padding: 0 11px;
    font-size: 11px;
    font-weight: 700;
    line-height: 1;
    border-radius: 999px;
    cursor: pointer;
    position: relative;
    z-index: 1;
    display: inline-flex;
    align-items: center;
    gap: 5px;
    transition: color 0.2s var(--ease), transform 0.2s var(--ease);
  }

  .lang-btn.active {
    color: var(--text);
  }

  .lang-btn .lbl {
    position: relative;
    top: 0.5px;
  }

  .win-btns {
    display: flex;
    gap: 4px;
    margin-left: 4px;
  }

  .win {
    width: 38px;
    height: 32px;
    border: none;
    border-radius: 8px;
    background: transparent;
    color: var(--muted);
    cursor: pointer;
    display: inline-grid;
    place-items: center;
    transition: background 0.2s var(--ease), color 0.2s var(--ease);
  }

  .win svg {
    width: 26px;
    height: 26px;
    display: block;
  }

  .win:hover {
    background: var(--picked-bg);
    color: var(--text);
  }

  .win.close:hover {
    background: #e11d48;
    color: white;
  }

  .progress {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px 22px 12px;
    --wails-draggable: no-drag;
  }

  .pip {
    width: 28px;
    height: 6px;
    border: none;
    border-radius: 99px;
    background: var(--hover-strong);
    cursor: pointer;
    padding: 0;
    transition: width 0.28s var(--ease), background 0.28s var(--ease), transform 0.2s var(--ease);
  }

  .pip:hover {
    transform: scaleY(1.25);
  }

  .pip.done {
    background: var(--pill-bg);
  }

  .pip.on {
    background: var(--accent);
    width: 42px;
  }

  .count {
    margin-left: auto;
    font-size: 12px;
    color: var(--muted);
    font-weight: 600;
  }

  .panel {
    background: var(--bg-elevated);
    border: 1px solid var(--border);
    border-radius: calc(var(--radius) + 2px);
    box-shadow: var(--shadow);
    --wails-draggable: no-drag;
  }

  .body {
    margin: 0 18px;
    padding: 24px 24px 20px;
    overflow: auto;
    min-height: 0;
  }

  .step-anim {
    animation: slide-step 0.32s var(--ease) both;
  }

  .eyebrow {
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--muted);
    margin-bottom: 10px;
  }

  h1 {
    font-size: 20px;
    font-weight: 700;
    line-height: 1.25;
    color: var(--text);
    letter-spacing: -0.02em;
    margin-bottom: 14px;
  }

  h1.warn {
    color: var(--beta-fg);
  }

  .text {
    font-size: 14px;
    line-height: 1.6;
    color: var(--text-secondary);
    white-space: pre-wrap;
  }

  .text.warn {
    color: var(--warn-text);
  }

  .callout {
    margin-top: 16px;
    padding: 12px 14px;
    border-radius: var(--radius-sm);
    background: rgba(251, 191, 36, 0.12);
    border: 1px solid rgba(251, 191, 36, 0.28);
    color: var(--warn-text);
    font-size: 13px;
    line-height: 1.45;
  }

  .callout.soft {
    background: var(--accent-soft);
    border-color: var(--mini-hover-border);
    color: var(--text-secondary);
  }

  .links {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 20px;
  }

  .link-btn {
    border: 1px solid var(--mini-hover-border);
    background: var(--accent-soft);
    color: var(--text);
    border-radius: var(--radius-sm);
    padding: 7px 12px;
    font-size: 12.5px;
    font-weight: 600;
    font-family: inherit;
    cursor: pointer;
    transition:
      background 0.2s var(--ease),
      border-color 0.2s var(--ease),
      color 0.2s var(--ease),
      transform 0.2s var(--ease);
  }

  .link-btn:hover {
    background: var(--hover-strong);
    border-color: var(--focus-border);
    transform: translateY(-1px);
  }

  .link-btn:active {
    transform: translateY(0) scale(0.98);
  }

  .foot {
    display: flex;
    gap: 10px;
    padding: 16px 18px 18px;
    --wails-draggable: no-drag;
  }

  .ghost {
    flex: 1;
    height: 42px;
    border-radius: var(--radius-sm);
    border: 1px solid var(--border);
    background: transparent;
    color: var(--text-secondary);
    font-family: inherit;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition:
      background 0.2s var(--ease),
      border-color 0.2s var(--ease),
      color 0.2s var(--ease),
      opacity 0.2s var(--ease),
      transform 0.2s var(--ease);
  }

  .ghost:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }

  .ghost:hover:not(:disabled) {
    background: var(--hover);
    border-color: var(--border-strong);
    color: var(--text);
    transform: translateY(-1px);
  }

  .ghost:active:not(:disabled) {
    transform: translateY(0) scale(0.98);
  }

  .primary {
    flex: 1.4;
    height: 42px;
    border: none;
    border-radius: var(--radius-sm);
    background: var(--accent);
    color: var(--accent-fg);
    font-family: inherit;
    font-weight: 700;
    font-size: 14px;
    cursor: pointer;
    box-shadow: var(--primary-shadow);
    transition:
      background 0.2s var(--ease),
      transform 0.2s var(--ease),
      box-shadow 0.2s var(--ease);
  }

  .primary:hover {
    background: var(--accent-hover);
    transform: translateY(-1px);
    box-shadow: var(--primary-shadow-hover);
  }

  .primary:active {
    background: var(--accent-active);
    transform: translateY(0) scale(0.98);
  }
</style>
