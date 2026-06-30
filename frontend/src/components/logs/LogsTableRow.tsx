import { Download, FileText } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { LogRowActions } from './LogRowActions';
import { formatModified, getLevelMeta } from '@/lib/format';
import type { LogFileView } from '@/api/types';

interface LogsTableRowProps {
  file: LogFileView;
  isAdmin: boolean;
  onDownload: (file: LogFileView) => void;
  onLink: (file: LogFileView) => void;
  isDownloading: (key: string) => boolean;
}

export function LogsTableRow({ file, isAdmin, onDownload, onLink, isDownloading }: LogsTableRowProps) {
  const meta = getLevelMeta(file.level);

  return (
    <tr className="border-t border-border transition-colors hover:bg-card-2">
      <th scope="row" className="px-[18px] py-[12px] text-left align-middle font-normal">
        <div className="flex min-w-0 items-center gap-2.5">
          <FileText size={18} className={`${meta.iconClass} shrink-0`} aria-hidden="true" />
          <span className="truncate font-mono text-name text-text">{file.filename}</span>
        </div>
      </th>
      <td className="px-[18px] py-[12px] align-middle text-cell text-muted">{file.size_human}</td>
      <td className="px-[18px] py-[12px] align-middle">
        <time
          dateTime={file.last_modified}
          title={new Date(file.last_modified).toLocaleString()}
          className="text-cell text-muted"
        >
          {formatModified(file.last_modified)}
        </time>
      </td>
      <td className="px-[18px] py-[12px] align-middle text-right">
        <div className="flex items-center justify-end gap-1.5">
          <Button
            variant="outline"
            size="sm"
            onClick={() => onDownload(file)}
            loading={isDownloading(file.key)}
            aria-label={`Download ${file.filename}`}
            leftIcon={<Download size={14} aria-hidden="true" />}
          >
            Download
          </Button>
          {isAdmin && <LogRowActions onLink={() => onLink(file)} filename={file.filename} />}
        </div>
      </td>
    </tr>
  );
}
