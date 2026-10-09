import type { SummaryResponse } from '../api/contract.gen';
import { formatCount, formatPercent } from '../lib/format';

export function Coverage({ summary }: { summary: SummaryResponse }) {
  const { records, sessions, source } = summary;
  const d = records.diagnostics;
  return (
    <section aria-labelledby="coverage-h">
      <h2 id="coverage-h">Coverage</h2>
      <p>
        Input: {formatCount(source.files)} file(s) under <span className="mono">{source.input_hint}</span>.
      </p>

      <h3>Records</h3>
      {records.total === 0 ? (
        <p>No records were found.</p>
      ) : (
        <>
          <table>
            <thead>
              <tr>
                <th>Outcome</th>
                <th className="num">Records</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>Total (nonblank)</td>
                <td className="num">{formatCount(records.total)}</td>
              </tr>
              <tr>
                <td>Parsed</td>
                <td className="num">{formatCount(records.parsed)}</td>
              </tr>
              <tr>
                <td>Malformed</td>
                <td className="num">{formatCount(records.malformed)}</td>
              </tr>
              <tr>
                <td>Unknown</td>
                <td className="num">{formatCount(records.unknown)}</td>
              </tr>
              <tr>
                <td>Duplicate</td>
                <td className="num">{formatCount(records.duplicate)}</td>
              </tr>
            </tbody>
          </table>
          <p>Each nonblank record is counted exactly once: total = parsed + malformed + unknown + duplicate.</p>
        </>
      )}
      <p>Unparsed fraction (malformed + unknown, over total): {formatPercent(records.unparsed_fraction)}.</p>

      <h3>Unclassified</h3>
      <p>
        Sessions: {formatCount(sessions.unclassified)} of {formatCount(sessions.total)} (
        {formatPercent(sessions.unclassified_fraction_sessions)}). Classified: {formatCount(sessions.classified)}.
      </p>
      <p>Records in unclassified sessions: {formatPercent(sessions.unclassified_fraction_records)}.</p>

      <h3>Diagnostics</h3>
      <table>
        <thead>
          <tr>
            <th>Diagnostic</th>
            <th className="num">Records</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>Missing usage</td>
            <td className="num">{formatCount(d.missing_usage)}</td>
          </tr>
          <tr>
            <td>Missing timestamp</td>
            <td className="num">{formatCount(d.missing_timestamp)}</td>
          </tr>
          <tr>
            <td>Missing cache fields</td>
            <td className="num">{formatCount(d.missing_cache_fields)}</td>
          </tr>
          <tr>
            <td>Sidechain</td>
            <td className="num">{formatCount(d.sidechain)}</td>
          </tr>
        </tbody>
      </table>
    </section>
  );
}
