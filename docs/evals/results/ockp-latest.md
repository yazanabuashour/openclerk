# OpenClerk Agent Eval Status

This file is a status pointer, not a raw eval report.

The previous contents recorded an older partial production run where
`production` passed 1/17 selected scenarios. That run is superseded and should
not be used as current OpenClerk AgentOps status.

## Current Full Production Gate

Use `docs/evals/results/ockp-agentops-production.md` as the current full
production gate artifact.

- Variant: `production`
- Gate: passing
- Scenario coverage: 18/18 production scenarios passed
- Recommendation: use the AgentOps runner for routine OpenClerk operations
- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`

That report is the current release-gate evidence for routine OpenClerk
knowledge workflows through the installed `openclerk` runner.

## Later Targeted POC Evidence

The following later reports are targeted POC evidence. Their production gates
may be false because they intentionally selected only a subset of scenarios;
use their selected scenario results and decision notes as POC evidence, not as
full release-gate replacements.

| Area | Report | Status |
| --- | --- | --- |
| Convention-first layout inspection | `docs/evals/results/ockp-layout-configuration.md` | `configured-layout-explain` and `invalid-layout-visible` completed |
| Synthesis compiler pressure | `docs/evals/results/ockp-synthesis-compiler-pressure.md` | Selected synthesis pressure and contract scenarios completed |
| Source URL ingestion POC | `docs/evals/results/ockp-source-url-ingestion-poc.md` | `ingest_source_url` promoted for PDF source URL ingestion |
| Source URL update mode | `docs/evals/results/ockp-source-url-update-mode.md` | Targeted AgentOps coverage passed duplicate create rejection, same-SHA no-op, changed-PDF stale projection visibility, and path-hint conflict behavior |
| Decision records POC | `docs/evals/results/ockp-decision-records-poc.md` | Selected decision lookup and supersession scenarios completed |
| Decision records hardening | `docs/evals/results/ockp-decision-records-hardening.md` | Real ADR migration and decision projection hardening scenarios completed |
| Source-sensitive audit | `docs/evals/results/ockp-source-sensitive-audit-poc.md` | Stale repair and unresolved conflict scenarios completed |
| Graph semantics reference | `docs/evals/results/ockp-graph-semantics-reference-poc.md` | Graph semantic-label pressure scenario completed as reference evidence |
| Memory/router reference | `docs/evals/results/ockp-memory-router-reference-poc.md` | Memory and router pressure scenario completed as reference evidence |
| Real-vault dogfooding | `docs/evals/results/ockp-real-vault-agentops-trial.md` | Sanitized aggregate evidence only; raw logs, paths, titles, snippets, citations, document IDs, and raw JSON remain local-only |
| Populated-vault synthetic pressure | `docs/evals/results/ockp-populated-vault-targeted.md`; follow-up `docs/evals/results/ockp-populated-vault-guidance-hardening.md` | Targeted run completed freshness/conflict and synthesis-update pressure; focused follow-up resolved the heterogeneous polluted-evidence guidance failure, with no runner capability gap or product/API promotion |
| Document-this intake pressure | `docs/evals/results/ockp-document-this-intake-pressure.md` | Targeted document-this intake lane covers missing-field clarification, explicit create/override, duplicate candidates, existing updates, and synthesis freshness |
| Document artifact candidate generation | `docs/evals/results/ockp-document-artifact-candidate-generation.md` | Targeted propose-before-create lane defers promotion pending repair; title/path, mixed-source, explicit-override, and body-faithfulness scenarios exposed `candidate_quality_gap`, and low-confidence clarification exposed `skill_guidance_or_eval_coverage` |

## Interpretation

OpenClerk is currently proven for the v1 AgentOps runner slice: canonical docs,
source-linked synthesis, promoted records, provenance events, projection
freshness, and final-answer-only rejection gates.

The later targeted POCs do not promote Mem0 or a memory API, an autonomous
router, a semantic graph truth layer, a broad contradiction engine, or new
public runner actions. Those capabilities remain deferred until eval evidence
shows the existing `openclerk document` and `openclerk retrieval` actions are
structurally insufficient.
