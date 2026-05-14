# Graph Context Report Implementation Eval

`oc-gy2s` implements the promoted read-only graph context report as an
`openclerk retrieval` action. This eval compares the current primitives plus
help path against the narrow report action and verifies that the implementation
solves routine relationship graph inspection ceremony without adding graph
truth.

## Surface

The implemented request shape is:

```json
{
  "action": "graph_context_report",
  "graph_context": {
    "path": "notes/graph/semantics/index.md",
    "limit": 20
  }
}
```

`graph_context` accepts exactly one selector among `doc_id`, `path`, and
`query`; `path_prefix` is valid only with `query`. Path selectors must remain
repo-relative markdown paths.

The result packages:

- source document identity and source selection
- cited canonical markdown relationship text
- outgoing links and incoming backlinks
- nearby graph neighborhood evidence
- graph projection freshness
- provenance refs
- candidate surfaces
- validation boundaries, authority limits, and `agent_handoff`

## Boundaries

The action is read-only and uses existing runner-visible document/retrieval
evidence only. Canonical markdown remains semantic relationship authority.
Graph edges, links, backlinks, provenance, and projection freshness are
derived navigation evidence, not independent truth or hidden ranking.

The action does not add semantic-label graph truth, graph memory, authority
ranking, writes, direct vault inspection, direct SQLite access, source-built
runner usage, HTTP/MCP bypasses, unsupported transports, storage changes,
schema changes, or migrations.

## Candidate Comparison

The targeted lane compares:

- `current_primitives_plus_help`: inspect retrieval help, search, list/get
  canonical markdown, links/backlinks, graph neighborhood, graph freshness,
  and provenance. This passes safety and capability but preserves high
  ceremony for routine callers.
- `graph_context_report`: one read-only retrieval action returning the same
  authority and provenance posture in a compact report.
- `no_new_surface`: safe by avoiding API growth, but does not solve the
  observed UX pressure.

## Targeted Run

```bash
mise exec -- go run ./scripts/agent-eval/ockp run \
  --parallel 1 \
  --scenario graph-context-current-primitives-plus-help,graph-context-report-action-natural,missing-document-path-reject,negative-limit-reject,unsupported-lower-level-reject,unsupported-transport-reject \
  --report-name ockp-graph-context-report-implementation
```

The committed evidence run serializes the targeted rows so the current
primitive control and promoted report action each exercise their runner path
without cross-row agent nondeterminism.

The reduced report must be published as:

- `docs/evals/results/ockp-graph-context-report-implementation.md`
- `docs/evals/results/ockp-graph-context-report-implementation.json`
