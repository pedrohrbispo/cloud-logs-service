import { useAuthStore } from '@/store/authStore';
import type { ApiErr, ApiOk } from './types';

export class ApiResponseError extends Error {
  constructor(
    public readonly code: string,
    message: string,
    public readonly status: number,
    public readonly requestId?: string,
  ) {
    super(message);
    this.name = 'ApiResponseError';
  }
}

const API_BASE = (import.meta.env['VITE_API_BASE'] as string | undefined) ?? '/api/v1';

interface RequestOptions extends RequestInit {
  skipAuthRedirect?: boolean;
}

/** Codes that mean "the session is gone" → hard redirect to /login?reason=expired. */
const SESSION_GONE = new Set(['TOKEN_EXPIRED', 'TOKEN_INVALID', 'TOKEN_MISSING']);

function redirectExpired(): void {
  // Never redirect away from the login page itself (would clobber an inline error).
  if (location.pathname === '/login') return;
  useAuthStore.getState().logout();
  const u = new URL('/login', location.href);
  u.searchParams.set('reason', 'expired');
  location.replace(u.toString());
}

async function parseError(res: Response): Promise<ApiResponseError> {
  let code = `HTTP_${res.status}`;
  let message = res.statusText || 'Request failed';
  let requestId: string | undefined;
  try {
    const body = (await res.json()) as ApiErr;
    if (body?.error) {
      code = body.error.code ?? code;
      message = body.error.message ?? message;
      requestId = body.error.request_id;
    }
  } catch {
    /* non-JSON error body — keep the HTTP fallback */
  }
  return new ApiResponseError(code, message, res.status, requestId);
}

/**
 * Decide what to do with a 401: redirect only for genuine session-loss codes
 * (and never on /login or when explicitly skipped). INVALID_CREDENTIALS from a
 * failed login propagates to the caller so the login form can show it inline.
 */
function handle401(err: ApiResponseError, skip: boolean | undefined): never {
  if (!skip && SESSION_GONE.has(err.code)) redirectExpired();
  throw err;
}

async function request<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const { skipAuthRedirect, headers, ...rest } = opts;
  const token = useAuthStore.getState().token;
  const res = await fetch(`${API_BASE}${path}`, {
    ...rest,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...headers,
    },
  });

  if (res.status === 401) handle401(await parseError(res), skipAuthRedirect);
  if (!res.ok) throw await parseError(res);

  const body = (await res.json()) as ApiOk<T>;
  return body.data;
}

/**
 * Authenticated streaming download (Bearer token → Blob → hidden-anchor save).
 * NOTE: `.blob()` buffers the whole file in memory — fine for log files
 * (<~100MB); production upgrade is ReadableStream + StreamSaver for true streaming.
 */
export async function downloadToBlob(path: string, fallbackName: string): Promise<void> {
  const token = useAuthStore.getState().token;
  const res = await fetch(`${API_BASE}${path}`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });

  if (res.status === 401) handle401(await parseError(res), false);
  if (!res.ok) throw await parseError(res);

  // Prefer the server's Content-Disposition filename, else the key basename.
  const cd = res.headers.get('Content-Disposition');
  const m = cd ? /filename\*?=(?:UTF-8'')?"?([^";]+)"?/i.exec(cd) : null;
  const filename = m?.[1] ? decodeURIComponent(m[1]) : fallbackName;

  const blob = await res.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}

export const apiClient = {
  get: <T>(path: string, opts?: RequestOptions) => request<T>(path, { ...opts, method: 'GET' }),
  post: <T>(path: string, body: unknown, opts?: RequestOptions) =>
    request<T>(path, { ...opts, method: 'POST', body: JSON.stringify(body) }),
  downloadToBlob,
};
