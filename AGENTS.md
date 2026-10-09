# AGENTS.md

Who edits what, and how parallel workers hand work back. Read with CLAUDE.md.

## Roster

| Agent | Kind | Definition |
| --- | --- | --- |
| coordinator | the Claude Code session opened in the repository | this file and CLAUDE.md |
| `go-worker` | subagent | `.claude/agents/go-worker.md` |
| `react-worker` | subagent | `.claude/agents/react-worker.md` |

## Ownership

| Path | Owner |
| --- | --- |
| `backend/**` | go-worker |
| `frontend/**` | react-worker |
| `contract/**`, `fixtures/**`, `evidence/**` | coordinator |
| `README.md`, `SPEC.md`, `PLAN.md`, `NOTES.md`, `PROMPTS.md`, `CLAUDE.md`, `AGENTS.md` | coordinator |
| `.claude/**`, `.cursor/**`, `docker-compose.yml`, git and GitHub | coordinator |

The coordinator writes the `backend/` and `frontend/` scaffolds in PR0c and never edits either
directory after that; from PR1 it integrates, verifies and commits.

## Frozen contract

`contract/*.schema.json` and `contract/examples/*.json` (added in PR0b) are frozen once PR0b is
merged. A worker
that finds a contract problem stops and reports it; it does not work around it. A change after
PR0b is a coordinator-only commit approved by the owner; both workers then restart from that
commit, and both drift tests (`TestExamplesDecodeStrictly` and `TestLiveSummaryMatchesContract`
in Go, `npm run contract:check` and `test/contract.test.ts` in the frontend) must pass.

## Worker protocol

- Edit only your own directory. Never run git or gh.
- Never read outside the repository. Never open `fixtures/labels/heldout.json`.
- Run the commands in your agent definition; do not install packages beyond the lockfile unless
  the task says so.
- Report back: every file touched, every command run with its exact result line, what you did
  not do, and any contract question. No claim beyond what a command showed.

## Cursor

`.cursor/rules/*.mdc` mirror these rules for parity. Cursor was not used to build this repository.
