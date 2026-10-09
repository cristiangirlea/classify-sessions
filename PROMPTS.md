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
