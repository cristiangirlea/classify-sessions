import { resolve } from 'node:path';
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vitest/config';
import { resolveProxyTarget } from './src/lib/proxyTarget';
import { mockApi } from './src/mock/vitePlugin';

const mock = process.env.VITE_API_MOCK === '1';
// Not VITE_-prefixed on purpose: it must never reach the client bundle. Fails at startup if not loopback.
const apiTarget = resolveProxyTarget(process.env.API_PROXY_TARGET);

export default defineConfig({
  plugins: [react(), ...(mock ? [mockApi(resolve(__dirname, '../contract/examples'))] : [])],
  server: {
    host: '127.0.0.1',
    port: 5173,
    strictPort: true,
    // No CORS anywhere: the browser only ever talks to 127.0.0.1:5173.
    proxy: mock ? undefined : { '/api': { target: apiTarget, changeOrigin: false } },
  },
  preview: { host: '127.0.0.1', port: 4173, strictPort: true },
  test: { include: ['test/**/*.test.ts', 'test/**/*.test.tsx'], passWithNoTests: true },
});
