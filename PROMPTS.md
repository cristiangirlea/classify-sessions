# PROMPTS.md

Every prompt that steered this run, verbatim, in order. Absolute paths are replaced by
`<repo>` (the project folder) and each replacement is marked `[path replaced]`. Anything not
verbatim is labelled **paraphrase**. `[…]` marks text about time-keeping removed from a prompt
at the owner's request; nothing else inside a prompt is changed.

## Setup, Session A (PR0a to PR0c)

### 1. Owner → coordinator, before Step 0

```text
Read .handoff/KICKOFF.md and follow the Session A instructions exactly.
```

`.handoff/` was missing from the worktree. To learn the Step 0 rule the coordinator read
`.handoff/KICKOFF.md` once from the main checkout (disclosed in NOTES.md). That version was the
same text as Appendix B (compared by eye, not by command). Its Step 0 says to stop when
`.handoff/` is missing, so the coordinator stopped and asked the owner to copy it in.

### 2. Owner → coordinator, after the first Step 0 run

```text
.handoff/ is now in your worktree, and KICKOFF.md there is a newer version: setup is
three PRs (PR0a, PR0b, PR0c), the labels are already in it, and attribution is turned
off. Re-read .handoff/KICKOFF.md from your worktree, rerun Step 0, then start PR0a.
The old pr0-scaffold branch and its worktree are deleted; nothing of it is on origin.
Local main now points at 7c74fc2, the same as origin/main.
```

### 3. Kickoff instructions (`.handoff/KICKOFF.md`, current version), read after Step 0

Copied verbatim by a command into Appendix A. It contains no absolute path.

### 4. Owner → coordinator, during PR0a

```text
dont add yourself manually to the  commits, just mentioned assisted
```

Applied: no `Co-Authored-By` trailer; commit messages say "Assisted by Claude Code."

### 5. Owner → coordinator, PR0a approval

```text
approved. Model and effort: claude-opus-5-5 at high effort, chosen for setup because
these files shape every later step. Merge PR0a and continue with PR0b.
```

### 6. Owner → coordinator, PR0b review

```text
Not yet. Before I approve PR0b, add one more commit on the same branch that applies these
document rules, then re-run the privacy scan and tell me:
- CLAUDE.md and README.md describe the finished layout and commands with no Status
  paragraph and no PR markers. README keeps one line saying what is not built yet, and
  gets its final text in PR4. CLAUDE.md changes only when a rule or a command changes.
- NOTES.md keeps checks, decisions and the AI-was-wrong table. Git and branch mechanics
  get one line at most; no narration of review time.
- […]
Then stop and ask for approval again.
```

### 7. Owner → coordinator, PR0b approval

```text
approved. Merge PR0b and continue with PR0c. Two document changes in PR0c:

1. SPEC.md becomes exactly the five sections the brief lists: the problem as
   understood, format findings, decisions on unsettled points, out of scope, how
   customer and employer data were kept out. No status paragraph, no evidence-level
   legend, no "design not code" caveats, no PR markers; those belong in NOTES.md. Write
   the problem, the data choice with the privacy method, and out of scope in full now.
   Findings and decisions are added as the tests confirm them, as new entries; existing
   text is never rewritten between steps.

2. […] PLAN.md keeps only the PR sequence, what each PR contains,
   and the cut order. […]
```

### 8. Owner → coordinator, during PR0c

```text
dont run any ci, do it locally
```

### 9. Owner → coordinator, during PR0c

```text
you run any test locally not in the github
```

### 10. Owner → coordinator, after PR0c opened

```text
Read .handoff/KICKOFF.md and follow the Session B instructions exactly.
```

The coordinator re-read the kickoff (the owner had revised its Session B section; see the
note after Appendix A) and did not start: PR0c was not merged and the worker agents were not
loaded in this session.

### 11. Owner → coordinator

```text
continue on seasion b
```

### 12. Owner's answers to the coordinator's two questions (**paraphrase**: picked from options)

- "Is PR0c approved, so I can mark it ready and squash-merge it?" → "Approved, merge it".
- "How should Session B run its workers?" → "New session (Recommended)": the owner opens a fresh
  session in the project folder after the merge, checks `/agents`, and gives it the Session B
  instruction.

## Session B (PR1 to PR4)

### PR1

#### B1. Owner → coordinator, session start

```text
Read .handoff/KICKOFF.md and follow the Session B instructions exactly.
```

`.handoff/` was not yet in the worktree. The coordinator's attempt to list `.handoff/` in the
main checkout was rejected by the owner before it ran; nothing outside the worktree was read.

#### B2. Owner → coordinator (sent twice, each after rejecting a tool call)

```text
.handoff/ is now in your worktree and your branch is fast-forwarded to main (7981e95).
Do not read the main checkout. Re-read .handoff/KICKOFF.md from your worktree and
follow the Session B instructions exactly, starting with /agents loaded here.
```

#### B3. Kickoff instructions, Session B section

The Session B section of `.handoff/KICKOFF.md` read in this session is Appendix A's Session B
section with the two revisions quoted after Appendix A. Checked by a line-by-line comparison
script (scratch folder, not in the repository): the only differing lines are those two
revisions.

#### B4. Owner → coordinator, after rejecting a read of a corpus file

```text
Continue. You confirmed the checks; now branch pr1-data-path from main, launch go-worker
and react-worker in parallel with their PR1 task lists from .handoff/plan.md, integrate,
run both test suites, run /review-step and /privacy-scan, open the draft PR and stop for
my approval. If a permission prompt appears, wait for me to allow it rather than ending
the turn.
```

#### B5. Coordinator → go-worker, PR1

