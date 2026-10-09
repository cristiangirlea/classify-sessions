// Scaffold: proves the generated types and the mock route work end to end.
// The four panels (Purpose, Spend, Coverage, Model choice) replace this page.
import { useEffect, useState } from 'react';
import type { SummaryResponse } from './api/contract.gen';

export function App() {
  const [summary, setSummary] = useState<SummaryResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetch('/api/summary')
      .then((r) => (r.ok ? r.json() : Promise.reject(new Error(`HTTP ${r.status}`))))
      .then((s: SummaryResponse) => setSummary(s))
      .catch((e: Error) => setError(e.message));
  }, []);

  if (error) return <p>Could not load the summary: {error}</p>;
  if (!summary) return <p>Loading…</p>;
  return (
    <main>
      <h1>Classify Your Sessions</h1>
      <p>
        {summary.records.total} records, {summary.sessions.total} sessions, known spend USD{' '}
        {summary.spend.total_usd}
      </p>
      <p>
        <code>{summary.example_request}</code>
      </p>
    </main>
  );
}
