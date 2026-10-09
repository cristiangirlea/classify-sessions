# NOTES.md

Working notes: sessions and models, tool use, evidence, decisions, what AI got wrong, and what
is left for another day. Times are `date +%H:%M` at the moment of writing.

## Prior knowledge disclosure

An earlier attempt at this brief exists outside this repository. This run is a clean-room
rebuild: **no code was copied from it**. What carried over is limited to facts about the
synthetic dataset and pitfalls, all restated in the plan and re-verified here by tests:
record counts and types, usage shapes, model mix, duplicates, out-of-order files, the pricing
reconciliation against the OTel file, and four pitfalls (colliding display ids when a uuid is
absent, `Host` header checks, CRLF in embedded assets, `GOWORK=off` under a parent workspace).
The `.claude/settings.json` here was written fresh from the plan's list of rules, not copied
from the earlier attempt's file, because no session in this run may read outside the repository.

Separately, inside this run, a **first setup attempt** (19:41 to 20:00, 19 minutes, counted in
the budget) produced all setup files as one 49-file pull request. The owner discarded it because
setup must land in small reviewable steps, and recreated the repository from the fixtures
commit. Its files were kept outside git in `.handoff/pr0-draft/` and are reused, re-sliced and
re-verified in PR0a to PR0c (see "Reuse from the discarded attempt" below).

## AI authorship

The code and documents are written by Claude Code sessions steered by the owner. Commits and
PRs carry no `Co-Authored-By` trailer or generated-by footer: `.claude/settings.json` sets
`"attribution": {"commit": "", "pr": ""}`, and the owner asked that commits only mention the
assistance. Authorship is documented here (sessions, models, tool use) and in PROMPTS.md
(every prompt verbatim).

## Sessions

### Session A (setup: PR0a, PR0b, PR0c)

