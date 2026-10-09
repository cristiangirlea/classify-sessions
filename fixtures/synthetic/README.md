# fixtures/synthetic

Hand-made edge cases, written by the coordinator for the PR1 parser tests. Not part of the
official dataset and not covered by `fixtures/claude-code.sha256`. The prompt strings are
invented. File name = `sessionId` inside it; session ids start with `5ea00000-`, record uuids
with `5ea0000N-` where N is the session number.

Prices used below: Sonnet 4.5 3/15, Opus 4.1 15/75, Haiku 4.5 1/5 USD per million input/output
tokens; cache read 0.1x input, cache write 1.25x input.

| Session | Edge case | Nonblank lines | Expected accounting | Known USD |
| --- | --- | --- | --- | --- |
| `…0001` | malformed lines | 5 (+1 blank line, not counted) | 3 parsed, 2 malformed: line 3 is truncated JSON, line 4 has `output_tokens: -3` | 0.00345; incomplete (malformed records in the session) |
| `…0002` | unknown type and unknown content block | 4 | 2 parsed, 2 unknown: line 2 `type: "telemetry"`, line 3 an assistant record with a `hologram` content block (whole record unknown, not billed) | 0.0018; incomplete (unknown records in the session) |
| `…0003` | duplicate vs conflicting uuid | 5 | 2 parsed, 2 duplicate, 1 malformed: line 3 is byte-identical to line 2, line 4 is line 2 with keys reordered (same canonical JSON), line 5 reuses line 2's uuid with different content (conflict) | 0.0018 (billed once); incomplete (malformed record) |
| `…0004` | first prompt in timestamp order; block content | 5 | 5 parsed; line 4 has no timestamp (`missing_timestamp`) | 0.0036; complete |
| `…0005` | tool results and `<system-reminder>` | 8 | 8 parsed; 2 tool calls, 3 tool results (line 1 answers a call made before the log starts), 1 `is_error` | 0.0054; complete |
| `…0006` | summary record and pricing flags | 7 | 7 parsed; line 1 is a summary record with no `sessionId`, `uuid` or timestamp | 0.0735; incomplete |

Whole directory: 34 nonblank = 27 parsed + 3 malformed + 2 unknown + 2 duplicate. Diagnostics:
`missing_usage` 1, `missing_timestamp` 2 (session 4 line 4, session 6 line 1),
`missing_cache_fields` 1, `sidechain` 1. Known USD 0.08955.

First genuine prompt per session (earliest dated user record whose text is non-empty after
removing `<system-reminder>…</system-reminder>`, with no `tool_result` block):

- `…0004`: line 3 (13:00, block content), not line 1 (first in file order, 13:05) and never the
  undated line 4.
- `…0005`: line 3, text after stripping `Write unit tests for the date parser`; line 1 is a
  tool result and line 2 is empty after stripping, so neither is a prompt.
- the others: line 1 (session 6: line 2, after the summary record).

Session 6 per request:

| Line | Model | Usage | USD | Flag |
| --- | --- | --- | --- | --- |
| 3 | `claude-sonnet-4-5-20250929[1m]` | in 1000, out 2000, cache read 10000, cache write 4000 | 0.051 (input 0.003, output 0.03, cache read 0.003, cache write 0.015) | `long_context_marker` |
| 4 | `claude-sonnet-4-5-20250929` | none | 0 known | `missing_usage` |
| 5 | `claude-opus-4-1-20250805` | in 100, out 200, no cache fields | 0.0165 | `no_cache_fields` |
| 6 | `claude-haiku-4-5-20251001`, `isSidechain: true`, no `requestId` | in 1000, out 1000 | 0.006 | `sidechain` |
| 7 | `claude-imaginary-9` | in 1000, out 1000 | 0 known, `price_key` null | `unknown_model` |

Checked by a one-off Node script (line count, JSON validity, record types, USD at the rates
above, no CRLF); the durable checks are the Go tests that read this directory.
