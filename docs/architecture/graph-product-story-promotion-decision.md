---
decision_id: decision-graph-product-story-promotion
decision_title: Graph Product Story Promotion Decision
decision_status: accepted
decision_scope: graph-product-stories
decision_owner: platform
source_refs: docs/evals/graph-product-story-exploration.md, docs/evals/results/ockp-graph-product-story-exploration.md, docs/architecture/graph-context-report-promotion-decision.md
follow_up_beads: oc-sl7c, oc-2hx7
---
# Decision: Graph Product Story Promotion

## Status

Accepted: promote only the existing narrow read-only `graph_context_report`
baseline for current graph product work. Defer adjacent read-only report and
approval-before-write maintenance-plan needs to linked candidate comparisons.
Kill durable semantic graph/schema/storage candidates for this track.

Evidence:

- [`docs/evals/graph-product-story-exploration.md`](../evals/graph-product-story-exploration.md)
- [`docs/evals/results/ockp-graph-product-story-exploration.md`](../evals/results/ockp-graph-product-story-exploration.md)
- [`docs/architecture/graph-context-report-promotion-decision.md`](graph-context-report-promotion-decision.md)
- [`docs/architecture/graph-relationship-report-promotion-decision.md`](graph-relationship-report-promotion-decision.md)

## Decision

`graph_context_report` remains the promoted graph product story: routine
read-only relationship graph explanation over runner-visible evidence.

No other graph story promotes from this track. The final outcomes are:

| Graph story | Outcome | Follow-up |
| --- | --- | --- |
| Read-only graph explanation | Promote via `graph_context_report`. | None. |
| Relationship/path finding | Defer. | `oc-sl7c` |
| Direct-vs-inferred relationship reporting | Defer. | `oc-sl7c` |
| Typed relationship candidates from canonical markdown | Defer. | `oc-sl7c` |
| Stale/contradictory/orphaned graph audits | Defer. | `oc-sl7c` |
| Approval-gated relationship annotation or maintenance plans | Defer. | `oc-2hx7` |
| Durable semantic graph/schema/storage candidates | Kill for this track. | None. |

There is no generic evidence-only outcome. Deferred rows have concrete
comparison work. The killed durable graph/storage row fails the current
promotion gate rather than waiting indefinitely.

## Safety, Capability, UX

Safety pass: `graph_context_report` passes because it is read-only,
runner-only, local-first, citation-bearing, and explicit that graph edges,
links, backlinks, provenance, and projection freshness are derived navigation
evidence. It does not write, rank authority, create graph memory, claim
semantic-label graph truth, inspect the vault or SQLite directly, use
source-built runner paths, or rely on unsupported transports.

Capability pass: the promoted baseline covers source identity, cited canonical
relationship text, outgoing links, incoming backlinks, nearby graph context,
graph projection freshness, provenance refs, candidate surfaces, validation
boundaries, and authority limits.

UX quality: normal users can reasonably expect a simpler surface than the
current primitive sequence for routine graph explanation. `graph_context_report`
meets that standard. Adjacent stories are plausible but need separate shape
comparisons before expanding public API or skill guidance.

## Authority Model

Canonical markdown remains semantic relationship authority. Derived graph
state, projections, links, backlinks, and provenance are evidence and
navigation context only. No evaluated candidate proved a safer durable
authority model.

Typed relationship labels, direct/inferred distinctions, and maintenance
recommendations must therefore be treated as cited candidates or plan output
until a later decision promotes a narrower surface.

## Provenance And Freshness

Promoted graph answers must expose source citations or source refs, provenance
refs, and graph projection freshness. Any future write, ranking, memory,
migration, schema, or durable semantic graph candidate must prove auditability,
rollback, provenance, freshness, duplicate handling, and failure-mode behavior
before promotion.

## Follow-Up Beads

Before closing the parent decision work, `bd search` found no existing graph
audit, typed relationship, or relationship annotation follow-up. The linked
follow-ups are:

- `oc-sl7c`: compare graph audit and typed relationship report candidates.
- `oc-2hx7`: compare approval-gated graph relationship maintenance plans.

Both follow-ups must compare 2-3 candidate shapes, choose the best, combine
useful behaviors if appropriate, defer or kill the track, or record `none
viable yet`.

`oc-sl7c` is now resolved by
[`docs/architecture/graph-relationship-report-promotion-decision.md`](graph-relationship-report-promotion-decision.md),
which promotes `graph_relationship_report` for the deferred read-only
relationship/path, direct-vs-derived, typed-candidate, and limited graph-audit
needs.
