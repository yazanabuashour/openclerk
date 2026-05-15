# Graph Relationship Maintenance Plan Implementation Eval

`oc-2hx7` compares approval-before-write graph relationship maintenance
candidate surfaces and implements the selected narrow read-only
`openclerk retrieval` action.

## Surface

The implemented request shape is:

```json
{
  "action": "graph_relationship_maintenance_plan",
  "graph_relationship_maintenance": {
    "path": "notes/graph/semantics/index.md",
    "limit": 20
  }
}
```

`graph_relationship_maintenance` accepts exactly one selector among `doc_id`,
`path`, and `query`; `path_prefix` is valid only with `query`. Path selectors
remain repo-relative markdown paths.

The result packages:

- source document identity and source selection
- `proposed_actions` for review-required relationship annotations or audit
  follow-up
- `candidate_section_content` for the canonical `Relationships` section
- `next_replace_section_request` and `next_append_document_request`
- `planned_no_write`, `approval_boundary`, duplicate handling, rollback/audit
  path, and failure modes
- graph projection freshness, provenance refs, validation boundaries,
  authority limits, and `agent_handoff`

## Candidate Comparison

| Candidate surface | Safety pass | Capability pass | UX quality | Outcome |
| --- | --- | --- | --- | --- |
| `current_primitives_plus_graph_relationship_report` | Pass: uses existing runner-only read actions and explicit document writes after approval. | Pass: evidence is visible, but the write plan, duplicate handling, rollback/audit path, and failure modes must be assembled manually. | Ceremonial for normal users asking how to maintain relationship markdown. | Keep as fallback/reference. |
| `graph_relationship_maintenance_plan` | Pass: read-only plan over `graph_relationship_report` evidence; exact write requests are returned but not executed. | Pass: combines proposed actions, candidate section content, next approved write requests, duplicate handling, rollback/audit path, freshness/provenance posture, and failure modes. | Best candidate: one action answers the maintenance intent without expanding durable graph authority. | Promote. |
| `durable_semantic_graph_maintenance` | Fails current promotion evidence because it would add durable graph authority and storage behavior. | Plausible only with future schema, migration, rollback, duplicate conflict, freshness, and failure-mode proof. | Too large and less inspectable than approval-gated markdown maintenance. | Do not select. |

## Boundaries

Canonical markdown remains relationship authority. Candidate maintenance text
and typed relationship annotations are suggestions from cited runner evidence,
not durable facts until an explicit approved document write.

The action does not add writes, durable semantic graph storage, graph schema,
migrations, semantic-label graph truth, hidden authority ranking, direct vault
inspection, direct SQLite access, source-built runner usage, HTTP/MCP bypasses,
unsupported transports, graph memory, or automatic repair.

## Targeted Run

```bash
mise exec -- go run ./scripts/agent-eval/ockp run \
  --parallel 1 \
  --scenario graph-relationship-maintenance-plan-action-control,missing-document-path-reject,negative-limit-reject,unsupported-lower-level-reject,unsupported-transport-reject \
  --report-name ockp-graph-relationship-maintenance-plan-implementation
```

The targeted lane requires `graph_relationship_maintenance_plan` to be the
first relevant runner workflow action for the implementation row and requires
final-answer-only negative controls.

The reduced report is published as:

- `docs/evals/results/ockp-graph-relationship-maintenance-plan-implementation.md`
- `docs/evals/results/ockp-graph-relationship-maintenance-plan-implementation.json`