- Model: Opus 5.5, `claude-opus-5-5` (from the session's own environment description).
  Effort as `/model` shows it: **pending, to be read out by the owner** (the coordinator cannot
  run `/model`). Why this choice for setup: **pending the owner's sentence**.
- Location: an app-created git worktree of this repository,
  `<repo>/.claude/worktrees/session-a-kickoff-e3a8f5`, branch `crs/session-a-kickoff-e3a8f5`.
  The worktree's root is "the project folder" for every rule. PR0a is on `pr0a-steering`,
  created from `main`.
- Step 0 ran twice. First run: `.handoff/` was missing from the worktree (it is excluded from
  git, so the app did not copy it), so the coordinator stopped and asked. It also reported that
  local `main` (`466eeeb`) was behind `origin/main` (`7c74fc2`, the recreated fixtures commit)
  and that the discarded attempt's branch still existed. The owner copied `.handoff/` in,
  deleted that branch and its worktree, and moved local `main` to `7c74fc2`. Second run passed:
  `pwd` and `git rev-parse --show-toplevel` = the worktree; `git worktree list` = main checkout
  and this worktree; origin = `https://github.com/cristiangirlea/classify-sessions.git`; clean
  tree; HEAD = `7c74fc2` = `main` = `origin/main`; `.handoff/` holds `plan.md`, `KICKOFF.md`,
  `pr0-draft/`.
- Disclosure: before the owner's copy, the coordinator read `.handoff/KICKOFF.md` once from the
  main checkout to learn the Step 0 rule, which itself forbids that; it was not repeated, and
  `plan.md` was read only from the worktree copy.

### Discarded first setup attempt (19:41 to 20:00), carried over from its own notes

Not re-verified by this session unless re-run below. Model and effort as the owner reported
them then: `claude-opus-5-5` at high effort, "chosen for PR0 because it writes the contract and
steering files that every later step depends on". It ran in another app-created worktree on
branch `pr0-scaffold`, opened a single draft PR at 19:55 and was ready for approval at 20:00.
Its notes disclosed two slips, both outside the worktree and removed within a minute: a dev
server log written one directory too high, and a temporary backup of a contract example in the
system temp folder. It looked up Claude Code settings, agent, skill and hook syntax through the
`claude-code-guide` subagent (prompt and answer summary in PROMPTS.md); those answers shaped the
settings and skills reused here and remain **unverified** beyond that docs answer.

### Workers

None in setup (coordinator only, per plan). `go-worker` (`model: opus`) and `react-worker`
(`model: sonnet`) are defined for Session B; the owner agreed to these aliases during the
discarded attempt: the Go side carries the accounting, pricing and evaluation logic where a
subtle error costs most; the React side is a typed rendering of a frozen contract. When
Session B launches the workers it records here the exact model id each alias resolved to.

## Reuse from the discarded attempt (`.handoff/pr0-draft/`)

| Draft file(s) | In | Changed |
| --- | --- | --- |
| CLAUDE.md | PR0a | status line; each path and command marked with the PR that adds it; contract frozen "once PR0b is merged"; attribution and generated-file rules added |
| AGENTS.md, `.cursor/rules/*.mdc` | PR0a | "after PR0" made precise (PR0b for the contract, PR0c for the scaffolds) |
| README.md, SPEC.md, PLAN.md | PR0a | rewritten to the current state: no claim about code, contract or evidence files that do not exist yet; held-out labels and provenance moved to PR0b; ledger restarted for this session with the discarded 19 minutes counted; PR0 row split into PR0a to PR0c |
| NOTES.md, PROMPTS.md | PR0a | this session's entries added; the discarded attempt's entries kept, labelled as carried over |
| `.claude/agents/*.md` | PR0a | unchanged |
| `.claude/settings.json` | PR0a | `attribution` block added |
| `.claude/hooks/guard-bash.mjs` | PR0a | unchanged (the settings hook needs it) |
| `.claude/skills/review-step/SKILL.md` | PR0a | PR names PR0a to PR0c; frozen-input rules per setup PR; "not applicable" before PR0c; trailer check |
| `.claude/skills/corpus-check/SKILL.md` | PR0a | tag check writes `evidence/tags.txt`; creates `evidence/`; matching `allowed-tools` |
| `.claude/skills/privacy-scan/*` | PR0a | `--sheet` labelling mode removed from `first-prompts.mjs` (labels now come from the kickoff, so nothing prints prompt text); in `scan.mjs` the stale `evidence/tags.txt` allowance removed and `corpus-check/SKILL.md` allowed to name the tag it greps for |
| LICENSE, `.editorconfig` | PR0a | unchanged |
| `.gitignore` | PR0a | unchanged from `main` (already has `.claude/worktrees/`); the draft's `evidence/*.tmp` line dropped (no step writes such files) |
| `.gitattributes` | PR0a | `linguist-generated=true` for `frontend/package-lock.json` and `frontend/src/api/contract.gen.ts` |
| `contract/**`, `fixtures/labels/heldout.json` | PR0b | to be re-verified there |
| `backend/**`, `frontend/**` | PR0c | to be re-verified there |
| `evidence/manifest.txt`, `evidence/tags.txt` | not reused | written by `/corpus-check` from PR1; PR0a records the same commands' output below |
| `docker-compose.yml` placeholder | not reused | not in any setup PR's file list; PR4 writes the real file |
| `fixtures/README.md`, `fixtures/claude-code.sha256` | not reused | identical to `main` (`diff` empty) |

## Tool use log

| When | Tool | What | Outcome |
| --- | --- | --- | --- |
| Discarded attempt, 19:41 to 19:51 | `claude-code-guide` subagent (docs lookup, not a worker) | settings keys, agent and skill frontmatter, rule syntax, hook exit codes | applied in the draft and reused here; **unverified** beyond the docs answer |
| PR0a | `/privacy-scan`, `/review-step`, `/corpus-check` | the skills are added by PR0a itself and load only in a session started after it is merged | their SKILL.md steps were run by hand; the privacy scan is the same script the skill runs |

## Evidence

### PR0a checks (2026-10-09, run by hand at 21:20)

| Check | Command | Result | Level |
| --- | --- | --- | --- |
| Fixture manifest | `cd fixtures && sha256sum -c claude-code.sha256 \| grep -vc ': OK$'` | `0` (34/34 OK) | command output |
| Files | `ls fixtures/claude-code/projects/*/*.jsonl \| wc -l` | `34` | command output |
| Nonblank lines | `cat fixtures/claude-code/projects/*/*.jsonl \| grep -c .` | `1107` | command output |
| Tags assumed by earlier planning | `grep -c` for `<system-reminder>`, `<command-`, `"isMeta"` over the same files | `0`, `0`, `0` | command output |
| Privacy scan (skill steps by hand) | `node .claude/skills/privacy-scan/scan.mjs` | first run: 1 hit, `.claude/skills/corpus-check/SKILL.md:21 system-reminder-tag` (the new tag check names the tag); file added to the scan's barrier-document list; rerun: `23 files scanned, 34 first-prompt fingerprints, 0 hit(s)`, exit 0 | command output |
| Privacy scan catches leaks | a temporary `planted-leak.md` with a Windows user path and a `.codex` path | 3 hits (`windows-user-path`, `windows-drive-path`, `codex-dir`), exit 1; file removed, rerun exit 0 | command output |
| `.handoff/` not tracked | `git ls-files .handoff \| wc -l` | `0` | command output |
| Settings parse, attribution off | `node -e` JSON.parse of `.claude/settings.json` | `attribution {"commit":"","pr":""}`, 17 deny rules, PreToolUse hook present | command output |
| Bash guard hook, three sample payloads | JSON piped to `node .claude/hooks/guard-bash.mjs` | `go test ./...` exit 0; a read under the Claude projects directory exit 2; `ls .claude/worktrees` exit 0 | command output (not yet observed firing inside Claude Code) |
| Whitespace and line endings | `git diff --check`; `file` on every changed file | clean; no CRLF | command output |

### Review PR0a (2026-10-09 21:25, `/review-step` steps run by hand)

Blocking: none.
Non-blocking: 1. Session A's effort and the reason for the model choice await the owner's
`/model` readout. 2. The three skills cannot be invoked in this session (they load at session
start); their steps were run by hand. 3. `.claude/hooks/guard-bash.mjs` is not named in the
kickoff's PR0a list; it is included because `.claude/settings.json` (which is) calls it.
Commands: `git diff --name-only main` → 22 paths, all coordinator-owned (documents, `.claude/`,
`.cursor/`, root config); `contract/` and `fixtures/` unchanged; no test named for PR0a in
PLAN.md; backend and frontend suites not applicable (directories do not exist); privacy scan
`0 hit(s)`; PROMPTS.md has the setup section; PLAN.md ledger row for PR0a stamped at start;
document numbers (34, 1107, 0/0/0) trace to the commands above, planning numbers in SPEC.md
are labelled unverified with the test that will confirm them.

## Decisions taken in setup

| Decision | Why |
| --- | --- |
| Setup in three PRs: documents and steering (PR0a), contract and labels (PR0b), scaffolds (PR0c) | owner's decision after discarding the single 49-file PR: each step reviewable on its own |
| MIT licence (owner agreed in the discarded attempt) | conventional default for a take-home repository |
| No `Co-Authored-By` trailer or PR footer; authorship documented in NOTES.md and PROMPTS.md | owner's instruction (kickoff and chat) |
| Generated files marked `linguist-generated`, generated line counts stated per PR | reviewers can see how much of a diff is hand-written |
| Bash guard hook parses the hook JSON and tests only `tool_input.command` | the payload always contains a transcript path under the user's Claude directory; grepping stdin would block every command |
| First-prompt extraction for the privacy scan lives in `.claude/skills/privacy-scan/first-prompts.mjs` (Node, repo-local) | the Go parser does not exist in setup; the scan needs the fingerprints at run time |
| Pricing source URL `https://docs.claude.com/en/docs/about-claude/pricing` allowed by the scan | **unverified** (web tools are denied by settings); PR1 must confirm the rates it encodes |

## AI was wrong

| # | Claim | Who | How caught | Fix |
| --- | --- | --- | --- | --- |
| 1 | Session logs contain `<system-reminder>` and `command-*` tags and `isMeta` records that must be stripped | earlier planning (AI-assisted) | grep: 0, 0, 0 (Evidence, PR0a) | stripping kept as defensive code with a synthetic test; SPEC does not claim the tags were observed |

## Another day

Drafted in setup, refined at close: real-session validation; an LLM or embedding classifier
compared against rules-v1 on the same held-out set; an explicit correction signal (OTel
`decision: reject`) for the model-choice analysis; about 30 real sessions per model per category.
