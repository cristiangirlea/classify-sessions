// Serves contract/examples/*.json at /api/* when VITE_API_MOCK=1 (`npm run dev:mock`),
// so the page can be built before the backend exists. Dev server only; never in a build.
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import type { Plugin } from 'vite';

const routes: Record<string, string> = {
  '/api/summary': 'summary.json',
  '/api/models': 'models.json',
};

export function mockApi(examplesDir: string): Plugin {
  return {
    name: 'mock-api',
    apply: 'serve',
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        const path = (req.url ?? '').split('?')[0];
        if (!path.startsWith('/api/')) return next();
        const file = routes[path];
        res.setHeader('Content-Type', 'application/json');
        if (!file) {
          res.statusCode = 404;
          res.end(JSON.stringify({ error: { code: 'not_found', message: 'no such endpoint' } }));
          return;
        }
        res.end(readFileSync(resolve(examplesDir, file)));
      });
    },
  };
}
