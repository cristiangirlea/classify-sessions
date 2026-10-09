import { useEffect, useState } from 'react';
import { getSummary } from './api/client';
import type { SummaryResponse } from './api/contract.gen';
import { Coverage } from './views/Coverage';
import { Purpose } from './views/Purpose';
import { Sessions } from './views/Sessions';
import { Spend } from './views/Spend';

type State =
  | { kind: 'loading' }
  | { kind: 'error'; message: string }
  | { kind: 'ready'; summary: SummaryResponse };

export function App() {
  const [state, setState] = useState<State>({ kind: 'loading' });

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
  if (state.kind === 'error') return <p role="alert">Could not load the summary: {state.message}</p>;

  const { summary } = state;
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
    </main>
  );
}
