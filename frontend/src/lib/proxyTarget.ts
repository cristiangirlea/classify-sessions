// Dev-server only (imported by vite.config.ts, never by the page): picks the /api proxy target.
// Nothing may leave the machine, so only a loopback http origin with an explicit port is accepted.

export const DEFAULT_API_PROXY_TARGET = 'http://127.0.0.1:8080';

const LOOPBACK = /^http:\/\/(127\.0\.0\.1|localhost):(\d{1,5})$/;

/** The proxy target from API_PROXY_TARGET (unset or empty means the default); throws on anything else. */
export function resolveProxyTarget(value: string | undefined): string {
  if (value === undefined || value === '') return DEFAULT_API_PROXY_TARGET;
  const m = LOOPBACK.exec(value);
  const port = m ? Number(m[2]) : 0;
  if (!m || port < 1 || port > 65535) {
    throw new Error(
      `API_PROXY_TARGET must be http://127.0.0.1:<port> or http://localhost:<port> (got "${value}"); nothing may leave the machine`,
    );
  }
  return value;
}
