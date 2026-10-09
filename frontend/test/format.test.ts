import { describe, expect, it } from 'vitest';
import {
  NOT_MEASURED,
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
});
