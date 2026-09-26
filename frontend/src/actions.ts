import { flushSync } from 'svelte';
import { Browser } from '@wailsio/runtime';
import { AppService } from '../bindings/nfa-tool';

// installLinkGuard keeps the webview locked to the app UI: external links
// open in the system browser, middle-click popups and drop-navigation are
// blocked. Call once per window (main app and drive guide).
export function installLinkGuard() {
  const openExternal = (url: string) => {
    void (async () => {
      try {
        await (AppService as any).OpenURL(url);
      } catch {
        await Browser.OpenURL(url);
      }
    })();
  };
  const linkAt = (e: MouseEvent): HTMLAnchorElement | null => {
    const a = (e.target as HTMLElement).closest?.('a[href]');
    return (a as HTMLAnchorElement) ?? null;
  };
  const onClick = (e: MouseEvent) => {
    const a = linkAt(e);
    if (a && /^https?:\/\//i.test(a.href)) {
      e.preventDefault();
      e.stopPropagation();
      openExternal(a.href);
    }
  };
  const onAuxClick = (e: MouseEvent) => {
    if (linkAt(e)) {
      e.preventDefault();
      e.stopPropagation();
    }
  };
  const onDrop = (e: DragEvent) => {
    const el = e.target as HTMLElement;
    if (el.closest?.('input, textarea, [contenteditable="true"]')) return;
    e.preventDefault();
  };
  document.addEventListener('click', onClick, true);
  document.addEventListener('auxclick', onAuxClick, true);
  window.addEventListener('drop', onDrop, true);
}

export function aimHint(node: HTMLElement) {
  const update = () => {
    try {
      const fab = node.querySelector<HTMLElement>('.help-fab');
      if (!fab) return;
      const centerX = fab.offsetLeft + fab.offsetWidth / 2;
      node.style.setProperty('--aim-r', `${node.offsetWidth - centerX}px`);
    } catch {
    }
  };
  update();
  const ro = new ResizeObserver(update);
  ro.observe(node);
  return {
    destroy() {
      ro.disconnect();
    },
  };
}

const FLIP_CLONE = 'pill-flip-clone';
const FLIP_HIDDEN = 'data-flip-hidden';

let calmTimer = 0;

export function calmUpdate(apply: () => void) {
  const root = document.documentElement;
  root.classList.add('calm-fade');
  window.clearTimeout(calmTimer);
  calmTimer = window.setTimeout(() => {
    apply();
    requestAnimationFrame(() =>
      requestAnimationFrame(() => {
        root.classList.remove('calm-fade');
      }),
    );
  }, 140);
}

export function pillTransition(btn: HTMLElement, update: () => void) {
  const sw = btn.closest('.lang-switch');
  if (!sw || window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
    update();
    return;
  }
  document.querySelectorAll(`.${FLIP_CLONE}`).forEach((c) => c.remove());
  sw.querySelectorAll<HTMLElement>(`.pill[${FLIP_HIDDEN}]`).forEach((p) => {
    p.style.opacity = '';
    p.removeAttribute(FLIP_HIDDEN);
  });
  const from = sw.querySelector('.lang-btn.active .pill')?.getBoundingClientRect();
  flushSync(update);
  const to = sw.querySelector<HTMLElement>('.lang-btn.active .pill');
  if (!from || !to) return;
  const target = to.getBoundingClientRect();
  if (
    from.left === target.left &&
    from.top === target.top &&
    from.width === target.width &&
    from.height === target.height
  ) {
    return;
  }
  const cs = getComputedStyle(to);
  const clone = document.createElement('div');
  clone.className = FLIP_CLONE;
  clone.style.cssText =
    `position:fixed;z-index:2147483647;pointer-events:none;border-radius:999px;` +
    `left:${target.left}px;top:${target.top}px;width:${target.width}px;height:${target.height}px;` +
    `background:${cs.backgroundColor};box-shadow:${cs.boxShadow};transform-origin:0 0;will-change:transform;`;
  document.body.appendChild(clone);
  to.style.opacity = '0';
  to.setAttribute(FLIP_HIDDEN, '');
  const finish = () => {
    clone.remove();
    to.style.opacity = '';
    to.removeAttribute(FLIP_HIDDEN);
  };
  // transform-only animation: under software rendering (--disable-gpu),
  // animating left/top/width/height forced a full layout per frame and
  // crashed the WebView2 renderer (0xC0000005 in msedgewebview2.dll).
  const anim = clone.animate(
    [
      {
        transform:
          `translate(${from.left - target.left}px, ${from.top - target.top}px) ` +
          `scale(${from.width / target.width}, ${from.height / target.height})`,
      },
      { transform: 'translate(0px, 0px) scale(1, 1)' },
    ],
    { duration: 300, easing: 'cubic-bezier(0.22, 1, 0.36, 1)' },
  );
  anim.onfinish = finish;
  anim.oncancel = finish;
}
