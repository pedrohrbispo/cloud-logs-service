import { useMutation } from '@tanstack/react-query';
import { createPresignedLink } from '@/api/presign';
import type { PresignData } from '@/api/types';

export interface TempLinkVars {
  provider: string;
  bucket: string;
  key: string;
  ttlSeconds: number;
}

export function useTempLink() {
  return useMutation<PresignData, Error, TempLinkVars>({
    mutationFn: ({ provider, bucket, key, ttlSeconds }) =>
      createPresignedLink(provider, bucket, key, ttlSeconds),
  });
}
