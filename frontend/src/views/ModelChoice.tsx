import { Fragment, type ReactNode } from 'react';
import type { ModelAdvice, ModelCell, ModelsResponse } from '../api/contract.gen';
import {
  NOT_MEASURED,
  categoryBadges,
  formatCountRate,
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

const COLUMNS = ['Category', 'Model', 'Sessions', 'Mean cost', 'Median cost', 'Corrections', 'Tool failures'];

function verdictTone(verdict: ModelAdvice['verdict']): string {
  if (verdict === 'fewer_corrections_at_premium') return 'badge ok';
  if (verdict === 'no_premium_benefit_observed') return 'badge';
  return 'badge warn';
}

/** "x/N (rate)" with the rate muted; the text stays "x/N (rate)" for readers and tests alike. */
function Rate({ n, base, rate }: { n: number; base: number; rate: number | null }) {
  const full = formatCountRate(n, base, rate);
  const cut = full.indexOf(' (');
  return (
    <>
      {full.slice(0, cut)} <span className="muted">{full.slice(cut + 1)}</span>
    </>
  );
}

function CellRow({ cell }: { cell: ModelCell }) {
  const complete = cell.complete_cost_sessions >= cell.sessions;
  return (
    <tr>
      <td />
      <td>{cell.model}</td>
      <td className="n">
        {cell.sessions}
        <br />
        <small className="muted">
          complete {cell.complete_cost_sessions}/{cell.sessions} · mixed {cell.mixed_model_sessions} · [1m]{' '}
          {cell.long_context_sessions}
        </small>
      </td>
      <td className="n">{formatUsdKnown(cell.mean_known_usd, complete)}</td>
      <td className="n">{formatUsdKnown(cell.median_known_usd, complete)}</td>
      <td className="n">
        <Rate n={cell.correction_messages} base={cell.user_messages} rate={cell.correction_rate} />
        <br />
        <small className="muted">
          corrected sessions {cell.corrected_sessions}/{cell.sessions}
        </small>
      </td>
      <td className="n">
        <Rate n={cell.tool_failures} base={cell.tool_results_known} rate={cell.tool_failure_rate} />
        {cell.unknown_outcome_results > 0 && (
          <>
            <br />
            <small className="muted">{cell.unknown_outcome_results} result(s) with unknown outcome</small>
          </>
        )}
      </td>
    </tr>
  );
}

function GroupRow({ children }: { children: ReactNode }) {
  return (
    <tr className="group">
      <td colSpan={COLUMNS.length}>{children}</td>
    </tr>
  );
}

export function ModelChoice({ choice }: { choice: ModelsResponse }) {
  const { advice, definitions } = choice;
  return (
    <section id="models" aria-labelledby="models-h">
      <h2 id="models-h">Is the expensive model worth it?</h2>
      <p className="muted">
        Model choice, per category: cost per session, how often the developer corrected the model, and how often a tool
        call failed. A comparison is supported only when every model compared has at least three sessions in the
        category. Models present: {choice.models.length === 0 ? 'none' : choice.models.join(', ')}.
      </p>

      {choice.categories.length === 0 ? (
        <p>No sessions to compare.</p>
      ) : (
        <div className="table-scroll">
          <table>
            <thead>
              <tr>
                {COLUMNS.map((c, i) => (
                  <th key={c} className={i >= 2 ? 'n' : undefined}>
                    {c}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {choice.categories.map((cat) => (
                <Fragment key={cat.category}>
                  <GroupRow>
                    {cat.category}
                    {categoryBadges(cat).map((b) => (
                      <span key={b} className="badge">
                        {b}
                      </span>
                    ))}
                    <span className="reason muted">{cat.reason}</span>
                  </GroupRow>
                  {cat.cells.map((cell) => (
                    <CellRow key={cell.model} cell={cell} />
                  ))}
                </Fragment>
              ))}
              {choice.model_totals.length > 0 && (
                <>
                  <GroupRow>All categories</GroupRow>
                  {choice.model_totals.map((cell) => (
                    <CellRow key={cell.model} cell={cell} />
                  ))}
                </>
              )}
            </tbody>
          </table>
        </div>
      )}

      <div className="advice">
        <span className={verdictTone(advice.verdict)}>Verdict: {advice.verdict}</span>
        <p>{advice.statement}</p>
        <p className="muted facts">
          Expensive model: {advice.expensive_model ?? NOT_MEASURED}. Comparator: {advice.comparator_model ?? NOT_MEASURED}.
          Cost ratio: {formatTimes(advice.cost_ratio)}. Shared categories:{' '}
          {advice.shared_categories.length === 0 ? 'none' : advice.shared_categories.join(', ')}.
        </p>
      </div>
      {advice.caveats.length === 0 ? (
        <p className="muted">No caveats.</p>
      ) : (
        <ul className="caveats">
          {advice.caveats.map((c) => (
            <li key={c}>{c}</li>
          ))}
        </ul>
      )}

      <p className="muted note">
        Excluded sessions (no assistant record): {choice.excluded_sessions.length}
        {choice.excluded_sessions.length > 0 && (
          <>
            : <span className="code">{choice.excluded_sessions.join(', ')}</span>
          </>
        )}
        .
      </p>
      <p className="muted note">{SYNTHETIC_NOTE}</p>

      <details>
        <summary>Definitions</summary>
        <ul className="defs">
          {DEFINITION_LABELS.map(([key, label]) => (
            <li key={key}>
              <strong>{label}</strong>: {definitions[key]}
            </li>
          ))}
        </ul>
      </details>
    </section>
  );
}
