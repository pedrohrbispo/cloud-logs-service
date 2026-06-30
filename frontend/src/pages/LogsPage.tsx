import { useMemo, useState } from 'react';
import { FileX2 } from 'lucide-react';
import { useUiStore } from '@/store/uiStore';
import { useAuthStore } from '@/store/authStore';
import { useIsAdmin } from '@/hooks/useIsAdmin';
import { useProviders } from '@/hooks/useProviders';
import { useLogs } from '@/hooks/useLogs';
import { useDownload } from '@/hooks/useDownload';
import { ApiResponseError } from '@/api/client';
import { providerLabel } from '@/lib/constants';
import { Button } from '@/components/ui/Button';
import { EmptyState } from '@/components/ui/EmptyState';
import { ErrorState } from '@/components/ui/ErrorState';
import { NoPermissionState } from '@/components/ui/NoPermissionState';
import { SkeletonTable } from '@/components/ui/SkeletonTable';
import { LogsToolbar } from '@/components/logs/LogsToolbar';
import { ViewToggle } from '@/components/logs/ViewToggle';
import { LogsTable } from '@/components/logs/LogsTable';
import { LogsCards } from '@/components/logs/LogsCards';
import { TempLinkModal } from '@/components/logs/TempLinkModal';
import type { LogFileView } from '@/api/types';

/**
 * Primary screen. Resolves the active provider's bucket, fetches its logs, and
 * renders exactly one mutually-exclusive surface (no-permission / skeleton /
 * error / empty / data) per the §8.5 state machine.
 */
export function LogsPage() {
  const provider = useUiStore((s) => s.provider);
  const search = useUiStore((s) => s.search);
  const logsView = useUiStore((s) => s.logsView);
  const setProvider = useUiStore((s) => s.setProvider);
  const clearSearch = useUiStore((s) => s.clearSearch);
  const setLogsView = useUiStore((s) => s.setLogsView);

  const user = useAuthStore((s) => s.user);
  const isAdmin = useIsAdmin();

  const { data: providers, isLoading: providersLoading, refetch: refetchProviders } = useProviders();

  const isProviderAllowed = !user?.allowed_providers || user.allowed_providers.includes(provider);
  const providerObj = providers?.find((p) => p.id === provider);
  // The /providers health probe is the fast signal — don't fire a doomed logs
  // request at a provider the backend already reports as unreachable.
  const providerUnhealthy = providerObj?.healthy === false;
  const bucket = providerObj?.buckets[0] ?? '';

  const q = useLogs(provider, bucket, isProviderAllowed && !providerUnhealthy);

  const filtered = useMemo(() => {
    const all = q.data?.files ?? [];
    const s = search.trim().toLowerCase();
    if (!s) return all;
    return all.filter((f) => f.filename.toLowerCase().includes(s));
  }, [q.data?.files, search]);

  const download = useDownload();
  const onDownload = (f: LogFileView) => download.mutate({ provider, bucket, key: f.key });
  const isDownloading = (key: string) => download.isPending && download.variables?.key === key;

  const [modalFile, setModalFile] = useState<LogFileView | null>(null);
  const onLink = (f: LogFileView) => setModalFile(f);

  const is403 = q.error instanceof ApiResponseError && q.error.code === 'PROVIDER_NOT_ALLOWED';
  // Allowed provider that resolved with no bucket → deterministic error instead
  // of a blank dead state (the §8.5 stale-bucket guard).
  const noBucket = isProviderAllowed && !providersLoading && !bucket && !providerUnhealthy;
  const showNoPermission = !isProviderAllowed || is403;
  const showSkeleton = !showNoPermission && (q.isLoading || providersLoading);
  const showError =
    !showNoPermission && !showSkeleton && (providerUnhealthy || (q.isError && !is403) || noBucket);
  const showEmpty = !showNoPermission && q.isSuccess && filtered.length === 0;
  const showData = !showNoPermission && q.isSuccess && filtered.length > 0;

  const firstAllowed = user?.allowed_providers?.[0];
  const hasSearch = search.trim().length > 0;
  const errorMessage = providerUnhealthy
    ? `${providerLabel(provider)} is currently unreachable. The storage backend may be down — try again.`
    : noBucket
      ? `No bucket is available for ${providerLabel(provider)} yet.`
      : undefined;

  return (
    <div className="animate-fade-in">
      <div className="mb-4 flex items-start justify-between">
        <div>
          <h1 className="text-title text-text">Logs</h1>
          <p className="text-body text-muted">
            {!showNoPermission && bucket ? (
              <>
                Browse and download log files from <span className="font-mono">{bucket}</span>
              </>
            ) : (
              'Browse and download log files'
            )}
          </p>
        </div>
        <ViewToggle
          view={logsView}
          onViewChange={setLogsView}
          onReload={() => q.refetch()}
          isFetching={q.isFetching}
        />
      </div>

      <LogsToolbar />

      {showNoPermission ? (
        <NoPermissionState
          providerLabel={providerLabel(provider)}
          {...(firstAllowed
            ? { switchToLabel: providerLabel(firstAllowed), onSwitch: () => setProvider(firstAllowed) }
            : {})}
        />
      ) : showSkeleton ? (
        <SkeletonTable />
      ) : showError ? (
        <ErrorState
          providerLabel={providerLabel(provider)}
          onRetry={() => (providerUnhealthy || noBucket ? refetchProviders() : q.refetch())}
          {...(errorMessage ? { message: errorMessage } : {})}
        />
      ) : showEmpty ? (
        <EmptyState
          icon={<FileX2 size={24} aria-hidden="true" />}
          title={hasSearch ? `No logs match "${search}"` : 'No log files found'}
          message={hasSearch ? 'Try a different search term.' : 'This bucket has no log files yet.'}
        >
          {hasSearch && (
            <Button variant="outline" onClick={clearSearch}>
              Clear search
            </Button>
          )}
        </EmptyState>
      ) : showData ? (
        logsView === 'table' ? (
          <LogsTable
            files={filtered}
            bucket={bucket}
            isAdmin={isAdmin}
            count={filtered.length}
            onDownload={onDownload}
            onLink={onLink}
            isDownloading={isDownloading}
          />
        ) : (
          <LogsCards
            files={filtered}
            bucket={bucket}
            isAdmin={isAdmin}
            onDownload={onDownload}
            onLink={onLink}
            isDownloading={isDownloading}
          />
        )
      ) : null}

      {isAdmin && (
        <TempLinkModal
          open={!!modalFile}
          file={modalFile}
          provider={provider}
          bucket={bucket}
          onClose={() => setModalFile(null)}
        />
      )}
    </div>
  );
}
