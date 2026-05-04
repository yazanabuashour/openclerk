# OpenClerk Next-Phase Maturity Evidence Inventory

## Status

Inventory for `oc-oa53`.

This note records the evidence baseline before representative real-vault and
scale-ladder validation. It is not a promotion decision, release gate, runner
API, storage change, or retrieval/indexing decision.

## Current Evidence

| Evidence | Current result | What it proves | Current limit |
| --- | --- | --- | --- |
| [`docs/evals/results/ockp-agentops-production.md`](../evals/results/ockp-agentops-production.md) | Release-blocking production gate passed: 30/30 scenarios, no direct SQLite, no broad repo search, no source-built runner, no module-cache inspection, validation rows final-answer-only. | The installed `openclerk document` and `openclerk retrieval` AgentOps path can cover the current v1 production scenario set safely. | Scenarios are targeted fixtures, not a representative private vault or scale test. |
| [`docs/evals/results/ockp-repo-docs-dogfood.md`](../evals/results/ockp-repo-docs-dogfood.md) | 7/7 repo-docs rows completed with safety pass, capability pass, and completed UX quality. | OpenClerk can import committed public repo markdown into an isolated eval vault and answer retrieval, synthesis, decision-record, release-readiness, tag, memory-router-report, and freshness questions through runner JSON. | The seed is public repo markdown only. Before this `oc-oa53` implementation, the eligible public markdown seed was about 145 files and 825 KB, excluding `docs/evals/results/` and `AGENTS.md`. |
| [`docs/evals/populated-vault-agentops-e2e.md`](../evals/populated-vault-agentops-e2e.md) and populated-vault reports | Synthetic populated-vault pressure covers mixed document families, stale sources, duplicate-looking docs, polluted/decoy evidence, and synthesis update pressure. | Current retrieval primitives plus compact authority policy can handle messy but synthetic corpus pressure without a new populated-vault-specific runner action. | It is still synthetic and small; it does not answer larger corpus behavior or real private workflow ergonomics. |
| [`docs/evals/results/ockp-real-vault-dogfood.md`](../evals/results/ockp-real-vault-dogfood.md) | Initial `oc-oa53` reduced real-vault timing report completed with reduced-report safety checks passing. | The maturity harness can report local vault counts and timings without emitting private paths, private queries, titles, snippets, document ids, chunk ids, raw roots, or raw logs. | It is maintainer-harness timing evidence, not Codex routine-agent command telemetry. |
| [`docs/evals/results/ockp-scale-ladder-10mb.md`](../evals/results/ockp-scale-ladder-10mb.md) | Initial `oc-oa53` 10 MB scale-ladder report completed with reduced-report safety checks passing. | The new maintainer harness can generate a deterministic synthetic corpus, sync it through the embedded OpenClerk runtime, and record reduced counts/timings without committing generated corpus content or machine-local paths. | 100 MB did not complete in-session and is tracked by `oc-oa53.12`; 1 GB remains unjustified. |
| [`docs/evals/results/ockp-scale-ladder-100mb-timeout.md`](../evals/results/ockp-scale-ladder-100mb-timeout.md) | Initial `oc-oa53` 100 MB attempts stalled before a completed reduced runtime report. | The scale ladder exposed a performance or harness cliff that must be investigated before 1 GB or release-gate promotion. | It is timeout/stall evidence only, not a successful 100 MB report. |
| [`CONTEXT.md`](../../CONTEXT.md) | Records the current domain vocabulary: AgentOps, runner, vault, canonical docs, source docs, synthesis docs, provenance, projection state, promoted records, and decision records. | Provides the local vocabulary and boundaries that maturity reports must use. | It is architecture context, not eval evidence. |

## Validated V1 Behavior

- Agent-facing work is the installed runner plus `skills/openclerk/SKILL.md`.
- Routine operations stay inside `openclerk document` and `openclerk retrieval`.
- Canonical markdown remains the source of truth; synthesis, graph, records,
  services, decisions, provenance, and projection state are derived evidence.
- Source-linked synthesis maintenance is validated for current scenarios,
  including source refs, freshness, duplicate avoidance, and provenance.
- Promoted workflow actions exist only where targeted evidence justified them,
  such as `compile_synthesis`, `source_audit_report`,
  `evidence_bundle_report`, `duplicate_candidate_report`,
  `memory_router_recall_report`, and `ingest_source_url` plan/create/update
  behavior.
- Reports use reduced artifacts with `<run-root>` placeholders and do not
  commit raw logs.

## Deferred Future Vision

- Representative private/real-vault dogfood is not yet proven.
- Larger corpus behavior at 10 MB, 100 MB, and 1 GB is not yet proven.
- Hybrid/vector retrieval is not justified by the current evidence. It remains
  a candidate only if real-vault or scale-ladder reports show lexical FTS
  relevance, latency, or workflow failures that tuning cannot address.
- LLM-wiki-like workflows remain mapped to OpenClerk source-linked synthesis,
  source intake, index/log equivalents, lint or health checks, and filed
  answers. New surfaces require candidate comparison and promotion evidence.

## Next Evidence Lanes

- Real-vault dogfood: run the reduced private-vault report described in
  [`docs/evals/real-vault-dogfood.md`](../evals/real-vault-dogfood.md), then
  record safety, capability, UX quality, performance, and evidence posture
  separately.
- Scale ladder: run the deterministic synthetic ladder described in
  [`docs/evals/scale-ladder-validation.md`](../evals/scale-ladder-validation.md)
  for 10 MB and 100 MB first. Run 1 GB only after smaller tiers show it will
  produce meaningful evidence.
- Retrieval/indexing decision: decide from those reports whether to keep
  lexical SQLite FTS, tune current indexes, create hybrid/vector
  candidate-comparison Beads, defer, or kill the scale track.

The first pass decision is recorded in
[`openclerk-next-phase-maturity-validation-decision.md`](openclerk-next-phase-maturity-validation-decision.md).
