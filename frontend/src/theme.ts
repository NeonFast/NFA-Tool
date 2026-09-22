/** Theme is the user's theme preference. */
export type Theme = 'auto' | 'dark' | 'light';

/** ResolvedTheme is the effective theme after resolving 'auto'. */
export type ResolvedTheme = 'dark' | 'light';

const STORAGE_KEY = 'nfa-tool-theme';

/** detectSystemTheme returns the OS-preferred color scheme. */
export function detectSystemTheme(): ResolvedTheme {
  try {
    if (window.matchMedia('(prefers-color-scheme: light)').matches) return 'light';
  } catch {
  }
  return 'dark';
}

/** loadTheme reads the saved theme preference, defaulting to 'auto'. */
export function loadTheme(): Theme {
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved === 'auto' || saved === 'dark' || saved === 'light') return saved;
  } catch {
  }
  return 'auto';
}

/** saveTheme persists the theme preference. */
export function saveTheme(theme: Theme) {
  try {
    localStorage.setItem(STORAGE_KEY, theme);
  } catch {
  }
}

/** resolveTheme maps a preference to the effective theme. */
export function resolveTheme(theme: Theme): ResolvedTheme {
  return theme === 'auto' ? detectSystemTheme() : theme;
}

function applyResolved(resolved: ResolvedTheme) {
  document.documentElement.dataset.theme = resolved;
}

let autoHooked = false;

/** Applies the theme and, for 'auto', keeps it in sync with the OS setting. */
export function applyTheme(theme: Theme) {
  applyResolved(resolveTheme(theme));
  if (theme !== 'auto' || autoHooked) return;
  let mq: MediaQueryList;
  try {
    mq = window.matchMedia('(prefers-color-scheme: light)');
  } catch {
    return;
  }
  autoHooked = true;
  const onChange = () => {
    if (loadTheme() === 'auto') applyResolved(mq.matches ? 'light' : 'dark');
  };
  if (typeof mq.addEventListener === 'function') mq.addEventListener('change', onChange);
  else mq.addListener(onChange);
}
