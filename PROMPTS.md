# PROMPTS.md

Every prompt that steered this run, verbatim, in order. Absolute paths are replaced by
`<repo>` (the project folder) and each replacement is marked `[path replaced]`. Anything not
verbatim is labelled **paraphrase**. Times are `date +%H:%M`; "not stamped" means no command
recorded the time.

## Setup, Session A (PR0a to PR0c)

### 1. Owner → coordinator, before T0, not stamped

```text
Read .handoff/KICKOFF.md and follow the Session A instructions exactly.
```

`.handoff/` was missing from the worktree. To learn the Step 0 rule the coordinator read
`.handoff/KICKOFF.md` once from the main checkout (disclosed in NOTES.md). That version was the
same text as Appendix B (compared by eye, not by command). Its Step 0 says to stop when
`.handoff/` is missing, so the coordinator stopped and asked the owner to copy it in.

### 2. Owner → coordinator, before T0 (T0 = 21:18), not stamped

```text
.handoff/ is now in your worktree, and KICKOFF.md there is a newer version: setup is
three PRs (PR0a, PR0b, PR0c), the labels are already in it, and attribution is turned
off. Re-read .handoff/KICKOFF.md from your worktree, rerun Step 0, then start PR0a.
The old pr0-scaffold branch and its worktree are deleted; nothing of it is on origin.
Local main now points at 7c74fc2, the same as origin/main.
```

### 3. Kickoff instructions (`.handoff/KICKOFF.md`, current version), read at 21:18

Copied verbatim by a command into Appendix A. It contains no absolute path.

### 4. Owner → coordinator, during PR0a (after 21:20), not stamped

```text
dont add yourself manually to the  commits, just mentioned assisted
```

Applied: no `Co-Authored-By` trailer; commit messages say "Assisted by Claude Code."

### 5. Owner → coordinator, after PR0a was ready (21:26), received by 21:33, not stamped

```text
approved. Model and effort: claude-opus-5-5 at high effort, chosen for setup because
these files shape every later step. Merge PR0a and continue with PR0b.
```

### 6. Owner → coordinator, after PR0b was ready (21:38), not stamped

```text
Not yet. Before I approve PR0b, add one more commit on the same branch that applies these
document rules, then re-run the privacy scan and tell me:
- CLAUDE.md and README.md describe the finished layout and commands with no Status
  paragraph and no PR markers. README keeps one line saying what is not built yet, and
  gets its final text in PR4. CLAUDE.md changes only when a rule or a command changes.
- NOTES.md keeps checks, decisions and the AI-was-wrong table. Git and branch mechanics
  get one line at most; no narration of review time.
- PLAN.md ledger: start, ready and minutes per PR, nothing else in the notes column.
Then stop and ask for approval again.
```

## Discarded first setup attempt (19:41 to 20:00), carried over verbatim from its PROMPTS.md

### D1. Owner → coordinator, before T0, not stamped

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

### D2. Owner → coordinator, before T0 (T0 = 19:41), not stamped

```text
Worktree confirmed as intended. I copied .handoff/ into your worktree and updated
Step 0 in .handoff/KICKOFF.md to accept a worktree of this repository. Re-read
.handoff/KICKOFF.md, run the new Step 0 from the start, and continue with PR0. Record in
NOTES.md that the session runs in an app-created worktree and that .handoff/ was copied
in by the owner.
```

### D3. Kickoff instructions (`.handoff/KICKOFF.md`, revised version), read at 19:41

Copied verbatim by a command into Appendix B, with the one absolute path replaced.

### D4. Coordinator → `claude-code-guide` subagent (docs lookup), between 19:41 and 19:51, not stamped

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

### D5. Owner → coordinator: labels, model and choices, received after the draft PR opened at 19:55, not stamped

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
JSON contract, the PR sequence, the minute budget, the stop rule and the cut order.

History you must record. A first PR0 attempt of this same run produced all setup files
as one 49-file pull request. The owner discarded it because setup must land in small
reviewable steps; the repository was recreated from the fixtures commit. The files from
that attempt are in `.handoff/pr0-draft/`. Reuse them: they were written for this
repository under this plan. Re-slice them into the three PRs below, re-verify everything,
and update what the rules below change. Record in the ledger that the discarded attempt
ran 19:41 to 20:00 (19 minutes) and count it as time spent.

Ground rules:
- The clock continues. Run `date +%H:%M` for a new T0 of this session and keep a ledger
  in PLAN.md with one row per PR, stamped from `date +%H:%M` at open and at "ready".
  Never reconstruct a time. The 180-minute budget includes the discarded 19 minutes.
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
README.md, SPEC.md, PLAN.md (with the ledger), NOTES.md, PROMPTS.md, CLAUDE.md,
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
final report what was reused from the draft, what changed, and the ledger so far.

## Session B: PR1 to PR4, parallel workers

Open only after PR0c is merged, in the project folder. The owner first runs `/agents`
and checks that go-worker and react-worker are listed.

You are the coordinator continuing the clean-room rebuild in this repository. Setup is
merged. Read CLAUDE.md, AGENTS.md, PLAN.md (the ledger holds the clock; keep using it
with `date +%H:%M`), SPEC.md, NOTES.md, PROMPTS.md, contract/ and `.handoff/plan.md`
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
- Watch the clock. At T+180 apply the plan's stop rule and cut order, mark the open PR
  "stopped at timebox", and write the remainder into NOTES.md.

Start PR1 now: create the branch, launch go-worker and react-worker in parallel with the
PR1 task lists from the plan, integrate, verify, open the draft PR, and report what you
did, what passed and what you need from the owner.
````

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
JSON contract, the PR sequence, the minute budget, the stop rule and the cut order.

Ground rules:
- The clock starts after Step 0. Run `date +%H:%M`, record it as T0 in a ledger you
  create in PLAN.md, and stamp every PR open and every "ready for approval" with
  `date +%H:%M`. Never reconstruct a time. Stop at T+180 per the plan's stop rule.
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
merged. Read CLAUDE.md, AGENTS.md, PLAN.md (the ledger holds T0; keep using that clock
with `date +%H:%M`), SPEC.md, NOTES.md, PROMPTS.md, contract/ and `.handoff/plan.md`
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
- Watch the clock. At T+180 apply the plan's stop rule and cut order, mark the open PR
  "stopped at timebox", and write the remainder into NOTES.md.

Start PR1 now: create the branch, launch go-worker and react-worker in parallel with the
PR1 task lists from the plan, integrate, verify, open the draft PR, and report what you
did, what passed and what you need from the owner.
````
