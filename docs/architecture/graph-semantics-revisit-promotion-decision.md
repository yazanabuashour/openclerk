---
decision_id: decision-graph-semantics-revisit-promotion
decision_title: Graph Semantics Revisit Promotion
decision_status: accepted
decision_scope: graph-semantics
decision_owner: platform
---
# Decision: Graph Semantics Revisit Promotion

## Status

Accepted: keep graph semantics as reference pressure. Do not promote a public
graph semantics surface from the refreshed evidence.

Evidence:

- [`graph-semantics-revisit-adr.md`](graph-semantics-revisit-adr.md)
- [`../evals/graph-semantics-revisit-comparison-poc.md`](../evals/graph-semantics-revisit-comparison-poc.md)
- [`../evals/results/ockp-graph-semantics-revisit-pressure.md`](../evals/results/ockp-graph-semantics-revisit-pressure.md)

## Decision

Do not promote a semantic-label graph layer, graph semantics runner action,
schema, migration, storage behavior, public API, or implementation follow-up
from this evidence.

The current public surface remains:

- `openclerk document`
- `openclerk retrieval`

Capability path: no promotion. The refreshed scripted-control row completed
with `none` failure classification, 22 tool/command calls, 6 assistant calls,
and 46.95s wall time. Current `openclerk document` and `openclerk retrieval`
primitives can safely express the relationship-shaped workflow while preserving
canonical markdown authority, citations, graph projection freshness, and
bypass boundaries.

Ergonomics path: no promotion. The refreshed natural-intent row completed with
`none` failure classification, 28 tool/command calls, 5 assistant calls, and
99.11s wall time. The row remains high-latency benchmark pressure, but it did
not fail and does not show repeated ergonomics evidence that justifies a new
public surface.

Validation controls passed final-answer-only for missing document path,
negative limit, unsupported lower-level workflow, and unsupported transport.
No bypass risk was observed in the selected rows.

## Promotion Rubric

| Outcome | Standard |
| --- | --- |
| Promote | Repeated targeted rows show `capability_gap`, or repeated natural rows show `ergonomics_gap` while scripted controls pass and an exact promoted surface preserves authority, citations, provenance, freshness, local-first operation, and AgentOps. |
| Defer | Failures are guidance, eval coverage, partial evidence, one-off ergonomics pressure, or insufficient scripted-control evidence. |
| Kill | The candidate creates independent graph truth, hides citations, hides provenance or freshness, weakens canonical markdown authority, or encourages bypasses. |
| Keep as reference | Existing document/retrieval workflows are sufficient enough, and the lane remains useful benchmark pressure. |

The current decision is **keep as reference pressure**. No implementation issue
is authorized.

## Follow-Up

Future work may rerun natural-intent pressure if additional graph workflows
show repeated UX or reliability failures. Any future promotion decision must
answer both:

- Can current primitives express relationship-shaped graph semantics safely?
- Is the current UX acceptable enough to keep without a promoted surface?

A future promoted design must name exact runner action names and JSON
request/response shapes, explain compatibility with current document and
retrieval workflows, and preserve canonical markdown authority, citations,
provenance, graph projection freshness, local-first operation, and no-bypass
boundaries.
