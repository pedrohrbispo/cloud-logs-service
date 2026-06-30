import { Card } from './Card';

const COLS = 'grid grid-cols-[minmax(0,2.4fr)_0.8fr_1.4fr_130px] items-center gap-3';

export function SkeletonTable({ rows = 5 }: { rows?: number }) {
  return (
    <Card className="overflow-hidden" aria-hidden="true">
      <div className={`${COLS} border-b border-border bg-card-2 px-[18px] py-[11px]`}>
        <span className="skeleton h-2.5 w-12" />
        <span className="skeleton h-2.5 w-10" />
        <span className="skeleton h-2.5 w-16" />
        <span className="skeleton h-2.5 w-14 justify-self-end" />
      </div>
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className={`${COLS} border-b border-border px-[18px] py-[14px] last:border-b-0`}>
          <div className="flex items-center gap-2.5">
            <span className="skeleton h-5 w-5 rounded" />
            <span className="skeleton h-3 w-40" />
          </div>
          <span className="skeleton h-3 w-12" />
          <span className="skeleton h-3 w-24" />
          <span className="skeleton h-7 w-[110px] justify-self-end rounded-sm" />
        </div>
      ))}
    </Card>
  );
}
