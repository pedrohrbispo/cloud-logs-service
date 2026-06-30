import { apiClient } from './client';
import type { AuthUser, LoginData, LoginReq } from './types';

export function login(email: string, password: string): Promise<LoginData> {
  const body: LoginReq = { email, password };
  // skipAuthRedirect: a 401 here is INVALID_CREDENTIALS, not a lost session —
  // it must surface inline on the login form, not trigger a redirect.
  return apiClient.post<LoginData>('/auth/login', body, { skipAuthRedirect: true });
}

/** Re-hydrate/validate the session on reload; a revoked/expired token → 401 routes to the session-expired redirect. */
export function getMe(): Promise<AuthUser> {
  return apiClient.get<AuthUser>('/me');
}

/** Fire-and-forget so the server emits its audit event before we clear state. */
export async function logout(): Promise<void> {
  try {
    await apiClient.post<unknown>('/auth/logout', {});
  } catch {
    /* best-effort; local session is cleared regardless */
  }
}
