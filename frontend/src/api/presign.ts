import { apiClient } from './client';
import type { PresignData, PresignReq } from './types';

/** Request body is snake_case `{ key, ttl_seconds }` per locked §0. */
export function createPresignedLink(
  provider: string,
  bucket: string,
  key: string,
  ttlSeconds: number,
): Promise<PresignData> {
  const body: PresignReq = { key, ttl_seconds: ttlSeconds };
  const qs = new URLSearchParams({ bucket });
  return apiClient.post<PresignData>(`/providers/${provider}/presign?${qs.toString()}`, body);
}