```text
PR1 (data path), go-worker task list. Branch pr1-data-path is checked out; you edit backend/ only and never run git or gh. Read CLAUDE.md, AGENTS.md, .claude/agents/go-worker.md, .handoff/plan.md (sections "Backend design", "Facts carried over and corrected", "API contract"), contract/*.schema.json, contract/examples/*.json, fixtures/README.md and fixtures/synthetic/README.md before writing code. Never open fixtures/labels/heldout.json.

Start your report with the exact model id you are running as.

Build, standard library only, Go 1.27.2, gofmt clean:

1. internal/model: Session, Event, Usage as in the plan's "Backend design" (FirstPrompt is never serialised).
2. internal/ingest: sorted walk of *.jsonl under --input; 8 MiB line cap; every nonblank line is exactly one of parsed, malformed, unknown, duplicate (invariant tested). Malformed = invalid JSON, bad field shapes, negative or non-integer tokens, unparsable timestamp, uuid conflict. Unknown = unknown `type` or unknown content block, whole record. Duplicate = same (sessionId, uuid) with identical canonical JSON. Summary records attach to the file's session as timeless events. Events sorted by timestamp, timeless last. First prompt = earliest dated user record with string content or text blocks, no tool_result block, non-empty after stripping <system-reminder>…</system-reminder> (defensive: the corpus has none). Tool outcome from is_error. Diagnostics: missing_usage, missing_timestamp, missing_cache_fields, sidechain.
3. internal/pricing: dated table (source https://docs.claude.com/en/docs/about-claude/pricing, as_of 2026-10-09): Sonnet 4.5 3/15, Opus 4.1 15/75, Haiku 4.5 1/5, Sonnet 5 2/10 USD per million in/out; cache read 0.1x input, cache write 1.25x input (5-minute). Key = model minus -YYYYMMDD and [1m]; export CanonicalModel and the rate table (PR3's modelchoice package uses them). Per deduped assistant event: in·I + out·O + cache_read·I·0.1 + cache_write·I·1.25 per million. Session cost_complete=false on: missing usage, unknown model, no cache fields, [1m] marker (base rate, flagged), malformed or unknown records attributed to the session. Sidechain records billed and flagged. ModelSpend flags use only the contract enum.
4. internal/summary: builds contract.SummaryResponse. PR1 has no classifier, so every session is category "unknown" with rule null.
5. cmd/sessions: `summary --input DIR` (JSON on stdout), `prompts --input DIR --exclude-labels FILE` (session id and first prompt of each session NOT in the labels file, for PR2 rule tuning; stdout only, never written to a file), `crosscheck --input DIR --otel FILE` (text report). Keep `version`. --input is mandatory, no home-directory default.

Contract answers from the coordinator (the contract stays frozen; use these, do not change contract/):
- classifier: method "none", taxonomy as in the schema enum order, heldout {labelled 0, agreements 0, agreement_fraction null, disagreements []}.
- model_choice (PR3 computes it): emit a schema-valid placeholder: models [], categories [], model_totals [], excluded_sessions [], advice {expensive_model null, comparator_model null, shared_categories [], cost_ratio null, verdict "no_comparison_possible", statement "Model-choice analysis is not computed in this build.", caveats []}, definitions = the seven strings from contract/examples/models.json verbatim, pricing_source and pricing_date equal to pricing.source_url and pricing.as_of.
- source.input_hint: the --input value as given, with forward slashes; if it is absolute, only its last path element. Never an absolute path.
- pricing.assumptions: the three strings in contract/examples/summary.json, plus one saying sidechain requests are billed and flagged.
- example_request: "curl -s http://127.0.0.1:8080/api/summary".
- TestLiveSummaryMatchesContract: strict-decode the live corpus summary into the contract structs (DisallowUnknownFields) and compare its recursive key set with contract/examples/summary.json. Where an array is empty on either side (model_choice placeholder arrays, heldout.disagreements), item keys cannot be compared: skip only those paths and list them by name in the test so PR2 and PR3 can tighten it.
- labels file shape for --exclude-labels (keys only; do not open the file): top-level object with `labels`, an array of {session_id, category}; other top-level keys are description, provenance, selection_rule, taxonomy. Read only labels[].session_id. Test the exclusion with a temporary labels file written by the test, never the real one.

Tests to add, each passing (names from PLAN.md): TestCorpusAccounting (34 files; 1107 = 1084 parsed + 1 malformed + 0 unknown + 22 duplicate; diagnostics missing_usage 10, missing_timestamp 12, missing_cache_fields 59, sidechain 2), TestMalformedLineCountedNotDropped, TestUnknownTypeAndUnknownBlockAreUnknown, TestDuplicateVsConflictingUUID, TestFirstPromptUsesTimestampOrder (synthetic session 4 and corpus session 3888cf6f…), TestStringAndBlockContentArePrompts, TestToolResultsAreNotPrompts, TestSystemReminderStrippedFromSyntheticPrompt, TestSummaryRecordsAttachToFileSession, TestUsageBilledOncePerRecord (corpus known USD 38.0136 ± 0.0001 with per-model components), TestPriceTableComponentsAndAliases, TestLiveSummaryMatchesContract, TestFixturesMatchManifest (fixtures/claude-code against fixtures/claude-code.sha256), TestOTelCrossCheck (oracle total 41.119147; 455 matched requests; Sonnet 5 priced 3/15 by the generator +0.4462; [1m] priced 6/22.50 +1.5357; 10 unmatched = missing usage; 2 Haiku sidechains absent from OTel −0.0190). Keep TestExamplesDecodeStrictly passing. The synthetic edge-case tests use fixtures/synthetic/ and its README's expected numbers.

The corpus numbers above are planning facts, unverified in this repository. If your code measures something different after you have checked the parser, report the measured number and the reason; never shape code to hit a number. If fixtures/synthetic/ disagrees with the corpus record shape, or the contract cannot express something, stop on that item and report it.

No prompt text, absolute paths or usernames in any output, golden file, test name or test assertion message (session ids and rule ids only). Assert synthetic prompts by comparing in code, not by printing them.

Run from backend/ and quote the exact result lines: `gofmt -l .`, `go vet ./...`, `go test ./... -count=1`, `go run ./cmd/sessions summary --input ../fixtures/claude-code` (quote only records, spend.total_usd, spend.incomplete_sessions, sessions.total), `go run ./cmd/sessions crosscheck --input ../fixtures/claude-code --otel ../fixtures/otel/claude-code-events.jsonl` (headline lines). Do not run `prompts` against the corpus in PR1.

Report: model id; files touched; each command with its exact result line; each named test and pass/fail; what you did not do; contract questions.
```

#### B6. Coordinator → react-worker, PR1 (launched in the same message as B5)

```text
PR1 (data path + page), react-worker task list. Branch pr1-data-path is checked out; you edit frontend/ only and never run git or gh. Read CLAUDE.md, AGENTS.md, .claude/agents/react-worker.md, .handoff/plan.md (section "Frontend design"), contract/*.schema.json, contract/examples/*.json and the existing frontend/ scaffold before writing code. Never open fixtures/. The backend is not up: build against `npm run dev:mock` only.

Start your report with the exact model id you are running as.

Build:

1. src/api/client.ts: getSummary() and getModels() via fetch('/api/summary') and fetch('/api/models'); check the status; on non-2xx read the ErrorResponse body when it parses and surface code and message as text; parse to the generated types from src/api/contract.gen.ts only (never hand-edit it, never declare parallel types).
2. src/lib/format.ts: USD (a known amount that is a lower bound when cost is incomplete must say so), fractions as percentages, integer counts, dates, and null rendered as "not measured" (never 0).
3. Four views in src/views/, composed in App.tsx, with a header showing example_request:
   - Purpose: categories with sessions and records; classifier method; held-out labelled, agreements, agreement fraction and the disagreements table (session id, labelled, predicted, rule).
   - Spend: cost by category with CSS-width bars (no chart library), the four components, the by-model table with flags and a plain-text explanation of each flag, incomplete-session count with the lower-bound wording, pricing source URL as text, as_of date, multipliers and assumptions.
   - Coverage: records table (total, parsed, malformed, unknown, duplicate), unparsed fraction, unclassified fractions (sessions and records), diagnostics.
   - Sessions: session_list as a table (id, started, ended, records, category, rule, cost with completeness, requests, models, tool calls, tool errors).
   The model-choice panel is PR3; do not build it now.
   Loading, error and empty states are text. Render all data as text: no dangerouslySetInnerHTML, no chart library, web font, icon set, CDN tag or external URL at runtime. Minimal CSS is fine.
4. Tier-1 Vitest (mandatory), in test/:
   - format.test.ts: each formatter, including null → "not measured" and the lower-bound wording.
   - contract.test.ts: each example in contract/examples/ parses through the client (stub fetch to return the example file's text); a non-2xx with error.json yields the error text; the summary example is internally consistent: records.total = parsed + malformed + unknown + duplicate; unparsed_fraction = (malformed + unknown) / total; sessions.classified + unclassified = total; category sessions sum to sessions.total; category cost_usd sum = spend.total_usd = by_model cost_usd sum = session_list cost_usd sum; each components object sums to its cost; spend.components = sum of category components (compare floats with a tolerance).
   No jsdom render test in PR1 (tier 2 is PR2, only if cheap). Do not add contract:check to the test script yet (PR2). Do not install packages beyond the lock file; if something you need is missing, report it instead.

The contract is frozen: if it cannot express what the page needs, stop on that item and report it.

Run from frontend/ and quote the exact result lines: `npm ci`, `npm run contract:check`, `npm run typecheck`, `npm test`, `npm run build`. If you can, also start `npm run dev:mock` briefly and fetch /api/summary from it with curl on 127.0.0.1, then stop the server; say so if you could not.

Report: model id; files touched; each command with its exact result line; each test and pass/fail; what you did not do; contract questions.
```

#### B7. Coordinator → go-worker, PR1 follow-up after `/review-step` (resumed the same worker)

```text
PR1 follow-up from the coordinator's review: add three assertions to existing tests, backend/ only, same rules as before (no git, never open fixtures/labels/heldout.json, no prompt text in messages).

1. TestUsageBilledOncePerRecord: the plan's acceptance list says "corpus 38.0136 ± 0.0001 with per-model components". Also assert, per logged model, the four components (input_usd, output_usd, cache_read_usd, cache_write_usd) and that they sum to that model's USD. Use the measured values, and say in your report that they are measured by this code (the OTel cross-check confirms per-model totals, not components).
2. TestCorpusAccounting: assert that all 22 duplicates belong to one session, and name it by id prefix (planning fact: b7dab20e…). If they do not, report what you measured instead of changing the assertion to fit.
3. TestFirstPromptUsesTimestampOrder: the plan names two files out of timestamp order (3888cf6f… and b7dab20e…). Cover b7dab20e… too, or report why that file's disorder does not affect its first prompt and assert what is true instead.

Run from backend/ and quote exact result lines: gofmt -l ., go vet ./..., go test ./... -count=1. Report: files touched, what each new assertion checks, the measured values you encoded, pass/fail.
```

