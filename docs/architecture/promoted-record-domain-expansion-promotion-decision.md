---
decision_id: decision-promoted-record-domain-expansion-promotion
decision_title: Promoted Record Domain Expansion Promotion
decision_status: accepted
decision_scope: promoted-records
decision_owner: platform
---
# Decision: Promoted Record Domain Expansion Promotion

## Status

Accepted: defer for guidance/eval repair. Do not promote a policy-specific or
other typed record-domain surface from the current evidence.

Evidence:

- [`promoted-record-domain-expansion-adr.md`](promoted-record-domain-expansion-adr.md)
- [`../evals/promoted-record-domain-expansion-comparison-poc.md`](../evals/promoted-record-domain-expansion-comparison-poc.md)
- [`../evals/results/ockp-promoted-record-domain-expansion-pressure.md`](../evals/results/ockp-promoted-record-domain-expansion-pressure.md)

## Decision

Do not promote a policy-specific record action, typed record-domain runner
surface, schema, migration, storage behavior, public API, or implementation
follow-up from this evidence.

The current public surface remains:

- `openclerk document`
- `openclerk retrieval`

Capability path: no promotion. The scripted-control row ran 16 tool/command
calls, 4 assistant calls, and 33.54s wall time. It failed with
`skill_guidance_or_eval_coverage`, not `capability_gap`: runner-visible
promoted-record evidence existed, and the failure was the assistant comparison
answer contract.

Ergonomics path: defer. The natural-intent row ran 28 tool/command calls, 8
assistant calls, and 70.68s wall time. It completed with `none` failure
classification while preserving canonical record authority, citations,
provenance, records projection freshness, and bypass boundaries. This is
high-latency benchmark pressure, but it is not repeated ergonomics-gap evidence.

Validation controls passed final-answer-only for missing document path,
negative limit, unsupported lower-level workflow, and unsupported transport.
No bypass risk was observed in the selected rows.

## Promotion Rubric

| Outcome | Standard |
| --- | --- |
| Promote | Repeated targeted rows show `capability_gap`, or repeated natural rows show `ergonomics_gap` while scripted controls pass and an exact promoted surface preserves authority, citations, provenance, freshness, local-first operation, and AgentOps. |
| Defer | Failures are guidance, answer-contract, eval coverage, partial evidence, one-off ergonomics pressure, or insufficient scripted-control evidence. |
| Kill | The candidate creates independent record truth, hides citations, hides provenance or freshness, weakens canonical markdown authority, silently routes around generic records, or encourages bypasses. |
| Keep as reference | Existing document/retrieval workflows are sufficient enough, natural UX is acceptable, and the lane remains useful benchmark pressure. |

The current decision is **defer for guidance/eval repair**. No implementation
issue is authorized.

## Follow-Up

Future work may repair the scripted answer contract and rerun this lane. Any
future promotion decision must answer both:

- Can current primitives express domain-shaped promoted record workflows safely?
- Is the current UX acceptable enough to keep without a promoted typed surface?

A future promoted design must name exact runner action names and JSON
request/response shapes, explain compatibility with current document/retrieval
and generic record workflows, and preserve canonical markdown authority,
citations, provenance, records projection freshness, local-first operation, and
no-bypass boundaries.

Follow-up `oc-biyn` repaired the scripted answer contract and reran
`docs/evals/results/ockp-promoted-record-domain-expansion-pressure.md`. The
lane now passes as `keep_as_reference`; no policy-specific record action, typed
record-domain runner surface, schema, migration, storage behavior, public API,
or implementation follow-up is promoted by that repaired evidence.
