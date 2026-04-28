---
decision_id: decision-memory-router-revisit-promotion
decision_title: Memory And Autonomous Router Revisit Promotion
decision_status: accepted
decision_scope: memory-router
decision_owner: platform
---
# Decision: Memory And Autonomous Router Revisit Promotion

## Status

Accepted: keep as reference pressure. Do not promote a public memory,
remember/recall, or autonomous router surface from the refreshed evidence.

Evidence:

- [`memory-router-revisit-adr.md`](memory-router-revisit-adr.md)
- [`../evals/memory-router-revisit-comparison-poc.md`](../evals/memory-router-revisit-comparison-poc.md)
- [`../evals/results/ockp-memory-router-revisit-pressure.md`](../evals/results/ockp-memory-router-revisit-pressure.md)

## Decision

Do not promote a memory API, remember/recall action, memory transport,
autonomous router API, schema, migration, storage behavior, public API, or
implementation follow-up from this evidence.

The current public surface remains:

- `openclerk document`
- `openclerk retrieval`

Capability path: no promotion. The scripted-control row ran 26 tool/command
calls, 7 assistant calls, and 45.43s wall time. It completed with
classification `none`, preserving canonical memory/router authority, source
refs, provenance, synthesis freshness, and bypass boundaries.

Ergonomics path: keep as reference. The natural-intent row ran 26 tool/command
calls, 5 assistant calls, and 66.91s wall time. It completed with
classification `none`, showing the existing workflow can handle the natural
revisit question after guidance/eval repair.

Validation controls passed final-answer-only for missing document path,
negative limit, unsupported lower-level workflow, and unsupported transport.
No bypass risk was observed in the selected rows.

## Promotion Rubric

| Outcome | Standard |
| --- | --- |
| Promote | Repeated targeted rows show `capability_gap`, or repeated natural rows show `ergonomics_gap` while scripted controls pass and an exact promoted surface preserves authority, citations/source refs, provenance, freshness, local-first operation, and AgentOps. |
| Defer | Failures are guidance, answer-contract, eval coverage, partial evidence, one-off ergonomics pressure, or insufficient scripted-control evidence. |
| Kill | The candidate creates independent memory authority, hides citations/source refs, hides provenance or freshness, weakens canonical markdown authority, silently routes across sources, or encourages bypasses. |
| Keep as reference | Existing document/retrieval workflows are sufficient enough, natural UX is acceptable, and the lane remains useful benchmark pressure. |

The current decision is **keep as reference pressure**. No implementation
issue is authorized.

## Follow-Up

Future work may rerun the targeted lane when memory/router evidence changes.
Any future promotion decision must answer both:

- Can current primitives express memory and autonomous routing safely?
- Is the current UX acceptable enough to keep without a promoted surface?

A future promoted design must name exact runner action names and JSON
request/response shapes, explain compatibility with current document and
retrieval workflows, and preserve canonical markdown authority, citations or
source refs, provenance, projection freshness, local-first operation, and
no-bypass boundaries.
