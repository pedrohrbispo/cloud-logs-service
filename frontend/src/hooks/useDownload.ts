import { useMutation } from '@tanstack/react-query';
import { downloadToBlob } from '@/api/client';
import { downloadPath } from '@/api/logs';
import { useToastStore } from '@/store/toastStore';
import { basename } from '@/lib/format';

export interface DownloadVars {
  provider: string;
  bucket: string;
  key: string;
}

export function useDownload() {
  const show = useToastStore((s) => s.show);
  const complete = useToastStore((s) => s.complete);
  const error = useToastStore((s) => s.error);

  return useMutation({
    mutationFn: async ({ provider, bucket, key }: DownloadVars) => {
      const filename = basename(key);
      await downloadToBlob(downloadPath(provider, bucket, key), filename);
      return filename;
    },
    onMutate: ({ key }) => show(basename(key)),
    onSuccess: (filename) => complete(filename),
    onError: (_err, { key }) => error(basename(key)),
  });
}
