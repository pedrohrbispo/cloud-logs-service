import { forwardRef, type InputHTMLAttributes, type ReactNode } from 'react';
import { cn } from '@/lib/cn';

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  invalid?: boolean;
  leadingIcon?: ReactNode;
}

/**
 * The `.input` class drives the focus ring (see base.css).
 * forwardRef so React Hook Form's `register()` can attach.
 */
export const Input = forwardRef<HTMLInputElement, InputProps>(function Input(
  { invalid, leadingIcon, className, ...rest },
  ref,
) {
  const field = (
    <input
      ref={ref}
      aria-invalid={invalid || undefined}
      className={cn(
        'input w-full rounded-sm border bg-bg px-[11px] py-[9px] text-body text-text placeholder:text-faint',
        invalid ? 'border-error' : 'border-border',
        leadingIcon && 'pl-9',
        className,
      )}
      {...rest}
    />
  );

  if (!leadingIcon) return field;
  return (
    <div className="relative w-full">
      <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-faint" aria-hidden="true">
        {leadingIcon}
      </span>
      {field}
    </div>
  );
});
