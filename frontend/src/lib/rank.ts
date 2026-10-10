// Ordering and colour assignment for the spend views. Everything is derived from the data.
import type { CategorySummary } from '../api/contract.gen';

/** Categories by known cost, largest first; ties by name so the order is stable. */
export function rankByCost(categories: CategorySummary[]): CategorySummary[] {
  return [...categories].sort((a, b) => b.cost_usd - a.cost_usd || a.category.localeCompare(b.category));
}

/** The CSS colour variable for a rank (0 = largest). Ranks past the ninth share the last colour. */
export function rankColor(rank: number): string {
  return `var(--c${Math.min(rank + 1, 9)})`;
}
