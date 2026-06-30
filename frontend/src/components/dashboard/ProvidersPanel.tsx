import { Card } from '@/components/ui/Card';
import type { ProviderInfo } from '@/api/types';

export function ProvidersPanel({ providers }: { providers: ProviderInfo[] }) {
  return (
    <Card className="overflow-hidden">
      <div className="border-b border-border p-[16px_18px]">
        <h2 className="text-section text-text">Providers</h2>
      </div>
      <div className="divide-y divide-border">
        {providers.map((p) => {
          const meta = p.buckets[0] ?? '—';
          return (
            <div key={p.id} className="flex items-center justify-between gap-3 px-[18px] py-3">
              <div className="min-w-0">
                <div className="text-body text-text">{p.name}</div>
                <div className="truncate font-mono text-meta text-muted">
                  {meta}
                  {p.region ? ` · ${p.region}` : ''}
                </div>
              </div>
              <span
                className={`inline-flex shrink-0 items-center gap-1.5 rounded-pill px-2 py-0.5 text-badge uppercase ${
                  p.healthy ? 'bg-success-soft text-success-strong' : 'bg-error-soft text-error-strong'
                }`}
              >
                <span
                  aria-hidden="true"
                  className={`inline-block h-1.5 w-1.5 rounded-pill ${p.healthy ? 'bg-success' : 'bg-error'}`}
                />
                {p.healthy ? 'Healthy' : 'Down'}
              </span>
            </div>
          );
        })}
      </div>
    </Card>
  );
}
