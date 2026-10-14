import react from '@vitejs/plugin-react';
import { loadEnv } from 'vite';
import { defineConfig } from 'vitest/config';

// One config for the dev server, the build and Vitest (docs/SPEC.md sections 2, 8, 10).
export default defineConfig(({ mode }) => {
  // VITE_API_PROXY_TARGET comes from the process env (Compose sets http://api:8080) or from
  // frontend/.env*.local for native dev; the process env wins. Empty or unset falls back.
  const env = loadEnv(mode, process.cwd(), 'VITE_');
  const target = env.VITE_API_PROXY_TARGET || 'http://localhost:8080';

  return {
    plugins: [react()],
    server: {
      proxy: {
        '/api': { target, changeOrigin: true },
        '/healthz': { target, changeOrigin: true },
      },
    },
    test: {
      environment: 'jsdom',
      setupFiles: ['./src/test/setup.ts'],
      include: ['src/**/*.test.{ts,tsx}'],
      // Undo vi.spyOn and vi.stubGlobal (e.g. a stubbed fetch) after every test.
      restoreMocks: true,
      unstubGlobals: true,
    },
  };
});
