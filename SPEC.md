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

1. **Every nonblank line is one JSON record, and the corpus has a few that are not clean.**
   34 files hold 1107 nonblank lines: 1084 parse, 1 is malformed, none has an unknown type or
   content block, and 22 repeat an earlier record with the same session and uuid and identical
   content. Confirmed by `TestCorpusAccounting`; the duplicate and conflict rules by
   `TestDuplicateVsConflictingUUID` on `fixtures/synthetic/`.
2. **Summary records have no identity.** They carry no session id, uuid or timestamp (12 in the
   corpus), so they belong to the session of the file they are in and sort last. Confirmed by
   `TestSummaryRecordsAttachToFileSession` and the `missing_timestamp` count of 12 in
   `TestCorpusAccounting`.
3. **Usage comes in three shapes.** Most assistant records carry input, output, cache-read and
   cache-write tokens; 59 carry only input and output; 10 carry no usage at all. Confirmed by
   `TestCorpusAccounting` (diagnostics) and `TestUsageBilledOncePerRecord`.
4. **File order is not time order.** In session `3888cf6f…` the first user record in the file is
   not the earliest one, so the first prompt must be chosen by timestamp. Confirmed by
   `TestFirstPromptUsesTimestampOrder`.
5. **Tool use spans two records.** A call is a `tool_use` block in an assistant record; its
   result is a `tool_result` block, with `is_error`, in the next user record. A user record that
   carries a tool result is not a prompt. Confirmed by `TestToolResultsAreNotPrompts`.
6. **User prompts are either a string or a list of text blocks.** Both forms are prompts.
   Confirmed by `TestStringAndBlockContentArePrompts`.
7. **Model ids are logged with a date suffix and sometimes a `[1m]` marker.** Five ids appear;
   two sidechain Haiku records have no request id and are absent from the OTel export.
   Confirmed by `TestOTelCrossCheck`.
8. **No harness tags appear in the corpus.** Zero lines contain `<system-reminder>`,
   `<command-` or `"isMeta"` (`evidence/tags.txt`). Stripping `<system-reminder>` from prompts is
   defensive only, tested on a synthetic record by
   `TestSystemReminderStrippedFromSyntheticPrompt`.
9. **All 22 duplicates are one replayed block.** Session `b7dab20e…` repeats its first 22 lines
   byte for byte at lines 45 to 66, so its timestamps jump back at line 45. Once the copies are
   counted as duplicates, its remaining records are in time order and its first prompt is line
   1 under either order. Confirmed by `TestCorpusAccounting` (all 22 duplicates in this session)
   and `TestFirstPromptUsesTimestampOrder` (subtest for this session).
10. **Sessions almost never switch models.** Every session has at least one assistant record, and
    each is dominated by one model family: Opus 4.1 in 12, Sonnet 4.5 in 18 (4 of them with
    `[1m]` requests), Sonnet 5 in 4. Only 2 sessions use a second model at all. Of the
    genuine user messages, 27 read as corrections, in 8 sessions. Confirmed by
    `TestCompareModelsSuppliedCorpus`.

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
3. **A repeated record is a duplicate only if its content is identical; otherwise it is
   malformed.** Two records with the same session and uuid and the same canonical JSON (key
   order ignored) count once as parsed and once as duplicate, and are billed once. Same uuid
   with different content is a conflict and counts as malformed. Why: a byte-for-byte rule
   would call a re-serialised copy a conflict, and silently keeping either version of a
   conflict would pick one at random. Enforced by `TestDuplicateVsConflictingUUID` and
   `TestUsageBilledOncePerRecord`.
4. **Known cost is a lower bound, and a session says when it is one.** A session's cost is
   incomplete when any of its requests has no usage, an unknown model, no cache fields or a
   `[1m]` marker, or when a malformed or unknown record belongs to it. The known part is still
   reported; unknown cost is never counted as zero. Sidechain requests are billed and flagged
   but do not make a session incomplete. Why: the lead needs the known total and to see how much
   of it is partial. Enforced by `TestUsageBilledOncePerRecord` and
   `TestPriceTableComponentsAndAliases`; on the corpus, 17 of 34 sessions are incomplete.
5. **`[1m]` requests are priced at base rates and flagged; the OTel generator priced them
   higher.** The `[1m]` marker names a long-context mode whose surcharge applies only above a
   context threshold; this repository prices them at the base rate and flags them. The OTel
   export priced them at 6/22.50, which accounts for USD 1.5357 of the difference between the
   two totals. Whether any request exceeds the threshold is not tested here. Enforced by
   `TestOTelCrossCheck`.
6. **Sonnet 5 is priced at 2/10 USD per million input/output tokens.** The OTel generator used
   3/15, which accounts for USD 0.4462 of the difference. The rates in the dated price table
   were not checked against the live pricing page from this repository (web access is denied
   by the settings). Enforced by `TestOTelCrossCheck` and `TestPriceTableComponentsAndAliases`.