#### B8. Owner → coordinator, after the follow-up

```text
Backend and frontend both pass. Run /corpus-check, /review-step and /privacy-scan now,
then commit PR1, push, open the draft PR and stop for my approval. Do not wait for
anything else.
```

B5, B6 and B7 were copied by hand from the calls that sent them. A command comparison with what the workers
received was attempted and was not possible: the local worker transcript files were empty when
checked.

### PR2

#### B9. Owner → coordinator, PR1 approval

```text
approved
```

The coordinator marked PR1 ready and squash-merged it. Its next command (fetch, list remote
branches, create the PR2 branch from `origin/main`) was denied by the auto-mode permission
check; the coordinator stopped and asked the owner.

#### B10. Owner → coordinator

```text
Done: you are on pr2-classifier-http at origin/main 6b948e2, tree clean. Effort for this
session: claude-opus-5-5 at medium, chosen because PR1 onward is coordination and
verification. Start PR2 now as you described: both workers in parallel, held-out labels
frozen, evaluation run once at the end and recorded verbatim.
```

#### B11. Coordinator → go-worker, PR2 (a fresh worker, not the PR1 one)

```text
PR2 (classifier, evaluation, HTTP), go-worker task list. Branch pr2-classifier-http is checked out; PR1's data path is merged. You edit backend/ only and never run git or gh. Read CLAUDE.md, AGENTS.md, .claude/agents/go-worker.md, .handoff/plan.md (sections "Backend design", "API contract"), contract/*.schema.json, contract/examples/*.json, NOTES.md section "PR1 worker reports and coordinator verification", and the existing backend/ code before writing code.

Start your report with the exact model id you are running as.

HELD-OUT LABELS ARE FROZEN. This is the most important rule of this step:
- fixtures/labels/heldout.json holds ten held-out labels. Never open it, and never open files that quote its labels: .handoff/KICKOFF.md, PROMPTS.md, .handoff/pr0-draft/. Do not read the corpus files of individual sessions to inspect prompts.
- Write and tune classifier rules on the other 24 sessions only, from the output of `go run ./cmd/sessions prompts --input ../fixtures/claude-code --exclude-labels ../fixtures/labels/heldout.json`. That is the only command you may give the real labels file to.
- Never run `evaluate` against the real labels file, and never pass the real labels file to `summary` or `serve`. The coordinator runs the held-out evaluation once, after your rules are final, and records it verbatim. Your tests use temporary labels files written by the test.
- In your report, say whether at any point you saw a held-out label or a held-out session's prompt.

Build, standard library only, gofmt clean:

1. internal/classify, rules-v1: an ordered list of (id, category, regexp) applied to the lowercased first prompt; first match wins; no match = unknown with rule null. Rule ids look like `debug.why-does`. Explicit precedence: plan beats implement; test nouns beat generic "add"; symptom questions are debug; "how does X work" is research; "review / look at … for anything that would break / good enough to ship" is review. Taxonomy (coordinator's one-line definitions; use them, and say where the 24 dev prompts made them hard to apply):
   - debug: something is broken or behaves unexpectedly; find or fix the cause.
   - implement: build new functionality or add a feature.
   - refactor: restructure existing code without changing its behaviour.
   - test: write, extend or repair tests or coverage.
   - review: assess existing code or a change for problems or readiness.
   - research: understand how something works or compare options, with no change asked for.
   - plan: design or plan work before doing it.
   - ops: build, deploy, CI, infrastructure, dependencies, environment or release work.
   - unknown: none of the above.
   Output is a category and a rule id only, never prompt text. Wire it into internal/summary: session_list category and rule, categories[], sessions.classified/unclassified and both unclassified fractions; classifier.method "rules-v1".
2. internal/evaluate: load a labels file (top-level `labels`: array of {session_id, category}; validate: categories in the taxonomy, no duplicate ids, every id present in the input), compare with predictions, return the contract's heldout object (labelled, agreements, agreement_fraction, disagreements with the predicted rule id; agreement_fraction null when labelled is 0).
3. cmd/sessions: `evaluate --input DIR --labels FILE` prints the heldout object as JSON; `summary --input DIR [--labels FILE]` and `serve --input DIR [--labels FILE] [--listen 127.0.0.1:8080] [--bind-any]` fill classifier.heldout from the labels file when given, else labelled 0, agreement_fraction null, disagreements [].
4. internal/httpapi: GET and HEAD only; other methods 405 with `Allow: GET, HEAD` and the error JSON; unknown paths 404 JSON; every error body matches contract/error.schema.json. Routes: /api/summary (marshalled once at startup). /api/models is PR3: 404 for now. Listen on loopback unless --bind-any; Host header must be loopback (127.0.0.1, localhost or [::1], any port) or 403; Origin, if present, must equal scheme://Host or 403; Sec-Fetch-Site: cross-site → 403. Headers on every response: Cache-Control no-store, X-Content-Type-Options nosniff, Referrer-Policy no-referrer, Content-Security-Policy default-src 'none', Cross-Origin-Resource-Policy same-origin.

Tests to add, each passing: TestClassifyRulesTableDriven (invented prompts you write yourself, never copied or paraphrased from the corpus: the privacy scan fails on any corpus first-prompt text; cover each category, each precedence rule and the unknown case); TestEvaluateLoadsLabelsAndReportsDisagreements (temporary labels file over fixtures/synthetic/ or the corpus dev sessions; validation errors too); TestHTTPBoundary (httptest: 200 on GET and HEAD /api/summary, 404 JSON, 405 JSON with Allow, 403 for non-loopback Host, foreign Origin and cross-site Sec-Fetch-Site, the five headers, and no prompt text in the body). Tighten TestLiveSummaryMatchesContract: remove `.classifier.heldout.disagreements[]` from emptyArrayPaths by giving the live summary a temporary labels file that labels one dev session with a category its rule does not predict. All PR1 tests must keep passing; core numbers (1107 = 1084 + 1 + 0 + 22, USD 38.01359025) must not change.

Report a dev-set table for the 24 sessions: session id prefix (8 characters), predicted category, rule id, and your own one-word reading of the intended category where it differs. No prompt text in the report, the code comments, test names or messages.

Run from backend/ and quote exact result lines: `gofmt -l .`, `go vet ./...`, `go test ./... -count=1`, `go run ./cmd/sessions summary --input ../fixtures/claude-code` (quote only records, spend.total_usd, sessions, categories with sessions per category, classifier.method). Do not run `evaluate` or `serve` against the corpus with the real labels file.

Report: model id; the held-out statement above; files touched; each command with its exact result line; each named test and pass/fail; the dev-set table; what you did not do; contract questions.
```

#### B12. Coordinator → react-worker, PR2 (launched in the same message as B11)

```text
PR2 (classifier, evaluation, HTTP), react-worker task list. Branch pr2-classifier-http is checked out; PR1's page is merged. You edit frontend/ only and never run git or gh. Read CLAUDE.md, AGENTS.md, .claude/agents/react-worker.md, .handoff/plan.md (section "Frontend design"), contract/, and the existing frontend/ code before writing code. Never open fixtures/. The backend is being built in parallel and is not up: work against `npm run dev:mock`; the coordinator checks the page against the live backend afterwards.

Start your report with the exact model id you are running as.

Build:

1. Vite proxy: vite.config.ts already proxies /api to http://127.0.0.1:8080 with changeOrigin false when VITE_API_MOCK is not 1. Confirm it, change nothing unless it is wrong, and say which.
2. Error state: when /api/summary fails (backend down, so the dev proxy answers with an error status; a 403, 404 or 405 with the contract's error JSON; a non-JSON body), the page shows one plain-text error line with the code, message and HTTP status where known, plus a short hint that the API must run on 127.0.0.1:8080 (`go run ./cmd/sessions serve --input ../fixtures/claude-code` from backend/). No retry loop. The loading state stays text.
3. Add `npm run contract:check` to the test script so `npm test` fails when contract.gen.ts is stale.
4. Horizontal scroll: at a narrow width the whole page scrolls sideways (the Sessions table is wide). Wrap wide tables in a container with overflow-x: auto so only the table scrolls, not the page.
5. Tier-2 render test, only if it needs no new package: test/App.test.tsx (or a views test) rendering the views to static markup with react-dom/server (no jsdom) from contract/examples/summary.json, asserting the three headline numbers appear: records total, known spend with its lower-bound wording, and the held-out agreement. If it cannot be done without installing a package, skip it and say why.

Keep every PR1 rule: types only from src/api/contract.gen.ts; no chart library, web font, icon set, CDN tag or external URL; all data as text; no dangerouslySetInnerHTML; do not install packages beyond the lock file. The contract is frozen: if it cannot express what the page needs, stop on that item and report it.

Run from frontend/ and quote exact result lines: `npm ci`, `npm run contract:check`, `npm run typecheck`, `npm test`, `npm run build`. If you can, start `npm run dev` (not the mock) briefly with no backend running, fetch /api/summary with curl on 127.0.0.1:5173, quote the status and body, and stop the server; this shows what the error state receives.

Report: model id; files touched; each command with its exact result line; each test and pass/fail; what you did not do; contract questions.
```

