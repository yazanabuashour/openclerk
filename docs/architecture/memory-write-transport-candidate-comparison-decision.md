---
decision_id: decision-memory-write-transport-candidate-comparison
decision_title: Memory Write Transport Candidate Comparison
decision_status: accepted
decision_scope: memory-write-transport-candidates
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/architecture/memory-architecture-recall-adr.md, docs/evals/results/ockp-memory-architecture-recall-track.md, docs/evals/results/ockp-memory-router-recall-report-implementation.md, docs/architecture/post-oc-uj2y-deferred-pipeline-reconciliation.md
---
# Decision: Memory Write Transport Candidate Comparison

## Status

Accepted as a non-promotion decision for `oc-rcfv`.

Do not add a memory write transport, `remember`/`recall` write action,
autonomous router API, Mem0 adapter, vector memory, graph memory, or hidden
authority ranking.

## Decision

Select the current combined shape:

- source-linked memory documents remain the durable authority pattern when a
  user explicitly approves a document write
- `memory_router_recall_report` remains the read-only recall/report surface
- derived memory projections and external adapters remain reference/future
  candidates only

Candidate comparison:

| Candidate | Safety | Capability | UX quality | Outcome |
| --- | --- | --- | --- | --- |
| Source-linked memory documents | Pass with approved writes, citations, provenance, and canonical markdown authority. | Good for durable inspected memory. | Acceptable through existing document workflows. | Keep as authority pattern. |
| Derived memory projections | Not proven for correction/delete, duplicate, freshness, or canonical conflicts. | Potentially useful later. | Risky if it looks like hidden truth. | Defer. |
| Explicit memory write action | Not proven for approval, privacy, source refs, lifecycle, or conflicts. | Could reduce ceremony, but current gap is not proven. | Too surprising without stronger evidence. | Defer. |
| Mem0/external adapter | Fails local-first/privacy posture for routine memory writes. | Useful as reference only. | Adds provider and lifecycle ceremony. | Reference only. |

## Safety, Capability, UX

Safety pass: pass. The selected outcome preserves canonical markdown/promoted
record authority, source citations, freshness, duplicate handling, privacy
posture, correction/delete lifecycle, canonical-conflict behavior,
installed-runner access, local-first behavior, and approval before durable
writes.

Capability pass: pass for current needs. Existing source-linked documents and
the read-only recall report cover the current memory workflows without adding
a second truth store.

UX quality: pass. The selected surface avoids surprising persistence while
keeping the lower-step read-only memory recall report for routine recall.

## Follow-Up

Search performed before closing `oc-rcfv`:

- `bd search "memory write transport source linked projection Mem0" --status all`: no existing issue found.

No new follow-up is required because this decision records no current
implementation need. Reopen only with new repeated evidence that existing
source-linked writes and read-only recall reports cannot safely express the
workflow.

## Compatibility

Existing behavior remains unchanged. No runner schema, storage schema, memory
transport, provider adapter, public API, or skill behavior is added.
