import { createPortal } from 'react-dom';
import { Check, X } from 'lucide-react';
import { useToastStore } from '@/store/toastStore';
import { Spinner } from './Spinner';

export function Toast() {
  const status = useToastStore((s) => s.status);
  const filename = useToastStore((s) => s.filename);
  const dismiss = useToastStore((s) => s.dismiss);

  if (!status || !filename) return null;

  const name = <span className="font-mono text-meta text-muted">{filename}</span>;

  return createPortal(
    <div
      role={status === 'error' ? 'alert' : 'status'}
      aria-live={status === 'error' ? 'assertive' : 'polite'}
      aria-atomic="true"
      className="fixed bottom-[22px] right-[22px] z-[90] flex min-w-[250px] animate-toast-in items-center gap-3 rounded-md border border-border bg-card p-[13px_16px] shadow-pop"
    >
      {status === 'progress' && (
        <>
          <Spinner size={18} />
          <span className="text-body text-text">Downloading {name}…</span>
        </>
      )}
      {status === 'done' && (
        <>
          <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-success text-white">
            <Check size={13} strokeWidth={3} aria-hidden="true" />
          </span>
          <span className="text-body text-text">{name} downloaded</span>
        </>
      )}
      {status === 'error' && (
        <>
          <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-error-soft text-error">
            <X size={13} strokeWidth={3} aria-hidden="true" />
          </span>
          <span className="text-body text-text">Couldn’t download {name}</span>
        </>
      )}
      <button
        type="button"
        onClick={dismiss}
        aria-label="Dismiss notification"
        className="ml-auto text-faint transition-colors hover:text-text"
      >
        <X size={15} aria-hidden="true" />
      </button>
    </div>,
    document.body,
  );
}
