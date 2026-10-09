# PLAN.md

The approved plan, reduced to what the repository needs to carry on: the PR sequence, the
minute budget, the stop rule, the cut order and the wall-clock ledger. The full planning note
lives outside git (`.handoff/`, excluded); scope decisions from it are restated in SPEC.md.
The plan's single setup step (PR0) is carried out as three PRs, PR0a to PR0c, each squash-merged
before the next starts.

## Wall-clock ledger

Times come from `date +%H:%M` when the event happened; none is reconstructed. The 180-minute
budget counts minutes worked across sessions, including the discarded first setup attempt.

- Discarded first setup attempt: 19:41 to 20:00 on 2026-10-09, 19 minutes, counted as spent.
  The minutes between 20:00 and this session's T0 are not counted: no session was working.
- Session A (setup, this run): T0 = **21:18** on 2026-10-09 (`date +%H:%M` after Step 0).
  Budget left at T0: 180 − 19 = 161 minutes, so if the work runs without breaks the timebox
  ends at **23:59**. Cumulative = 19 + minutes since T0.

| Step | Planned min | Start hh:mm | Ready hh:mm | Actual | Cumulative | Slipped or cut |
| --- | --- | --- | --- | --- | --- | --- |
| Pre-repo planning (plan.md) | n/a | not measured | not measured | not measured; disclosed | n/a | n/a |
| PR0 first attempt, one 49-file PR (discarded) | 25 | 19:41 | 20:00 | 19 | 19 | discarded by the owner: setup must land in small reviewable steps; repository recreated from the fixtures commit |
| Step 0, Session A (two runs, see NOTES) | 0 | before T0 | 21:18 | n/a | 19 | first run stopped: `.handoff/` missing from the worktree; owner copied it in |
| PR0a documents and steering | 10 | 21:18 | 21:26 | 8 | 27 | none; draft PR opened 21:26, ready for approval 21:26 |
| PR0b contract and held-out labels | 5 | | | | | |
| PR0c backend and frontend scaffolds | 10 | | | | | |
| PR1 data path + page | 40 | | | | | |
| PR2 classifier, evaluation, HTTP | 40 | | | | | |
| PR3 extra (model choice) | 35 | | | | | |
| PR4 docker, documents, close | 30 | | | | | |
| Buffer | 10 | | | | | |

Planned minutes for PR0a to PR0c are the coordinator's estimate for the re-slice; the plan
gave setup 25 minutes in total, 19 of which the discarded attempt used. The overrun is taken
from the buffer first, then shown per row in "Slipped or cut".

## Hand-off between sessions

PR0a creates `.claude/agents/`, `.claude/skills/` and `.claude/settings.json`. Claude Code loads
these when a session starts, so **PR1 onward runs in a new session (Session B) opened in the
project folder after PR0c is merged**. The owner first runs `/agents` and checks that
`go-worker` and `react-worker` are listed; the new session keeps this ledger's clock. When it
launches the workers, it records in NOTES.md the exact model id each frontmatter alias
(`opus`, `sonnet`) resolved to. The concrete hand-off note is written here after PR0c merges.

## PR sequence

