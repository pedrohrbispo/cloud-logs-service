import { cn } from '@/lib/cn';

interface SkeletonProps {
  className?: string;
}

export function Skeleton({ className }: SkeletonProps) {
  return <span aria-hidden="true" className={cn('skeleton block h-3 w-full', className)} />;
}
