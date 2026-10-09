# SPEC.md

For the human reviewer: what the problem is, what the tool promises, how data flows, what the
corpus shows and which decisions were taken. Who edits which file is in AGENTS.md, not here.

Evidence levels used below: **unit-tested**, **live-HTTP-checked**, **browser-checked**,
**docker-checked**, **owner-confirmed**, **unverified**.

Status (PR0b): the contract (`contract/`) and the held-out labels exist; no code or scaffold
exists yet. Everything below about the backend, the frontend and the endpoints is the
**design** from the approved plan, not a description of running code. Most corpus facts are **unverified in this repository**: they come
from the read-only planning check and are confirmed or corrected by the PR1 tests named beside
them.

## Problem

A platform lead wants to know what developers use Claude Code for and what each kind of use
costs. The input is a folder of Claude Code JSONL session logs (here: the official synthetic
dataset, 34 files). The output is a classification of each session by purpose, the spend per
purpose at public prices, an honest account of what could not be parsed or priced, and one
extra analysis: whether the expensive model buys anything the cheaper one does not.

## Taxonomy

`debug`, `implement`, `refactor`, `test`, `review`, `research`, `plan`, `ops`, and `unknown`
for no match. One category per session, from its first genuine prompt (rules in PR2).

## Held-out evaluation set

Ten sessions, labelled in `fixtures/labels/heldout.json` (PR0b), before any classifier rule
exists. Provenance, in the same words as the file and NOTES.md:

> AI-proposed by the planning session (claude-opus-5-5) from the first prompt only, before any rule existed in this repository; the owner accepted all ten unchanged. Not hand-labelled. Selection: the first ten of every third sorted session id from index 2.

Rules are written from the taxonomy and the other 24 sessions only; the held-out run happens
once, at the end of PR2. Read the agreement figure accordingly: it measures agreement with an
AI reading of the same first prompt, not with a human judgement (**owner-confirmed**
provenance).

## Data flow and contract (contract exists; code from PR1)

```text
fixtures/claude-code/**/*.jsonl ──► backend: ingest ► pricing ► classify ► summary ► httpapi
                                                   │                                  │
                                                   └► modelchoice ───────────────────►│ GET /api/summary
                                                                                      │ GET /api/models
contract/*.schema.json (frozen) ◄── drift tests on both sides ──► frontend (Vite, generated types)
```

- `GET /api/summary` → `SummaryResponse`; `GET /api/models` → `ModelsResponse` (also embedded in
  the summary as `model_choice`). GET and HEAD only; other methods 405 with `Allow`; unknown
  paths 404; every error is `{"error":{"code","message"}}`. Schemas and examples: `contract/`.
- The backend listens on 127.0.0.1:8080; the page on 127.0.0.1:5173 proxies `/api` there. No CORS.
- The summary is computed once at startup. It carries session ids and rule ids, never prompt
  text, absolute paths or usernames.

## Accounting rules

- Every nonblank line is exactly one of parsed, malformed, unknown, duplicate.
  Malformed: invalid JSON, wrong field shapes, negative or non-integer tokens, unparsable
  timestamp, uuid conflict. Unknown: unknown `type` or content block. Duplicate: same
  `(sessionId, uuid)` with identical canonical JSON.
- `unparsed_fraction = (malformed + unknown) / total`.
- First genuine prompt: earliest dated user record whose content is a string or text blocks
  and has no `tool_result` block. Tag stripping (`<system-reminder>`, `<command-*>`) is
  defensive only; the corpus contains none of these tags (grep output in NOTES.md; written to
  `evidence/tags.txt` by `/corpus-check` from PR1).

## Pricing rules

Price key = model id minus `-YYYYMMDD` and `[1m]`. Per deduped assistant event:
`in·I + out·O + cache_read·I·0.1 + cache_write·I·1.25`, per million tokens, 5-minute cache
write assumed. Rates (USD per million input/output) from the public pricing page, dated
2026-10-09: Sonnet 4.5 3/15, Opus 4.1 15/75, Haiku 4.5 1/5, Sonnet 5 2/10 (**unverified** in
this repository until PR1's `TestPriceTableComponentsAndAliases`). A session's cost is
incomplete when any request has missing usage, an unknown model, no cache fields, or a `[1m]`
marker; incomplete cost is reported as a lower bound, never as zero. `[1m]` requests are priced
at base rates because no request exceeds the long-context threshold; they are flagged.

## Expected corpus facts (to be confirmed in PR1)

Confirmed now by command (output in NOTES.md, PR0a): 34 files, 1107 nonblank lines, manifest
34/34 OK.

From planning, **unverified** here until the named test passes:

| Fact | Confirmed by |
| --- | --- |
| 1107 = 1084 parsed + 1 malformed + 0 unknown + 22 duplicate | `TestCorpusAccounting` |
| diagnostics: 10 missing usage, 12 summary records without identity, 59 without cache fields, 2 sidechain | `TestCorpusAccounting` |
| 22 byte-identical duplicates, all in one session | `TestDuplicateVsConflictingUUID` |
| two files out of timestamp order | `TestFirstPromptUsesTimestampOrder` |
| known spend USD 38.0136 at base rates | `TestUsageBilledOncePerRecord` |
| OTel oracle 41.119147, difference explained by Sonnet 5 priced 3/15, `[1m]` priced 6/22.50, 10 missing-usage requests, 2 Haiku sidechains absent | `TestOTelCrossCheck` |

## Extra: model-choice analysis

Descriptive, not causal. For each category with sessions from two or more models, compare
mean and median known cost per session, correction-message rate and tool-failure rate per
dominant model, then give a deterministic verdict (`insufficient_evidence`,
`no_premium_benefit_observed`, `fewer_corrections_at_premium`, `no_comparison_possible`). The
definitions are in `contract/examples/models.json` → `definitions` and will be stated
verbatim by the endpoint (PR3). Planning expects the verdict on this corpus to be `insufficient_evidence`
(**unverified**; written up in PR3 only after reading the endpoint's output).

## Decisions

| Decision | Why |
| --- | --- |
| Go standard library + React/Vite, separate trees, JSON contract as the only boundary | lets two workers build in parallel without touching each other's files |
| Contract frozen once PR0b is merged, drift-tested on both sides (PR0c onward) | the page can be built against examples before the backend exists |
| Nullable numbers (`agreement_fraction`, rates, `cost_ratio`) are `null`, not 0, when the denominator is 0 | 0 would claim a measurement that was not made |
| Held-out set fixed in setup (PR0b) by a stated selection rule, before any classifier rule exists | rules cannot be tuned to the evaluation set |
| Setup in three small PRs (documents and steering, contract and labels, scaffolds) | each step is reviewable on its own; a first single 49-file setup PR was discarded for that reason (NOTES.md) |
| `[1m]` priced at base rate and flagged | no request crosses the long-context threshold |
| OTel file used only as a pricing cross-check | it is a different format and a different pricing generator |
