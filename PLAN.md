# PLAN.md

The approved plan, reduced to the PR sequence, what each PR contains, and the cut order. Each
PR is a draft pull request, approved by the owner and squash-merged before the next starts.
Setup is PR0a to PR0c. PR1 to PR4 run in a new session opened after PR0c is merged, so the
agents, skills and settings added in PR0a load; it records the exact model id each worker
alias (`opus`, `sonnet`) resolves to.

## PR sequence

| PR | Coordinator | go-worker | react-worker |
| --- | --- | --- | --- |
| PR0a documents and steering | README, SPEC, PLAN, NOTES, PROMPTS, CLAUDE.md, AGENTS.md, both agent definitions, three skills, settings and Bash guard hook, Cursor rules, LICENSE, `.editorconfig`, `.gitignore`, `.gitattributes` | idle | idle |
| PR0b contract and held-out labels | `contract/README.md`, three schemas, three examples; `fixtures/labels/heldout.json` with its provenance statement | idle | idle |
| PR0c backend and frontend scaffolds | `backend/` (go.mod 1.27.2, `cmd/sessions version`, `internal/contract` structs, strict-decode example test); `frontend/` (Vite + React + TS, scripts, mock plugin, generated types, lock file); SPEC.md in its five sections | idle | idle |
| PR1 data path + page | synthetic edge-case fixtures, contract answers, SPEC findings from worker reports, PROMPTS | `model`, `ingest`, `pricing`, `summary` (all categories unknown), `cmd summary\|prompts`, tests 1–10, strict-decode, manifest, cross-check | `client.ts`, `format.ts`, four views on `dev:mock`, tier-1 Vitest |
| PR2 classifier, evaluation, HTTP | held-out evaluation once, recorded verbatim; SPEC classifier decisions; browser check with the live backend; curl 403/404/405 | `classify` rules-v1 from the 24 dev prompts, `evaluate`, `httpapi`, `cmd serve\|evaluate`, tests 11, 12, 14 | Vite proxy, error state, `contract:check` in `test`, tier-2 render test if cheap |
| PR3 extra (model choice) | SPEC entries for the extra after reading the endpoint output | `modelchoice` (minimum scope first), `/api/models`, metric tests | `ModelChoice.tsx` on the example, then live |
| PR4 docker, documents, close | Dockerfiles, compose on 127.0.0.1:8088, README final text, SPEC, NOTES, PROMPTS; final review and privacy scan | only a finding in `backend/` | only a finding in `frontend/` |

Named Go tests (acceptance list): `TestCorpusAccounting`, `TestMalformedLineCountedNotDropped`,
`TestUnknownTypeAndUnknownBlockAreUnknown`, `TestDuplicateVsConflictingUUID`,
`TestFirstPromptUsesTimestampOrder`, `TestStringAndBlockContentArePrompts`,
`TestToolResultsAreNotPrompts`, `TestSystemReminderStrippedFromSyntheticPrompt`,
`TestSummaryRecordsAttachToFileSession`, `TestUsageBilledOncePerRecord`,
`TestPriceTableComponentsAndAliases`, `TestClassifyRulesTableDriven`,
`TestEvaluateLoadsLabelsAndReportsDisagreements`, `TestExamplesDecodeStrictly`,
`TestLiveSummaryMatchesContract`, `TestHTTPBoundary`, `TestFixturesMatchManifest`,
`TestOTelCrossCheck`; model choice: `TestCorrectionRateIsCorrectionsOverUserMessages`,
`TestCorrectionRulesIgnoreNewTaskPromptsAndApprovals`,
`TestDominantModelAttributesMixedSessionToMajority`, `TestSingleModelCategoryReportedNotCompared`,
`TestIncompleteCostCountedSeparately`, `TestAdviceDeterministicAndVerdictThresholds`,
`TestCompareModelsSuppliedCorpus`.

## Cut order

If scope must shrink, cut in this order: tier-2 render test → compose and nginx →
`TestOTelCrossCheck` and `crosscheck` (numbers stay in SPEC as "computed once, not tested") →
page styling → the model-choice panel (endpoint stays; README shows the curl) → `prompts`
subcommand → `session_list` on the page.

Never cut: record counts on the page, the held-out evaluation, pricing notes, endpoint plus
page, the five documents, the privacy scan.
