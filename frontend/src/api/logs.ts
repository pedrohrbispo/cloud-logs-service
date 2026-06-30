import { apiClient } from './client';
import type { ListLogsData } from './types';

export interface ListLogsParams {
  prefix?: string;
  cursor?: string;
  limit?: number;
}

export function listLogs(
  provider: string,
  bucket: string,
  params: ListLogsParams = {},
): Promise<ListLogsData> {
  const qs = new URLSearchParams({ bucket });
  if (params.prefix) qs.set('prefix', params.prefix);
  if (params.cursor) qs.set('cursor', params.cursor);
  qs.set('limit', String(params.limit ?? 100));
  return apiClient.get<ListLogsData>(`/providers/${provider}/logs?${qs.toString()}`);
}

export function downloadPath(provider: string, bucket: string, key: string): string {
  const qs = new URLSearchParams({ bucket, key });
  return `/providers/${provider}/download?${qs.toString()}`;
}
