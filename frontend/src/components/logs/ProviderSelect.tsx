import { Select } from '@/components/ui/Select';
import { useUiStore } from '@/store/uiStore';
import { useAuthStore } from '@/store/authStore';
import { PROVIDERS } from '@/lib/constants';

/** Unreachable providers keep a "(no access)" suffix but stay selectable, in the locked order, so the RBAC boundary (→ NoPermissionState) is demonstrable. */
export function ProviderSelect() {
  const provider = useUiStore((s) => s.provider);
  const setProvider = useUiStore((s) => s.setProvider);
  const allowed = useAuthStore((s) => s.user?.allowed_providers ?? null);

  return (
    <div className="w-[170px]">
      <Select
        value={provider}
        onChange={(e) => setProvider(e.target.value)}
        aria-label="Cloud provider"
      >
        {PROVIDERS.map((p) => {
          const noAccess = allowed !== null && !allowed.includes(p.id);
          return (
            <option key={p.id} value={p.id}>
              {p.label}
              {noAccess ? ' (no access)' : ''}
            </option>
          );
        })}
      </Select>
    </div>
  );
}
