import { useAuthStore } from '@/store/authStore';
import { RoleBadge } from './RoleBadge';

export function UserChip() {
  const user = useAuthStore((s) => s.user);
  if (!user) return null;

  const initial = user.email.charAt(0).toUpperCase();
  const localPart = user.email.split('@')[0] ?? user.email;

  return (
    <div className="flex items-center gap-2.5">
      <span className="flex h-[30px] w-[30px] items-center justify-center rounded-pill bg-primary text-[13px] font-semibold text-primary-contrast">
        {initial}
      </span>
      <div className="hidden flex-col items-start leading-tight sm:flex">
        <span className="text-body font-medium leading-tight text-text">{localPart}</span>
        <RoleBadge role={user.role} />
      </div>
    </div>
  );
}