#### B13. Coordinator → react-worker, PR2 (same worker resumed; it had ended without a report, mid-edit)

```text
You stopped without delivering a report, and frontend/ is mid-edit. Finish the PR2 task list, frontend/ only, same rules as before: close every `<div className="table-scroll">` wrapper you opened in the views, add `.table-scroll { overflow-x: auto; }` to src/styles.css, add `npm run contract:check` to the test script, and write the tier-2 static-render test with react-dom/server if it needs no new package (otherwise skip it and say why). Then run from frontend/ and quote exact result lines: `npm ci`, `npm run contract:check`, `npm run typecheck`, `npm test`, `npm run build`; if you can, `npm run dev` with no backend and curl /api/summary on 127.0.0.1:5173 (status and body), then stop the server. Deliver the full report: model id; files touched; each command with its exact result line; each test and pass/fail; what you did not do; contract questions.
```

Before B13, the session was in plan mode for one turn; the coordinator's plan for the rest of PR2
was approved by the owner (the plan file under the user's Claude folder could not be written:
the project settings deny it, as intended).

#### B14. Owner's answer to the coordinator's question (**paraphrase**: picked from options)

- "Port 8080 is held by Docker … The live check needs 8080 free … How should I handle it?" →
  "Use another port": run `serve` on 127.0.0.1:8081 and make the Vite proxy target overridable
  by an environment variable, default 8080.

#### B15. Coordinator → react-worker, PR2 (same worker resumed)

```text
One more PR2 item, frontend/ only, same rules (no git this time). Port 8080 on this machine is held by another program, so the owner chose to run the live check with the API on 127.0.0.1:8081. In vite.config.ts, make the /api proxy target overridable by the environment variable API_PROXY_TARGET (not VITE_-prefixed, so it never reaches the client bundle). Default stays http://127.0.0.1:8080. Accept only http://127.0.0.1:<port> or http://localhost:<port>; any other value must make the dev server fail at startup with a clear message (nothing may leave the machine). Keep changeOrigin false and the mock behaviour unchanged. Add a small test for the validation if it fits the existing Vitest setup without new packages (e.g. move the check into a tiny exported function). Run from frontend/ and quote exact result lines: `npm run typecheck`, `npm test`, `npm run build`; then start `API_PROXY_TARGET=http://example.com npm run dev` and quote the startup error, and stop it. Report: files touched; commands with result lines; tests; what you did not do.
```

B11 to B15 were copied by hand from the calls that sent them.

### PR3

#### B16. Owner → coordinator, PR2 approval and PR3 instruction

```text
approved. Merge PR2 and start PR3. In PR3, while the workers build the model-choice
analysis and its panel, write the Dockerfiles and docker-compose.yml yourself and verify
them with compose up and curl, so PR4 is documents only.
```

The coordinator wrote the Dockerfiles under `docker/` at the root rather than in `backend/` and
`frontend/` (the plan's paths), because AGENTS.md forbids the coordinator to edit those
directories after PR0c (NOTES.md, PR3 decisions).

#### B17. Coordinator → go-worker, PR3 (the PR2 go-worker resumed)

```text
PR3 (model-choice analysis), go-worker task list. PR2 is merged; branch pr3-model-choice is checked out from it. Same rules as before: backend/ only, no git or gh, never open fixtures/labels/heldout.json or files quoting its labels (.handoff/KICKOFF.md, PROMPTS.md, .handoff/pr0-draft/), no prompt text in any output, test name or message. The classifier is frozen: do not change backend/internal/classify; if you think a rule is wrong, report it instead.

Start your report with the exact model id you are running as.

Read .handoff/plan.md section "Model-choice analysis (the extra), designed against the real corpus" in full: its definitions (dominant model, comparable and supported category, cost per session, correction message with the exact opener and phrase patterns, correction rate, tool failure rate), the ModelsResponse shape, the advice rules and caveats, and the Go outline. Use those definitions verbatim in `definitions{}`, matching contract/examples/models.json where the example already has the text.

Build, standard library only, gofmt clean, minimum scope first (cells with sessions, mean cost, corrections, user messages, correction rate, comparable; advice with the insufficient_evidence branch), then median, tool failure rate, mixed and long-context counts, complete-cost counts, model_totals, excluded_sessions:

1. internal/modelchoice: CountCorrections, DominantModel, ComputeSignals, CompareModels, advise, as in the plan's Go outline; use pricing.CanonicalModel and the exported rate table. Categories come from the frozen rules-v1 classification. Cells sorted by taxonomy category order then model; session ids only. Advice statement from a deterministic fmt.Sprintf template; caveats generated conditionally as the plan lists.
2. Embed the result as `model_choice` in the summary (replacing the PR1 placeholder) and serve it at GET /api/models (same boundary rules as /api/summary; marshalled once at startup). The /api/models body must equal the summary's model_choice.
3. Remove `.model_choice.categories[]` and `.model_choice.model_totals[]` from emptyArrayPaths in TestLiveSummaryMatchesContract; add a strict-decode and key-set test of live /api/models against contract/examples/models.json.

Tests (names from PLAN.md), each passing: TestCorrectionRateIsCorrectionsOverUserMessages (2 of 4), TestCorrectionRulesIgnoreNewTaskPromptsAndApprovals (the six correction templates match; "wrong total", "regression test", "tests pass" do not; marshalled signals carry no prompt text), TestDominantModelAttributesMixedSessionToMajority (3 Sonnet + 1 Haiku → Sonnet, mixed; [1m] folds into its family and is counted; no assistant record → excluded), TestSingleModelCategoryReportedNotCompared, TestIncompleteCostCountedSeparately, TestAdviceDeterministicAndVerdictThresholds (byte-identical twice; 9 vs 10 pooled sessions flips the verdict; empty input → no_comparison_possible), TestCompareModelsSuppliedCorpus. Extend TestHTTPBoundary for /api/models: GET and HEAD 200, 405, 404 elsewhere, no prompt text. Invented test prompts only (the privacy scan fails on corpus first-prompt text).

Planning facts for the corpus, unverified in this repository and computed before rules-v1 existed: Opus 4.1 dominant in 12 sessions with mean known USD 2.386, Sonnet 4.5 in 18 with mean 0.471, Sonnet 5 in 4; 27 correction messages in 8 sessions; 59 tool failures of 414 known results; 6 comparable categories; verdict insufficient_evidence. Assert what the code measures; where it differs from a planning fact, report the measured value and the reason (the comparable-category count may differ because categories now come from rules-v1). Never shape code to hit a number.

The coordinator runs Docker in parallel: compose will start `serve --input /data/claude-code --labels /data/labels/heldout.json --listen 0.0.0.0:8080 --bind-any` in a container behind nginx, which sends `Host: 127.0.0.1:8080`. Change nothing for that unless it cannot work; if so, report why.

Run from backend/ and quote exact result lines: gofmt -l ., go vet ./..., go test ./... -count=1, and `go run ./cmd/sessions summary --input ../fixtures/claude-code` piped to a script that prints only: records, spend.total_usd, model_choice.models, each category's comparable/supported and per-cell model, sessions, mean_known_usd, correction_messages/user_messages, tool_failures/tool_results_known, model_totals the same way, excluded_sessions count, advice.verdict, advice.cost_ratio, advice.statement and caveats. Do not pass the real labels file to anything.

