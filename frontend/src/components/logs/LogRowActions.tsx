import { Link2 } from 'lucide-react';
import { IconButton } from '@/components/ui/IconButton';

interface LogRowActionsProps {
  onLink: () => void;
  filename: string;
  size?: 30 | 34;
}

/** Admin-only: callers gate with `isAdmin &&`, so this is DOM-absent for viewers (RBAC). */
export function LogRowActions({ onLink, filename, size = 30 }: LogRowActionsProps) {
  return (
    <IconButton
      size={size}
      aria-label={`Create temporary link for ${filename}`}
      onClick={onLink}
    >
      <Link2 size={15} aria-hidden="true" />
    </IconButton>
  );
}
