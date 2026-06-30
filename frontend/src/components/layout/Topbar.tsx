import { ThemeToggle } from '@/components/ui/ThemeToggle';
import { Breadcrumb } from './Breadcrumb';
import { UserChip } from './UserChip';

export function Topbar() {
  return (
    <header className="sticky top-0 z-20 flex h-topbar items-center justify-between border-b border-border bg-card px-page">
      <Breadcrumb />
      <div className="flex items-center gap-3">
        <ThemeToggle variant="icon" />
        <div className="h-[22px] w-px bg-border" aria-hidden="true" />
        <UserChip />
      </div>
    </header>
  );
}
