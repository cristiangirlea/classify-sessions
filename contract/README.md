# contract

The only boundary between `backend/` and `frontend/`. Added in PR0b and frozen once PR0b is
merged; a later change is a coordinator-only commit approved by the owner (AGENTS.md).

| File | Endpoint |
| --- | --- |
| `summary.schema.json` | `GET /api/summary` → `SummaryResponse` (embeds `model_choice`) |
| `models.schema.json` | `GET /api/models` → `ModelsResponse` |
| `error.schema.json` | every non-2xx body: `{"error":{"code","message"}}` |
| `examples/*.json` | hand-written, internally consistent, **illustrative**: three fake session ids, not corpus numbers |

Rules: GET and HEAD only; other methods 405 with `Allow: GET, HEAD`; unknown paths 404. All
objects are closed (`additionalProperties: false`) and every key is required; a value that
cannot be measured is `null`, never 0. No prompt text, absolute paths or usernames.

Drift checks:

- Go: `TestExamplesDecodeStrictly` (PR0c; strict decode + key-set round trip of each example);
  `TestLiveSummaryMatchesContract` (PR1) compares the live summary's key set with the example.
- Frontend (PR0c): `npm run contract:gen` writes `src/api/contract.gen.ts` (committed);
  `npm run contract:check` fails if it is stale; `tsc --noEmit` catches field misuse;
  `test/contract.test.ts` (PR1) parses the examples through the client.
