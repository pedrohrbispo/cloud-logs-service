import { useEffect, useRef, useState } from 'react';
import { Check, Copy, FileText, Link2, X } from 'lucide-react';
import { Modal } from '@/components/ui/Modal';
import { IconButton } from '@/components/ui/IconButton';
import { Button } from '@/components/ui/Button';
import { Select } from '@/components/ui/Select';
import { useTempLink } from '@/hooks/useTempLink';
import { TTL_OPTIONS } from '@/lib/constants';
import { formatRelativeTime } from '@/lib/format';
import { cn } from '@/lib/cn';
import type { LogFileView, PresignData } from '@/api/types';

interface TempLinkModalProps {
  open: boolean;
  file: LogFileView | null;
  provider: string;
  bucket: string;
  onClose: () => void;
}

const DEFAULT_TTL = 900;

/**
 * Admin-only: generates a presigned URL; the helper text derives the blast radius
 * from the server's actual `expires_at`, and the URL is never fetch()'d (CORS).
 */
export function TempLinkModal({ open, file, provider, bucket, onClose }: TempLinkModalProps) {
  const mutation = useTempLink();
  const [ttl, setTtl] = useState<number>(DEFAULT_TTL);
  const [data, setData] = useState<PresignData | null>(null);
  const [copied, setCopied] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const copyTimer = useRef<number | null>(null);

  const fileKey = file?.key;

  useEffect(() => {
    setTtl(DEFAULT_TTL);
    setData(null);
    setCopied(false);
    setError(null);
    mutation.reset();
    return () => {
      if (copyTimer.current !== null) window.clearTimeout(copyTimer.current);
    };
    // mutation is a fresh object each render; reset is stable. Deps are open/file.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, fileKey]);

  if (!file) return null;

  const onGenerate = async () => {
    setError(null);
    try {
      const result = await mutation.mutateAsync({ provider, bucket, key: file.key, ttlSeconds: ttl });
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not generate the link. Please try again.');
    }
  };

  const onCopy = async () => {
    if (!data) return;
    try {
      await navigator.clipboard.writeText(data.url);
      setCopied(true);
      if (copyTimer.current !== null) window.clearTimeout(copyTimer.current);
      copyTimer.current = window.setTimeout(() => setCopied(false), 1600);
    } catch {
      setError('Could not copy to the clipboard.');
    }
  };

  const capped = data !== null && data.ttl_seconds < ttl;

  return (
    <Modal open={open} onClose={onClose} titleId="templink-title">
      <div className="flex items-center justify-between border-b border-border p-[18px]">
        <div className="flex items-center gap-2.5">
          <span className="flex h-9 w-9 items-center justify-center rounded-sm bg-primary-soft text-primary">
            <Link2 size={18} aria-hidden="true" />
          </span>
          <h2 id="templink-title" className="text-section text-text">
            Temporary link
          </h2>
        </div>
        <IconButton size={30} aria-label="Close dialog" onClick={onClose}>
          <X size={16} aria-hidden="true" />
        </IconButton>
      </div>

      <div className="space-y-4 p-[18px]">
        <div className="flex items-center gap-2.5 rounded-sm bg-card-2 px-3 py-2.5">
          <FileText size={16} className="shrink-0 text-faint" aria-hidden="true" />
          <span className="truncate font-mono text-name text-text">{file.filename}</span>
        </div>

        <div className="space-y-1.5">
          <label htmlFor="templink-ttl" className="block text-label text-text">
            Link expires in
          </label>
          <Select
            id="templink-ttl"
            value={ttl}
            onChange={(e) => {
              // Changing the TTL invalidates any already-generated link, so the
              // "capped" caption never describes a stale link.
              setTtl(Number(e.target.value));
              setData(null);
              setError(null);
            }}
          >
            {TTL_OPTIONS.map((o) => (
              <option key={o.seconds} value={o.seconds}>
                {o.label}
              </option>
            ))}
          </Select>
        </div>

        {data === null ? (
          <Button fullWidth loading={mutation.isPending} onClick={onGenerate}>
            Generate link
          </Button>
        ) : (
          <div className="space-y-1.5">
            <span className="block text-label text-text">Shareable link</span>
            <div className="flex gap-2">
              <input
                readOnly
                value={data.url}
                onFocus={(e) => e.currentTarget.select()}
                aria-label="Shareable link"
                className="input w-full min-w-0 rounded-sm border border-border bg-bg px-[11px] py-[9px] font-mono text-hint text-text"
              />
              <Button
                variant="outline"
                onClick={onCopy}
                className={cn(copied && 'border-success bg-success-soft text-success-strong')}
                leftIcon={
                  copied ? <Check size={15} aria-hidden="true" /> : <Copy size={15} aria-hidden="true" />
                }
              >
                {copied ? 'Copied' : 'Copy'}
              </Button>
            </div>
            <p className="text-hint text-faint">
              {`Anyone with this link can download the file until it expires ${formatRelativeTime(
                data.expires_at,
              )}.`}
              {capped ? ' (capped at the maximum allowed TTL)' : ''}
            </p>
          </div>
        )}

        {error && (
          <p role="alert" className="text-hint text-error-strong">
            {error}
          </p>
        )}
      </div>

      <span className="sr-only" aria-live="polite">
        {copied ? 'Link copied' : ''}
      </span>
    </Modal>
  );
}
