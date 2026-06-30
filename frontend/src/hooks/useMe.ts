import { useQuery } from '@tanstack/react-query';
import { getMe } from '@/api/auth';
import { useAuthStore } from '@/store/authStore';

/** Validates the persisted session on mount; a 401 (revoked/expired) is handled by the API client's session-expired redirect. */
export function useMe() {
  const token = useAuthStore((s) => s.token);
  return useQuery({
    queryKey: ['me'],
    queryFn: getMe,
    enabled: !!token,
    staleTime: 5 * 60_000,
    retry: false,
  });
}
