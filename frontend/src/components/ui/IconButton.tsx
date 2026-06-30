import { forwardRef, type ButtonHTMLAttributes } from 'react';
import { cn } from '@/lib/cn';

type Size = 30 | 34;

interface IconButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  /** Required for screen readers — icon-only buttons have no text. */
  'aria-label': string;
  size?: Size;
  active?: boolean;
}

export const IconButton = forwardRef<HTMLButtonElement, IconButtonProps>(function IconButton(
  { size = 34, active, className, children, ...rest },
  ref,
) {
  return (
    <button
      ref={ref}
      className={cn(
        'inline-flex shrink-0 items-center justify-center rounded-sm border border-border bg-card text-muted transition-colors hover:bg-card-2 hover:text-text disabled:cursor-not-allowed disabled:opacity-50',
        active && 'border-primary bg-primary-soft text-primary-strong',
        className,
      )}
      style={{ width: size, height: size }}
      {...rest}
    >
      {children}
    </button>
  );
});
