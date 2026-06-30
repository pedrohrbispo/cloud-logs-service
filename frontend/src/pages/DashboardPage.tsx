import { Database, FileText, Folder, HardDrive } from 'lucide-react';
import { useProviders } from '@/hooks/useProviders';
import { useLogs } from '@/hooks/useLogs';
import { StatCard } from '@/components/dashboard/StatCard';
import { RecentFiles } from '@/components/dashboard/RecentFiles';
import { ProvidersPanel } from '@/components/dashboard/ProvidersPanel';
import { Card } from '@/components/ui/Card';
import { Skeleton } from '@/components/ui/Skeleton';
import { providerLabel } from '@/lib/constants';
import { formatBytes } from '@/lib/format';
import type { ProviderInfo, LogFileView } from '@/api/types';

/**
 * Overview composed entirely from real endpoints (/providers + the primary
 * provider's /logs) — there is no /summary endpoint on the BFF.
 */
export function DashboardPage() {
  const { data: providers, isLoading: providersLoading, isError } = useProviders();
  const primary = providers?.[0];
  const primaryBucket = primary?.buckets[0] ?? '';
  const logs = useLogs(primary?.id ?? '', primaryBucket, Boolean(primary) && Boolean(primaryBucket));

  const files = logs.data?.files ?? [];
  const totalBytes = files.reduce((sum, f) => sum + f.size, 0);
  const loading = providersLoading || (Boolean(primary) && Boolean(primaryBucket) && logs.isLoading);

  return (
    <div className="animate-fade-in">
      <header className="mb-5">
        <h1 className="text-title text-text">Dashboard</h1>
        <p className="text-body text-muted">Overview of your cloud log storage</p>
      </header>

      {isError ? (
        <Card className="p-6 text-center text-muted">Couldn’t load provider information.</Card>
      ) : loading ? (
        <DashboardSkeleton />
      ) : !primary ? (
        <Card className="p-6 text-center text-muted">No providers available for your account.</Card>
      ) : (
        <DashboardContent
          primary={primary}
          providers={providers ?? []}
          files={files}
          totalBytes={totalBytes}
        />
      )}
    </div>
  );
}

function DashboardContent({
  primary,
  providers,
  files,
  totalBytes,
}: {
  primary: ProviderInfo;
  providers: ProviderInfo[];
  files: LogFileView[];
  totalBytes: number;
}) {
  const bucket = primary.buckets[0] ?? '—';
  return (
    <>
      <div className="mb-3.5 grid grid-cols-2 gap-3.5 cards:grid-cols-4">
        <StatCard
          eyebrow="Cloud Provider"
          icon={<Database size={18} strokeWidth={1.8} />}
          value={providerLabel(primary.id)}
          sub={
            <span className="inline-flex items-center gap-1.5">
              <span
                aria-hidden="true"
                className={`inline-block h-1.5 w-1.5 rounded-pill ${primary.healthy ? 'bg-success' : 'bg-error'}`}
              />
              {primary.healthy ? 'Connected' : 'Unreachable'}
              {primary.region ? ` · ${primary.region}` : ''}
            </span>
          }
        />
        <StatCard
          eyebrow="Bucket"
          icon={<Folder size={18} strokeWidth={1.8} />}
          value={<span className="font-mono text-bucket">{bucket}</span>}
          sub="Primary bucket"
        />
        <StatCard
          eyebrow="Log Files"
          icon={<FileText size={18} strokeWidth={1.8} />}
          value={files.length}
          sub={`in ${bucket}`}
        />
        <StatCard
          eyebrow="Total Size"
          icon={<HardDrive size={18} strokeWidth={1.8} />}
          value={formatBytes(totalBytes)}
          sub={`across ${files.length} file${files.length === 1 ? '' : 's'}`}
        />
      </div>

      <div className="grid grid-cols-1 gap-3.5 cards:grid-cols-[1.6fr_1fr]">
        <RecentFiles files={files} />
        <ProvidersPanel providers={providers} />
      </div>
    </>
  );
}

const STAT_KEYS = ['provider', 'bucket', 'files', 'size'] as const;
const ROW_KEYS = ['a', 'b', 'c', 'd'] as const;

function DashboardSkeleton() {
  return (
    <>
      <div className="mb-3.5 grid grid-cols-2 gap-3.5 cards:grid-cols-4">
        {STAT_KEYS.map((key) => (
          <Card key={key} className="p-[16px_18px]">
            <div className="flex items-start justify-between gap-3">
              <Skeleton className="h-3 w-20" />
              <Skeleton className="h-4 w-4 rounded-sm" />
            </div>
            <Skeleton className="mt-3 h-6 w-28" />
            <Skeleton className="mt-3 h-3 w-24" />
          </Card>
        ))}
      </div>

      <div className="grid grid-cols-1 gap-3.5 cards:grid-cols-[1.6fr_1fr]">
        {[0, 1].map((panel) => (
          <Card key={panel} className="overflow-hidden">
            <div className="border-b border-border p-[16px_18px]">
              <Skeleton className="h-4 w-32" />
            </div>
            <div className="space-y-3.5 p-[16px_18px]">
              {ROW_KEYS.map((key) => (
                <div key={key} className="flex items-center gap-3">
                  <Skeleton className="h-9 w-9 shrink-0 rounded-sm" />
                  <Skeleton className="h-3 flex-1" />
                  <Skeleton className="h-3 w-12 shrink-0" />
                </div>
              ))}
            </div>
          </Card>
        ))}
      </div>
    </>
  );
}
