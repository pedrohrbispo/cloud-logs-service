import { CloudRain, LayoutGrid, List, LogOut } from 'lucide-react';
import { NavLink, useNavigate } from 'react-router-dom';
import { logout } from '@/api/auth';
import { useAuthStore } from '@/store/authStore';
import { cn } from '@/lib/cn';

const NAV_ITEMS = [
  { to: '/dashboard', label: 'Dashboard', icon: LayoutGrid },
  { to: '/logs', label: 'Logs', icon: List },
] as const;

export function Sidebar() {
  const navigate = useNavigate();
  const logoutStore = useAuthStore((s) => s.logout);

  const onLogout = async () => {
    await logout(); // fire-and-forget audit event; best-effort
    logoutStore();
    navigate('/login');
  };

  return (
    <nav
      aria-label="Primary"
      className="sticky top-0 flex h-screen w-sidebar shrink-0 flex-col border-r border-border bg-sidebar"
    >
      <div className="flex items-center gap-2.5 p-[18px]">
        <CloudRain size={22} className="text-primary" aria-hidden="true" />
        <span className="text-[15px] font-bold text-text">Cloud Log Access</span>
      </div>

      <p className="my-2.5 px-[18px] text-eyebrow uppercase text-faint">Navigation</p>

      <ul className="flex flex-col gap-[3px]">
        {NAV_ITEMS.map(({ to, label, icon: Icon }) => (
          <li key={to} className="mx-2.5">
            <NavLink
              to={to}
              className={({ isActive }) =>
                cn(
                  'flex items-center gap-2.5 rounded-sm px-2.5 py-2 text-body transition-colors',
                  isActive
                    ? 'bg-primary-soft font-semibold text-primary-strong'
                    : 'font-medium text-muted hover:bg-card-2 hover:text-text',
                )
              }
            >
              <Icon size={18} aria-hidden="true" />
              {label}
            </NavLink>
          </li>
        ))}
      </ul>

      <div className="flex-1" />

      <div className="border-t border-border p-2.5">
        <button
          type="button"
          onClick={onLogout}
          className="flex w-full items-center gap-2.5 rounded-sm px-2.5 py-2 text-body font-medium text-muted transition-colors hover:bg-card-2 hover:text-error"
        >
          <LogOut size={18} aria-hidden="true" />
          Logout
        </button>
      </div>
    </nav>
  );
}
