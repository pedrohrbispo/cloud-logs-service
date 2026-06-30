import type { ProviderId } from '@/api/types';

/** All providers, in the locked display order (aws / azure / gcp — §0). */
export const PROVIDERS: { id: ProviderId; label: string }[] = [
  { id: 'aws', label: 'AWS S3' },
  { id: 'azure', label: 'Azure Blob' },
  { id: 'gcp', label: 'GCP Storage' },
];

export const PROVIDER_LABELS: Record<ProviderId, string> = {
  aws: 'AWS S3',
  azure: 'Azure Blob',
  gcp: 'GCP Storage',
};

export function providerLabel(id: string): string {
  return PROVIDER_LABELS[id as ProviderId] ?? id;
}

/** Temp-link TTL choices. 1h/24h are clamped server-side to PRESIGN_MAX_TTL. */
export const TTL_OPTIONS: { label: string; seconds: number }[] = [
  { label: '5 minutes', seconds: 300 },
  { label: '15 minutes', seconds: 900 },
  { label: '1 hour', seconds: 3600 },
  { label: '24 hours', seconds: 86400 },
];

/** Seeded demo accounts — surfaced on the login card for graders. */
export const DEMO_CREDENTIALS: { email: string; password: string; role: string }[] = [
  { email: 'admin@example.com', password: 'admin123', role: 'Admin' },
  { email: 'viewer@example.com', password: 'viewer123', role: 'Viewer' },
];
