import { resolve } from 'node:path';
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vitest/config';
import { mockApi } from './src/mock/vitePlugin';

const mock = process.env.VITE_API_MOCK === '1';

export default defineConfig({
  plugins: [react(), ...(mock ? [mockApi(resolve(__dirname, '../contract/examples'))] : [])],
  server: {
    host: '127.0.0.1',
    port: 5173,
    strictPort: true,
    // No CORS anywhere: the browser only ever talks to 127.0.0.1:5173.
    proxy: mock ? undefined : { '/api': { target: 'http://127.0.0.1:8080', changeOrigin: false } },
  },
  preview: { host: '127.0.0.1', port: 4173, strictPort: true },
  test: { include: ['test/**/*.test.ts', 'test/**/*.test.tsx'], passWithNoTests: true },
});
