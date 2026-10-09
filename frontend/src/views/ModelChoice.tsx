import type { ModelCell, ModelsResponse } from '../api/contract.gen';
import {
  NOT_MEASURED,
  categoryBadges,
  formatCountRate,
  formatSessionsCell,
  formatTimes,
  formatUsdKnown,
} from '../lib/format';

export const SYNTHETIC_NOTE =
  'This fixture is synthetic: its prompts are templated, so correction counts describe the generator, not real developers.';

const DEFINITION_LABELS: [keyof ModelsResponse['definitions'], string][] = [
  ['dominant_model', 'Dominant model'],
  ['comparable_category', 'Comparable category'],
  ['supported_category', 'Supported category'],
  ['cost_per_session', 'Cost per session'],
  ['correction_message', 'Correction message'],
  ['correction_rate', 'Correction rate'],
  ['tool_failure_rate', 'Tool failure rate'],
];

const COLUMNS = ['Model', 'Sessions', 'Mean known cost', 'Median known cost', 'Correction rate', 'Tool failure rate'];

function CellRow({ cell }: { cell: ModelCell }) {
  const complete = cell.complete_cost_sessions >= cell.sessions;
  return (
    <tr>
      <td className="mono">{cell.model}</td>
      <td>{formatSessionsCell(cell)}</td>
      <td>{formatUsdKnown(cell.mean_known_usd, complete)}</td>
      <td>{formatUsdKnown(cell.median_known_usd, complete)}</td>
      <td>
        {formatCountRate(cell.correction_messages, cell.user_messages, cell.correction_rate)}
        <br />
        corrected sessions {cell.corrected_sessions}/{cell.sessions}
      </td>
      <td>
        {formatCountRate(cell.tool_failures, cell.tool_results_known, cell.tool_failure_rate)}
        {cell.unknown_outcome_results > 0 && (
          <>
            <br />
            {cell.unknown_outcome_results} result(s) with unknown outcome
          </>
        )}
      </td>
    </tr>
  );
}

function Head() {
  return (
    <thead>
      <tr>
        {COLUMNS.map((c) => (
          <th key={c}>{c}</th>
        ))}
      </tr>
    </thead>
  );
}

export function ModelChoice({ choice }: { choice: ModelsResponse }) {
  const { advice, definitions } = choice;
  return (
    <section aria-labelledby="models-h">
      <h2 id="models-h">Model choice</h2>
      <p>
        Models present: {choice.models.length === 0 ? 'none' : choice.models.join(', ')}. Descriptive only: no
        causal claim.
      </p>

      <h3>Definitions</h3>
      <ul>
        {DEFINITION_LABELS.map(([key, label]) => (
          <li key={key}>
            <strong>{label}</strong>: {definitions[key]}
          </li>
        ))}
      </ul>

      <h3>By category</h3>
      {choice.categories.length === 0 ? (
        <p>No sessions to compare.</p>
      ) : (
        <div className="table-scroll">
          <table>
            <Head />
            {choice.categories.map((cat) => (
              <tbody key={cat.category}>
                <tr>
                  <th colSpan={COLUMNS.length}>
                    {cat.category}{' '}
                    {categoryBadges(cat).map((b) => (
                      <span key={b} className="badge">
                        {b}
                      </span>
                    ))}{' '}
                    <span className="reason">{cat.reason}</span>
                  </th>
                </tr>
                {cat.cells.map((cell) => (
                  <CellRow key={cell.model} cell={cell} />
                ))}
              </tbody>
            ))}
          </table>
        </div>
      )}

      <h3>Totals by model</h3>
      {choice.model_totals.length === 0 ? (
        <p>No model totals.</p>
      ) : (
        <div className="table-scroll">
          <table>
            <Head />
            <tbody>
              {choice.model_totals.map((cell) => (
                <CellRow key={cell.model} cell={cell} />
              ))}
            </tbody>
          </table>
        </div>
      )}

      <h3>Excluded sessions</h3>
      {choice.excluded_sessions.length === 0 ? (
        <p>Excluded sessions (no assistant record): 0.</p>
      ) : (
        <p>
          Excluded sessions (no assistant record): {choice.excluded_sessions.length}:{' '}
          <span className="mono">{choice.excluded_sessions.join(', ')}</span>.
        </p>
      )}

      <h3>Advice</h3>
      <p>{advice.statement}</p>
      <p>
        Verdict: <span className="badge">{advice.verdict}</span>
      </p>
      <p>
        Expensive model: {advice.expensive_model ?? NOT_MEASURED}. Comparator: {advice.comparator_model ?? NOT_MEASURED}.
        Cost ratio: {formatTimes(advice.cost_ratio)}. Shared categories:{' '}
        {advice.shared_categories.length === 0 ? 'none' : advice.shared_categories.join(', ')}.
      </p>
      <h4>Caveats</h4>
      {advice.caveats.length === 0 ? (
        <p>No caveats.</p>
      ) : (
        <ul>
          {advice.caveats.map((c) => (
            <li key={c}>{c}</li>
          ))}
        </ul>
      )}
      <p>{SYNTHETIC_NOTE}</p>
      <p>
        Prices: <span className="mono">{choice.pricing_source}</span> (text only, not fetched), as of{' '}
        {choice.pricing_date}.
      </p>
    </section>
  );
}
