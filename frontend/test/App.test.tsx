import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { renderToStaticMarkup } from 'react-dom/server';
import { describe, expect, it } from 'vitest';
import { ErrorState, Report } from '../src/App';
import type { SummaryResponse } from '../src/api/contract.gen';

const summary = JSON.parse(
  readFileSync(fileURLToPath(new URL('../../contract/examples/summary.json', import.meta.url)), 'utf8'),
) as SummaryResponse;

// Static markup keeps tags between texts, so strip tags and collapse whitespace first.
function text(markup: string): string {
  return markup.replace(/<[^>]*>/g, ' ').replace(/&#x27;/g, "'").replace(/\s+/g, ' ');
}

describe('Report renders the three headline numbers from the summary example', () => {
  const page = text(renderToStaticMarkup(<Report summary={summary} />));

  it('shows the records total', () => {
    expect(page).toContain('Total (nonblank) 40');
  });

  it('shows known spend as a lower bound while a session is incomplete', () => {
    expect(summary.spend.incomplete_sessions).toBeGreaterThan(0);
    expect(page).toContain('at least USD 1.4130 (lower bound, cost incomplete)');
  });

  it('shows the held-out agreement', () => {
    expect(page).toContain('Agreement fraction: 50.0%');
  });

  it('wraps every table in a scroll container', () => {
    const markup = renderToStaticMarkup(<Report summary={summary} />);
    const tables = markup.match(/<table>/g)?.length ?? 0;
    const wrapped = markup.match(/<div class="table-scroll"><table>/g)?.length ?? 0;
    expect(tables).toBeGreaterThan(0);
    expect(wrapped).toBe(tables);
  });
});

describe('ErrorState', () => {
  it('shows code, message, status and the hint to start the API', () => {
    const message = 'method_not_allowed: only GET and HEAD are allowed (HTTP 405)';
    const markup = renderToStaticMarkup(<ErrorState message={message} />);
    const t = text(markup);
    expect(markup).toContain('role="alert"');
    expect(t).toContain('method_not_allowed');
    expect(t).toContain('only GET and HEAD are allowed');
    expect(t).toContain('HTTP 405');
    expect(t).toContain('127.0.0.1:8080');
    expect(t).toContain('go run ./cmd/sessions serve --input ../fixtures/claude-code');
  });
});
