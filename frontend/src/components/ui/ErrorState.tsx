import { AlertTriangle, RefreshCw } from 'lucide-react';
import { EmptyState } from './EmptyState';
import { Button } from './Button';

interface ErrorStateProps {
  providerLabel: string;
  onRetry: () => void;
  message?: string;
}

/** Transient connection failure (503 / network) — offers Retry. */
export function ErrorState({ providerLabel, onRetry, message }: ErrorStateProps) {
  return (
    <EmptyState
      tone="error"
      role="status"
      icon={<AlertTriangle size={24} strokeWidth={1.9} aria-hidden="true" />}
      title={`Unable to connect to ${providerLabel}`}
      message={message ?? 'We couldn’t reach the storage backend. This may be temporary — try again.'}
    >
      <Button variant="outline" onClick={onRetry} leftIcon={<RefreshCw size={15} aria-hidden="true" />}>
        Retry
      </Button>
    </EmptyState>
  );
}
