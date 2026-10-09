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

- Model and effort, as the owner read them from `/model`: `claude-opus-5-5` at high effort.
- Why (owner): chosen for setup because these files shape every later step.
- Location: app-created worktree `<repo>/.claude/worktrees/session-a-kickoff-e3a8f5`; its root
  is "the project folder" for every rule.
- Git: one branch per PR from `origin/main`, squash-merged; `main` is never checked out here.
- Disclosure: before the owner copied `.handoff/` into the worktree, the coordinator read
  `.handoff/KICKOFF.md` once from the main checkout, which the kickoff forbids; not repeated,
  and `plan.md` was read only from the worktree copy.

### Discarded first setup attempt (19:41 to 20:00), carried over from its own notes

Not re-verified by this session unless re-run below. Model and effort as the owner reported
them then: `claude-opus-5-5` at high effort, "chosen for PR0 because it writes the contract and
steering files that every later step depends on". Its notes disclosed two slips, both outside the worktree and removed within a minute: a dev
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

### Held-out labels (PR0b)

Written into `fixtures/labels/heldout.json` directly from the kickoff block. Provenance, in the
same words as the file and SPEC.md:

> AI-proposed by the planning session (claude-opus-5-5) from the first prompt only, before any rule
> existed in this repository; the owner accepted all ten unchanged. Not hand-labelled. Selection:
> the first ten of every third sorted session id from index 2.

Checked by command (Evidence, PR0b): the ten ids are exactly the selection rule's ids, all ten
categories are in the taxonomy, and the file is identical to the kickoff block. Frozen.

