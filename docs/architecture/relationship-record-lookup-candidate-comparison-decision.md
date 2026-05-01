---
decision_id: decision-relationship-record-lookup-candidate-comparison
decision_title: Relationship-Record Lookup Candidate Comparison
decision_status: accepted
decision_scope: relationship-record-lookup
decision_owner: platform
---
# Decision: Relationship-Record Lookup Candidate Comparison

## Status

Accepted: select a future narrow relationship-record lookup helper or report
candidate for targeted promotion evidence.

This decision does not add a runner action, schema, migration, storage
behavior, public API, public OpenClerk interface, product behavior, or shipped
skill behavior. It does not authorize implementation work.

Evidence:

- [`docs/evals/relationship-record-lookup-candidate-comparison-poc.md`](../evals/relationship-record-lookup-candidate-comparison-poc.md)
- [`docs/architecture/relationship-record-ceremony-promotion-decision.md`](relationship-record-ceremony-promotion-decision.md)
- [`docs/evals/results/ockp-high-touch-relationship-record-ceremony.md`](../evals/results/ockp-high-touch-relationship-record-ceremony.md)
- [`docs/architecture/graph-semantics-revisit-promotion-decision.md`](graph-semantics-revisit-promotion-decision.md)
- [`docs/architecture/promoted-record-domain-expansion-promotion-decision.md`](promoted-record-domain-expansion-promotion-decision.md)

## Decision

Select the candidate shape: a future narrow relationship-record lookup helper
or report surface that exposes canonical relationship evidence, document
links/backlinks, graph projection freshness, record lookup/entity evidence,
citations, provenance, records projection freshness, no-bypass boundaries, and
authority limits. Do not implement the candidate yet.

Rejected alternatives:

- Guidance-only repair is too weak as the next step because the `oc-oowv`
  natural row preserved safety and capability but failed with an
  `ergonomics_gap` after 86 tools/commands, 13 assistant calls, and 108.82 wall
  seconds.
- No new surface is premature because the combined relationship plus record
  lookup need remains real, and a normal user would reasonably expect
  OpenClerk to produce one safe relationship-record answer without a manually
  stitched graph and records ceremony.

## Safety, Capability, UX

Safety pass: pass. Existing evidence preserved canonical markdown authority,
`document_links`, incoming backlinks, `graph_neighborhood`, graph projection
freshness, generic `records_lookup`, `record_entity`, source citations, entity
provenance, records projection freshness, local-first runner-only access,
validation controls, no-bypass boundaries, and no durable-write shortcut.

Capability pass: pass for current primitives. The `oc-oowv` scripted control
completed with classification `none`, and current `openclerk document` plus
`openclerk retrieval` primitives can express the combined workflow safely.

UX quality: not acceptable enough to stop at reference pressure. The
`oc-oowv` natural row failed with `ergonomics_gap` despite passing safety and
capability checks. The scripted control also remained high-touch at 34
tools/commands, 5 assistant calls, and 60.75 wall seconds.

## Follow-Up

File one follow-up Bead for targeted eval and promotion evidence for the
selected narrow relationship-record lookup candidate. Do not file an
implementation Bead.

Follow-up `oc-t7ob` must compare the selected candidate against current
primitives and guidance-only repair, then either promote an exact
request/response contract, defer, kill, or record `none viable yet`. Any later
promotion decision must name the exact response fields, compatibility
expectations, failure modes, validation behavior, and gates.

## Compatibility

Existing behavior remains unchanged:

- `openclerk document` and `openclerk retrieval` remain the public relationship
  and record lookup surfaces.
- Canonical markdown remains authority for relationship wording and promoted
  record facts.
- Graph state and record projections remain derived evidence, not independent
  truth surfaces.
- Relationship-record candidate work remains read-only unless a later
  promotion decision explicitly authorizes durable write behavior.
- Committed evidence must continue to use repo-relative paths or neutral
  placeholders such as `<run-root>`.
