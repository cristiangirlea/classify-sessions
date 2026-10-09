# SPEC.md

## The problem as understood

A platform lead wants to know what developers use Claude Code for and what each kind of use
costs. The input is a folder of Claude Code session logs: JSON Lines files, one file per
session, one record per line. The answer has three parts.

- **Purpose.** Each session gets exactly one category from a fixed taxonomy: `debug`,
  `implement`, `refactor`, `test`, `review`, `research`, `plan`, `ops`, or `unknown` when
  nothing matches. The category is decided from the session's first genuine prompt, the first
  thing the developer asked for, chosen in timestamp order rather than file order. How well
  the categories match a human-style reading is measured on a held-out set that no rule was
  written from.
- **Cost.** Each session's requests are priced at the public per-token rates for the model
  that served them, including cache reads and cache writes, and the totals are reported per
  category and per model. Where usage is missing or a model is unknown, the cost is reported
  as incomplete and the known part as a lower bound, never as zero.
- **Coverage.** The lead must be able to see how much of the input the answer rests on: every
  nonblank line is counted once as parsed, malformed, unknown or duplicate, and the report
  states the unparsed fraction and the share of sessions and records left unclassified.

The answer is served as JSON by a local API and shown on a single local page. One extra
analysis is in scope: whether the most expensive model buys anything the cheaper models do
not, measured per category by cost per session, correction messages and tool failures. It is
descriptive and makes no causal claim.

## Format findings

Each entry is a fact about the session-log format, with the test that confirms it.

## Decisions on unsettled points

Each entry is a decision the brief leaves open, why it was taken, and what enforces it.

1. **The backend and the page share only `contract/`.** The Go API and the React page never
   import each other; the JSON Schemas and hand-written examples in `contract/` are the only
   boundary, every object is closed and every key required, and a value that cannot be
   measured is `null` rather than 0. Why: the page can be built and tested against the
   examples before the backend exists, and 0 would claim a measurement that was not made.
   Enforced by `TestExamplesDecodeStrictly` (each example decodes into the Go structs with
   unknown fields disallowed and round-trips to the same key set) and `npm run
   contract:check` (the committed TypeScript types match the schemas).
2. **The classifier is judged on ten held-out sessions fixed before any rule existed.** The
   labels are in `fixtures/labels/heldout.json`. Provenance, in the same words as the file
   and NOTES.md:

   > AI-proposed by the planning session (claude-opus-5-5) from the first prompt only, before
   > any rule existed in this repository; the owner accepted all ten unchanged. Not
   > hand-labelled. Selection: the first ten of every third sorted session id from index 2.

   Why: rules written from the other 24 sessions cannot be tuned to the evaluation set. The
   agreement figure measures agreement with an AI reading of the same first prompt, not with
   a human judgement. Checked by a command that recomputes the selection rule from the
   fixture file names (NOTES.md).

## Out of scope

- Real session logs from any person or machine, including the owner's. Only the official
  synthetic dataset is read.
- Classifying with a language model or embeddings, more than one category per session, and
  classifying individual messages rather than sessions.
- Ingesting the OpenTelemetry export. The OTel file in `fixtures/otel/` is a cross-check for
  pricing only.
- Pricing tiers the logs do not identify: batch and priority pricing, regional pricing,
  one-hour cache writes, and long-context surcharges for requests that do not exceed the
  threshold.
- Any causal claim about models: the model-choice analysis describes what the logs show and
  says when they are too thin to support advice.
- Accounts, authentication, multiple users, persistence, network deployment, and any
  listener other than loopback.
- Live updates: the summary is computed once from a folder when the server starts.
- Building with Cursor. The Cursor rules exist for parity; Cursor was not used.

## How customer and employer data were kept out

**Data choice.** The only input is the official synthetic Claude Code dataset, copied into
`fixtures/claude-code/` and pinned by `fixtures/claude-code.sha256` (34 files); the OTel file is
from the same synthetic archive. No customer data, employer code, employer data or real
developer session was used to build, tune or test anything here.

**Keeping real logs out of reach.** `.claude/settings.json` denies reads of the AI tools' local
data folders (`~/.claude`, `~/.codex`, `~/.cursor`, Cursor's application data, `~/.hrn`) and
of `.env` files, blocks reads outside the working directories, and denies edits to the
fixture data and its manifest. A PreToolUse hook (`.claude/hooks/guard-bash.mjs`) rejects any
shell command that names those folders or the earlier attempt's folder. The `go-worker` and
`react-worker` definitions forbid reading outside the repository, and the CLI has no default
input directory: `--input` must be given.

**Keeping identifying text out of outputs.** The contract has no field for prompt text,
absolute paths or usernames; sessions and classifier rules are identified by id only, and
`source.input_hint` is the relative directory given on the command line. The server listens
on loopback, and the page loads no font, script or asset from outside `127.0.0.1`.

**Checking it.** `node .claude/skills/privacy-scan/scan.mjs` runs before every pull request. It
scans every tracked and untracked-not-ignored file outside `fixtures/` for Windows, macOS and
Linux user paths, the current OS username and git email (read at run time so they never enter
the repository), AI tool log locations, external URLs in `backend/` and `frontend/src/`, and
the first 40 characters of every session's first prompt (computed at run time, never
printed). It exits non-zero on any hit, and confirms that nothing under `.handoff/` is tracked.
