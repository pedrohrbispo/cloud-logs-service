import { cn } from '@/lib/cn';

interface SpinnerProps {
  size?: number;
  className?: string;
}

/** Decorative spinner — callers carry the accessible loading text. */
export function Spinner({ size = 16, className }: SpinnerProps) {
  return (
    <span
      aria-hidden="true"
      className={cn(
        'inline-block shrink-0 animate-spin rounded-full border-2 border-border border-t-primary',
        className,
      )}
      style={{ width: size, height: size }}
    />
  );
}
