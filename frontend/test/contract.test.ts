import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ApiError, getModels, getSummary } from '../src/api/client';
import type { Components, ErrorResponse, ModelsResponse, SummaryResponse } from '../src/api/contract.gen';

function example(name: string): string {
  return readFileSync(fileURLToPath(new URL(`../../contract/examples/${name}`, import.meta.url)), 'utf8');
}

function stubFetch(body: string, status = 200): ReturnType<typeof vi.fn> {
  const fn = vi.fn(async () => new Response(body, { status }));
  vi.stubGlobal('fetch', fn);
  return fn;
}

afterEach(() => {
  vi.unstubAllGlobals();
});

const TOL = 1e-9;
const sum = (xs: number[]) => xs.reduce((a, b) => a + b, 0);
const componentsSum = (c: Components) => c.input_usd + c.output_usd + c.cache_read_usd + c.cache_write_usd;

describe('client parses the contract examples', () => {
  it('getSummary parses summary.json and requests /api/summary', async () => {
    const text = example('summary.json');
    const fn = stubFetch(text);
    const got = await getSummary();
    expect(fn).toHaveBeenCalledWith('/api/summary');
    expect(got).toEqual(JSON.parse(text));
    expect(got.records.total).toBe(40);
  });

  it('getModels parses models.json and requests /api/models', async () => {
    const text = example('models.json');
    const fn = stubFetch(text);
    const got = await getModels();
    expect(fn).toHaveBeenCalledWith('/api/models');
    expect(got).toEqual(JSON.parse(text));
    expect(got.advice.verdict).toBe('no_comparison_possible');
  });

  it('the summary embeds the same model_choice as the models example', () => {
    const summary = JSON.parse(example('summary.json')) as SummaryResponse;
    const models = JSON.parse(example('models.json')) as ModelsResponse;
    expect(summary.model_choice).toEqual(models);
  });

  it('error.json yields its code and message as text on a non-2xx', async () => {
    const text = example('error.json');
    const err = (JSON.parse(text) as ErrorResponse).error;
    stubFetch(text, 405);
    const caught = await getSummary().catch((e: unknown) => e);
    expect(caught).toBeInstanceOf(ApiError);
    const apiErr = caught as ApiError;
    expect(apiErr.message).toContain(err.code);
    expect(apiErr.message).toContain(err.message);
    expect(apiErr.message).toContain('405');
    expect(apiErr.status).toBe(405);
    expect(apiErr.code).toBe(err.code);
  });

  it('a non-2xx without a readable error body still yields status text', async () => {
    stubFetch('<html>bad gateway</html>', 502);
    const caught = (await getModels().catch((e: unknown) => e)) as ApiError;
    expect(caught).toBeInstanceOf(ApiError);
    expect(caught.message).toContain('502');
    expect(caught.code).toBeNull();
  });

  it('a 2xx body that is not JSON is an error, not a page of zeros', async () => {
    stubFetch('not json', 200);
    await expect(getSummary()).rejects.toBeInstanceOf(ApiError);
  });

  it('a network failure becomes an ApiError with text', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => {
        throw new TypeError('connection refused');
      }),
    );
    const caught = (await getSummary().catch((e: unknown) => e)) as ApiError;
    expect(caught).toBeInstanceOf(ApiError);
    expect(caught.message).toContain('connection refused');
    expect(caught.status).toBeNull();
  });
});

describe('summary example is internally consistent', () => {
  const s = JSON.parse(example('summary.json')) as SummaryResponse;

  it('records.total = parsed + malformed + unknown + duplicate', () => {
    const r = s.records;
    expect(r.total).toBe(r.parsed + r.malformed + r.unknown + r.duplicate);
  });

  it('unparsed_fraction = (malformed + unknown) / total', () => {
    const r = s.records;
    expect(Math.abs(r.unparsed_fraction - (r.malformed + r.unknown) / r.total)).toBeLessThan(TOL);
  });

  it('sessions.classified + unclassified = total', () => {
    expect(s.sessions.classified + s.sessions.unclassified).toBe(s.sessions.total);
  });

  it('category sessions sum to sessions.total', () => {
    expect(sum(s.categories.map((c) => c.sessions))).toBe(s.sessions.total);
  });

  it('category cost = spend total = by_model cost = session_list cost', () => {
    const total = s.spend.total_usd;
    expect(Math.abs(sum(s.categories.map((c) => c.cost_usd)) - total)).toBeLessThan(TOL);
    expect(Math.abs(sum(s.spend.by_model.map((m) => m.cost_usd)) - total)).toBeLessThan(TOL);
    expect(Math.abs(sum(s.session_list.map((r) => r.cost_usd)) - total)).toBeLessThan(TOL);
  });

  it('each components object sums to its cost', () => {
    for (const c of s.categories) {
      expect(Math.abs(componentsSum(c.components) - c.cost_usd)).toBeLessThan(TOL);
    }
    expect(Math.abs(componentsSum(s.spend.components) - s.spend.total_usd)).toBeLessThan(TOL);
  });

  it('spend.components = sum of category components', () => {
    const keys: (keyof Components)[] = ['input_usd', 'output_usd', 'cache_read_usd', 'cache_write_usd'];
    for (const k of keys) {
      expect(Math.abs(sum(s.categories.map((c) => c.components[k])) - s.spend.components[k])).toBeLessThan(TOL);
    }
  });

  it('incomplete sessions agree between spend, categories and session_list', () => {
    const incompleteRows = s.session_list.filter((r) => !r.cost_complete).length;
    expect(s.spend.incomplete_sessions).toBe(incompleteRows);
    expect(sum(s.categories.map((c) => c.cost_incomplete_sessions))).toBe(incompleteRows);
  });
});
