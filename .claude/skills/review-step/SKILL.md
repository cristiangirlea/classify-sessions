---
name: review-step
description: Review the current PR branch against SPEC.md, the frozen contract, the ownership map and the hard barriers, run both test suites and the privacy scan, and list findings as blocking or non-blocking. Never fixes code. Run before asking the owner to approve a PR.
allowed-tools: Read Grep Glob Bash(git diff *) Bash(git status *) Bash(git log *) Bash(go test *) Bash(go vet *) Bash(npm test *) Bash(npm test) Bash(npm run *) Bash(node .claude/skills/privacy-scan/scan.mjs*)
---

# review-step

Argument: the PR name (PR0a, PR0b, PR0c, PR1 … PR4). Base is `main` unless told otherwise.

1. `git status --short` and `git diff --name-only main...HEAD` plus uncommitted changes.
2. Ownership (AGENTS.md): for each changed path, name its owner. Blocking if a worker-owned
   directory was changed outside that worker's task, or if `backend/` or `frontend/` was
   changed by the coordinator after PR0c.
3. Frozen inputs: blocking if `contract/**` changed in any PR other than PR0b without an
   owner-approved note in NOTES.md; blocking if `fixtures/claude-code/**`, `fixtures/otel/**`
   or `fixtures/claude-code.sha256` changed at all, or `fixtures/labels/heldout.json` changed
   after PR0b.
4. Decisions: for each SPEC.md decision touched by the diff, a matching NOTES.md entry exists.
5. Tests named for this PR in PLAN.md exist (`grep -rn "func TestName" backend/`); list missing ones.
6. Run locally `cd backend && go vet ./... && go test ./... -count=1` and, from `frontend/`,
   `npm run contract:check && npm run typecheck && npm test && npm run build`. Quote result lines.
   Before PR0c the directories do not exist; say "not applicable" for this step.
7. Run `node .claude/skills/privacy-scan/scan.mjs`; any hit is blocking.
8. PROMPTS.md has a section for this PR; no document, commit message or PR description records a time.
   Commits carry no `Co-Authored-By` trailer (`git log main..HEAD --format=%B`).
9. Documents: any number in README/SPEC/NOTES not traceable to `evidence/` or a quoted command
   output is non-blocking unless it is a headline number, then blocking.

Output:

```markdown
### Review PRn
Blocking: <numbered list or "none">
Non-blocking: <numbered list or "none">
Commands: <each command and its result line>
```

Do not edit any file. Do not commit.
