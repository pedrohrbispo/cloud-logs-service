import { useRef, type KeyboardEvent, type ReactNode } from 'react';
import { cn } from '@/lib/cn';

export interface Segment<T extends string> {
  value: T;
  label: string;
  icon?: ReactNode;
}

interface SegmentedControlProps<T extends string> {
  options: Segment<T>[];
  value: T;
  onChange: (value: T) => void;
  'aria-label': string;
  size?: 'sm' | 'md';
  className?: string;
}

export function SegmentedControl<T extends string>({
  options,
  value,
  onChange,
  size = 'sm',
  className,
  ...rest
}: SegmentedControlProps<T>) {
  const refs = useRef<(HTMLButtonElement | null)[]>([]);

  const move = (delta: number) => {
    const idx = options.findIndex((o) => o.value === value);
    const next = (idx + delta + options.length) % options.length;
    const opt = options[next];
    if (!opt) return;
    onChange(opt.value);
    refs.current[next]?.focus();
  };

  const onKeyDown = (e: KeyboardEvent<HTMLButtonElement>) => {
    if (e.key === 'ArrowRight' || e.key === 'ArrowDown') {
      e.preventDefault();
      move(1);
    } else if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') {
      e.preventDefault();
      move(-1);
    }
  };

  return (
    <div
      role="radiogroup"
      aria-label={rest['aria-label']}
      className={cn('inline-flex gap-1 rounded-sm border border-border bg-card-2 p-1', className)}
    >
      {options.map((o, i) => {
        const selected = o.value === value;
        return (
          <button
            key={o.value}
            ref={(el) => {
              refs.current[i] = el;
            }}
            type="button"
            role="radio"
            aria-checked={selected}
            tabIndex={selected ? 0 : -1}
            onKeyDown={onKeyDown}
            onClick={() => onChange(o.value)}
            className={cn(
              'inline-flex items-center gap-1.5 rounded-[5px] font-medium transition-colors',
              size === 'sm' ? 'px-3 py-1.5 text-meta' : 'px-3.5 py-2 text-body',
              selected ? 'bg-card text-text shadow-card' : 'text-muted hover:text-text',
            )}
          >
            {o.icon}
            {o.label}
          </button>
        );
      })}
    </div>
  );
}
