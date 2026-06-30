export type ThemePref = 'light' | 'dark' | 'system';
export type ResolvedTheme = 'light' | 'dark';

const KEY = 'cla_theme';
const mql = () => matchMedia('(prefers-color-scheme: dark)');

export const resolveTheme = (p: ThemePref): ResolvedTheme =>
  p === 'system' ? (mql().matches ? 'dark' : 'light') : p;

export function applyTheme(p: ThemePref): void {
  const r = resolveTheme(p);
  const el = document.documentElement;
  el.setAttribute('data-theme', r);
  el.style.colorScheme = r;
}

export const getPref = (): ThemePref => (localStorage.getItem(KEY) as ThemePref | null) ?? 'system';

export function setPref(p: ThemePref): void {
  localStorage.setItem(KEY, p);
  applyTheme(p);
}

let off: (() => void) | null = null;

export function bindSystem(p: ThemePref, onChange: () => void): void {
  off?.();
  off = null;
  if (p !== 'system') return;
  const m = mql();
  const h = () => {
    applyTheme('system');
    onChange();
  };
  m.addEventListener('change', h);
  off = () => m.removeEventListener('change', h);
}

export const THEME_KEY = KEY;
