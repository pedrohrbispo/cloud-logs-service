import { forwardRef, type ButtonHTMLAttributes, type ReactNode } from 'react';
import { cn } from '@/lib/cn';
import { Spinner } from './Spinner';

type Variant = 'primary' | 'outline' | 'ghost' | 'danger';
type Size = 'sm' | 'md';

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  size?: Size;
  fullWidth?: boolean;
  loading?: boolean;
  leftIcon?: ReactNode;
}

const VARIANTS: Record<Variant, string> = {
  primary:
    'bg-primary text-primary-contrast hover:bg-primary-hover border border-transparent shadow-card',
  outline: 'border border-border bg-card text-text hover:border-primary hover:bg-primary-soft hover:text-primary-strong',
  ghost: 'border border-transparent text-muted hover:bg-card-2 hover:text-text',
  danger: 'bg-error text-white hover:opacity-90 border border-transparent',
};

const SIZES: Record<Size, string> = {
  sm: 'h-8 px-3 text-meta gap-1.5 rounded-sm',
  md: 'h-[38px] px-4 text-body gap-2 rounded-sm',
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(function Button(
  { variant = 'primary', size = 'md', fullWidth, loading, leftIcon, className, children, disabled, ...rest },
  ref,
) {
  return (
    <button
      ref={ref}
      disabled={disabled || loading}
      aria-busy={loading || undefined}
      className={cn(
        'inline-flex items-center justify-center font-medium transition-colors disabled:cursor-not-allowed disabled:opacity-60',
        VARIANTS[variant],
        SIZES[size],
        fullWidth && 'w-full',
        className,
      )}
      {...rest}
    >
      {loading ? <Spinner size={size === 'sm' ? 14 : 16} /> : leftIcon}
      {children}
    </button>
  );
});
