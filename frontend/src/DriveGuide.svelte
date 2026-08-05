<script lang="ts">
  import { Browser } from '@wailsio/runtime';
  import { AppService } from '../bindings/nfa-tool';
  import { type Lang, loadLang, saveLang, t } from './i18n';

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
    lang = next;
    saveLang(next);
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
  <header class="bar">
    <div class="brand">
      <span class="dot"></span>
      <span>{t(lang, 'driveTutorialTitle')}</span>
    </div>
    <div class="bar-right">
      <div class="lang-switch">
        <button type="button" class="lang-btn" class:active={lang === 'ru'} onclick={() => setLang('ru')}>RU</button>
        <button type="button" class="lang-btn" class:active={lang === 'en'} onclick={() => setLang('en')}>EN</button>
      </div>
      <button class="x" type="button" onclick={closeWin} aria-label={t(lang, 'close')}>✕</button>
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

  <main class="body" style="animation: none">
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
    --text: #fafafa;
    --muted: #71717a;
    --accent: #ffffff;
    --accent-strong: #ffffff;
    --cyan: #e4e4e7;
    --warn: #fbbf24;
    --card: #141416;
    --border: rgba(255, 255, 255, 0.08);
    --radius: 14px;
    --ease: cubic-bezier(0.22, 1, 0.36, 1);
    font-family: 'Segoe UI Variable', 'Segoe UI', system-ui, sans-serif;
    color-scheme: dark;
  }

  :global(*) {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
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

  @keyframes fade-up {
    from { opacity: 0; transform: translateY(10px); }
    to { opacity: 1; transform: translateY(0); }
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

  .bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 16px 8px 18px;
    gap: 12px;
  }

  .brand {
    display: flex;
    align-items: center;
    gap: 10px;
    font-weight: 650;
    color: var(--text);
    font-size: 14px;
  }

  .dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: #ffffff;
    box-shadow: 0 0 0 4px rgba(255, 255, 255, 0.1);
  }

  .bar-right {
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
  }

  .lang-btn {
    border: none;
    background: transparent;
    color: var(--muted);
    padding: 6px 10px;
    font-size: 11.5px;
    font-weight: 700;
    cursor: pointer;
  }

  .lang-btn {
    transition: background 0.2s var(--ease), color 0.2s var(--ease);
  }

  .lang-btn.active {
    background: #ffffff;
    color: #09090b;
  }

  .x {
    width: 34px;
    height: 28px;
    border: none;
    border-radius: 8px;
    background: transparent;
    color: var(--muted);
    cursor: pointer;
  }

  .x:hover {
    background: #e11d48;
    color: white;
  }

  .progress {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 22px 14px;
    --wails-draggable: no-drag;
  }

  .pip {
    width: 28px;
    height: 6px;
    border: none;
    border-radius: 99px;
    background: rgba(255, 255, 255, 0.12);
    cursor: pointer;
    padding: 0;
    transition: width 0.28s var(--ease), background 0.28s var(--ease), transform 0.2s var(--ease);
  }

  .pip:hover {
    transform: scaleY(1.25);
  }

  .pip.done {
    background: rgba(255, 255, 255, 0.35);
  }

  .pip.on {
    background: #ffffff;
    width: 42px;
  }

  .count {
    margin-left: auto;
    font-size: 12px;
    color: var(--muted);
    font-weight: 600;
  }

  .body {
    margin: 0 18px;
    padding: 28px 28px 22px;
    border-radius: var(--radius);
    background: var(--card);
    border: 1px solid var(--border);
    box-shadow: 0 12px 40px rgba(0, 0, 0, 0.35);
    overflow: auto;
    min-height: 0;
    --wails-draggable: no-drag;
  }

  .step-anim {
    animation: slide-step 0.32s var(--ease) both;
  }

  .eyebrow {
    font-size: 12px;
    letter-spacing: 0.6px;
    text-transform: uppercase;
    color: var(--muted);
    font-weight: 700;
    margin-bottom: 10px;
  }

  h1 {
    font-size: 24px;
    font-weight: 700;
    line-height: 1.25;
    color: var(--text);
    letter-spacing: -0.02em;
    margin-bottom: 16px;
  }

  h1.warn {
    color: var(--warn);
  }

  .text {
    font-size: 15.5px;
    line-height: 1.6;
    color: var(--text);
    white-space: pre-wrap;
  }

  .text.warn {
    color: #fde68a;
  }

  .callout {
    margin-top: 18px;
    padding: 12px 14px;
    border-radius: 12px;
    background: rgba(251, 191, 36, 0.12);
    border: 1px solid rgba(251, 191, 36, 0.28);
    color: #fde68a;
    font-size: 13.5px;
    line-height: 1.45;
  }

  .callout.soft {
    background: rgba(167, 139, 250, 0.1);
    border-color: rgba(167, 139, 250, 0.25);
    color: var(--muted);
  }

  .links {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 22px;
  }

  .link-btn {
    border: 1px solid rgba(255, 255, 255, 0.22);
    background: rgba(255, 255, 255, 0.08);
    color: #ffffff;
    border-radius: 10px;
    padding: 10px 14px;
    font-size: 13px;
    font-weight: 650;
    cursor: pointer;
    transition: background 0.2s var(--ease), border-color 0.2s var(--ease), transform 0.2s var(--ease);
  }

  .link-btn:hover {
    background: rgba(255, 255, 255, 0.14);
    border-color: rgba(255, 255, 255, 0.4);
  }

  .link-btn:active {
    transform: scale(0.98);
  }

  .foot {
    display: flex;
    gap: 10px;
    padding: 16px 18px 18px;
    --wails-draggable: no-drag;
  }

  .ghost {
    flex: 1;
    height: 46px;
    border-radius: 12px;
    border: 1px solid rgba(255, 255, 255, 0.1);
    background: rgba(255, 255, 255, 0.04);
    color: var(--text);
    font-weight: 650;
    cursor: pointer;
    transition: background 0.2s var(--ease), border-color 0.2s var(--ease), transform 0.2s var(--ease);
  }

  .ghost:disabled {
    opacity: 0.35;
    cursor: not-allowed;
  }

  .ghost:hover:not(:disabled) {
    border-color: rgba(255, 255, 255, 0.28);
    background: rgba(255, 255, 255, 0.07);
  }

  .ghost:active:not(:disabled),
  .primary:active {
    transform: scale(0.98);
  }

  .primary {
    flex: 1.4;
    height: 46px;
    border: none;
    border-radius: 12px;
    background: #ffffff;
    color: #09090b;
    font-weight: 750;
    font-size: 15px;
    cursor: pointer;
    transition: background 0.2s var(--ease), transform 0.2s var(--ease);
  }

  .primary:hover {
    background: #f4f4f5;
  }
</style>
