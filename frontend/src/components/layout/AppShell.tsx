import { useEffect } from 'react';
import { Outlet } from 'react-router-dom';
import { SkipLink } from '@/components/ui/SkipLink';
import { Toast } from '@/components/ui/Toast';
import { useMe } from '@/hooks/useMe';
import { useAuthStore } from '@/store/authStore';
import { Sidebar } from './Sidebar';
import { Topbar } from './Topbar';

export function AppShell() {
  // Re-validate the persisted session and refresh the user (e.g. allowed_providers)
  // from the server on load. A revoked token 401s → client redirects to /login.
  const { data: me } = useMe();
  useEffect(() => {
    if (!me) return;
    const cur = useAuthStore.getState().user;
    const changed =
      !cur ||
      cur.id !== me.id ||
      cur.role !== me.role ||
      JSON.stringify(cur.allowed_providers) !== JSON.stringify(me.allowed_providers);
    if (changed) useAuthStore.setState({ user: me });
  }, [me]);

  return (
    <>
      <div className="flex min-h-screen">
        <SkipLink />
        <Sidebar />
        <div className="flex min-w-0 flex-1 flex-col">
          <Topbar />
          <main id="main" tabIndex={-1} className="flex-1 bg-bg px-page py-page-y outline-none">
            <div className="mx-auto w-full max-w-content">
              <Outlet />
            </div>
          </main>
        </div>
      </div>
      <Toast />
    </>
  );
}
