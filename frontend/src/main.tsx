import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { RouterProvider } from 'react-router-dom';
import { router } from '@/router';
import { useUiStore } from '@/store/uiStore';
import { applyTheme, bindSystem, getPref, THEME_KEY } from '@/lib/theme';
import './styles/index.css';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { refetchOnWindowFocus: false, retry: 1 },
  },
});

// Keep <html data-theme> in sync with the store after the no-flash script's
// initial paint, and react to OS changes while pref === 'system'.
useUiStore.subscribe((state) => applyTheme(state.theme));
bindSystem(getPref(), () => applyTheme('system'));

// Cross-tab theme sync (the no-flash key is canonical).
window.addEventListener('storage', (e) => {
  if (e.key === THEME_KEY) applyTheme(getPref());
});

// Enable theme cross-fade only after first paint (no load flash).
requestAnimationFrame(() => document.documentElement.classList.add('theme-ready'));

async function bootstrap() {
  // Optional in-browser mock BFF for backend-less dev/demo runs (`npm run dev:mock`
  // sets MODE=mock; VITE_USE_MOCKS=true also works).
  if (import.meta.env.MODE === 'mock' || import.meta.env['VITE_USE_MOCKS'] === 'true') {
    const { startMocks } = await import('./mocks/browser');
    await startMocks();
  }
  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </StrictMode>,
  );
}

void bootstrap();