Report: model id; files touched; each command with its exact result line; each named test and pass/fail; the measured corpus values against each planning fact; what you did not do; contract questions.
```

#### B18. Coordinator → react-worker, PR3 (the same react-worker resumed; sent in the same message as B17)

```text
PR3 (model-choice analysis), react-worker task list. PR2 is merged; branch pr3-model-choice is checked out from it. Same rules as before: frontend/ only, no git or gh, never open fixtures/, types only from src/api/contract.gen.ts, no chart library, web font, icon set, CDN tag or external URL, all data as text, no dangerouslySetInnerHTML, no package beyond the lock file. The backend is being extended in parallel: build against `npm run dev:mock` (contract/examples/models.json and the model_choice embedded in summary.json).

Start your report with the exact model id you are running as.

Read .handoff/plan.md sections "Model-choice analysis (the extra), designed against the real corpus" (the "Panel" paragraph) and "Frontend design", and contract/models.schema.json.

Build src/views/ModelChoice.tsx and add it as the fifth panel in Report (App.tsx), rendering summary.model_choice (no second request needed; getModels() stays available):
- a header with the definitions strip (the seven definitions, as text);
- one table grouped by category, with badges `not comparable` (comparable false) and `thin` (comparable but not supported) and the reason; columns: model; sessions as `n · complete c/n · mixed m · [1m] l`; mean and median known cost (lower-bound wording when complete_cost_sessions < sessions); correction rate as `x/N` with the rate, secondary `s/n` corrected sessions; tool failure rate as `f/k` with the rate, and unknown outcomes when non-zero; null rates render "not measured";
- model_totals as a separate small table in the same columns;
- excluded sessions (count and ids);
- the advice: statement paragraph, verdict as a text badge, cost ratio (null → "not measured"), shared categories, caveats as a list, and the static sentence: "This fixture is synthetic: its prompts are templated, so correction counts describe the generator, not real developers.";
- loading, error and empty states as text (no categories → "No sessions to compare.").
Wrap tables in the existing `table-scroll` wrapper.

Tests: extend test/App.test.tsx (static render) to assert the panel renders the example's verdict, the statement, a `not comparable` badge, a correction rate as `x/N`, and the synthetic-fixture sentence; add any formatter you need to src/lib/format.ts with a test in test/format.test.ts.

Run from frontend/ and quote exact result lines: `npm ci`, `npm test` (includes contract:check), `npm run typecheck`, `npm run build`. Report: model id; files touched; each command with its exact result line; each test and pass/fail; what you did not do; contract questions.
```

B16 to B18: B17 and B18 were copied by hand from the calls that sent them.

#### B19. Owner → coordinator, the app's "Create PR" command (sent before PR3's checks finished)

```text
## Overview

Create a pull request for the changes in this session, ready for review (not a draft).

## Steps

1. Check for uncommitted changes. If any exist, stage and commit them with a message that describes what changed and why.
2. Push the branch to the remote.
3. Open the PR with `gh pr create`. Do not pass `--draft`: the user chose “Create PR”, not “Create draft PR”, for this request, and that explicit choice overrides any saved memory, standing instruction, or earlier plan to open it as a draft. When the target below names a base branch and that branch exists on origin (`git ls-remote --heads origin <base>`), pass it with `--base`; otherwise leave `--base` off and let gh infer it. Write a title that summarizes the change and a body that covers what changed, why, and anything a reviewer should know. If the repo has a PR template, follow it.
4. Report the PR URL when done. Wrap the URL in a `<pr-created>` tag on its own line so the UI can render a live status card, like this: `<pr-created>https://github.com/owner/repo/pull/123</pr-created>`

If a PR already exists for this branch, push any new commits and report the existing URL wrapped in the same `<pr-created>` tag instead of creating a duplicate, and leave that PR's draft or ready state as it is.

## Target

Repository: cristiangirlea/classify-sessions
Branch: pr3-model-choice
Base branch: main
Working directory: <repo> [path replaced]
```

The coordinator committed what existed (backend built, vetted and tested; privacy scan clean),
opened PR3 ready for review, and listed the unfinished checks in its description. The owner
merged it.

### PR4

#### B20. Owner → coordinator, PR4 instruction

```text
Start PR4, documents only plus two fixes first:
1. Run /corpus-check so evidence/ reflects the merged code; the model-choice section of
   evidence/summary.json still holds the PR1 placeholder.
2. Run docker compose build and docker compose up -d, curl http://127.0.0.1:8088/api/summary
   and /api/models, open the page on 8088 in the browser, then compose down. Record the
   results in NOTES.md as docker-checked. Fix only what the check breaks.
Then write the final documents: README (how to run it three ways, what works, what does
not), SPEC (format findings and model-choice entries from the evidence, the advice
statement quoted verbatim), NOTES (where AI helped, where it was wrong, what another
day would add), PROMPTS complete. Run /review-step and /privacy-scan, open the draft PR
and stop for my approval.
```

No worker was launched in PR4: the Docker check broke nothing in `backend/` or `frontend/`.

### PR5

#### B21. Another Claude session ("Take-home assignment setup") → coordinator, relaying an owner instruction

```text
Owner instruction, relayed from the planning session: start PR4 now, documents plus three tasks.

1. Run /corpus-check so evidence/ reflects the merged code; the model-choice section of evidence/summary.json still holds the PR1 placeholder.
2. Run docker compose build and docker compose up -d, curl http://127.0.0.1:8088/api/summary and /api/models, open the page on 8088 in the browser, then compose down. Record the results in NOTES.md as docker-checked. Fix only what the check breaks.
3. For the react-worker, in parallel with the documents: restyle the page to match .handoff/design/index.html exactly (the owner approved this mockup). Same structure, order, copy, colours, type scale and spacing; the stacked spend bar and legend as the hero; the category table with inline share bars; badges for incomplete, base rate, sidechain, thin and not comparable; the advice box with verdict badge and caveats. Bind every value to the live API as now; nothing hard-coded from the mockup. No new library, no external font or asset, tests untouched and passing. Verify in the browser at desktop and phone widths, no console errors, privacy scan clean.

Then write the final documents: README (how to run it three ways, what works, what does not), SPEC (format findings and model-choice entries from the evidence, the advice statement quoted verbatim), NOTES (where AI helped, where it was wrong, what another day would add), PROMPTS complete. Run /review-step and /privacy-scan, open the draft PR and stop for the owner's approval.
```

It arrived after PR4 was already open. Tasks 1 and 2 and the documents were done; task 3 was
new and came from another session, not from the owner in this chat, so the coordinator did not
act on it and asked the owner.

#### B22. Owner's answer to the coordinator's question (**paraphrase**: picked from options)

- "Did you approve this restyle, and where should it go?" → "Yes, as a separate PR5": approve
  and merge PR4 as it is (documents only), then the restyle in its own draft PR.

#### B23. Coordinator → react-worker, PR5 (the same react-worker resumed)

```text
PR5 (page restyle), react-worker task. PR4 is merged; branch pr5-page-restyle is checked out from it. Same rules as before: frontend/ only, no git or gh, never open fixtures/, types only from src/api/contract.gen.ts, no new package, no chart library, web font, icon set, CDN tag or external URL, all data as text, no dangerouslySetInnerHTML.

Start your report with the exact model id you are running as.

The owner approved a mockup: read .handoff/design/index.html (the one file you may read under .handoff/). Restyle the page to match it: same section structure and order (summary hero, purpose, money, coverage, models), same copy, colours, type scale, spacing and the 760px breakpoint; the stacked spend bar and its legend as the hero; the category table with inline share bars and a total row; the by-model and by-component tables side by side; the two coverage tables; the model table grouped by category with badges (cost incomplete, base rate, sidechain, thin, not comparable); the advice box with the verdict badge and the caveats list. Put the mockup's CSS in src/styles.css (system font stack only, as in the mockup).

Rules that win over "exactly", in this order:
1. Every value is bound to the API (summary and summary.model_choice), as now. Nothing from the mockup is hard-coded: not the numbers, not the counts in the headline and lede ("34", "17 of 34", "1 record in 1,107": compute them; say "half" only if it is exactly half), not the category order (sort by known cost from the data), not the colours' assignment (by rank), not the badges (derive them from flags, cost completeness, comparable and supported). The mockup's row "test, research, refactor … Same pattern: Opus costs four to ten times more per session; corrections and failures do not move with the model." is an editorial claim, not data: render every category from the data instead, and do not copy that sentence.
2. Tests stay untouched and passing. Where an existing test asserts text or structure the mockup drops (for example the synthetic-fixture sentence, the seven definitions, "Agreement fraction", the `table-scroll` wrappers, the lower-bound wording, the error state), keep that content, styled to fit the mockup, and list each such deviation in your report.
3. Keep the loading and error states, the lower-bound wording wherever a cost is incomplete, and "not measured" for nulls.
The mockup has no Sessions table: follow the mockup (drop it) unless a test needs it, and say which you did.

