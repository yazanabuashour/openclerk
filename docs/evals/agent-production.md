# Production Agent Evaluation Protocol

OpenClerk agent evals measure the same production skill a real agent receives.
Do not add hidden evaluator-only instructions to improve a result; if an
instruction is needed, put it in `skills/openclerk` first.

## Active Surfaces

- `production`: the installed runner-first `skills/openclerk` skill using
  `cmd/openclerk-agentops`.
- `sdk-baseline`: an archived SDK-oriented skill retained only as a comparison
  surface.

Generated clients, HTTP server calls, and backend-specific variants are not
active production agent surfaces.

## Harness

Use `scripts/agent-eval/ockp`:

```bash
go run ./scripts/agent-eval/ockp run
go run ./scripts/agent-eval/ockp run --parallel 1
```

`--parallel` defaults to `4`. Each `(variant, scenario)` job gets an isolated
repo copy, data directory, Go cache, and raw event log path under `<run-root>`.
Reports preserve deterministic ordering even when jobs finish out of order.

## Metrics

Reports include correctness status when a verifier is available, tool calls,
assistant calls, wall time, non-cache input tokens, output tokens, direct
generated-file inspection, module-cache inspection, broad repo search, direct
SQLite access, configured parallelism, harness elapsed seconds, and raw log
references with `<run-root>` placeholders.

## Core Scenarios

- Create canonical notes with stable paths, frontmatter, headings, and body
  content, then verify they exist in the vault and document registry.
- Search before creating a synthesis page, then create or update source-linked
  synthesis with citations/source refs to canonical evidence.
- Repeat create requests and ensure duplicate paths fail clearly instead of
  silently overwriting canonical Markdown.
- Append a new section and replace an existing section without losing unrelated
  document content.
- Search for a user concept and verify citation paths and line ranges point back
  to the correct Markdown source.
- Inspect document links and graph neighborhood output for linked notes.
- Create a promoted-record-shaped document and verify records lookup returns the
  expected entity, facts, and citations.
- Inspect provenance events and projection states after document create/update
  workflows.
