import { useLocation } from 'react-router-dom';
import { useProviders } from '@/hooks/useProviders';
import { useUiStore } from '@/store/uiStore';
import { useAuthStore } from '@/store/authStore';
import { providerLabel } from '@/lib/constants';

export function Breadcrumb() {
  const { pathname } = useLocation();
  const provider = useUiStore((s) => s.provider);
  const allowedProviders = useAuthStore((s) => s.user?.allowed_providers);
  const { data: providers } = useProviders();

  if (!pathname.startsWith('/logs')) {
    return <span className="text-body font-semibold text-text">Dashboard</span>;
  }

  // Omit the bucket segment when the provider is not permitted (the no-permission
  // surface is showing) — there is no bucket to browse. (Plan §7.)
  const isAllowed = !allowedProviders || allowedProviders.includes(provider);
  const current = providers?.find((p) => p.id === provider);
  const bucket = isAllowed ? (current?.buckets[0] ?? '') : '';

  return (
    <nav aria-label="Breadcrumb" className="flex items-center gap-1.5 text-body">
      <span className="font-semibold text-text">Logs</span>
      <span className="text-faint">/</span>
      <span className="text-muted">{providerLabel(provider)}</span>
      {bucket && (
        <>
          <span className="text-faint">/</span>
          <span className="font-mono text-[12px] text-faint">{bucket}</span>
        </>
      )}
    </nav>
  );
}
