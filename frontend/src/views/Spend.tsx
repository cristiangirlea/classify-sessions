import type { Components, ModelSpend, SummaryResponse } from '../api/contract.gen';
import { formatAmount, formatCount, formatUsdKnown } from '../lib/format';

type Flag = ModelSpend['flags'][number];

// Plain-text meaning of each flag the contract can carry; shown as the badge tooltip.
export const FLAG_EXPLANATIONS: Record<Flag, string> = {
  missing_usage: 'Some requests carry no token usage, so they contribute no cost and their sessions are incomplete.',
  unknown_model: 'The model is not in the price table, so its requests cannot be priced and its sessions are incomplete.',
  no_cache_fields: 'Some usage objects have no cache fields; cache cost is unknown and treated as incomplete, not as zero.',
  long_context_marker: 'The model id carries the [1m] marker; requests are priced at base rates because none exceeds the long-context threshold.',
  sidechain: 'Some requests were made by sidechain agents; they are billed and counted, and flagged for visibility.',
};

interface Badge {
  label: string;
  tone: 'warn' | 'plain';
  title: string;
}

/** Badges for a model row, derived from its flags only. */
export function modelBadges(flags: Flag[]): Badge[] {
  const out: Badge[] = [];
  const incomplete = flags.filter((f) => f === 'missing_usage' || f === 'unknown_model' || f === 'no_cache_fields');
  if (incomplete.length > 0) {
    out.push({ label: 'cost incomplete', tone: 'warn', title: incomplete.map((f) => FLAG_EXPLANATIONS[f]).join(' ') });
  }
  if (flags.includes('unknown_model')) {
    out.push({ label: 'not in price table', tone: 'plain', title: FLAG_EXPLANATIONS.unknown_model });
  }
  if (flags.includes('long_context_marker')) {
    out.push({ label: 'base rate', tone: 'plain', title: FLAG_EXPLANATIONS.long_context_marker });
  }
  if (flags.includes('sidechain')) {
    out.push({ label: 'sidechain', tone: 'plain', title: FLAG_EXPLANATIONS.sidechain });
  }
  return out;
}

export function Spend({ summary }: { summary: SummaryResponse }) {
  const { spend, pricing } = summary;
  const complete = spend.incomplete_sessions === 0;
  const models = [...spend.by_model].sort((a, b) => b.cost_usd - a.cost_usd || a.model.localeCompare(b.model));

  const rows: { key: keyof Components; label: string; rate: string }[] = [
    { key: 'input_usd', label: 'Fresh input', rate: 'list price' },
    { key: 'output_usd', label: 'Output', rate: 'list price' },
    { key: 'cache_read_usd', label: 'Cache reads', rate: `${pricing.cache_read_multiplier}× input` },
    { key: 'cache_write_usd', label: 'Cache writes', rate: `${pricing.cache_write_multiplier}× input` },
  ];
  rows.sort((a, b) => spend.components[b.key] - spend.components[a.key]);

  return (
    <section id="money" aria-labelledby="spend-h">
      <h2 id="spend-h">Where the money goes</h2>
      <p className="muted">
        Total known cost: <strong>{formatUsdKnown(spend.total_usd, complete)}</strong>.{' '}
        {complete
          ? 'Every session has complete cost.'
          : `${formatCount(spend.incomplete_sessions)} session(s) have incomplete cost, so the total is the known part only: a lower bound, not the full cost. Unknown cost is never counted as zero.`}
      </p>
      <div className="two">
        <div>
          <h3>By model</h3>
          {models.length === 0 ? (
            <p>No model usage was found.</p>
          ) : (
            <div className="table-scroll">
              <table>
                <thead>
                  <tr>
                    <th>Model</th>
                    <th className="n">Requests</th>
                    <th className="n">Known cost</th>
                  </tr>
                </thead>
                <tbody>
                  {models.map((m) => (
                    <tr key={m.model}>
                      <td>
                        {m.model}
                        {modelBadges(m.flags).map((b) => (
                          <span key={b.label}>
                            {' '}
                            <span className={b.tone === 'warn' ? 'badge warn' : 'badge'} title={b.title}>
                              {b.label}
                            </span>
                          </span>
                        ))}
                      </td>
                      <td className="n">{formatCount(m.requests)}</td>
                      <td className="n">{formatAmount(m.cost_usd)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
        <div>
          <h3>By component</h3>
          <div className="table-scroll">
            <table>
              <thead>
                <tr>
                  <th>Component</th>
                  <th className="n">Known cost</th>
                  <th className="n">Rate</th>
                </tr>
              </thead>
              <tbody>
                {rows.map((r) => (
                  <tr key={r.key}>
                    <td>{r.label}</td>
                    <td className="n">{formatAmount(spend.components[r.key])}</td>
                    <td className="n muted">{r.rate}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <p className="muted note">
            Public rates as of {pricing.as_of}
            {pricing.assumptions.length > 0 ? `. ${pricing.assumptions.join(' ')}` : '.'} Source:{' '}
            <span className="code">{pricing.source_url}</span> (text only, not fetched).
          </p>
        </div>
      </div>
    </section>
  );
}
