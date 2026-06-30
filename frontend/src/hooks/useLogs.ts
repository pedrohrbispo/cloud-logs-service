import { useQuery } from '@tanstack/react-query';
import { listLogs } from '@/api/logs';
import { toLogFileView } from '@/lib/format';
import type { LogFileView } from '@/api/types';

export interface LogsResult {
  files: LogFileView[];
  count: number;
  next_cursor: string | null;
}

/**
 * Disabled until both `provider` and `bucket` are known and the provider is
 * allowed. `retry: false` — the user retries manually via the Retry button.
 */
export function useLogs(provider: string, bucket: string, enabled: boolean) {
  return useQuery({
    queryKey: ['logs', provider, bucket],
    queryFn: () => listLogs(provider, bucket),
    enabled: enabled && !!provider && !!bucket,
    staleTime: 30_000,
    retry: false,
    select: (data): LogsResult => ({
      files: data.files.map(toLogFileView),
      count: data.count,
      next_cursor: data.next_cursor,
    }),
  });
}
