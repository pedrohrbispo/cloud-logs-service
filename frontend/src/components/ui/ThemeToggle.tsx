import { useEffect, useRef, useState, type KeyboardEvent } from 'react';
import { Monitor, Moon, Sun } from 'lucide-react';
import { useUiStore } from '@/store/uiStore';
import type { ThemePref } from '@/lib/theme';
import { SegmentedControl, type Segment } from './SegmentedControl';
import { cn } from '@/lib/cn';

const OPTIONS: Segment<ThemePref>[] = [
  { value: 'light', label: 'Light', icon: <Sun size={15} aria-hidden="true" /> },
  { value: 'dark', label: 'Dark', icon: <Moon size={15} aria-hidden="true" /> },
  { value: 'system', label: 'System', icon: <Monitor size={15} aria-hidden="true" /> },
];

function PrefIcon({ pref, size = 16 }: { pref: ThemePref; size?: number }) {
  const Icon = pref === 'light' ? Sun : pref === 'dark' ? Moon : Monitor;
  return <Icon size={size} aria-hidden="true" />;
}

interface ThemeToggleProps {
  variant: 'icon' | 'pill';
}

export function ThemeToggle({ variant }: ThemeToggleProps) {
  const theme = useUiStore((s) => s.theme);
  const setTheme = useUiStore((s) => s.setTheme);

  if (variant === 'pill') {
    return (
      <SegmentedControl
        aria-label="Color theme"
        options={OPTIONS}
        value={theme}
        onChange={setTheme}
        className="rounded-pill"
      />
    );
  }

  return <ThemeIconMenu theme={theme} setTheme={setTheme} />;
}

function ThemeIconMenu({
  theme,
  setTheme,
}: {
  theme: ThemePref;
  setTheme: (p: ThemePref) => void;
}) {
  const [open, setOpen] = useState(false);
  const wrapRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const radioRefs = useRef<(HTMLButtonElement | null)[]>([]);

  const close = (returnFocus: boolean) => {
    setOpen(false);
    if (returnFocus) triggerRef.current?.focus();
  };

  useEffect(() => {
    if (!open) return;
    const idx = OPTIONS.findIndex((o) => o.value === theme);
    radioRefs.current[idx === -1 ? 0 : idx]?.focus();

    const onDoc = (e: MouseEvent) => {
      if (!wrapRef.current?.contains(e.target as Node)) setOpen(false);
    };
    const onKey = (e: globalThis.KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.preventDefault();
        close(true);
      }
    };
    document.addEventListener('mousedown', onDoc);
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('mousedown', onDoc);
      document.removeEventListener('keydown', onKey);
    };
  }, [open, theme]);

  const onRadioKeyDown = (e: KeyboardEvent<HTMLButtonElement>) => {
    const keys = ['ArrowDown', 'ArrowRight', 'ArrowUp', 'ArrowLeft'];
    if (!keys.includes(e.key)) return;
    e.preventDefault();
    const delta = e.key === 'ArrowDown' || e.key === 'ArrowRight' ? 1 : -1;
    const idx = OPTIONS.findIndex((o) => o.value === theme);
    const next = (idx + delta + OPTIONS.length) % OPTIONS.length;
    const opt = OPTIONS[next];
    if (!opt) return;
    setTheme(opt.value);
    radioRefs.current[next]?.focus();
  };

  return (
    <div ref={wrapRef} className="relative">
      <button
        ref={triggerRef}
        type="button"
        aria-label="Change theme"
        aria-haspopup="true"
        aria-expanded={open}
        onClick={() => setOpen((v) => !v)}
        className="flex h-[34px] w-[34px] items-center justify-center rounded-sm border border-border bg-card text-muted transition-colors hover:bg-card-2 hover:text-text"
      >
        <PrefIcon pref={theme} />
      </button>
      {open && (
        <div
          role="radiogroup"
          aria-label="Color theme"
          className="absolute right-0 top-[42px] z-50 w-40 animate-fade-in rounded-md border border-border bg-card p-1 shadow-pop"
        >
          {OPTIONS.map((o, i) => {
            const selected = o.value === theme;
            return (
              <button
                key={o.value}
                ref={(el) => {
                  radioRefs.current[i] = el;
                }}
                type="button"
                role="radio"
                aria-checked={selected}
                tabIndex={selected ? 0 : -1}
                onKeyDown={onRadioKeyDown}
                onClick={() => {
                  setTheme(o.value);
                  close(true);
                }}
                className={cn(
                  'flex w-full items-center gap-2.5 rounded-sm px-2.5 py-2 text-body transition-colors',
                  selected ? 'bg-primary-soft text-primary-strong' : 'text-muted hover:bg-card-2 hover:text-text',
                )}
              >
                {o.icon}
                {o.label}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
