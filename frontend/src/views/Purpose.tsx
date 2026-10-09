import type { SummaryResponse } from '../api/contract.gen';
import { NOT_MEASURED, formatCount, formatPercent, formatRule } from '../lib/format';

export function Purpose({ summary }: { summary: SummaryResponse }) {
  const { categories, classifier } = summary;
  const { heldout } = classifier;
  return (
    <section aria-labelledby="purpose-h">
      <h2 id="purpose-h">Purpose</h2>
      <p>
        Classifier method: <strong>{classifier.method}</strong>. Taxonomy: {classifier.taxonomy.join(', ')}.
      </p>

      <h3>Categories</h3>
      {categories.length === 0 ? (
        <p>No categories: no sessions were found.</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>Category</th>
              <th className="num">Sessions</th>
              <th className="num">Records</th>
            </tr>
          </thead>
          <tbody>
            {categories.map((c) => (
              <tr key={c.category}>
                <td>{c.category}</td>
                <td className="num">{formatCount(c.sessions)}</td>
                <td className="num">{formatCount(c.records)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      <h3>Held-out check</h3>
      {heldout.labelled === 0 ? (
        <p>No held-out sessions were evaluated. Agreement: {NOT_MEASURED}.</p>
      ) : (
        <p>
          Labelled sessions: {formatCount(heldout.labelled)}. Agreements: {formatCount(heldout.agreements)}.
          Agreement fraction: {formatPercent(heldout.agreement_fraction)}.
        </p>
      )}

      <h3>Disagreements</h3>
      {heldout.disagreements.length === 0 ? (
        <p>No disagreements.</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>Session</th>
              <th>Labelled</th>
              <th>Predicted</th>
              <th>Rule</th>
            </tr>
          </thead>
          <tbody>
            {heldout.disagreements.map((d) => (
              <tr key={d.session_id}>
                <td className="mono">{d.session_id}</td>
                <td>{d.labelled}</td>
                <td>{d.predicted}</td>
                <td>{formatRule(d.rule)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </section>
  );
}
