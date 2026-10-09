import type { SummaryResponse } from '../api/contract.gen';
import { formatCount, formatDate, formatRule, formatUsdKnown } from '../lib/format';

export function Sessions({ summary }: { summary: SummaryResponse }) {
  const rows = summary.session_list;
  return (
    <section aria-labelledby="sessions-h">
      <h2 id="sessions-h">Sessions</h2>
      {rows.length === 0 ? (
        <p>No sessions were found.</p>
      ) : (
        <div className="table-scroll">
          <table>
            <thead>
              <tr>
                <th>Session</th>
                <th>Started</th>
                <th>Ended</th>
                <th className="num">Records</th>
                <th>Category</th>
                <th>Rule</th>
                <th>Known cost</th>
                <th className="num">Requests</th>
                <th>Models</th>
                <th className="num">Tool calls</th>
                <th className="num">Tool errors</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((s) => (
                <tr key={s.id}>
                  <td className="mono">{s.id}</td>
                  <td>{formatDate(s.started_at)}</td>
                  <td>{formatDate(s.ended_at)}</td>
                  <td className="num">{formatCount(s.records)}</td>
                  <td>{s.category}</td>
                  <td>{formatRule(s.rule)}</td>
                  <td>{formatUsdKnown(s.cost_usd, s.cost_complete)}</td>
                  <td className="num">{formatCount(s.requests)}</td>
                  <td className="mono">{s.models.length === 0 ? 'none' : s.models.join(', ')}</td>
                  <td className="num">{formatCount(s.tool_calls)}</td>
                  <td className="num">{formatCount(s.tool_errors)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}
