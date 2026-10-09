// Display formatting. A value that was not measured is never shown as 0.

export const NOT_MEASURED = 'not measured';

/** A known USD amount, four decimals so sub-cent request costs stay visible. */
export function formatUsd(value: number | null): string {
  if (value === null) return NOT_MEASURED;
  return `USD ${value.toFixed(4)}`;
}

/**
 * A known USD amount that may be a lower bound. When the cost is incomplete the
 * known part is stated as "at least", never as the full cost.
 */
export function formatUsdKnown(value: number | null, complete: boolean): string {
  if (value === null) return NOT_MEASURED;
  return complete ? formatUsd(value) : `at least ${formatUsd(value)} (lower bound, cost incomplete)`;
}

/** A fraction in [0, 1] as a percentage with one decimal. */
export function formatPercent(fraction: number | null): string {
  if (fraction === null) return NOT_MEASURED;
  return `${(fraction * 100).toFixed(1)}%`;
}

/** An integer count with thousands separators. */
export function formatCount(value: number | null): string {
  if (value === null) return NOT_MEASURED;
  return value.toLocaleString('en-US');
}

/** An ISO-8601 timestamp as "YYYY-MM-DD HH:MM UTC"; unparsable input is shown as given. */
export function formatDate(iso: string | null): string {
  if (iso === null) return NOT_MEASURED;
  const t = Date.parse(iso);
  if (Number.isNaN(t)) return iso;
  return `${new Date(t).toISOString().slice(0, 16).replace('T', ' ')} UTC`;
}

/** A rule id, or a plain statement that no rule matched. */
export function formatRule(rule: string | null): string {
  return rule === null ? 'no rule matched' : rule;
}
