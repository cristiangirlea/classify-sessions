import { describe, expect, it } from 'vitest';
import {
  NOT_MEASURED,
  categoryBadges,
  formatCountRate,
  formatSessionsCell,
  formatTimes,
  formatCount,
  formatDate,
  formatPercent,
  formatRule,
  formatUsd,
  formatUsdKnown,
} from '../src/lib/format';

describe('format', () => {
  it('formats a known USD amount with four decimals', () => {
    expect(formatUsd(1.413)).toBe('USD 1.4130');
    expect(formatUsd(0)).toBe('USD 0.0000');
  });

  it('states that an incomplete cost is a lower bound', () => {
    expect(formatUsdKnown(1.1475, true)).toBe('USD 1.1475');
    const text = formatUsdKnown(1.1475, false);
    expect(text).toContain('at least USD 1.1475');
    expect(text).toContain('lower bound');
  });

  it('formats fractions as percentages', () => {
    expect(formatPercent(0.05)).toBe('5.0%');
    expect(formatPercent(0.3333333333333333)).toBe('33.3%');
    expect(formatPercent(0)).toBe('0.0%');
    expect(formatPercent(1)).toBe('100.0%');
  });

  it('formats integer counts', () => {
    expect(formatCount(40)).toBe('40');
    expect(formatCount(1107)).toBe('1,107');
    expect(formatCount(0)).toBe('0');
  });

  it('formats dates as UTC minutes', () => {
    expect(formatDate('2026-08-01T09:00:00Z')).toBe('2026-08-01 09:00 UTC');
    expect(formatDate('2026-08-01T11:30:45+02:00')).toBe('2026-08-01 09:30 UTC');
  });

  it('shows an unparsable date as given', () => {
    expect(formatDate('yesterday')).toBe('yesterday');
  });

  it('renders null as "not measured", never 0', () => {
    expect(NOT_MEASURED).toBe('not measured');
    expect(formatUsd(null)).toBe('not measured');
    expect(formatUsdKnown(null, true)).toBe('not measured');
    expect(formatUsdKnown(null, false)).toBe('not measured');
    expect(formatPercent(null)).toBe('not measured');
    expect(formatCount(null)).toBe('not measured');
    expect(formatDate(null)).toBe('not measured');
  });

  it('renders a missing rule as a statement, not an empty cell', () => {
    expect(formatRule(null)).toBe('no rule matched');
    expect(formatRule('debug.why-does')).toBe('debug.why-does');
  });

  it('formats the sessions cell of the model-choice table', () => {
    expect(
      formatSessionsCell({ sessions: 5, complete_cost_sessions: 4, mixed_model_sessions: 1, long_context_sessions: 2 }),
    ).toBe('5 · complete 4/5 · mixed 1 · [1m] 2');
  });

  it('formats a count over its base with the rate, and a null rate as not measured', () => {
    expect(formatCountRate(1, 3, 1 / 3)).toBe('1/3 (33.3%)');
    expect(formatCountRate(0, 0, null)).toBe('0/0 (not measured)');
    expect(formatCountRate(0, 4, 0)).toBe('0/4 (0.0%)');
  });

  it('formats a cost ratio, null as not measured', () => {
    expect(formatTimes(5.123)).toBe('5.1x');
    expect(formatTimes(null)).toBe('not measured');
  });

  it('badges non-comparable and thin categories only', () => {
    expect(categoryBadges({ comparable: false, supported: false })).toEqual(['not comparable']);
    expect(categoryBadges({ comparable: true, supported: false })).toEqual(['thin']);
    expect(categoryBadges({ comparable: true, supported: true })).toEqual([]);
  });
});
