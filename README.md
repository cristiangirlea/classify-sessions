# classify-sessions

Classifies Claude Code session logs by purpose (debug, implement, refactor, test, review,
research, plan, ops) and reports what each purpose costs at public prices, with an honest
account of what could not be parsed or priced. Extra: model-choice analysis. Built on the
official synthetic dataset only.

Status: setup, PR0b of three setup PRs. The repository holds the fixtures, the held-out
labels, the frozen JSON contract, the documents and the AI steering files. There is no code
yet: the `backend/` and `frontend/` scaffolds arrive in PR0c, the product in PR1 to PR4.
Every command below is marked with the PR that makes it work.

## Run

Go only (Go 1.27.2; `GOTOOLCHAIN=auto` downloads it from `backend/go.mod`; untested fallback if
the download fails: change `go 1.27.2` to `go 1.26` in `go.mod`):

```bash
cd backend
go test ./... -count=1                                         # PR0c
go run ./cmd/sessions serve --input ../fixtures/claude-code    # PR2, then: curl -s http://127.0.0.1:8080/api/summary
```

Go + Node 22:

```bash
cd frontend && npm ci && npm run dev     # PR2: page on http://127.0.0.1:5173, proxies /api to :8080
npm run dev:mock                         # PR0c: page from contract/examples, no backend needed
```

Docker *(PR4)*: `docker compose up --build`, then http://127.0.0.1:8088.

Works now: `node .claude/skills/privacy-scan/scan.mjs` (privacy scan) and
`cd fixtures && sha256sum -c claude-code.sha256` (fixture manifest).

## Layout

`fixtures/` synthetic data and held-out labels · `contract/` JSON Schemas and examples, the
only boundary · `.claude/` agents, skills, settings · `.cursor/rules/` (written, unused).
Planned: `backend/` Go API and CLI *(PR0c)* · `frontend/` React page *(PR0c)* · `evidence/`
regenerated outputs *(PR1)*.

Documents: [SPEC.md](SPEC.md) (what and why) · [PLAN.md](PLAN.md) (steps and ledger) ·
[NOTES.md](NOTES.md) (evidence, tool use, what went wrong) · [PROMPTS.md](PROMPTS.md) (every
prompt verbatim) · [CLAUDE.md](CLAUDE.md) and [AGENTS.md](AGENTS.md) (AI steering).

Built with AI assistance (Claude Code); how is documented in NOTES.md and PROMPTS.md.