7. **A session's purpose is the first rule that matches its first genuine prompt (rules-v1).**
   The prompt is lowercased and tried against 24 ordered regular-expression rules; the first
   match gives the category and a rule id, and no match gives `unknown` with no rule. The
   categories, as defined for this repository: *debug*, something is broken or behaves
   unexpectedly; *implement*, build new functionality; *refactor*, restructure code without
   changing behaviour; *test*, write, extend or repair tests; *review*, assess code or a change
   for problems or readiness; *research*, understand how something works or compare options with
   no change asked for; *plan*, design work before doing it; *ops*, build, deploy, CI,
   infrastructure, dependencies or release. Where a prompt fits two, the order decides: plan
   rules come first (so plan beats implement), then review ("look for anything that would
   break", "ready to ship"); "why does …" and pasted error text are debug; "how does X work or handle …" is
   research; CI work is ops even when it names tests; a test noun with a verb is test, but a
   path under a test directory is not; fix verbs and symptom words are debug; implement is
   last. Why: a rule list is deterministic, explains every decision by its rule id, and needs no
   network or model. The rules were written from the 24 sessions outside the held-out set only.
   Enforced by `TestClassifyRulesTableDriven` (every rule and every precedence case, on invented
   prompts).
8. **The held-out evaluation ran once, after the rules were frozen, and agrees 10 of 10.** The
   labels file is read only when a command is given `--labels` explicitly; the rules were tuned
   without it (`prompts --exclude-labels`) and their file's SHA-256 was recorded before the one
   evaluation run (NOTES.md). Result: 10 labelled, 10 agreements, no disagreement
   (`evidence/evaluation.json`). What it does not show: the labels were proposed by a Claude
   model and the rules written by one, the prompts are synthetic templates, and ten sessions
   cannot reveal where the rules fail. Where the rules are weakest, from the dev set: prompts
   that sit between two categories (CI work that runs tests, a feature placed in a test file, a
   regression test for a fixed bug, a readiness check on a test file); every one is decided by
   rule order, not by understanding. `unknown` is never reached on this corpus, so real prompts
   in other wording would likely fall to `unknown` or to the first keyword they happen to
   contain. Enforced by `TestEvaluateLoadsLabelsAndReportsDisagreements`.
9. **The API answers only loopback requests for its own origin.** It listens on 127.0.0.1
   unless `--bind-any` is given; the `Host` header must be a loopback name, an `Origin` header
   must equal the request's own origin, and `Sec-Fetch-Site: cross-site` is refused, each with
   403. Only GET and HEAD are served (405 with `Allow: GET, HEAD` otherwise), unknown paths are
   404, every error body follows `contract/error.schema.json`, and every response carries
   `Cache-Control: no-store`, `nosniff`, `no-referrer`, `default-src 'none'` and a same-origin
   resource policy. Why: the summary describes the developer's own work, so another website
   open in the same browser must not be able to read it, and a DNS-rebinding name must not pass
   as local. Enforced by `TestHTTPBoundary`.
10. **The dev page's proxy target can move, but only to another loopback port.**
    `API_PROXY_TARGET` overrides the default `http://127.0.0.1:8080`; any value other than
    `http://127.0.0.1:<port>` or `http://localhost:<port>` stops the dev server at startup. Why:
    port 8080 may be taken on a developer's machine, and nothing may leave the machine. Enforced
    by `test/proxyTarget.test.ts`.
11. **The model-choice analysis compares models within a category, never across, and says when
    it cannot judge.** Each session is attributed whole to its dominant model family (most
    assistant records; `[1m]` folded into its family and counted). Per category and model it
    reports sessions, mean and median known cost, correction messages over genuine user
    messages (keyword rules on messages after the first), and tool failures over known results.
    A category is comparable when two models have a session in it, and supported when each has at
    least three; only supported categories may be pooled into advice, and fewer than 10 pooled
    sessions per model gives `insufficient_evidence`. Why: the corpus has 34 sessions, so a
    comparison that did not count its own support would read noise as a finding. On this corpus
    (`evidence/summary.json`): six categories are comparable (debug, implement, refactor, test,
    review, research), none is supported, plan and ops have Sonnet 4.5 sessions only; over all
    categories Opus 4.1 has 10 corrections in 77 user messages and 22 tool failures in 153
    results, Sonnet 4.5 has 13 in 97 and 30 in 230. The advice, generated by the code and quoted
    verbatim:

    > claude-opus-4-1 has the highest mean known cost per session, USD 2.386 over 12 sessions,
    > against USD 0.471 over 18 sessions for claude-sonnet-4-5, a ratio of 5.1x. They share 5
    > comparable categories (debug, implement, test, review, research), of which 0 have at least
    > three sessions per model. No shared category is supported, so no correction rates are
    > pooled. With fewer than 10 pooled sessions for at least one model, the logs cannot show
    > whether the premium buys fewer corrections.

    What to tell the platform lead: the 5.1x premium for Opus 4.1 is a price difference, and
    these logs cannot show that it buys fewer corrections or that it does not. Capturing an
    explicit correction signal and about 30 sessions per model per category would let the same
    code answer it. Enforced by `TestCompareModelsSuppliedCorpus`,
    `TestAdviceDeterministicAndVerdictThresholds` and
    `TestSingleModelCategoryReportedNotCompared`.
12. **Docker serves page and API on 127.0.0.1:8088 only, and the API has no route out.** The
    page container (nginx) is the only published port, on loopback; the API container sits on an
    internal network with no outbound route, reads the fixtures through a read-only mount, and
    runs on a read-only, non-root, shell-less image. nginx refuses a foreign `Host` or `Origin`
    itself before it rewrites `Host` for the API, so the API's loopback check is not bypassed by
    the proxy. Why: the container must keep the same boundary as the local run. Checked by
    `docker compose up` and curl (NOTES.md, docker-checked).

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
