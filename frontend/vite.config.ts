import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { fileURLToPath, URL } from 'node:url';

// Dev: same-origin `/api/*` is proxied to the Go BFF so there is zero CORS in
// development — identical to the nginx `/api` proxy used in production.
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  build: {
    target: 'es2022',
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: process.env['VITE_PROXY_TARGET'] ?? 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
});