| PR | Window | Coordinator | go-worker | react-worker | Ready when |
| --- | --- | --- | --- | --- | --- |
| PR0a documents and steering | setup | README, SPEC, PLAN (ledger), NOTES, PROMPTS, CLAUDE.md, AGENTS.md, both agent definitions, three skills, settings and Bash guard hook, Cursor rules, LICENSE, `.editorconfig`, `.gitignore`, `.gitattributes`; no contract, scaffold or generated file | idle | idle | privacy scan clean; every document is a true statement of the current state; draft PR; owner approves |
| PR0b contract and held-out labels | setup | `contract/README.md`, three schemas, three examples; `fixtures/labels/heldout.json` with its provenance statement, same wording in SPEC and NOTES | idle | idle | a command confirms the ten ids match the selection rule and every label is in the taxonomy; examples internally consistent; draft PR; owner approves |
| PR0c backend and frontend scaffolds | setup | `backend/` (go.mod 1.27.2, `cmd/sessions version`, `internal/contract` structs, strict-decode example test); `frontend/` (Vite + React + TS, scripts, mock plugin, generated types, lock file) | idle | idle | `go vet` and `go test` pass on go1.27.2; `npm ci && npm run contract:gen && npm run contract:check && npm run typecheck && npm test && npm run build` exit 0; `dev:mock` serves the examples; draft PR; owner approves |
| PR1 data path + page | T+25 → T+65 | synthetic fixtures, contract answers, SPEC findings from worker reports, PROMPTS | `model`, `ingest`, `pricing`, `summary` (all categories unknown), `cmd summary\|prompts`, tests 1–10, strict-decode, manifest, cross-check | `client.ts`, `format.ts`, four views on `dev:mock`, tier-1 Vitest | `go test` green incl. accounting and billing; `summary` shows 1107/1084/1/0/22 and 38.0136; `npm test && npm run build` green; page renders from mock; `/review-step` no blockers; owner approves |
| PR2 classifier, evaluation, HTTP | T+65 → T+105 | held-out evaluation once, recorded verbatim; SPEC classifier section; browser check with live backend; curl 403/404/405 | `classify` rules-v1 from the 24 dev prompts, `evaluate`, `httpapi`, `cmd serve\|evaluate`, tests 11, 12, 14 | Vite proxy, error state, `contract:check` in `test`, tier-2 render test if cheap | five core items on page and by curl; `/corpus-check` in NOTES; privacy scan clean; owner approves |
| PR3 extra | T+105 → T+140 | SPEC extra section after reading the endpoint output | `modelchoice` (minimum scope first), `/api/models`, metric tests | `ModelChoice.tsx` on the example, then live | `curl /api/models` matches the contract; panel renders; core numbers unchanged; owner approves |
| PR4 docker, documents, close | T+140 → T+170 | Dockerfiles, compose on 8088, README, SPEC, NOTES, PLAN ledger, PROMPTS; final review and privacy scan | idle unless a finding is in `backend/` | idle unless a finding is in `frontend/` | documents cite only existing commands and `evidence/` numbers; docker page equals dev page; owner approves |
| Buffer | T+170 → T+180 | | | | |

Named Go tests (acceptance list): `TestCorpusAccounting`, `TestMalformedLineCountedNotDropped`,
`TestUnknownTypeAndUnknownBlockAreUnknown`, `TestDuplicateVsConflictingUUID`,
`TestFirstPromptUsesTimestampOrder`, `TestStringAndBlockContentArePrompts`,
`TestToolResultsAreNotPrompts`, `TestSystemReminderStrippedFromSyntheticPrompt`,
`TestSummaryRecordsAttachToFileSession`, `TestUsageBilledOncePerRecord`,
`TestPriceTableComponentsAndAliases`, `TestClassifyRulesTableDriven`,
`TestEvaluateLoadsLabelsAndReportsDisagreements`, `TestExamplesDecodeStrictly` (PR0c),
`TestLiveSummaryMatchesContract`, `TestHTTPBoundary`, `TestFixturesMatchManifest`,
`TestOTelCrossCheck`; model choice: `TestCorrectionRateIsCorrectionsOverUserMessages`,
`TestCorrectionRulesIgnoreNewTaskPromptsAndApprovals`,
`TestDominantModelAttributesMixedSessionToMajority`, `TestSingleModelCategoryReportedNotCompared`,
`TestIncompleteCostCountedSeparately`, `TestAdviceDeterministicAndVerdictThresholds`,
`TestCompareModelsSuppliedCorpus`.

## Stop rule and cut order

At T+180: stop workers, keep only what compiles and passes, mark the open PR "stopped at
timebox", move the remainder to NOTES.md "Another day", fill the last ledger row.

Cut order if the clock runs out: tier-2 render test → compose and nginx → `TestOTelCrossCheck`
and `crosscheck` (numbers stay in SPEC as "computed once, not tested") → page styling → the
model-choice panel (endpoint stays; README shows the curl) → `prompts` subcommand →
`session_list` on the page.

Never cut: record counts on the page, the held-out evaluation, pricing notes, endpoint plus
page, the five documents, the privacy scan.
