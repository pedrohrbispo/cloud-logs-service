import { Lock } from 'lucide-react';
import { EmptyState } from './EmptyState';
import { Button } from './Button';

interface NoPermissionStateProps {
  providerLabel: string;
  switchToLabel?: string;
  onSwitch?: () => void;
}

/**
 * RBAC boundary (403 PROVIDER_NOT_ALLOWED or a preemptive client check).
 * Amber + lock distinguishes "restricted" from red "broken" / gray "empty".
 * No Retry — this is an authorization boundary, not a transient failure.
 */
export function NoPermissionState({ providerLabel, switchToLabel, onSwitch }: NoPermissionStateProps) {
  return (
    <EmptyState
      tone="warn"
      role="status"
      icon={<Lock size={24} strokeWidth={1.9} aria-hidden="true" />}
      title={`You don’t have access to ${providerLabel}`}
      message="Your role doesn’t permit access to this provider. Contact an administrator to request access."
    >
      {onSwitch && switchToLabel && (
        <Button variant="outline" onClick={onSwitch}>
          Switch to {switchToLabel}
        </Button>
      )}
    </EmptyState>
  );
}
