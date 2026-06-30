import { http, HttpResponse, delay } from 'msw';
import type {
  ApiOk,
  AuthUser,
  ListLogsData,
  ListProvidersData,
  LoginData,
  PresignData,
} from '@/api/types';

// ─────────────────────────────────────────────────────────────────────────────
// In-browser mock of the Go BFF /api/v1 contract. DEV/DEMO ONLY — enabled when
// VITE_USE_MOCKS=true so the SPA runs end-to-end with no backend. Mirrors the
// real envelopes, error codes, and the prototype's per-provider state demo
// (aws → data, azure → connection error, gcp → empty).
// ─────────────────────────────────────────────────────────────────────────────

const BASE = '/api/v1';
const meta = () => ({ request_id: `mock-${Math.random().toString(36).slice(2, 10)}`, timestamp: new Date().toISOString() });
const ok = <T>(data: T) => HttpResponse.json<ApiOk<T>>({ data, meta: meta() });
const fail = (status: number, code: string, message: string) =>
  HttpResponse.json({ error: { code, message, request_id: meta().request_id } }, { status });

const USERS: Record<string, { password: string; user: AuthUser }> = {
  'admin@example.com': {
    password: 'admin123',
    user: { id: 'u-admin', email: 'admin@example.com', role: 'admin', allowed_providers: null },
  },
  'viewer@example.com': {
    password: 'viewer123',
    user: { id: 'u-viewer', email: 'viewer@example.com', role: 'viewer', allowed_providers: ['aws'] },
  },
};

const ago = (ms: number) => new Date(Date.now() - ms).toISOString();
const MIN = 60_000;
const HOUR = 60 * MIN;
const DAY = 24 * HOUR;

const AWS_FILES = [
  { key: 'production-logs/payment.log', size: 5_452_595, last_modified: ago(2 * HOUR), etag: '"a1"', content_type: 'text/plain' },
  { key: 'production-logs/auth.log', size: 2_202_010, last_modified: ago(1 * DAY), etag: '"a2"', content_type: 'text/plain' },
  { key: 'production-logs/gateway.log', size: 8_808_038, last_modified: ago(3 * HOUR), etag: '"a3"', content_type: 'text/plain' },
  { key: 'production-logs/orders.log', size: 3_145_728, last_modified: ago(6 * HOUR), etag: '"a4"', content_type: 'text/plain' },
  { key: 'production-logs/worker.log', size: 1_363_149, last_modified: ago(2 * DAY), etag: '"a5"', content_type: 'text/plain' },
];

const requireAuth = (request: Request) => request.headers.get('Authorization')?.startsWith('Bearer ');

export const handlers = [
  http.post(`${BASE}/auth/login`, async ({ request }) => {
    await delay(450);
    const { email, password } = (await request.json()) as { email: string; password: string };
    const record = USERS[email];
    if (!record || record.password !== password) {
      return fail(401, 'INVALID_CREDENTIALS', 'Invalid email or password.');
    }
    const data: LoginData = {
      token: `mock.${record.user.role}.jwt`,
      token_type: 'Bearer',
      expires_in: 3600,
      user: record.user,
    };
    return ok(data);
  }),

  http.post(`${BASE}/auth/logout`, () => ok({ status: 'logged_out' })),

  http.get(`${BASE}/me`, ({ request }) => {
    const auth = request.headers.get('Authorization') ?? '';
    if (!auth.startsWith('Bearer ')) return fail(401, 'TOKEN_MISSING', 'Missing bearer token.');
    const record = auth.includes('viewer') ? USERS['viewer@example.com'] : USERS['admin@example.com'];
    return ok(record!.user);
  }),

  http.get(`${BASE}/providers`, ({ request }) => {
    if (!requireAuth(request)) return fail(401, 'TOKEN_MISSING', 'Missing bearer token.');
    const data: ListProvidersData = {
      providers: [
        { id: 'aws', name: 'AWS S3', buckets: ['production-logs'], healthy: true, region: 'us-east-1' },
        { id: 'azure', name: 'Azure Blob', buckets: ['app-logs-prod'], healthy: false, region: 'eastus', error: 'STORAGE_UNAVAILABLE' },
        { id: 'gcp', name: 'GCP Storage', buckets: ['gcs-prod-logs'], healthy: true, region: 'us-central1' },
      ],
    };
    return ok(data);
  }),

  http.get(`${BASE}/providers/:provider/logs`, async ({ request, params }) => {
    if (!requireAuth(request)) return fail(401, 'TOKEN_MISSING', 'Missing bearer token.');
    await delay(650);
    const provider = params['provider'] as string;
    if (provider === 'azure') return fail(503, 'STORAGE_UNAVAILABLE', 'The storage backend is unavailable.');
    if (provider === 'gcp') return ok<ListLogsData>({ files: [], next_cursor: null, count: 0 });
    return ok<ListLogsData>({ files: AWS_FILES, next_cursor: null, count: AWS_FILES.length });
  }),

  http.get(`${BASE}/providers/:provider/download`, async ({ request }) => {
    if (!requireAuth(request)) return fail(401, 'TOKEN_MISSING', 'Missing bearer token.');
    const url = new URL(request.url);
    const key = url.searchParams.get('key') ?? 'file.log';
    const name = key.split('/').pop() ?? 'file.log';
    await delay(800);
    const body = `# Mock log file: ${name}\n${Array.from({ length: 40 }, (_, i) => `${ago(i * MIN)}  INFO  line ${i + 1} — demo content`).join('\n')}\n`;
    return new HttpResponse(body, {
      headers: {
        'Content-Type': 'text/plain',
        'Content-Disposition': `attachment; filename="${name}"`,
      },
    });
  }),

  http.post(`${BASE}/providers/:provider/presign`, async ({ request, params }) => {
    if (!requireAuth(request)) return fail(401, 'TOKEN_MISSING', 'Missing bearer token.');
    await delay(500);
    const provider = params['provider'] as string;
    const url = new URL(request.url);
    const bucket = url.searchParams.get('bucket') ?? 'production-logs';
    const { key, ttl_seconds } = (await request.json()) as { key: string; ttl_seconds: number };
    const clamped = Math.min(ttl_seconds, 900); // PRESIGN_MAX_TTL = 15m
    const data: PresignData = {
      url: `https://logs.cloudaccess.io/${provider}/${bucket}/${key}?X-Sig=${Math.random().toString(36).slice(2)}&expires=${clamped}`,
      expires_at: new Date(Date.now() + clamped * 1000).toISOString(),
      ttl_seconds: clamped,
    };
    return ok(data);
  }),
];
