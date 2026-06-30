import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { useAuthStore } from '@/store/authStore';
import type { Role } from '@/api/types';

interface ProtectedRouteProps {
  requiredRole?: Role;
}

export function ProtectedRoute({ requiredRole }: ProtectedRouteProps) {
  const location = useLocation();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const role = useAuthStore((s) => s.user?.role);

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }
  if (requiredRole && role !== requiredRole) {
    return <Navigate to="/dashboard" replace />;
  }
  return <Outlet />;
}
