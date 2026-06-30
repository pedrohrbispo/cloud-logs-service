import { LogCard } from './LogCard';
import type { LogFileView } from '@/api/types';

interface LogsCardsProps {
  files: LogFileView[];
  bucket: string;
  isAdmin: boolean;
  onDownload: (file: LogFileView) => void;
  onLink: (file: LogFileView) => void;
  isDownloading: (key: string) => boolean;
}

export function LogsCards({ files, bucket, isAdmin, onDownload, onLink, isDownloading }: LogsCardsProps) {
  return (
    <div className="flex flex-col gap-2.5">
      {files.map((file) => (
        <LogCard
          key={file.key}
          file={file}
          bucket={bucket}
          isAdmin={isAdmin}
          onDownload={onDownload}
          onLink={onLink}
          isDownloading={isDownloading}
        />
      ))}
    </div>
  );
}
