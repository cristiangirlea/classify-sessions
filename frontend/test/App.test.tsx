import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { renderToStaticMarkup } from 'react-dom/server';
import { describe, expect, it } from 'vitest';
import { ErrorState, Report } from '../src/App';
import { ModelChoice, SYNTHETIC_NOTE } from '../src/views/ModelChoice';
import type { ModelsResponse, SummaryResponse } from '../src/api/contract.gen';

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

describe('Model choice panel', () => {
  const choice = summary.model_choice;
  const markup = renderToStaticMarkup(<Report summary={summary} />);
  const page = text(markup);

  it('is the fifth panel of the report', () => {
    expect(page).toContain('Model choice');
    expect(markup.indexOf('id="sessions-h"')).toBeLessThan(markup.indexOf('id="models-h"'));
  });

  it('shows the example verdict and statement', () => {
    expect(choice.advice.verdict).toBe('no_comparison_possible');
    expect(page).toContain('Verdict: no comparison possible');
    expect(page).toContain(choice.advice.statement);
    expect(page).toContain('Cost ratio: not measured');
  });

  it('shows the not comparable badge with its reason', () => {
    expect(page).toContain('not comparable');
    expect(page).toContain('only one model has a session in this category');
  });

  it('shows a correction rate as x/N with the rate', () => {
    expect(page).toContain('1 of 3 (33%)');
    expect(page).toContain('0 of 1 (0%)');
  });

  it('shows a null tool failure rate as not measured', () => {
    expect(page).toContain('0 of 0 (not measured)');
  });

  it('flags a row with cost incomplete and shows two-decimal money where complete_cost_sessions < sessions', () => {
    expect(page).toContain('claude-opus-4-1 cost incomplete');
    expect(page).not.toContain('at least USD 1.1475');
    expect(page).toContain('1.15');
  });

  it('shows the seven definitions, the caveats and the synthetic-fixture sentence', () => {
    for (const d of Object.values(choice.definitions)) expect(page).toContain(d);
    for (const c of choice.advice.caveats) expect(page).toContain(c);
    expect(page).toContain(SYNTHETIC_NOTE);
    expect(page).toContain(
      'This fixture is synthetic: its prompts are templated, so correction counts describe the generator, not real developers.',
    );
  });

  it('shows the thin badge, the cost ratio, shared categories and excluded ids', () => {
    const thin: ModelsResponse = {
      ...choice,
      categories: [{ ...choice.categories[0], category: 'review', comparable: true, supported: false, reason: 'fewer than three sessions per model' }],
      excluded_sessions: ['00000000-0000-4000-8000-0000000000ff'],
      advice: { ...choice.advice, cost_ratio: 5.1, shared_categories: ['review'], verdict: 'insufficient_evidence' },
    };
    const t = text(renderToStaticMarkup(<ModelChoice choice={thin} />));
    expect(t).toContain('thin');
    expect(t).not.toContain('not comparable');
    expect(t).toContain('Cost ratio: 5.1x');
    expect(t).toContain('Shared categories: review');
    expect(t).toContain('00000000-0000-4000-8000-0000000000ff');
    expect(t).toContain('Verdict: insufficient evidence');
  });

  it('shows a text empty state with no categories', () => {
    const empty: ModelsResponse = { ...choice, categories: [], model_totals: [], models: [] };
    expect(text(renderToStaticMarkup(<ModelChoice choice={empty} />))).toContain('No sessions to compare.');
  });
});
