import { Link } from 'react-router-dom';
import { FileText } from 'lucide-react';
import { Card } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { formatModified, getLevelMeta } from '@/lib/format';
import type { LogFileView } from '@/api/types';

export function RecentFiles({ files }: { files: LogFileView[] }) {
  const recent = [...files]
    .sort((a, b) => b.last_modified.localeCompare(a.last_modified))
    .slice(0, 6);

  return (
    <Card className="overflow-hidden">
      <div className="flex items-center justify-between border-b border-border p-[16px_18px]">
        <h2 className="text-section text-text">Recent log files</h2>
        <Link to="/logs" className="text-meta text-primary-strong hover:underline">
          View all logs
        </Link>
      </div>

      {recent.length === 0 ? (
        <div className="px-[18px] py-6 text-center text-muted">No log files</div>
      ) : (
        <div className="divide-y divide-border">
          {recent.map((f) => {
            const meta = getLevelMeta(f.level);
            return (
              <div key={f.key} className="flex items-center gap-3 px-[18px] py-3">
                <span
                  className={`flex h-9 w-9 shrink-0 items-center justify-center rounded-sm ${meta.softClass}`}
                >
                  <FileText size={16} className={meta.iconClass} aria-hidden="true" />
                </span>
                <span className="flex min-w-0 flex-1 items-center gap-2">
                  <span className="truncate font-mono text-name text-text">{f.filename}</span>
                  <Badge level={f.level} />
                </span>
                <span className="shrink-0 whitespace-nowrap text-meta text-muted">
                  {f.size_human} · {formatModified(f.last_modified)}
                </span>
              </div>
            );
          })}
        </div>
      )}
    </Card>
  );
}
