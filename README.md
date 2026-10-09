# classify-sessions

Classifies Claude Code session logs by purpose (debug, implement, refactor, test, review,
research, plan, ops) and reports what each purpose costs at public prices, with an honest
account of what could not be parsed or priced. Extra: model-choice analysis. Built on the
official synthetic dataset only.

Not built yet: `backend/`, `frontend/`, `evidence/` and the Docker files; only the contract,
fixtures, documents and steering files exist.

## Run

Go only (Go 1.27.2; `GOTOOLCHAIN=auto` downloads it from `backend/go.mod`; untested fallback if
the download fails: change `go 1.27.2` to `go 1.26` in `go.mod`):

```bash
cd backend
go test ./... -count=1
go run ./cmd/sessions serve --input ../fixtures/claude-code    # then: curl -s http://127.0.0.1:8080/api/summary
```

Go + Node 22:

```bash
cd frontend && npm ci && npm run dev     # page on http://127.0.0.1:5173, proxies /api to :8080
npm run dev:mock                         # page from contract/examples, no backend needed
```

Docker: `docker compose up --build`, then http://127.0.0.1:8088.

Checks: `node .claude/skills/privacy-scan/scan.mjs` (privacy scan);
`cd fixtures && sha256sum -c claude-code.sha256` (fixture manifest).

## Layout

`backend/` Go API and CLI · `frontend/` React page · `contract/` JSON Schemas and examples (the
only boundary) · `fixtures/` synthetic data and held-out labels · `evidence/` regenerated
outputs · `.claude/` agents, skills, settings · `.cursor/rules/` (written, unused).

Documents: [SPEC.md](SPEC.md) (what and why) · [PLAN.md](PLAN.md) (steps and ledger) ·
[NOTES.md](NOTES.md) (evidence, tool use, what went wrong) · [PROMPTS.md](PROMPTS.md) (every
prompt verbatim) · [CLAUDE.md](CLAUDE.md) and [AGENTS.md](AGENTS.md) (AI steering).

Built with AI assistance (Claude Code); how is documented in NOTES.md and PROMPTS.md.
