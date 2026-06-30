import { useAuthStore } from '@/store/authStore';

export const useIsAdmin = (): boolean => useAuthStore((s) => s.user?.role === 'admin');