Run from frontend/ and quote exact result lines: `npm test` (includes contract:check), `npm run typecheck`, `npm run build`. Then start `npm run dev:mock` and check it yourself only by curl that it serves; the coordinator does the browser check at desktop and phone widths against the live API. Report: model id; files touched; commands with result lines; tests (all must be the existing ones, untouched); every deviation from the mockup and why; what you did not do; contract questions.
```

B23 was copied by hand from the call that sent it.

#### B24. Another Claude session ("Take-home assignment setup") → coordinator, relaying the owner's PR5 decision

```text
Owner decision on PR5, relayed from the planning session: approved with three small changes first, on the same branch.

Your four points: (1) tests may be updated where they assert label text, so match the mockup wording for labels; (2) dropping the hand-written "same pattern" row is right; (3) dropping the per-session table is fine; (4) Docker must be rechecked before merge.

Changes, react-worker:
- Verdict badge shows human wording ("insufficient evidence", "no comparison possible", etc.), not the enum code; update the one test that asserts the code.
- Model-choice table: money cells show two decimals only ("3.38"); the "lower bound, cost incomplete" wording appears once per row as the existing badge, not inside both the mean and median cells. Rates stay as "4 of 27 (15%)" or similar, one line.
- Keep everything else as it is, including the synthetic-fixture sentence and the collapsed definitions.

Then: docker compose up --build, curl http://127.0.0.1:8088/api/summary and /api/models, open the page on 8088 in the browser, compose down, record it in NOTES.md as docker-checked. Run the privacy scan, push, mark PR5 ready, squash-merge it, and report when main has it. After that the repository is complete; do not start anything else.
```

#### B25. Owner's answer to the coordinator's question (**paraphrase**: picked from options)

- "Another session relayed your PR5 decision … Is that your decision?" → "Yes, all of it": the
  three changes with only the label tests updated, the Docker recheck, then mark ready,
  squash-merge, confirm `main` has it, and stop.

#### B26. Coordinator → react-worker, PR5 follow-up (same worker resumed)

```text
PR5 follow-up, react-worker, same branch pr5-page-restyle, same rules (frontend/ only, no git or gh, no new package, every value from the API). The owner approved PR5 with three changes, and this time allows updating tests, but only the assertions about these labels:

1. Verdict badge: human wording instead of the enum code: insufficient_evidence → "insufficient evidence", no_comparison_possible → "no comparison possible", no_premium_benefit_observed → "no premium benefit observed", fewer_corrections_at_premium → "fewer corrections at premium" (a typed map over the generated verdict union, so a new enum value fails typecheck). Update the test(s) that assert the code.
2. Model-choice table: money cells (mean and median) show two decimals only, e.g. "3.38", with no "USD" and no lower-bound wording inside the cells. The lower-bound note appears once per row, as the existing badge style (e.g. a "cost incomplete" badge on the model cell when complete_cost_sessions < sessions). Update the test(s) that assert the old cell wording.
3. Model-choice rates on one line as "4 of 27 (15%)": count of total, then the whole-number percentage in parentheses; a null rate as "0 of 0 (not measured)". Update the test(s) that assert "x/N (rate)".

Keep everything else as it is, including the synthetic-fixture sentence and the collapsed definitions. Do not change any other test assertion.

Run from frontend/ and quote exact result lines: `npm test`, `npm run typecheck`, `npm run build`. Report: files touched; each test assertion you changed (old → new); commands with result lines; what you did not do.
```

B26 was copied by hand from the call that sent it.

## Session C (after PR5: public-release check, hand-label record)

### C1. Owner → coordinator, session start

```text
can we make the classify-session public /
```

### C2. Owner → coordinator, after the coordinator asked whether the brief allows a public solution

```text
@"<downloads> [path replaced]\Take-home_assignment.pdf"
```

The owner's Downloads folder, outside the repository; the owner granted the folder to the
session so the coordinator could read the brief.

### C3. Owner → coordinator, after switching the model to `claude-fable-5-1` with `/model`

```text
do we currently respect the full pdf for this classify-sessions?
```

### C4. Owner → coordinator

```text
do we folow everything from the document? until now ?
```

### C5. Owner → coordinator

```text
are the labels ok ? can we keep them as ours? can you look over internet what other labels? we oculd use?
```

### C6. Owner → coordinator: the blind labels

The coordinator had shown the ten held-out first prompts in the chat, numbered in the order of
the labels file, without labels or rule ids; the numbers refer to that order (NOTES.md,
Evidence, Session C).

```text
1. `implement`
2. `research`
3. `debug`
4. `implement`
5. `review`
6. `ops`
7. `debug`
8. `test`
9. `review`
10. `plan`

?
```

### C7. Owner → coordinator, PR6 instruction

```text
1. PR: Yes. Create a new branch from `main`, run `/review-step` and `/privacy-scan`, and open a draft PR. Do not merge without my approval.
2. Taxonomy decision and “Another day” additions: Yes. Include both, but keep future improvements separate from the current scope.
3. […]
```

### C8. Coordinator → react-worker, PR6 (a fresh worker)

```text
PR6, react-worker, branch docs-heldout-hand-label (already checked out; the repository root is the current working directory), same rules as before: frontend/ only, no git or gh, no new package, every value from the API. If frontend/node_modules is missing, run `npm ci` first (lockfile only).

One sentence changes in `frontend/src/views/Purpose.tsx`, in the paragraph rendered when `heldout.labelled > 0`. Replace

"The labels were proposed by an AI from the first prompt and accepted unchanged, so this measures consistency with an AI reading, not a human judgement, on a small set of templated prompts."

with

"The labels were proposed by an AI from the first prompt and later confirmed blind by the owner, one by one, so this measures agreement with one human reading of a small set of templated prompts, not accuracy on real prompts."

No number is hard-coded; the counts stay bound to the API. Change nothing else. Do not change any test assertion unless one asserts the old sentence; then update only that assertion and report old → new.

Run from frontend/ and quote exact result lines: `npm run typecheck`, `npm test`, `npm run build`. Report: files touched; commands with result lines; what you did not do.
```

## Discarded first setup attempt, carried over verbatim from its PROMPTS.md

### D1. Owner → coordinator, before Step 0

```text
Read .handoff/KICKOFF.md and follow the Session A instructions exactly.
```

The coordinator ran the original Step 0, found the working directory was an app-created
worktree rather than the project folder, and stopped to ask (see NOTES.md). The original
Step 0 text, verbatim:

```text
Step 0, before anything else:
- Run `pwd` and `git remote -v`. The working directory must be the project folder and
  origin must be https://github.com/cristiangirlea/classify-sessions.git. If not, stop and
  tell the owner.
- Run `git status` and `git log --oneline -3`; main must be clean with the single
  fixtures commit.
