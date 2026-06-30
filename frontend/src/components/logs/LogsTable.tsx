import { Card } from '@/components/ui/Card';
import { LogsTableRow } from './LogsTableRow';
import type { LogFileView } from '@/api/types';

interface LogsTableProps {
  files: LogFileView[];
  bucket: string;
  isAdmin: boolean;
  count: number;
  onDownload: (file: LogFileView) => void;
  onLink: (file: LogFileView) => void;
  isDownloading: (key: string) => boolean;
}

const TH = 'px-[18px] py-[11px] text-eyebrow uppercase text-faint';

export function LogsTable({ files, bucket, isAdmin, count, onDownload, onLink, isDownloading }: LogsTableProps) {
  return (
    <Card className="overflow-hidden">
      <table className="w-full table-fixed border-collapse" aria-label={`Log files in ${bucket}`}>
        <colgroup>
          <col style={{ width: '50%' }} />
          <col style={{ width: '14%' }} />
          <col style={{ width: '22%' }} />
          <col style={{ width: '130px' }} />
        </colgroup>
        <thead>
          <tr className="bg-card-2">
            <th scope="col" className={`text-left ${TH}`}>
              Name
            </th>
            <th scope="col" className={`text-left ${TH}`}>
              Size
            </th>
            <th scope="col" className={`text-left ${TH}`}>
              Modified
            </th>
            <th scope="col" className={`text-right ${TH}`}>
              <span className="sr-only">Actions</span>
            </th>
          </tr>
        </thead>
        <tbody>
          {files.map((file) => (
            <LogsTableRow
              key={file.key}
              file={file}
              isAdmin={isAdmin}
              onDownload={onDownload}
              onLink={onLink}
              isDownloading={isDownloading}
            />
          ))}
        </tbody>
      </table>
      <div className="border-t border-border px-[18px] py-[10px] text-meta text-faint">{count} files</div>
    </Card>
  );
}
