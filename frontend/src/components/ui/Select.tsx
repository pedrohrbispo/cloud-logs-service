import { forwardRef, type SelectHTMLAttributes } from 'react';
import { ChevronDown } from 'lucide-react';
import { cn } from '@/lib/cn';

type SelectProps = SelectHTMLAttributes<HTMLSelectElement>;

export const Select = forwardRef<HTMLSelectElement, SelectProps>(function Select(
  { className, children, ...rest },
  ref,
) {
  return (
    <div className="relative inline-flex w-full items-center">
      <select
        ref={ref}
        className={cn(
          'input w-full appearance-none rounded-sm border border-border bg-card py-[9px] pl-3 pr-9 text-body text-text',
          className,
        )}
        {...rest}
      >
        {children}
      </select>
      <ChevronDown
        size={16}
        className="pointer-events-none absolute right-3 text-faint"
        aria-hidden="true"
      />
    </div>
  );
});
