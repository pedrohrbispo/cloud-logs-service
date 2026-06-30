import type { ReactNode } from 'react';
import { Card } from '@/components/ui/Card';

interface StatCardProps {
  eyebrow: string;
  icon: ReactNode;
  value: ReactNode;
  sub?: ReactNode;
}

export function StatCard({ eyebrow, icon, value, sub }: StatCardProps) {
  return (
    <Card className="p-[16px_18px]">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="text-eyebrow uppercase text-faint">{eyebrow}</div>
          <div className="mt-2 text-stat text-text">{value}</div>
        </div>
        <span aria-hidden="true" className="shrink-0 text-faint">
          {icon}
        </span>
      </div>
      {sub ? <div className="mt-2 text-meta text-muted">{sub}</div> : null}
    </Card>
  );
}
