import { Download, FileText } from 'lucide-react';
import { Card } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { LogRowActions } from './LogRowActions';
import { formatModified, getLevelMeta } from '@/lib/format';
import { cn } from '@/lib/cn';
import type { LogFileView } from '@/api/types';

interface LogCardProps {
  file: LogFileView;
  bucket: string;
  isAdmin: boolean;
  onDownload: (file: LogFileView) => void;
  onLink: (file: LogFileView) => void;
  isDownloading: (key: string) => boolean;
}

export function LogCard({ file, bucket, isAdmin, onDownload, onLink, isDownloading }: LogCardProps) {
  const meta = getLevelMeta(file.level);

  return (
    <Card className="flex items-center gap-3.5 p-[14px_16px] transition-colors hover:border-primary">
      <div
        className={cn(
          'flex h-[42px] w-[42px] shrink-0 items-center justify-center rounded-[9px]',
          meta.softClass,
        )}
      >
        <FileText size={20} className={meta.iconClass} aria-hidden="true" />
      </div>

      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2.5">
          <span className="truncate font-mono text-[14px] font-medium text-text">{file.filename}</span>
          <Badge level={file.level} className="shrink-0" />
        </div>
        <div className="text-meta text-muted">
          {`${file.size_human} · Modified ${formatModified(file.last_modified)} · `}
          <span className="font-mono">{bucket}</span>
        </div>
      </div>

      <div className="flex shrink-0 items-center gap-2">
        {isAdmin && <LogRowActions onLink={() => onLink(file)} filename={file.filename} size={34} />}
        <Button
          onClick={() => onDownload(file)}
          loading={isDownloading(file.key)}
          aria-label={`Download ${file.filename}`}
          leftIcon={<Download size={15} aria-hidden="true" />}
        >
          Download
        </Button>
      </div>
    </Card>
  );
}
