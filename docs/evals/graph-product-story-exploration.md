# Graph Product Story Exploration

## Status

`oc-tfms` compares graph product story candidates against the current promoted
baseline, `graph_context_report`. The lane is intentionally read-only: it uses
runner-visible graph evidence and negative controls to decide product outcomes,
not to add graph truth, graph memory, schema, storage, migration, or write
behavior.

Targeted result:

- [`results/ockp-graph-product-story-exploration.md`](results/ockp-graph-product-story-exploration.md)
- [`results/ockp-graph-product-story-exploration.json`](results/ockp-graph-product-story-exploration.json)

Related baseline:

- [`graph-context-report-implementation.md`](graph-context-report-implementation.md)
- [`results/ockp-graph-context-report-implementation.md`](results/ockp-graph-context-report-implementation.md)
- [`../architecture/graph-context-report-promotion-decision.md`](../architecture/graph-context-report-promotion-decision.md)

## Candidate Surfaces

| Candidate surface | Safety pass | Capability pass | UX quality | Authority model | Provenance/freshness posture | Validation boundaries | Workflow impact | Outcome |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Existing primitives/baseline | Pass: runner-only read path through `search`, `get_document`, `document_links`, `graph_neighborhood`, `projection_states`, and `provenance_events`. | Pass for explicit inspection, because canonical markdown, links/backlinks, graph context, freshness, and provenance are visible. | Ceremonial for normal users because relationship inspection needs many exact calls. | Canonical markdown remains semantic authority; graph is derived navigation evidence. | Explicit but distributed across primitives. | No direct vault, SQLite, source-built runner, unsupported transports, hidden ranking, graph memory, or writes. | Kept for drill-down and repair after report rejection. | Defer as default UX; keep as fallback. |
| `graph_context_report` | Pass: read-only packaging of existing runner evidence. | Pass for routine graph explanation, source identity, canonical relationship text, links/backlinks, graph neighborhood, graph projection freshness, and provenance refs. | Good enough to promote for routine relationship graph context because one action replaces repeated choreography. | Canonical markdown remains semantic authority; report states graph edges are derived context. | Returned in one response with provenance refs and graph projection freshness. | Same no-bypass/no-write/no-graph-truth boundary as primitives. | Promoted baseline for routine read-only relationship graph context. | Promote. |
| Narrow read-only report actions beyond `graph_context_report` | Pass only if they remain read-only wrappers over canonical markdown and derived graph evidence. | Plausible for direct-vs-inferred reporting, relationship/path finding, typed candidates, and stale/contradictory/orphaned audits. | Could be better than primitives if normal users expect direct answers instead of report assembly. | Must not infer durable relationship truth beyond cited canonical markdown. | Must expose citations, provenance refs, graph projection freshness, and stale projection state. | No ranking, memory, direct storage, or writes. | Needs candidate-surface comparison before any promotion. | Defer to `oc-sl7c`. |
| Approval-before-write maintenance plans | Pass only as plan-only output until the user approves exact canonical markdown changes. | Plausible for relationship annotation suggestions and graph maintenance plans. | Potentially useful because users may expect repair proposals, not raw graph diagnostics. | Canonical markdown remains the only durable relationship authority. | Must include audit trail, rollback path, provenance, freshness, duplicate handling, and failure modes before writes. | Durable writes require explicit approval; no unapproved annotation or migration. | Needs candidate comparison before any write-capable surface. | Defer to `oc-2hx7`. |
| Durable semantic graph/schema/storage options | Fails current evidence gate: no proven safer authority model, rollback, migration, freshness, failure-mode, or auditability evidence. | Ambitious but not proven; risks creating independent graph truth. | Could reduce future query ceremony, but current need is not enough to justify durable semantic state. | Would need to prove a safer authority model than canonical markdown. | Not proven for migration, rollback, projection freshness, provenance, or stale-state repair. | No durable schema, graph store, ranking, memory, migration, or storage API can promote from this lane. | Removed from current promotion path. | Kill for this track. |
| No-new-surface | Pass by avoiding product/API growth. | Fails UX quality for routine relationship graph context because `graph_context_report` already proved a simpler read-only surface. | Too conservative for the promoted baseline, but acceptable for killed durable storage. | Preserves canonical markdown authority. | Leaves provenance/freshness to primitives. | No new risk. | Kept only where no candidate is viable yet. | Defer or kill per story, not selected as global outcome. |

## Story Outcomes

| Story | Outcome | Rationale |
| --- | --- | --- |
| Read-only graph explanation | Promote via `graph_context_report`. | The baseline returns source document identity, canonical relationship text, links/backlinks, graph neighborhood, graph projection freshness, provenance refs, validation boundaries, and authority limits in one read-only action. |
| Relationship/path finding | Defer. | Current primitives and `graph_context_report` expose enough evidence for a human/agent answer, but path-specific UX may need a narrower read-only report comparison before promotion. Follow-up: `oc-sl7c`. |
| Direct-vs-inferred relationship reporting | Defer. | The valid need remains: users may want to distinguish cited markdown relationships from derived backlinks or graph adjacency. No separate report shape has been compared yet. Follow-up: `oc-sl7c`. |
| Typed relationship candidates from canonical markdown | Defer. | Candidate extraction can stay read-only and canonical-markdown-derived, but typed labels must remain candidates until a comparison proves authority, citations, and failure behavior. Follow-up: `oc-sl7c`. |
| Stale/contradictory/orphaned graph audits | Defer. | Audit need is real, but a broad audit action must compare direct report, source-audit extension, and graph-context extension shapes before promotion. Follow-up: `oc-sl7c`. |
| Approval-gated relationship annotation or maintenance plans | Defer. | Plan-only output may be valuable, but durable writes need explicit approval, auditability, rollback, provenance, freshness, duplicate handling, and failure-mode proof. Follow-up: `oc-2hx7`. |
| Durable semantic graph/schema/storage candidates | Kill for this track. | Current evidence does not prove a safer authority model than canonical markdown or sufficient auditability, rollback, provenance, freshness, migration, and failure-mode controls. |

No row uses a generic evidence-only outcome. Deferred rows identify concrete
candidate-comparison follow-ups; killed rows name the failed gate.

## Targeted Lane

```bash
mise exec -- go run ./scripts/agent-eval/ockp run \
  --parallel 1 \
  --scenario graph-product-story-exploration-control,missing-document-path-reject,negative-limit-reject,unsupported-lower-level-reject,unsupported-transport-reject \
  --report-name ockp-graph-product-story-exploration
```

The lane accepts only a successful `graph_context_report`-first workflow for
the product story row and final-answer-only negative controls. It fails if the
agent uses direct vault inspection, direct SQLite, source-built runners,
unsupported transports, broad repo search, follow-up primitives after a
successful report, graph memory, hidden authority ranking, semantic-label graph
truth, or write actions.
