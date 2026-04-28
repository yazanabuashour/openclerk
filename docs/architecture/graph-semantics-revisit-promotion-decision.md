---
decision_id: decision-graph-semantics-revisit-promotion
decision_title: Graph Semantics Revisit Promotion
decision_status: accepted
decision_scope: graph-semantics
decision_owner: platform
---
# Decision: Graph Semantics Revisit Promotion

## Status

Accepted: defer graph semantics promotion for guidance or eval repair. Keep
richer graph semantics as reference pressure for now.

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

Capability path: no promotion. The scripted-control row failed, but its
database evidence passed. The failure classification was
`skill_guidance_or_eval_coverage`, not `capability_gap`: runner-visible graph
evidence existed, but the assistant answer or required runner steps did not
satisfy the scenario. This does not prove current primitives are structurally
insufficient.

Ergonomics path: no promotion yet. The natural-intent row reported
`ergonomics_gap`, with 26 tool/command calls, 7 assistant calls, and 80.88s
wall time. That is real pressure, but it is not repeated successful
scripted-control-backed ergonomics evidence. Because the scripted control still
needs guidance or eval repair, the current evidence is too narrow to justify a
new public surface.

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

The current decision is **defer** for guidance or eval repair and **keep as
reference pressure**. No implementation issue is authorized.

## Follow-Up

Future work may repair the graph semantics revisit eval or guidance so the
scripted-control row passes cleanly, then rerun natural-intent pressure. Any
future promotion decision must answer both:

- Can current primitives express relationship-shaped graph semantics safely?
- Is the current UX acceptable enough to keep without a promoted surface?

A future promoted design must name exact runner action names and JSON
request/response shapes, explain compatibility with current document and
retrieval workflows, and preserve canonical markdown authority, citations,
provenance, graph projection freshness, local-first operation, and no-bypass
boundaries.
