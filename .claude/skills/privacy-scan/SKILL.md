---
name: privacy-scan
description: Scan the repository for leaked absolute paths, usernames, the owner's email, AI tool log locations, real prompt text and external URLs, and confirm nothing under .handoff/ is tracked. Run before every commit and before asking for PR approval.
allowed-tools: Bash(node .claude/skills/privacy-scan/scan.mjs*) Bash(git ls-files *) Read Grep
---

# privacy-scan

Run from the repository root:

```bash
node .claude/skills/privacy-scan/scan.mjs
```

What it checks (`scan.mjs`), over tracked and untracked-not-ignored files, excluding
`fixtures/`, `node_modules/` and `package-lock.json`:

1. `git ls-files .handoff` is empty.
2. Identity: `C:\Users\`, `/Users/`, `/home/`, the current OS username and `git config user.email`,
   both read at run time so they never appear in the repository.
3. Tool-log locations outside the files that describe the barrier: `.claude/projects`, `.codex`,
   `Cursor/User`, `<system-reminder>`.
4. Prompt text: the first 40 normalised characters of each of the 34 sessions' first genuine
   prompt (computed by `first-prompts.mjs` at run time, never printed) must not appear anywhere.
5. `http(s)://` in `backend/` and `frontend/src/` other than loopback and the pricing source URL.

Output is `file:line  rule`; matched prompt text is never printed. Exit 0 = clean, 1 = hits.

Report the summary line and every hit verbatim. Do not fix anything inside this skill: list the
hits and let the coordinator decide. A clean run is evidence level "unit-tested" only for the
patterns above; say so.
