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

/** "n · complete c/n · mixed m · [1m] l" for a model-choice cell. */
export function formatSessionsCell(c: {
  sessions: number;
  complete_cost_sessions: number;
  mixed_model_sessions: number;
  long_context_sessions: number;
}): string {
  return `${c.sessions} · complete ${c.complete_cost_sessions}/${c.sessions} · mixed ${c.mixed_model_sessions} · [1m] ${c.long_context_sessions}`;
}

/** "x/N (rate)": a count over its base, then the rate; a null rate is "not measured", never 0. */
export function formatCountRate(numerator: number, base: number, rate: number | null): string {
  return `${numerator}/${base} (${formatPercent(rate)})`;
}

/** A cost ratio such as "5.1x"; null is "not measured". */
export function formatTimes(value: number | null): string {
  if (value === null) return NOT_MEASURED;
  return `${value.toFixed(1)}x`;
}

/** The badges a model-choice category carries: "not comparable", or "thin" when comparable but unsupported. */
export function categoryBadges(c: { comparable: boolean; supported: boolean }): string[] {
  if (!c.comparable) return ['not comparable'];
  return c.supported ? [] : ['thin'];
}
