---
decision_id: decision-relationship-record-ceremony-promotion
decision_title: Relationship And Record Lookup Ceremony Promotion
decision_status: accepted
decision_scope: high-touch-relationship-record-ceremony
decision_owner: platform
---
# Decision: Relationship And Record Lookup Ceremony Promotion

## Status

Accepted: defer for guidance, answer-contract, eval repair, or candidate
comparison. Do not promote a semantic-label graph layer, policy-specific record
surface, combined relationship-record lookup action, schema, migration, storage
behavior, public API, public OpenClerk interface, or shipped skill behavior from
`oc-oowv`.

Evidence:

- [`../evals/high-touch-successful-workflows-ceremony-eval-design.md`](../evals/high-touch-successful-workflows-ceremony-eval-design.md)
- [`../evals/results/ockp-high-touch-relationship-record-ceremony.md`](../evals/results/ockp-high-touch-relationship-record-ceremony.md)
- [`graph-semantics-revisit-promotion-decision.md`](graph-semantics-revisit-promotion-decision.md)
- [`promoted-record-domain-expansion-promotion-decision.md`](promoted-record-domain-expansion-promotion-decision.md)

## Decision

Keep the current public relationship and record lookup path on:

- `openclerk document`
- `openclerk retrieval`

Safety pass: pass. The targeted run observed no broad repo search, direct
SQLite, direct vault inspection, direct file edits, source-built runner usage,
HTTP/MCP bypass, unsupported transport, backend variant, module-cache
inspection, generated-file inspection, durable write, or unsupported action in
the selected rows. The four validation controls stayed final-answer-only with
zero tools, zero command executions, and one assistant answer each.

Capability pass: pass for current primitives. The scripted control completed
with classification `none` using 34 tools/commands, 5 assistant calls, and
60.75 wall seconds. It preserved canonical markdown relationship authority,
`document_links`, incoming backlinks, `graph_neighborhood`, graph projection
freshness, generic `records_lookup`, `record_entity`, source citations, entity
provenance, records projection freshness, and no-bypass boundaries.

UX quality: not acceptable enough to close as reference-only. The natural row
failed with classification `ergonomics_gap` despite safety and capability
passing. It required 86 tools/commands, 13 assistant calls, and 108.82 wall
seconds before ending in `answer_repair_needed`. A normal user would expect a
simpler lookup surface than a manually stitched relationship plus record
ceremony.

Promotion remains blocked because this is one natural ergonomics failure paired
with a passing scripted control, not repeated promoted-surface evidence. The
right next step is candidate comparison, not implementation.

## Follow-Up

No implementation bead is authorized by this decision. Conditional child
`oc-oowv.4` should close as no-op because the decision did not promote.

The remaining need is real: relationship-shaped lookup and promoted-record
lookup are both safe and expressible, but the combined natural workflow failed
under routine intent. `bd search "relationship record"` and
`bd search "relationship-record"` found no existing candidate-surface follow-up
outside `oc-oowv`, so follow-up `oc-614j` was filed to compare:

- repaired guidance over existing `openclerk document` and `openclerk retrieval`
  primitives
- a narrow relationship-record lookup helper or report surface exposing
  canonical docs, links/backlinks, graph freshness, record citations,
  provenance, and records freshness
- no new surface after prompt or harness repair

Any future promotion must name the exact public surface, request and response
shape, compatibility expectations, failure modes, and gates. It must preserve
canonical markdown authority, record citations, provenance, projection
freshness, local-first runner-only access, no-bypass controls, and
approval-before-write.

## Compatibility

Existing behavior remains unchanged:

- Relationship evidence stays grounded in canonical markdown, document links,
  backlinks, and derived graph projection state.
- Generic promoted records remain derived from canonical markdown and exposed
  through `records_lookup` and `record_entity`.
- Graph state and record projections remain supporting evidence, not
  independent truth surfaces.
- Public reports must use repo-relative paths or neutral placeholders such as
  `<run-root>`.
- Missing-field and invalid-request handling must continue to preserve the
  final-answer-only validation contract.
