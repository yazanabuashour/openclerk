# Production Agent Evaluation Protocol

OpenClerk agent evals measure the same AgentOps surface a real agent receives:
the production skill plus the installed `openclerk` runner.
Do not add hidden evaluator-only instructions to improve a result; if an
instruction is needed, put it in `skills/openclerk` first.

## Active Surfaces

- `production`: the installed runner-first `skills/openclerk` skill using the
  `openclerk` binary.

OpenClerk runner is the production semantic contract for routine agent work. The
machine-facing runner is the supported transport for that contract today.

HTTP server calls, direct SQLite access, ad hoc runtime programs, repo-wide
spelunking, module-cache inspection, stale API paths, and
backend-specific variants are not active production agent surfaces.

## Evaluation Purpose

The eval harness validates the AgentOps contract and the knowledge-model
behaviors implemented behind it. It checks that routine tasks use runner JSON
requests, that bypass attempts are rejected before tools, and that synthesis,
records, provenance, and freshness behavior remain reliable.

## Harness

Use `scripts/agent-eval/ockp`:

```bash
go run ./scripts/agent-eval/ockp run
go run ./scripts/agent-eval/ockp run --parallel 1
go run ./scripts/agent-eval/ockp run --cache-mode isolated
go run ./scripts/agent-eval/ockp run --report-name ockp-final
```

`--parallel` defaults to `4`. The harness runs independent
`(variant, scenario)` jobs with deterministic report ordering even when jobs
finish out of order.

Single-turn scenarios use `codex exec --ephemeral`. Multi-turn scenarios use one
persisted eval session per variant/scenario: the first turn creates a session in
the throwaway copied repo, and later turns use `codex exec resume` with explicit
writable roots for the scenario run directory and shared Go cache. Per-turn raw
logs live under `<run-root>/<variant>/<scenario>/turn-N/`.

Each job gets an isolated copied repo and OpenClerk storage rooted inside that
copy:

- `OPENCLERK_DATA_DIR=<run-root>/<variant>/<scenario>/repo/.openclerk-eval/data`
- `OPENCLERK_DATABASE_PATH=<run-root>/<variant>/<scenario>/repo/.openclerk-eval/openclerk.db`
- `OPENCLERK_VAULT_ROOT=<run-root>/<variant>/<scenario>/repo/.openclerk-eval/vault`

The copied repo omits root `AGENTS.md`, VCS metadata, Beads metadata, eval
artifacts, and the eval harness itself before installing the selected variant
instructions. Raw event logs are not committed; reduced reports refer to them
with `<run-root>` placeholders.

The harness defaults to `--cache-mode shared`, which prewarms one shared Go
module/build cache under `<run-root>/shared-cache` while keeping OpenClerk
databases, vaults, temporary directories, copied repos, and raw logs isolated
per job. Use `--cache-mode isolated` for apples-to-apples comparison with older
per-job-cache reports.

## Metrics

Reports include:

- database/vault verification and assistant-answer verification
- configured harness parallelism and elapsed harness wall time
- cache mode, cache prewarm time, effective parallel speedup, and parallel
  efficiency
- per-job phase timing totals for setup, repo copy, variant install, cache warm,
  seed data, agent run, metrics parsing, and verification
- per-turn metrics and raw log references for multi-turn scenarios
- tool calls, command executions, assistant calls, wall time, non-cached input
  tokens, cached input tokens, input tokens, and output tokens
- stale surface inspection, module-cache inspection, broad repo search,
  direct SQLite access, and legacy source-built runner usage

Legacy source-built runner usage is counted only for executed source-tree command paths, not installed OpenClerk runner calls or documentation text containing command strings.

## Scenario Coverage

The `ockp` harness covers routine local knowledge-plane workflows:

- canonical note creation with stable paths, frontmatter, headings, body content,
  and vault/document-registry verification
- source-grounded search followed by source-linked synthesis creation or update
  under `notes/synthesis/` with `type: synthesis`, `status: active`,
  `freshness: fresh`, `source_refs`, `## Sources`, and `## Freshness`
- durable answer filing into source-linked markdown
- contradiction or stale synthesis repair against newer canonical sources
- append and replace-section workflows that preserve unrelated document content
- promoted-record-shaped document creation, records lookup, provenance events,
  projection states, and freshness-aware synthesis when records are summarized
- service registry lookup comparison against plain docs retrieval for
  service-centric tasks
- duplicate canonical path rejection without overwrite
- mixed document/retrieval workflows that require both runner domains
- final-answer-only direct rejections for missing required fields, invalid
  limits, unsupported lower-level routine workflows, and bypass attempts through
  legacy source-built command paths or alternate transports
- true multi-turn workflows that require resumed context across ordered turns

## Production Gate

Production OpenClerk AgentOps is release-ready only when:

- production passes every selected scenario
- production has no stale surface inspection, module-cache inspection,
  broad repo search, direct SQLite access, or legacy source-built runner usage
- rule-covered validation scenarios are final-answer-only: no tools, no command
  executions, and at most one assistant answer
- the eval context preflight confirms the model-visible agent context is the
  shipped skill and runner, not hidden evaluator-only instructions

Current reduced eval reports are written under `docs/evals/results/`.
