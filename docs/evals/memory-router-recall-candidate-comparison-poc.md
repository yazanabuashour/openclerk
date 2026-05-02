# Memory/Router Recall Candidate Comparison POC

## Status

Implemented candidate-comparison framing for `oc-ge4p`.

This document compares candidate shapes for reducing memory/router recall
ceremony after `oc-nu12`. It does not add runner actions, schemas, migrations,
storage behavior, public API behavior, product behavior, memory transports,
remember/recall actions, autonomous router APIs, or shipped skill behavior.

Governing evidence:

- [`docs/evals/results/ockp-high-touch-memory-router-recall-ceremony.md`](results/ockp-high-touch-memory-router-recall-ceremony.md)
- [`docs/architecture/memory-router-recall-ceremony-promotion-decision.md`](../architecture/memory-router-recall-ceremony-promotion-decision.md)
- [`docs/evals/results/ockp-memory-router-revisit-pressure.md`](results/ockp-memory-router-revisit-pressure.md)
- [`docs/architecture/memory-router-revisit-promotion-decision.md`](../architecture/memory-router-revisit-promotion-decision.md)

## Candidate Workflows

| Candidate | Shape | Strength | Risk |
| --- | --- | --- | --- |
| Guidance-only repair | Keep existing `openclerk document` and `openclerk retrieval` calls; repair skill or prompt guidance for memory/router recall. | No API or response change; preserves all current safety boundaries. | The `oc-nu12` natural row failed after 32 tools/commands and 5 assistant calls, and the scripted row still missed required answer/step evidence. Guidance alone may preserve the fragile ceremony. |
| Narrow memory/router recall candidate | Evaluate a future read-only helper or report surface that packages temporal status, canonical evidence, stale/session status, source refs, provenance, synthesis freshness, feedback weighting, routing rationale, validation boundaries, and authority limits. | Directly targets the routine user expectation for one recall and routing answer while keeping safety evidence visible. | Must not introduce hidden memory authority, stale-evidence ranking, memory transports, remember/recall actions, autonomous routing, or provenance/freshness hiding. |
| No new surface after prompt or harness repair | Treat `oc-nu12` as repair pressure only and keep all work on existing primitives after prompt or harness adjustment. | Avoids over-promoting from one failed natural row and one scripted answer-shape miss. | Leaves a real UX need unresolved: normal users should not need a high-step ceremony for temporal recall and routing rationale. |

## Selected Candidate

Select the narrow memory/router recall candidate for future targeted evidence,
not implementation.

The future candidate should evaluate a read-only helper or report surface that
accepts routine temporal recall and routing intent and returns the evidence
needed to answer safely without requiring a manually stitched retrieval
sequence. A future response candidate should expose:

- query summary
- temporal status
- canonical evidence refs
- stale or session-observation status
- feedback weighting as advisory
- routing rationale through existing document/retrieval authority
- provenance refs
- synthesis freshness
- validation and no-bypass boundaries
- authority limits explaining that canonical markdown remains durable memory
  authority

The candidate must not promote memory transports, remember/recall actions,
autonomous router APIs, vector stores, embedding stores, graph memory, direct
SQLite, direct vault inspection, source-built runners, HTTP/MCP bypasses,
unsupported transports, hidden authority ranking, or durable write shortcuts.

## Evidence Scorecard

| Evidence | Safety | Capability | UX quality |
| --- | --- | --- | --- |
| `oc-nu12` natural row | Passed. No bypass, direct storage, unsupported transport, memory transport, autonomous router, or durable write risk was observed. | Passed. Runner-visible evidence existed and current primitives could express the workflow. | Failed with `ergonomics_gap`: 32 tools/commands, 5 assistant calls, 50.35s, and `answer_repair_needed`. |
| `oc-nu12` scripted control | Passed. Preserved runner-only access, canonical memory/router authority, source refs, provenance, and synthesis freshness. | Passed with repair need. The row failed with `skill_guidance_or_eval_coverage`, not `capability_gap`, after missing required synthesis `get_document` evidence. | Still ceremonial and fragile: 34 tools/commands, 8 assistant calls, 60.32s, and `answer_repair_needed`. |
| Prior memory/router revisit pressure | Passed. Preserved canonical memory/router authority, source refs, provenance, synthesis freshness, and no-bypass boundaries. | Passed with `none` for natural and scripted rows using current primitives. | Completed but high-touch: natural row used 26 tools/commands, 5 assistant calls, and 66.91s. |

## Conclusion

Do not file an implementation bead from this comparison. File targeted
eval/promotion evidence for the selected narrow memory/router recall
candidate.

Follow-up `oc-fnhj` should compare the selected candidate against current
primitives and guidance-only repair. Promotion remains blocked until evidence
shows the candidate reduces ceremony while preserving canonical markdown
authority, temporal status, source refs or citations, provenance, synthesis
freshness, advisory feedback weighting, routing rationale, local-first
runner-only access, no-bypass controls, approval-before-write, and validation
controls.
