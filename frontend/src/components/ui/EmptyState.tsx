import type { ReactNode } from 'react';
import { Card } from './Card';
import { cn } from '@/lib/cn';

export type StateTone = 'muted' | 'primary' | 'warn' | 'error';

interface EmptyStateProps {
  icon: ReactNode;
  tone?: StateTone;
  title: string;
  message?: string;
  /** 'status' announces the panel to screen readers (used by no-permission / error). */
  role?: 'status';
  children?: ReactNode;
}

const TONE_CIRCLE: Record<StateTone, string> = {
  muted: 'bg-card-2 text-faint',
  primary: 'bg-primary-soft text-primary',
  warn: 'bg-warn-soft text-warn',
  error: 'bg-error-soft text-error',
};

export function EmptyState({ icon, tone = 'muted', title, message, role, children }: EmptyStateProps) {
  return (
    <Card
      className="flex animate-fade-in flex-col items-center px-6 py-[52px] text-center"
      {...(role ? { role } : {})}
    >
      <div
        className={cn('mb-4 flex h-[52px] w-[52px] items-center justify-center rounded-pill', TONE_CIRCLE[tone])}
      >
        {icon}
      </div>
      <h3 className="text-section text-text">{title}</h3>
      {message && <p className="mx-auto mt-1.5 max-w-[360px] text-body text-muted">{message}</p>}
      {children && <div className="mt-5 flex items-center gap-2.5">{children}</div>}
    </Card>
  );
}
