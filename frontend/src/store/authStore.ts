import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { AuthUser } from '@/api/types';

const STORAGE_KEY = 'cla_session';

interface AuthState {
  token: string | null;
  user: AuthUser | null;
  isAuthenticated: boolean;
  login: (token: string, user: AuthUser) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      user: null,
      isAuthenticated: false,
      login: (token, user) => set({ token, user, isAuthenticated: true }),
      logout: () => set({ token: null, user: null, isAuthenticated: false }),
    }),
    {
      name: STORAGE_KEY,
      partialize: (s) => ({ token: s.token, user: s.user, isAuthenticated: s.isAuthenticated }),
    },
  ),
);

/**
 * Cross-tab session sync. zustand/persist does not propagate across tabs on its
 * own, so we mirror writes to `cla_session`: a logout in tab A clears tab B, and
 * a login in tab A authenticates tab B. Registered once at module scope.
 *
 * XSS tradeoff of a localStorage JWT is accepted and documented (README) —
 * mitigated by the 1h token TTL and the strict CSP.
 */
if (typeof window !== 'undefined') {
  window.addEventListener('storage', (e) => {
    if (e.key !== STORAGE_KEY) return;
    if (!e.newValue) {
      useAuthStore.setState({ token: null, user: null, isAuthenticated: false });
      return;
    }
    try {
      const parsed = JSON.parse(e.newValue) as { state?: Partial<AuthState> };
      const next = parsed.state;
      if (!next) return;
      useAuthStore.setState({
        token: next.token ?? null,
        user: next.user ?? null,
        isAuthenticated: Boolean(next.isAuthenticated),
      });
    } catch {
      /* ignore malformed storage payloads */
    }
  });
}
