import { useQuery } from '@tanstack/react-query';
import { listProviders } from '@/api/providers';

export function useProviders() {
  return useQuery({
    queryKey: ['providers'],
    queryFn: listProviders,
    staleTime: 60_000,
  });
}
