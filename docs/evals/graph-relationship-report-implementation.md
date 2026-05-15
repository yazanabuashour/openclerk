# Graph Relationship Report Implementation Eval

`oc-sl7c` compares the deferred graph audit and typed relationship report
candidates from `oc-tfms` and implements the selected narrow read-only
`openclerk retrieval` action.

## Surface

The implemented request shape is:

```json
{
  "action": "graph_relationship_report",
  "graph_relationship": {
    "path": "notes/graph/semantics/index.md",
    "limit": 20
  }
}
```

`graph_relationship` accepts exactly one selector among `doc_id`, `path`, and
`query`; `path_prefix` is valid only with `query`. Path selectors remain
repo-relative markdown paths.

The result packages:

- source document identity and source selection
- one-hop `relationship_paths` from outgoing links and incoming backlinks
- `direct_relationships` from cited canonical markdown and markdown links
- `derived_relationships` from cited graph edges
- `typed_relationship_candidates` suggested from canonical markdown wording
- limited `audit_findings` for stale graph projection, orphaned graph context,
  and simple contradictory relationship wording
- graph projection freshness, provenance refs, validation boundaries,
  authority limits, and `agent_handoff`

## Candidate Comparison

| Candidate surface | Safety pass | Capability pass | UX quality | Outcome |
| --- | --- | --- | --- | --- |
| `current_primitives_plus_graph_context_report` | Pass: uses existing runner-only read actions and `graph_context_report`; no writes or bypasses. | Pass: evidence is visible, but typed labels, direct-vs-derived distinctions, paths, and audit notes must be assembled by the agent. | Ceremonial for normal users asking for relationship paths or audits. | Keep as fallback/reference. |
| `graph_relationship_report` | Pass: read-only wrapper over canonical markdown, links/backlinks, graph edges, provenance, and graph projection freshness. | Pass: combines the deferred path, direct-vs-derived, typed-candidate, and limited audit needs in one cited report. | Best candidate: one action answers the adjacent user intent without expanding durable graph authority. | Promote. |
| `split_specialized_reports` | Could pass with the same boundaries. | Plausible, but the path, candidate, and audit evidence share the same source data. | Worse public API taste: users and agents must choose among adjacent reports before seeing evidence. | Do not select. |

## Boundaries

Canonical markdown remains relationship authority. Typed relationship
candidates are suggestions from cited markdown wording, not durable semantic
truth. Derived graph edges and backlinks are navigation evidence, not
independent inference, contradiction proof, authority ranking, or graph memory.

The action does not add writes, durable semantic graph storage, graph schema,
migrations, semantic-label graph truth, hidden authority ranking, direct vault
inspection, direct SQLite access, source-built runner usage, HTTP/MCP bypasses,
or unsupported transports.

## Targeted Run

```bash
mise exec -- go run ./scripts/agent-eval/ockp run \
  --parallel 1 \
  --scenario graph-relationship-report-action-control,missing-document-path-reject,negative-limit-reject,unsupported-lower-level-reject,unsupported-transport-reject \
  --report-name ockp-graph-relationship-report-implementation
```

The targeted lane requires `graph_relationship_report` to be the first runner
workflow action for the implementation row and requires final-answer-only
negative controls.

The reduced report must be published as:

- `docs/evals/results/ockp-graph-relationship-report-implementation.md`
- `docs/evals/results/ockp-graph-relationship-report-implementation.json`
