import { describe, expect, it } from 'vitest';
import { DEFAULT_API_PROXY_TARGET, resolveProxyTarget } from '../src/lib/proxyTarget';

describe('resolveProxyTarget', () => {
  it('defaults to 127.0.0.1:8080 when unset or empty', () => {
    expect(DEFAULT_API_PROXY_TARGET).toBe('http://127.0.0.1:8080');
    expect(resolveProxyTarget(undefined)).toBe('http://127.0.0.1:8080');
    expect(resolveProxyTarget('')).toBe('http://127.0.0.1:8080');
  });

  it('accepts loopback http targets with a port', () => {
    expect(resolveProxyTarget('http://127.0.0.1:8081')).toBe('http://127.0.0.1:8081');
    expect(resolveProxyTarget('http://localhost:9000')).toBe('http://localhost:9000');
  });

  it('rejects everything else with a clear message', () => {
    const bad = [
      'http://example.com',
      'http://example.com:8080',
      'https://127.0.0.1:8080',
      'http://127.0.0.1',
      'http://127.0.0.1:0',
      'http://127.0.0.1:70000',
      'http://127.0.0.1:8080/',
      'http://127.0.0.1:8080/api',
      'http://127.0.0.1.example.com:8080',
      'http://user@127.0.0.1:8080',
      'http://0.0.0.0:8080',
      '127.0.0.1:8080',
    ];
    for (const v of bad) {
      expect(() => resolveProxyTarget(v), v).toThrow(/API_PROXY_TARGET must be/);
    }
  });
});
