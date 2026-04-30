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

The populated-vault e2e lane is defined in
[`populated-vault-agentops-e2e.md`](populated-vault-agentops-e2e.md). It is an
implemented targeted eval lane for heterogeneous synthetic vault pressure, not
part of the current release-blocking production gate.

The repo-docs dogfood lane is defined in
[`repo-docs-dogfood.md`](repo-docs-dogfood.md). It imports this repository's
committed public markdown into an isolated eval vault and provides recurring
non-release-blocking pressure on retrieval, synthesis maintenance, and
decision-record explainability without private vault evidence.

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

The harness creates an isolated Codex home for each job and seeds it with the
user's `auth.json` only. Single-turn scenarios use
`codex exec --ephemeral --ignore-user-config`. Multi-turn scenarios use one
persisted eval session per variant/scenario inside the isolated eval Codex home:
the first turn creates a session in the throwaway copied repo, and later turns
use `codex exec resume --ignore-user-config` with explicit writable roots for
the scenario run directory and shared Go cache. Per-turn raw logs live under
`<run-root>/<variant>/<scenario>/turn-N/`.

Each job gets an isolated copied repo and OpenClerk storage rooted inside that
copy:

- `OPENCLERK_DATABASE_PATH=<run-root>/<variant>/<scenario>/repo/.openclerk-eval/openclerk.db`

The scenario setup initializes that database with the copied repo's
`.openclerk-eval/vault` as the configured vault root.

The copied repo omits root `AGENTS.md`, stale `.agents` contents, VCS metadata,
Beads/Dolt metadata, eval artifacts, and the eval harness itself before
installing the shipped `skills/openclerk` skill into
`.agents/skills/openclerk/`. The harness does not generate evaluator-only
OpenClerk `AGENTS.md` instructions. Before each job runs, it preflights the
rendered Codex context with `codex debug prompt-input` to verify that the
project skill points at `.agents/skills/openclerk/SKILL.md` and that no
OpenClerk product instructions leak through an `AGENTS.md` block. The preflight
uses the same isolated `CODEX_HOME`; `codex debug prompt-input` does not expose
an `--ignore-user-config` flag in the current CLI. Raw event logs are not
committed; reduced reports refer to them with `<run-root>` placeholders.

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

## Targeted Report Template

Targeted ADR, POC, eval, promotion, and deferred-capability reports should
separate correctness from taste review. The reduced report or companion
decision note should record three lenses:

- **Safety pass:** whether the workflow preserved authority, citations,
  provenance, freshness, local-first behavior, duplicate handling,
  runner-only access, and approval-before-write.
- **Capability pass:** whether current `openclerk document` and
  `openclerk retrieval` primitives can technically express the workflow
  without a new runner action, schema, storage behavior, transport, or public
  API.
- **UX quality:** whether the workflow is acceptable for routine use, or
  whether it is taste debt because it completed only through high step count,
  long latency, exact prompt choreography, repeated assistant turns, or
  surprising clarification.

A scenario can pass safety and capability while still recording UX quality as
taste debt. That outcome supports follow-up audit, design, or eval work, but it
does not authorize implementation without targeted evidence and an explicit
promotion decision naming the exact surface and gates.

Companion eval-design follow-ups for completed-but-high-touch workflows are
recorded in
[`high-touch-successful-workflows-ceremony-eval-design.md`](high-touch-successful-workflows-ceremony-eval-design.md).
They are future targeted pressure designs, not production-gate scenarios.

Committed reports must continue to use repo-relative paths and neutral
placeholders such as `<run-root>` for raw logs. This template guidance does not
require a generated report JSON field, harness schema change, runner action, or
storage migration.

## Scenario Coverage

The `ockp` harness covers routine local knowledge-plane workflows:

- canonical note creation with stable paths, frontmatter, headings, body content,
  and vault/document-registry verification
- convention-first layout inspection with `inspect_layout`, including resolved
  configuration mode, no committed manifest requirement, conventional prefixes,
  first-class document kinds, and runner-visible invalid layout checks
- canonical directory and link navigation with path-prefix listing, exact
  document heading retrieval, outgoing links, incoming backlinks, graph
  neighborhood expansion, and graph projection freshness
- RAG-style retrieval-only answering with unfiltered search, path-prefix
  filtering, metadata filtering, repeated-query stability, chunk citations, and
  no implicit synthesis filing
- source-grounded search followed by source-linked synthesis creation or update
  under `synthesis/` with `type: synthesis`, `status: active`,
  `freshness: fresh`, `source_refs`, `## Sources`, and `## Freshness`
- durable answer filing into source-linked markdown
- contradiction or stale synthesis repair against newer canonical sources
- synthesis compiler pressure checks for candidate selection with decoys,
  multi-source source-ref preservation, resumed drift repair, projection
  freshness inspection, and duplicate prevention
- append and replace-section workflows that preserve unrelated document content
- promoted-record-shaped document creation, records lookup, provenance events,
  projection states, and freshness-aware synthesis when records are summarized
- service registry lookup comparison against plain docs retrieval for
  service-centric tasks
- decision record lookup comparison, supersession freshness, and migrated ADR
  markdown projection with citations, provenance, and projection freshness
- duplicate canonical path rejection without overwrite
- mixed document/retrieval workflows that require both runner domains
- one no-tools clarification response for missing required fields, plus
  final-answer-only direct rejections for invalid limits, unsupported
  lower-level routine workflows, and bypass attempts through legacy
  source-built command paths or alternate transports
- true multi-turn workflows that require resumed context across ordered turns

## Production Gate

Production OpenClerk AgentOps is release-ready only when:

- production passes every selected scenario
- production has no stale surface inspection, module-cache inspection,
  broad repo search, direct SQLite access, or legacy source-built runner usage
- rule-covered validation scenarios use no tools, no command executions, and at
  most one assistant answer; missing-field scenarios clarify while invalid
  limits and bypass requests reject
- the eval context preflight confirms the model-visible agent context is the
  shipped skill and runner, not hidden evaluator-only instructions

Current reduced eval reports are written under `docs/evals/results/`.