Caveat for reading the PR2 agreement number (carried over from the discarded attempt's notes):
proposer and rule-writer are both Claude models, so agreement measures consistency with an AI
reading of the same first prompt, not with a human judgement; and the planning session also
read the earlier attempt (read-only, for data facts), so independence from that attempt's
taxonomy is not guaranteed.

## Reuse from the discarded attempt (`.handoff/pr0-draft/`)

| Draft file(s) | In | Changed |
| --- | --- | --- |
| CLAUDE.md | PR0a | attribution and generated-file rules added; describes the finished layout and commands (owner's document rules) |
| AGENTS.md, `.cursor/rules/*.mdc` | PR0a | "after PR0" made precise (PR0b for the contract, PR0c for the scaffolds) |
| README.md | PR0a | finished layout and commands, one line on what is not built yet; final text in PR4 |
| SPEC.md, PLAN.md | PR0a | no claim about code that does not exist yet; held-out labels and provenance moved to PR0b; ledger restarted with the discarded 19 minutes counted; PR0 row split into PR0a to PR0c |
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
| `contract/*.schema.json`, `contract/examples/models.json`, `contract/examples/error.json` | PR0b | unchanged; re-verified (Evidence, PR0b) |
| `contract/examples/summary.json` | PR0b | `"sidechain"` flag added to the Sonnet `by_model` row: the draft reported 1 sidechain record with no model flagged (AI was wrong, #2) |
| `contract/README.md` | PR0b | "frozen at the end of PR0" → "once PR0b is merged"; Go and frontend drift checks marked PR0c |
| `fixtures/labels/heldout.json` | PR0b | "Frozen in PR0" → "PR0b"; provenance statement extended with the selection sentence so it matches the kickoff wording exactly |
| `backend/**`, `frontend/**` | PR0c | to be re-verified there |
| `evidence/manifest.txt`, `evidence/tags.txt` | not reused | written by `/corpus-check` from PR1; PR0a records the same commands' output below |
| `docker-compose.yml` placeholder | not reused | not in any setup PR's file list; PR4 writes the real file |
| `fixtures/README.md`, `fixtures/claude-code.sha256` | not reused | identical to `main` (`diff` empty) |

## Tool use log

| When | Tool | What | Outcome |
| --- | --- | --- | --- |
| Discarded attempt | `claude-code-guide` subagent (docs lookup, not a worker) | settings keys, agent and skill frontmatter, rule syntax, hook exit codes | applied in the draft and reused here; **unverified** beyond the docs answer |
| PR0b | one-off Node script in the session's scratch folder (not in the repository) | validates the three examples against their schemas (draft-07 subset: type, required, closed objects, enum, `$ref` across files, min/max, pattern), checks every object schema is closed with every key required, recomputes every derived number in the examples, compares the labels with the kickoff block | found AI was wrong #2; after the fix all checks pass. The durable checks are PR0c's `TestExamplesDecodeStrictly` and `contract:check` |
| PR0a | `/privacy-scan`, `/review-step`, `/corpus-check` | the skills are added by PR0a itself and load only in a session started after it is merged | their SKILL.md steps were run by hand; the privacy scan is the same script the skill runs |

## Evidence

### PR0a checks (2026-10-09, run by hand)

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

### PR0b checks (2026-10-09, run by hand)

| Check | Command | Result | Level |
| --- | --- | --- | --- |
| Held-out ids follow the selection rule | `node -e` over the sorted fixture file names (indices 2, 5, …, 29) compared with `heldout.json` | `sessions 34 \| ids match selection rule: true \| count 10 \| categories valid: true \| taxonomy equals contract: true` | command output |
| File name = `sessionId` inside it | `grep -o '"sessionId":"…"'` per file, compared with the file name | no mismatch | command output |
| Examples valid against schemas | scratch `check-contract.mjs` | `schema: all three examples valid; every object closed, every key required` | command output |
| Examples internally consistent | same script | first run `sidechain diagnostic without a sidechain flag`; after the fix `consistency: all derived numbers agree` | command output |
| Labels equal the kickoff block | same script | `labels: identical to the kickoff block (10/10)` | command output |
| Validator catches errors | `error.json` temporarily given an extra key and a bad `code` | `error.error: extra extra`, `error.error.code: pattern`; restored, exit 0 | command output |

### Reviews (`/review-step` steps run by hand; the skills load only in a later session)

| PR | Blocking | Non-blocking | Checks |
| --- | --- | --- | --- |
| PR0a | none | `.claude/hooks/guard-bash.mjs` is not in the kickoff's PR0a list; included because `.claude/settings.json` calls it | 22 paths, all coordinator-owned; `contract/` and `fixtures/` unchanged; suites not applicable (no code); privacy scan 0 hits; document numbers trace to the PR0a checks; SPEC planning numbers labelled unverified |
| PR0b | none | the example consistency check is a one-off scratch script; the durable drift tests (PR0c) check shape, not arithmetic | paths: `contract/**`, `fixtures/labels/heldout.json` and documents only; no fixture data, `backend/` or `frontend/` path; suites not applicable; privacy scan 0 hits; `.handoff/` untracked; `git diff --check` clean |

## Decisions taken in setup

| Decision | Why |
| --- | --- |
| Setup in three PRs: documents and steering (PR0a), contract and labels (PR0b), scaffolds (PR0c) | owner's decision after discarding the single 49-file PR: each step reviewable on its own |
| MIT licence (owner agreed in the discarded attempt) | conventional default for a take-home repository |
| No `Co-Authored-By` trailer or PR footer; authorship documented in NOTES.md and PROMPTS.md | owner's instruction (kickoff and chat) |
| CLAUDE.md and README.md describe the finished layout and commands, without status paragraphs or PR markers; README keeps one "not built yet" line and gets its final text in PR4; CLAUDE.md changes only when a rule or command changes | owner's document rules |
| NOTES.md holds checks, decisions and the AI-was-wrong table; PLAN.md's ledger holds start, ready and minutes per PR | owner's document rules |
| Generated files marked `linguist-generated`, generated line counts stated per PR | reviewers can see how much of a diff is hand-written |
| Bash guard hook parses the hook JSON and tests only `tool_input.command` | the payload always contains a transcript path under the user's Claude directory; grepping stdin would block every command |
| First-prompt extraction for the privacy scan lives in `.claude/skills/privacy-scan/first-prompts.mjs` (Node, repo-local) | the Go parser does not exist in setup; the scan needs the fingerprints at run time |
| Contract: every key required, objects closed, unmeasurable values `null` (carried from the draft) | strict drift tests in both directions; 0 would be a false measurement |
| `unparsed_fraction = (malformed + unknown) / total`; duplicates count as parsed-equivalent for this fraction | duplicates were read correctly, they are just not new |
| `ModelSpend.flags` is a closed enum: `missing_usage`, `unknown_model`, `no_cache_fields`, `long_context_marker`, `sidechain` | the page can explain every flag |
| Contract examples use three fake session ids and small invented numbers | examples must not look like corpus findings |
| Held-out provenance in the labels file, SPEC.md and NOTES.md in identical words | owner's instruction; the agreement number is read against it |
| Pricing source URL `https://docs.claude.com/en/docs/about-claude/pricing` allowed by the scan | **unverified** (web tools are denied by settings); PR1 must confirm the rates it encodes |

## AI was wrong

| # | Claim | Who | How caught | Fix |
| --- | --- | --- | --- | --- |
| 1 | Session logs contain `<system-reminder>` and `command-*` tags and `isMeta` records that must be stripped | earlier planning (AI-assisted) | grep: 0, 0, 0 (Evidence, PR0a) | stripping kept as defensive code with a synthetic test; SPEC does not claim the tags were observed |
| 2 | The contract examples are internally consistent | discarded setup attempt (its checks passed the strict-decode test, which tests shape, not arithmetic) | PR0b consistency script: 1 sidechain record in `diagnostics`, no `sidechain` flag on any model | flag added to the Sonnet row of `contract/examples/summary.json` |

## Another day

Drafted in setup, refined at close: real-session validation; an LLM or embedding classifier
compared against rules-v1 on the same held-out set; an explicit correction signal (OTel
`decision: reject`) for the model-choice analysis; about 30 real sessions per model per category.