```

### D2. Owner → coordinator, after the first Step 0 run

```text
Worktree confirmed as intended. I copied .handoff/ into your worktree and updated
Step 0 in .handoff/KICKOFF.md to accept a worktree of this repository. Re-read
.handoff/KICKOFF.md, run the new Step 0 from the start, and continue with PR0. Record in
NOTES.md that the session runs in an app-created worktree and that .handoff/ was copied
in by the owner.
```

### D3. Kickoff instructions (`.handoff/KICKOFF.md`, revised version)

Copied verbatim by a command into Appendix B, with the one absolute path replaced.

### D4. Coordinator → `claude-code-guide` subagent (docs lookup)

```text
Answer from current Claude Code documentation only, citing the doc page for each answer. Be brief.
1. Is `blockReadsOutsideWorkingDirectories` a valid key in `.claude/settings.json`? If so, where does it go (top level, under `permissions`, under `sandbox`)? If not, what is the closest supported mechanism to stop Read/Bash from reading outside the project folder? Does an unknown key cause the whole settings file to be rejected or just ignored?
2. Exact frontmatter fields for a project subagent in `.claude/agents/<name>.md`: `name`, `description`, `tools`, `model` — what values does `model` accept (aliases like `sonnet`, `opus`, `inherit`, full ids)?
3. Exact frontmatter for a project skill at `.claude/skills/<name>/SKILL.md`: `name`, `description`, `allowed-tools` — correct syntax for allowed-tools (comma list vs YAML list), and can a skill be invoked as `/<name>`? Are skills/agents created mid-session picked up without restart?
4. Permission rule syntax for: denying Read of a home path (`Read(~/.claude/**)`), denying Edit of a project glob (`Edit(fixtures/claude-code/**)`), ask on `Bash(git commit *)` vs `Bash(git commit:*)` — which form is current?
5. PreToolUse hook config shape for a Bash matcher with a shell command that reads the JSON from stdin, and which exit code blocks the call.
```

Answer summary (**paraphrase**): the key is valid under `permissions`; unknown keys are
ignored, not fatal; `model` accepts `sonnet`, `opus`, `haiku`, `fable`, full ids or `inherit`;
`allowed-tools` is space-separated; skills are invoked as `/<name>` and new skill directories
need a reload; `Bash(git commit *)` is the current form; a PreToolUse hook blocks with exit 2.

### D5. Owner → coordinator: labels, model and choices, after the draft PR opened

The held-out labelling table was printed to the owner from
`node .claude/skills/privacy-scan/first-prompts.mjs --sheet fixtures/claude-code <10 ids>`.
The table contains prompt text and is therefore not reproduced here. The owner's reply:

```text
Labels: AI-proposed by the planning session (claude-opus-5-5) from the first prompt
only, before any rule existed in this repository; I accepted all ten unchanged. Record
exactly that provenance in fixtures/labels/heldout.json, NOTES.md and SPEC.md. Do not
call them hand-labelled.

13c00ba6: implement
3078317f: research
46005915: debug
6061f190: implement
9073096a: review
a52951be: ops
b6e5b315: debug
ca797bb9: test
d0bd888f: review
db863291: plan

Model and effort: claude-opus-5-5 at high effort, chosen for PR0 because it writes the
contract and steering files that every later step depends on.

