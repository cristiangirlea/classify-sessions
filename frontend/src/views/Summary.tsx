import type { SummaryResponse } from '../api/contract.gen';
import { formatAmount, formatCount, formatDate, formatPercentWhole, plural } from '../lib/format';
import { rankByCost, rankColor } from '../lib/rank';

export function Summary({ summary }: { summary: SummaryResponse }) {
  const { sessions, spend, records, categories } = summary;
  const ranked = rankByCost(categories);
  const sum = ranked.reduce((acc, c) => acc + c.cost_usd, 0);
  const unparsed = records.malformed + records.unknown;
  const incomplete = spend.incomplete_sessions;
  const half = sessions.total > 0 && incomplete * 2 === sessions.total;

  return (
    <section id="summary" aria-labelledby="summary-h">
      <h1 id="summary-h">What {formatCount(sessions.total)} coding sessions cost, and what they were for</h1>
      <p className="lede">
        The logs account for <strong>USD {formatAmount(spend.total_usd)}</strong> of API spend at public list prices.{' '}
        {incomplete === 0 ? (
          <>Every session is fully priced.</>
        ) : half ? (
          <>
            Half of the sessions, <strong>{formatCount(incomplete)} of {formatCount(sessions.total)}</strong>, have
            requests the logs cannot fully price, so that figure is a lower bound.
          </>
        ) : (
          <>
            In <strong>{formatCount(incomplete)} of {formatCount(sessions.total)}</strong> sessions some requests cannot
            be fully priced, so that figure is a lower bound.
          </>
        )}{' '}
        {records.total === 0 ? (
          <>No records were found.</>
        ) : unparsed === 0 ? (
          <>Every record could be read.</>
        ) : (
          <>
            <strong>
              {plural(unparsed, 'record')} in {formatCount(records.total)}
            </strong>{' '}
            could not be read and {unparsed === 1 ? 'is' : 'are'} counted, not dropped.
          </>
        )}
      </p>

      {sum > 0 ? (
        <>
          <div className="bar" role="img" aria-label="Spend by category">
            {ranked.map((c, i) =>
              c.cost_usd > 0 ? (
                <span
                  key={c.category}
                  style={{ width: `${(c.cost_usd / sum) * 100}%`, background: rankColor(i) }}
                  title={c.category}
                />
              ) : null,
            )}
          </div>
          <ul className="legend">
            {ranked.map((c, i) => (
              <li key={c.category}>
                <i style={{ background: rankColor(i) }} />
                <b>{c.category}</b> {formatAmount(c.cost_usd)} <small>{formatPercentWhole(c.cost_usd / sum)}</small>
              </li>
            ))}
          </ul>
        </>
      ) : (
        <p className="muted">No known cost to show by category.</p>
      )}

      <p className="identity">
        Synthetic dataset, {formatCount(sessions.total)} sessions, generated {formatDate(summary.generated_at)}. An
        agent can ask the same question: <span className="request code">{summary.example_request}</span>
      </p>
    </section>
  );
}
