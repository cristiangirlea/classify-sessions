import { useEffect, useState } from 'react';
import { getSummary } from './api/client';
import type { SummaryResponse } from './api/contract.gen';
import { Coverage } from './views/Coverage';
import { ModelChoice } from './views/ModelChoice';
import { Purpose } from './views/Purpose';
import { Sessions } from './views/Sessions';
import { Spend } from './views/Spend';

type State =
  | { kind: 'loading' }
  | { kind: 'error'; message: string }
  | { kind: 'ready'; summary: SummaryResponse };

/** One plain-text error line (code, message and HTTP status are inside `message`) plus a hint. */
export function ErrorState({ message }: { message: string }) {
  return (
    <main>
      <p role="alert">
        Could not load the summary: {message}. The API must be running on 127.0.0.1:8080: from backend/ run{' '}
        <code>go run ./cmd/sessions serve --input ../fixtures/claude-code</code>, then reload this page.
      </p>
    </main>
  );
}

/** The four panels for a loaded summary. */
export function Report({ summary }: { summary: SummaryResponse }) {
  return (
    <main>
      <header>
        <h1>Classify Your Sessions</h1>
        <p>
          Generated {summary.generated_at}. Request: <code>{summary.example_request}</code>
        </p>
      </header>
      <Purpose summary={summary} />
      <Spend summary={summary} />
      <Coverage summary={summary} />
      <Sessions summary={summary} />
      <ModelChoice choice={summary.model_choice} />
    </main>
  );
}

export function App() {
  const [state, setState] = useState<State>({ kind: 'loading' });

  // One request per page load; a failure is shown, never retried.
  useEffect(() => {
    let live = true;
    getSummary().then(
      (summary) => {
        if (live) setState({ kind: 'ready', summary });
      },
      (e: unknown) => {
        if (live) setState({ kind: 'error', message: e instanceof Error ? e.message : String(e) });
      },
    );
    return () => {
      live = false;
    };
  }, []);

  if (state.kind === 'loading') return <p>Loading the summary…</p>;
  if (state.kind === 'error') return <ErrorState message={state.message} />;
  return <Report summary={state.summary} />;
}
