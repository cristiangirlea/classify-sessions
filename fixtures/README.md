# Fixtures

Everything in this directory is synthetic. It comes from the dataset supplied with the
take-home brief, which states that it contains no real prompts and is safe to commit.

Source archive: session-logs-dataset.zip, as linked in the brief.
Size: 191448 bytes. SHA-256: b36e189ce9468dc6720acddc1b3cdbf7a7fb1b8f32361ffbab7abf9384c22142
(the brief abbreviates it as b36e189c…c22142).

| Path | What it is | Used for |
| --- | --- | --- |
| claude-code/projects/*/*.jsonl | the 34 Claude Code-shaped session files, byte-identical to the archive | the only input the tool ingests |
| claude-code.sha256 | SHA-256 of each of the 34 files, relative to this directory | checked by a test; run `sha256sum -c claude-code.sha256` here |
| otel/claude-code-events.jsonl | the archive's OTel events file | cross-check oracle for pricing only; never ingested |
| labels/heldout.json | ten labels AI-proposed by the planning session (claude-opus-5-5) from the first prompt only, before any rule existed in this repository, and accepted unchanged by the owner; not hand-labelled. Selection: the first ten of every third sorted session id from index 2 | held-out evaluation (added in the scaffold step) |
| synthetic/*.jsonl | hand-made edge cases written for tests | parser tests (added later) |

Do not edit, reformat or clean the files under claude-code/ or otel/. The malformed line,
the duplicate records and the out-of-order timestamps are intentional inputs.
The Codex files in the archive are not included: the tool reads one format.
