import type { LogFile, LogFileView, LogLevel } from '@/api/types';

export function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  const unit = units[i] ?? 'B';
  const val = bytes / 1024 ** i;
  const rounded = val >= 10 || i === 0 ? Math.round(val) : Number(val.toFixed(1));
  return `${rounded} ${unit}`;
}

/** Calendar-relative "Modified" label: "Today, 14:32" / "Yesterday, 09:14" / "2 days ago" / locale date beyond a week. */
export function formatModified(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  const now = new Date();
  const time = d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false });
  const startOf = (x: Date) => new Date(x.getFullYear(), x.getMonth(), x.getDate()).getTime();
  const dayDiff = Math.round((startOf(now) - startOf(d)) / 86_400_000);
  if (dayDiff <= 0) return `Today, ${time}`;
  if (dayDiff === 1) return `Yesterday, ${time}`;
  if (dayDiff < 7) return `${dayDiff} days ago`;
  return d.toLocaleDateString([], { year: 'numeric', month: 'short', day: 'numeric' });
}

const RTF = new Intl.RelativeTimeFormat('en', { numeric: 'auto', style: 'long' });
const DIVISIONS: { amount: number; unit: Intl.RelativeTimeFormatUnit }[] = [
  { amount: 60, unit: 'second' },
  { amount: 60, unit: 'minute' },
  { amount: 24, unit: 'hour' },
  { amount: 7, unit: 'day' },
  { amount: 4.34524, unit: 'week' },
  { amount: 12, unit: 'month' },
  { amount: Number.POSITIVE_INFINITY, unit: 'year' },
];

/** Relative time, signed: past → "2 min ago", future → "in 15 min". */
export function formatRelativeTime(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  let duration = (d.getTime() - Date.now()) / 1000;
  for (const division of DIVISIONS) {
    if (Math.abs(duration) < division.amount) {
      return RTF.format(Math.round(duration), division.unit);
    }
    duration /= division.amount;
  }
  return iso;
}

/** Deterministic level heuristic on the object key (locked §0). */
export function inferLogLevel(key: string): LogLevel {
  if (/error|fail|worker/i.test(key)) return 'error';
  if (/warn|auth/i.test(key)) return 'warn';
  return 'info';
}

export interface LevelMeta {
  level: LogLevel;
  label: string;
  iconClass: string;
  softClass: string;
  /** AA-passing accent text (badge label) */
  textClass: string;
}

const LEVEL_META: Record<LogLevel, LevelMeta> = {
  info: {
    level: 'info',
    label: 'INFO',
    iconClass: 'text-primary',
    softClass: 'bg-primary-soft',
    textClass: 'text-primary-strong',
  },
  warn: {
    level: 'warn',
    label: 'WARN',
    iconClass: 'text-warn',
    softClass: 'bg-warn-soft',
    textClass: 'text-warn-strong',
  },
  error: {
    level: 'error',
    label: 'ERROR',
    iconClass: 'text-error',
    softClass: 'bg-error-soft',
    textClass: 'text-error-strong',
  },
};

export function getLevelMeta(level: LogLevel): LevelMeta {
  return LEVEL_META[level];
}

export function basename(key: string): string {
  const parts = key.split('/').filter(Boolean);
  return parts.length > 0 ? parts[parts.length - 1]! : key;
}

export function toLogFileView(f: LogFile): LogFileView {
  return {
    ...f,
    filename: basename(f.key),
    size_human: formatBytes(f.size),
    level: inferLogLevel(f.key),
  };
}
