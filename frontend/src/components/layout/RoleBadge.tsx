import { cn } from '@/lib/cn';

interface RoleBadgeProps {
  role: 'admin' | 'viewer';
}

export function RoleBadge({ role }: RoleBadgeProps) {
  const isAdmin = role === 'admin';
  return (
    <span
      className={cn(
        'inline-flex rounded-pill px-1.5 py-0.5 text-badge uppercase',
        isAdmin ? 'bg-success-soft text-success-strong' : 'bg-card-2 text-muted',
      )}
    >
      {isAdmin ? 'Admin' : 'Viewer'}
    </span>
  );
}