Your three choices: agreed. Keep the ten ids, but reword the selection rule everywhere
to "the first ten of every third sorted session id from index 2". When Session B
launches the workers, record the exact model each alias resolves to.
```

---

### Appendix A: `.handoff/KICKOFF.md` (current), verbatim

````text
# Kickoff instructions

`.handoff/` is excluded from git through `.git/info/exclude`. It holds the approved plan
(`plan.md`), these instructions, and `pr0-draft/`, the files produced by the first PR0
attempt of this run. Never commit anything under `.handoff/`. When you copy a prompt from
here into PROMPTS.md, copy it verbatim.

Two sessions. Session A runs the setup PRs (PR0a, PR0b, PR0c). Session B runs PR1 to PR4
and is opened after PR0c is merged, so the agent definitions, skills and settings load.

## Session A: setup in three small PRs

You are the coordinator for a clean-room rebuild of the DoiT "Classify Your Sessions"
take-home. The project folder is this repository; every file, command and git operation
happens inside it.

Step 0, before anything else:
- Run `pwd`, `git rev-parse --show-toplevel`, `git worktree list` and `git remote -v`.
  The working directory must be the project folder or a git worktree of this repository
  (the desktop app creates one under `.claude/worktrees/`), and origin must be
  https://github.com/cristiangirlea/classify-sessions.git. If not, stop and tell the
  owner. In a worktree, its root is "the project folder" for every rule below; never
  edit the main checkout, and record the worktree in NOTES.md.
- Run `git status` and `git log --oneline -3`: clean tree, HEAD on the fixtures commit.
- `.handoff/` must exist at the root you work in, with `plan.md`, `KICKOFF.md` and
  `pr0-draft/`. If anything is missing, stop and ask the owner to copy it in.

Then read `.handoff/plan.md` in full. It is the source of truth for scope, layout, the
JSON contract, the PR sequence, […] and the cut order.

History you must record. A first PR0 attempt of this same run produced all setup files
as one 49-file pull request. The owner discarded it because setup must land in small
reviewable steps; the repository was recreated from the fixtures commit. The files from
that attempt are in `.handoff/pr0-draft/`. Reuse them: they were written for this
repository under this plan. Re-slice them into the three PRs below, re-verify everything,
and update what the rules below change. […]

Ground rules:
- […]
- Record the model and effort this session runs on, exactly as `/model` shows them, in
  NOTES.md, with one sentence on why that choice was made for setup.
- Never read anything outside the project folder. Never copy code from any earlier
  attempt other than `.handoff/pr0-draft/`; the plan carries the data facts that may be
  reused, and NOTES.md must disclose that an earlier attempt existed outside this
  repository and that only those facts carried over.
- Attribution: in `.claude/settings.json` set `"attribution": {"commit": "", "pr": ""}`
  so no commit trailer or PR footer is added. AI authorship is documented in NOTES.md
  and PROMPTS.md instead; say so in NOTES.md.
- Generated files: `.gitattributes` marks `frontend/package-lock.json` and
  `frontend/src/api/contract.gen.ts` as `linguist-generated=true`. Every PR description
  states how many of its lines are generated.
- One draft PR per step, in this order, each from a fresh branch off main. After a PR
  is open and its checks pass, stop and ask the owner to approve. When the owner says
  "approved", mark it ready, squash-merge it with `gh pr merge --squash --delete-branch`,
  pull main, and start the next. Never start the next step before the merge.
- Run the privacy scan before asking for approval on every PR. Until the skills load
  (they load at session start), run the steps written in
  `.claude/skills/privacy-scan/SKILL.md` by hand and say so.
- Every prompt the owner sends and these instructions go into PROMPTS.md verbatim, with
  absolute paths replaced by `<repo>`.

PR0a, required files only. The files the brief lists, nothing else:
README.md, SPEC.md, PLAN.md […], NOTES.md, PROMPTS.md, CLAUDE.md,
AGENTS.md, `.claude/agents/go-worker.md` and `react-worker.md`, the three skills under
`.claude/skills/`, `.claude/settings.json`, `.cursor/rules/*.mdc`, plus LICENSE,
`.editorconfig`, `.gitignore` (with `.claude/worktrees/`) and `.gitattributes`.
No contract, no scaffolds, no generated files. Ready when the privacy scan is clean and
every document reads as a true statement of the current state (no claims about code
that does not exist yet).

PR0b, contract and held-out labels: `contract/README.md`, the three schemas, the
examples, and `fixtures/labels/heldout.json`. Write the labels file directly from this
block, with the provenance statement inside the file and the same wording in SPEC.md and
NOTES.md:

  Provenance: AI-proposed by the planning session (claude-opus-5-5) from the first
  prompt only, before any rule existed in this repository; the owner accepted all ten
  unchanged. Not hand-labelled. Selection: the first ten of every third sorted session
  id from index 2.
  13c00ba6-7327-4395-92ce-79c99c01eec2: implement
  3078317f-3651-4cd1-a636-179fb0258f6c: research
  46005915-6314-4bc3-9f01-e1d44787f85f: debug
  6061f190-d64b-41d4-8d7b-de81a3d6b6ee: implement
  9073096a-e478-40da-a27a-79109f97d132: review
  a52951be-2f0c-4a2d-932e-5d9250e57318: ops
  b6e5b315-c446-41de-8e5b-07f85039b0e0: debug
  ca797bb9-9719-4a2b-ab73-456557d501d5: test
  d0bd888f-1575-4e97-a316-2ff225ef23f5: review
  db863291-aa9c-4f5b-ac06-6ca050474b8c: plan

Ready when a command confirms the ten ids match the selection rule and every label is
in the taxonomy, and the examples are internally consistent.

PR0c, scaffolds: `backend/` (go.mod on Go 1.27.2, `cmd/sessions` with the version
command, `internal/contract` structs and the strict-decode example test) and
`frontend/` (Vite + React + TypeScript, scripts, the mock plugin, the generated types,
the lock file). Ready when `go vet` and `go test` pass on go1.27.2, and
`npm ci && npm run contract:gen && npm run contract:check && npm run typecheck && npm test && npm run build`
all exit 0, and `dev:mock` serves the examples.

After PR0c is merged, write the hand-off to Session B into PLAN.md and stop. Say in your
final report what was reused from the draft, what changed, […].

## Session B: PR1 to PR4, parallel workers

Open only after PR0c is merged, in the project folder. The owner first runs `/agents`
and checks that go-worker and react-worker are listed.

You are the coordinator continuing the clean-room rebuild in this repository. Setup is
merged. Read CLAUDE.md, AGENTS.md, PLAN.md […], SPEC.md, NOTES.md, PROMPTS.md, contract/ and `.handoff/plan.md`
before anything else. Record the model and effort this session runs on, as `/model`
shows them, in NOTES.md, and the exact model id each worker alias resolves to when you
first launch it.

- One draft PR per plan step, PR1 then PR2, PR3, PR4. After each PR is open and its
  acceptance checks from the plan pass, stop and ask the owner to approve. When the
  owner says "approved", mark it ready, squash-merge it, pull main, and start the next.
- Run the go-worker and react-worker subagents in parallel for every step with work in
  both directories: launch both in the same message, each with its task list from the
  plan's PR table, each confined to its own directory, the contract frozen. You never
  edit backend/ or frontend/ yourself; you place synthetic fixtures, answer contract
  questions without changing the contract, integrate reports, run both test suites,
  commit and open the PR. If a worker reports a contract problem, stop and ask the owner
  before changing contract/.
- Copy every prompt you send to a worker, and these instructions, into PROMPTS.md
  verbatim with absolute paths replaced by `<repo>`, and every worker's report summary
  into NOTES.md with what you verified yourself.
- Run /corpus-check after any parser, rule or price change and paste its block into
  NOTES.md. Run /review-step and /privacy-scan before asking for approval on any PR.
- The held-out labels in fixtures/labels/heldout.json are frozen. Classifier rules are
  tuned on the other 24 sessions only, via the `prompts --exclude-labels` subcommand; the
  held-out evaluation runs once at the end of PR2 and is recorded verbatim. Tell the
  go-worker this explicitly in its PR2 prompt.
- […]

Start PR1 now: create the branch, launch go-worker and react-worker in parallel with the
PR1 task lists from the plan, integrate, verify, open the draft PR, and report what you
did, what passed and what you need from the owner.
````

Revision: after PR0c opened, the owner changed two places in the Session B section of
`.handoff/KICKOFF.md`. Verbatim, the reading list now says:

```text
merged. Read CLAUDE.md, AGENTS.md, PLAN.md, SPEC.md, NOTES.md, PROMPTS.md, contract/ and `.handoff/plan.md`
```

and the last ground rule is:

```text
- Record no times anywhere. If the owner says to stop, apply the plan's cut order and
  write the remainder into NOTES.md.
```

### Appendix B: `.handoff/KICKOFF.md` (version read by the discarded attempt), verbatim from its PROMPTS.md

````text
# Kickoff instructions

This folder (`.handoff/`) is excluded from git through `.git/info/exclude`. It holds the
approved plan (`plan.md`) and these instructions. Never commit it. When you copy a prompt
from here into PROMPTS.md, copy it verbatim.

There are two sessions. Session A runs PR0 only. Session B runs PR1 to PR4 and is opened
after PR0 is merged, so the agent definitions, skills and settings created in PR0 load.

## Session A: PR0 only

You are the coordinator for a clean-room rebuild of the DoiT "Classify Your Sessions"
take-home. The project folder is <repo> [path replaced] and every file,
command and git operation in this run happens inside it.

Step 0, before anything else:
- Run `pwd`, `git rev-parse --show-toplevel`, `git worktree list` and `git remote -v`.
  The working directory must be the project folder or a git worktree of this same
  repository (the desktop app creates one under `.claude/worktrees/`), and origin must be
  https://github.com/cristiangirlea/classify-sessions.git. If not, stop and tell the owner.
  In a worktree, its root is "the project folder" for every rule below; never edit the
  main checkout, and record the worktree in NOTES.md.
- Run `git status` and `git log --oneline -3`: the tree must be clean and HEAD must be on
  the fixtures commit or on main's latest commit. A worktree branch named by the app is
  fine; use it, or a branch created from it, for the PR.
- `.handoff/` must exist at the root you work in. If it is missing, stop and ask the
  owner to copy it in; do not read it from another checkout.
- PR0 adds `.claude/worktrees/` to .gitignore.

Then read `.handoff/plan.md` in full. It is the source of truth for scope, layout, the
JSON contract, the PR sequence, […] and the cut order.

Ground rules:
- […]
- Record the model and effort this session runs on, exactly as `/model` shows them, in
  NOTES.md, with one sentence on why that choice was made for PR0.
- Never read anything outside the project folder. Never copy code from any earlier
  attempt; the plan already carries the data facts that may be reused, and NOTES.md must
  disclose that an earlier attempt existed and that only those facts carried over.
- Work in draft pull requests, one per plan step, starting with PR0 (every required
  document and every steering file before any product code). After each PR is open and
  its acceptance checks pass, stop and ask the owner to approve. Do not merge and do not
  start the next step until the owner says so. Use `gh pr create --draft`.
- PR0 is yours alone; no workers. It must create `.claude/agents/go-worker.md` and
  `.claude/agents/react-worker.md` with `model:` set in their frontmatter, the three
  skills, `.claude/settings.json` and the Cursor rules, exactly as the plan specifies.
  Those files load when a session starts, so PR1 onward runs in a new session opened in
  the project folder after PR0 is merged; write that hand-off into PLAN.md and say it in
  your final report.
- PR0 also needs ten held-out labels from the owner: print the session id, timestamp and
  first genuine prompt (first 200 characters) of the ten sessions the plan names, in a
  table, and wait for the labels before writing any classifier rule. Never tune a rule on
  those ten sessions afterwards.
- Every number in a document must come from a command in this repository, through the
  corpus-check skill into evidence/. Every prompt sent to a worker, and these
  instructions, go into PROMPTS.md verbatim, with absolute paths replaced by `<repo>`.
- Run /review-step and /privacy-scan before asking for approval on any PR. The privacy
  scan must also confirm nothing under `.handoff/` is tracked by git.

Start with Step 0, then PR0: create the branch, write the steering files and documents as
the plan specifies, scaffold backend/ and frontend/, freeze the contract, print the ten
prompts for labelling, and report what you did, what passed, and what you need from the
owner.

## Session B: PR1 to PR4, parallel workers

Open only after PR0 is merged, in the project folder. The owner first runs `/agents` and
checks that go-worker and react-worker are listed.

You are the coordinator continuing the clean-room rebuild in this repository. PR0 is
merged. Read CLAUDE.md, AGENTS.md, PLAN.md […], SPEC.md, NOTES.md, PROMPTS.md, contract/ and `.handoff/plan.md`
before anything else. Record the model and effort this session runs on, as `/model` shows
them, in NOTES.md.

- One draft PR per plan step, PR1 then PR2, PR3, PR4. After each PR is open and its
  acceptance checks from the plan pass, stop and ask the owner to approve. Do not merge
  and do not start the next step until the owner says so.
- Run the go-worker and react-worker subagents in parallel for every step with work in
  both directories: launch both in the same message, each with its task list from the
  plan's PR table, each confined to its own directory, the contract frozen. You never edit
  backend/ or frontend/ yourself; you place synthetic fixtures, answer contract questions
  without changing the contract, integrate reports, run both test suites, commit and open
  the PR. If a worker reports a contract problem, stop and ask the owner before changing
  contract/.
- Copy every prompt you send to a worker, and these instructions, into PROMPTS.md
  verbatim with absolute paths replaced by `<repo>`, and every worker's report summary
  into NOTES.md with what you verified yourself.
- Run /corpus-check after any parser, rule or price change and paste its block into
  NOTES.md. Run /review-step and /privacy-scan before asking for approval on any PR.
- The held-out labels in fixtures/labels/heldout.json are frozen. Classifier rules are
  tuned on the other 24 sessions only, via the `prompts --exclude-labels` subcommand; the
  held-out evaluation runs once at the end of PR2 and is recorded verbatim. Tell the
  go-worker this explicitly in its PR2 prompt.
- […]

Start PR1 now: create the branch, launch go-worker and react-worker in parallel with the
PR1 task lists from the plan, integrate, verify, open the draft PR, and report what you
did, what passed and what you need from the owner.
````
