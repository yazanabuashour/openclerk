---
decision_id: decision-graph-relationship-maintenance-plan-promotion
decision_title: Graph Relationship Maintenance Plan Promotion
decision_status: accepted
decision_scope: graph-relationship-maintenance-plan
decision_owner: platform
source_refs: docs/evals/graph-relationship-maintenance-plan-implementation.md, docs/evals/results/ockp-graph-relationship-maintenance-plan-implementation.md, docs/architecture/graph-product-story-promotion-decision.md
follow_up_work: none
---
# Decision: Graph Relationship Maintenance Plan Promotion

## Status

Accepted: promote the narrow read-only
`graph_relationship_maintenance_plan` retrieval action.

This closes `oc-2hx7` by selecting the approval-before-write plan surface for
canonical markdown relationship annotation and maintenance. Durable writes
remain explicit `replace_section` or `append_document` requests after review.

Evidence:

- [`docs/evals/graph-relationship-maintenance-plan-implementation.md`](../evals/graph-relationship-maintenance-plan-implementation.md)
- [`docs/evals/results/ockp-graph-relationship-maintenance-plan-implementation.md`](../evals/results/ockp-graph-relationship-maintenance-plan-implementation.md)
- [`docs/architecture/graph-product-story-promotion-decision.md`](graph-product-story-promotion-decision.md)

## Decision

Promote `graph_relationship_maintenance_plan` under `openclerk retrieval`.

The action packages proposed maintenance actions, candidate relationship
section content, next approved `replace_section` and `append_document`
requests, `planned_no_write`, approval boundary, duplicate handling,
rollback/audit path, failure modes, graph projection freshness, provenance
refs, validation boundaries, authority limits, and
`graph_relationship_maintenance.agent_handoff`.

`graph_relationship_report` remains the read-only evidence report. Existing
document writes remain the durable approval point.

## Safety, Capability, UX

Safety pass: pass. The selected action is read-only, local-first,
runner-only, citation-bearing, and explicit that read/fetch/inspect planning is
not durable-write approval. It does not write, inspect the vault or SQLite
directly, use source-built runners, use unsupported transports, create graph
memory, rank authority, add durable semantic graph storage, or claim
semantic-label graph truth.

Capability pass: pass. The selected action covers approval-gated relationship
maintenance with candidate annotations, candidate section content, exact next
write requests, duplicate handling, rollback/audit path, failure modes,
freshness, and provenance posture.

UX quality: pass. A normal user asking how to maintain relationship markdown
should not need to manually bridge `graph_relationship_report` output into a
write plan. One plan action is simpler while preserving the durable-write
approval boundary.

## Authority Model

Canonical markdown remains semantic relationship authority. Candidate
maintenance text and typed relationship annotations are review-required
suggestions from cited runner evidence. They do not become durable facts until
the user approves the exact document write request.

## Candidate Outcome

| Candidate | Outcome | Rationale |
| --- | --- | --- |
| `current_primitives_plus_graph_relationship_report` | Keep as fallback/reference. | Safe and capable, but too ceremonial because approval, duplicate handling, rollback/audit, and failure-mode text must be assembled manually. |
| `graph_relationship_maintenance_plan` | Promote. | Best plan-only surface: packages the write candidate and evidence posture without crossing the durable-write approval boundary. |
| `durable_semantic_graph_maintenance` | Do not select. | Too much authority and storage surface for the observed need; schema, rollback, duplicate conflict, freshness, and failure-mode evidence remain unproven. |
