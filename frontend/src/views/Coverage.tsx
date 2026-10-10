import type { SummaryResponse } from '../api/contract.gen';
import { formatCount, formatPercent, formatPercentFine } from '../lib/format';

export function Coverage({ summary }: { summary: SummaryResponse }) {
  const { records, sessions } = summary;
  const d = records.diagnostics;
  const share = (n: number) => (records.total > 0 ? formatPercent(n / records.total) : formatPercent(null));
  const outcomes: [string, number][] = [
    ['Parsed', records.parsed],
    ['Duplicate, read once', records.duplicate],
    ['Malformed', records.malformed],
    ['Unknown type', records.unknown],
  ];

  return (
    <section id="coverage" aria-labelledby="coverage-h">
      <h2 id="coverage-h">What could not be counted</h2>
      <p className="muted">
        Every nonblank line in the logs ends in exactly one of these four outcomes. Input: {formatCount(summary.source.files)}{' '}
        file(s) under <span className="code">{summary.source.input_hint}</span>.
      </p>
      <div className="two">
        <div>
          {records.total === 0 ? (
            <p>No records were found.</p>
          ) : (
            <div className="table-scroll">
              <table>
                <thead>
                  <tr>
                    <th>Outcome</th>
                    <th className="n">Records</th>
                    <th className="n">Share</th>
                  </tr>
                </thead>
                <tbody>
                  {outcomes.map(([label, n]) => (
                    <tr key={label}>
                      <td>{label}</td>
                      <td className="n">{formatCount(n)}</td>
                      <td className="n">{share(n)}</td>
                    </tr>
                  ))}
                  <tr className="total">
                    <td>Total (nonblank)</td>
                    <td className="n">{formatCount(records.total)}</td>
                    <td className="n">100%</td>
                  </tr>
                </tbody>
              </table>
            </div>
          )}
        </div>
        <div>
          <div className="table-scroll">
            <table>
              <thead>
                <tr>
                  <th>Among parsed records</th>
                  <th className="n">Records</th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td>Requests with no token usage</td>
                  <td className="n">{formatCount(d.missing_usage)}</td>
                </tr>
                <tr>
                  <td>Requests without cache fields</td>
                  <td className="n">{formatCount(d.missing_cache_fields)}</td>
                </tr>
                <tr>
                  <td>Records without a timestamp</td>
                  <td className="n">{formatCount(d.missing_timestamp)}</td>
                </tr>
                <tr>
                  <td>Sidechain requests</td>
                  <td className="n">{formatCount(d.sidechain)}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <p className="muted note">
            Unparsed: {formatPercentFine(records.unparsed_fraction)} of records. Unclassified:{' '}
            {formatCount(sessions.unclassified)} {sessions.unclassified === 1 ? 'session' : 'sessions'}
            {sessions.unclassified > 0
              ? ` (${formatPercent(sessions.unclassified_fraction_sessions)} of sessions, ${formatPercent(sessions.unclassified_fraction_records)} of records)`
              : ''}
            .
          </p>
        </div>
      </div>
    </section>
  );
}
