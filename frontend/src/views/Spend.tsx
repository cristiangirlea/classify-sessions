import type { Components, ModelSpend, SummaryResponse } from '../api/contract.gen';
import { formatCount, formatUsd, formatUsdKnown } from '../lib/format';

type Flag = ModelSpend['flags'][number];

// Plain-text meaning of each flag the contract can carry.
export const FLAG_EXPLANATIONS: Record<Flag, string> = {
  missing_usage: 'Some requests carry no token usage, so they contribute no cost and their sessions are incomplete.',
  unknown_model: 'The model is not in the price table, so its requests cannot be priced and its sessions are incomplete.',
  no_cache_fields: 'Some usage objects have no cache fields; cache cost is unknown and treated as incomplete, not as zero.',
  long_context_marker: 'The model id carries the [1m] marker; requests are priced at base rates because none exceeds the long-context threshold.',
  sidechain: 'Some requests were made by sidechain agents; they are billed and counted, and flagged for visibility.',
};

const COMPONENT_LABELS: [keyof Components, string][] = [
  ['input_usd', 'Input'],
  ['output_usd', 'Output'],
  ['cache_read_usd', 'Cache read'],
  ['cache_write_usd', 'Cache write'],
];

function ComponentsTable({ components }: { components: Components }) {
  return (
    <div className="table-scroll">
      <table>
        <thead>
          <tr>
            <th>Component</th>
            <th className="num">Known USD</th>
          </tr>
        </thead>
        <tbody>
          {COMPONENT_LABELS.map(([key, label]) => (
            <tr key={key}>
              <td>{label}</td>
              <td className="num">{formatUsd(components[key])}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

export function Spend({ summary }: { summary: SummaryResponse }) {
  const { spend, categories, pricing } = summary;
  const complete = spend.incomplete_sessions === 0;
  const maxCost = Math.max(0, ...categories.map((c) => c.cost_usd));
  const usedFlags = Array.from(new Set(spend.by_model.flatMap((m) => m.flags)));

  return (
    <section aria-labelledby="spend-h">
      <h2 id="spend-h">Spend</h2>
      <p>
        Total: <strong>{formatUsdKnown(spend.total_usd, complete)}</strong>.
      </p>
      {complete ? (
        <p>Every session has complete cost.</p>
      ) : (
        <p>
          {formatCount(spend.incomplete_sessions)} session(s) have incomplete cost. The total is the known part
          only: a lower bound, not the full cost. Unknown cost is never counted as zero.
        </p>
      )}

      <h3>Cost by category</h3>
      {categories.length === 0 ? (
        <p>No categories: no spend to show.</p>
      ) : (
        <div className="table-scroll">
          <table>
            <thead>
              <tr>
                <th>Category</th>
                <th className="num">Known USD</th>
                <th>Share of largest</th>
                <th className="num">Incomplete sessions</th>
              </tr>
            </thead>
            <tbody>
              {categories.map((c) => {
                const width = maxCost > 0 ? (c.cost_usd / maxCost) * 100 : 0;
                return (
                  <tr key={c.category}>
                    <td>{c.category}</td>
                    <td className="num">{formatUsdKnown(c.cost_usd, c.cost_incomplete_sessions === 0)}</td>
                    <td className="barcell">
                      <div className="bar" style={{ width: `${width}%` }} />
                    </td>
                    <td className="num">{formatCount(c.cost_incomplete_sessions)}</td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      <h3>Cost components</h3>
      <ComponentsTable components={spend.components} />

      <h3>By model</h3>
      {spend.by_model.length === 0 ? (
        <p>No model usage was found.</p>
      ) : (
        <div className="table-scroll">
          <table>
            <thead>
              <tr>
                <th>Model</th>
                <th>Price key</th>
                <th className="num">Requests</th>
                <th className="num">Known USD</th>
                <th>Flags</th>
              </tr>
            </thead>
            <tbody>
              {spend.by_model.map((m) => (
                <tr key={m.model}>
                  <td className="mono">{m.model}</td>
                  <td className="mono">{m.price_key ?? 'not in price table'}</td>
                  <td className="num">{formatCount(m.requests)}</td>
                  <td className="num">{formatUsd(m.cost_usd)}</td>
                  <td>{m.flags.length === 0 ? 'none' : m.flags.join(', ')}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      {usedFlags.length > 0 && (
        <>
          <h4>What the flags mean</h4>
          <ul>
            {usedFlags.map((f) => (
              <li key={f}>
                <span className="mono">{f}</span>: {FLAG_EXPLANATIONS[f]}
              </li>
            ))}
          </ul>
        </>
      )}

      <h3>Pricing</h3>
      <p>
        Source: <span className="mono">{pricing.source_url}</span> (text only, not fetched). As of {pricing.as_of}.
      </p>
      <p>
        Cache read multiplier: {pricing.cache_read_multiplier}x input. Cache write multiplier:{' '}
        {pricing.cache_write_multiplier}x input.
      </p>
      <h4>Assumptions</h4>
      {pricing.assumptions.length === 0 ? (
        <p>No assumptions listed.</p>
      ) : (
        <ul>
          {pricing.assumptions.map((a) => (
            <li key={a}>{a}</li>
          ))}
        </ul>
      )}
    </section>
  );
}
