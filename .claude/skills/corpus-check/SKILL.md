---
name: corpus-check
description: Re-run every corpus number into evidence/ (manifest, Go tests, summary, held-out evaluation, OTel cross-check) and print a NOTES.md block. Run after any parser, rule or price change and before quoting any number in a document.
allowed-tools: Bash(go *) Bash(sha256sum *) Bash(git status *) Bash(ls *) Bash(cat *) Bash(grep *) Bash(mkdir -p evidence) Read
---

# corpus-check

You regenerate evidence; you never edit classifier rules, prices, labels, fixtures or the
contract. If a step fails, stop and report the failure verbatim.

Steps, from the repository root (skip a step whose subcommand does not exist yet and say
"not available in this PR"):

1. Manifest: `cd fixtures && sha256sum -c claude-code.sha256 | grep -vc ': OK$'` must print 0.
   Raw counts, independent of the Go parser: `ls fixtures/claude-code/projects/*/*.jsonl | wc -l`
   and `cat fixtures/claude-code/projects/*/*.jsonl | grep -c .` (nonblank lines).
   Write these three result lines to `evidence/manifest.txt` (create `evidence/` if missing).
   Tag check (the "assumption caught by grep" in NOTES.md): count lines containing
   `<system-reminder>`, `<command-` and `"isMeta"` across the same files with `grep -c` and
   write the three counts to `evidence/tags.txt`.
2. `cd backend && go vet ./... && go test ./... -count=1`; keep the `ok`/`FAIL` lines.
3. `go run ./cmd/sessions summary --input ../fixtures/claude-code --labels ../fixtures/labels/heldout.json > ../evidence/summary.json`
   (the held-out result in it is recomputed from the frozen rules; it is not a tuning signal)
4. `go run ./cmd/sessions evaluate --input ../fixtures/claude-code --labels ../fixtures/labels/heldout.json > ../evidence/evaluation.json`
5. If `fixtures/otel/claude-code-events.jsonl` exists:
   `go run ./cmd/sessions crosscheck --input ../fixtures/claude-code --otel ../fixtures/otel/claude-code-events.jsonl > ../evidence/crosscheck.txt`
6. Read the files you wrote and print:
   - accounting: `total = parsed + malformed + unknown + duplicate` with the numbers
   - known spend USD and incomplete-session count
   - held-out agreement `agreements/labelled` and each disagreement as `session id: labelled → predicted (rule)`
   - cross-check headline lines

Then print this block for the coordinator to paste into NOTES.md, filling every value from
the files, never from memory:

```markdown
### Corpus check (PRn)

- Manifest: 34/34 OK
- Go tests: <ok/FAIL lines>
- Records: <total> = <parsed> parsed + <malformed> malformed + <unknown> unknown + <duplicate> duplicate
- Known spend: USD <total_usd> (<incomplete_sessions> sessions incomplete)
- Held-out: <agreements>/<labelled> (<fraction>); disagreements: <list or none>
- Cross-check: <headline or "not available in this PR">
- Evidence level: unit-tested (Go tests) + command output in evidence/
```

Never paste prompt text into the block. Session ids and rule ids only.
