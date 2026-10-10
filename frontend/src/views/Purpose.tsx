import type { SummaryResponse } from '../api/contract.gen';
import { NOT_MEASURED, formatAmount, formatCount, formatPercent, formatRule } from '../lib/format';
import { rankByCost } from '../lib/rank';

export function Purpose({ summary }: { summary: SummaryResponse }) {
  const { classifier, spend, sessions } = summary;
  const { heldout } = classifier;
  const ranked = rankByCost(summary.categories);
  const maxCost = Math.max(0, ...ranked.map((c) => c.cost_usd));
  const totalComplete = spend.incomplete_sessions === 0;

  return (
    <section id="purpose" aria-labelledby="purpose-h">
      <h2 id="purpose-h">What the sessions were used for</h2>
      <p className="muted">
        One category per session, decided by the first thing the developer asked for. Costs are the known part; a
        session is marked incomplete when some of its requests carry no usage.
      </p>
      {ranked.length === 0 ? (
        <p>No categories: no sessions were found.</p>
      ) : (
        <div className="table-scroll">
          <table>
            <thead>
              <tr>
                <th>Category</th>
                <th className="n">Sessions</th>
                <th className="n">Known cost</th>
                <th className="bar-col">Share of spend</th>
                <th className="n">Cost incomplete</th>
              </tr>
            </thead>
            <tbody>
              {ranked.map((c) => (
                <tr key={c.category}>
                  <td>{c.category}</td>
                  <td className="n">{formatCount(c.sessions)}</td>
                  <td className="n">
                    {formatAmount(c.cost_usd)}
                    {c.cost_incomplete_sessions > 0 && <small className="muted"> lower bound</small>}
                  </td>
                  <td className="bar-col">
                    <span className="inbar">
                      <span style={{ width: `${maxCost > 0 ? (c.cost_usd / maxCost) * 100 : 0}%` }} />
                    </span>
                  </td>
                  <td className="n">
                    {formatCount(c.cost_incomplete_sessions)} of {formatCount(c.sessions)}
                  </td>
                </tr>
              ))}
              <tr className="total">
                <td>All sessions</td>
                <td className="n">{formatCount(sessions.total)}</td>
                <td className="n">
                  {formatAmount(spend.total_usd)}
                  {!totalComplete && <small className="muted"> lower bound</small>}
                </td>
                <td />
                <td className="n">
                  {formatCount(spend.incomplete_sessions)} of {formatCount(sessions.total)}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      )}

      <h3>How the classifier was checked</h3>
      {heldout.labelled === 0 ? (
        <p>No held-out sessions were evaluated. Agreement: {NOT_MEASURED}.</p>
      ) : (
        <p>
          {formatCount(heldout.labelled)} sessions were labelled before any rule was written and never used for tuning.
          The rules agree on {formatCount(heldout.agreements)} of them (method {classifier.method}). Agreement fraction:{' '}
          {formatPercent(heldout.agreement_fraction)}. The labels were proposed by an AI from the first prompt and
          accepted unchanged, so this measures consistency with an AI reading, not a human judgement, on a small set of
          templated prompts.
        </p>
      )}
      {heldout.disagreements.length > 0 && (
        <div className="table-scroll">
          <table>
            <thead>
              <tr>
                <th>Disagreement (session)</th>
                <th>Labelled</th>
                <th>Predicted</th>
                <th>Rule</th>
              </tr>
            </thead>
            <tbody>
              {heldout.disagreements.map((d) => (
                <tr key={d.session_id}>
                  <td className="code">{d.session_id}</td>
                  <td>{d.labelled}</td>
                  <td>{d.predicted}</td>
                  <td>{formatRule(d.rule)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}
