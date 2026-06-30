import { getLevelMeta } from '@/lib/format';
import { cn } from '@/lib/cn';
import type { LogLevel } from '@/api/types';

interface BadgeProps {
  level: LogLevel;
  className?: string;
}

export function Badge({ level, className }: BadgeProps) {
  const meta = getLevelMeta(level);
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-pill px-2 py-0.5 text-badge uppercase',
        meta.softClass,
        meta.textClass,
        className,
      )}
    >
      {meta.label}
    </span>
  );
}
