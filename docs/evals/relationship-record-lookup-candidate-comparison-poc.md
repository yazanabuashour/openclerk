# Relationship-Record Lookup Candidate Comparison POC

## Status

Implemented candidate-comparison framing for `oc-614j`.

This document compares candidate shapes for reducing relationship-record lookup
ceremony after `oc-oowv`. It does not add runner actions, schemas, migrations,
storage behavior, public API behavior, product behavior, or shipped skill
behavior.

Governing evidence:

- [`docs/evals/results/ockp-high-touch-relationship-record-ceremony.md`](results/ockp-high-touch-relationship-record-ceremony.md)
- [`docs/architecture/relationship-record-ceremony-promotion-decision.md`](../architecture/relationship-record-ceremony-promotion-decision.md)
- [`docs/architecture/graph-semantics-revisit-promotion-decision.md`](../architecture/graph-semantics-revisit-promotion-decision.md)
- [`docs/architecture/promoted-record-domain-expansion-promotion-decision.md`](../architecture/promoted-record-domain-expansion-promotion-decision.md)

## Candidate Workflows

| Candidate | Shape | Strength | Risk |
| --- | --- | --- | --- |
| Guidance-only repair | Keep existing `openclerk document` and `openclerk retrieval` calls; repair skill or prompt guidance for combined relationship and record lookup. | No API or response change; preserves all current safety boundaries. | The `oc-oowv` natural row failed after 86 tools/commands and 13 assistant calls, so guidance alone may preserve the surprising ceremony. |
| Narrow relationship-record lookup candidate | Evaluate a future read-only helper/report surface that packages relationship evidence, links/backlinks, graph freshness, record lookup/entity evidence, citations, provenance, records freshness, and authority limits. | Directly targets the user expectation for one routine relationship plus record answer while keeping safety evidence visible. | Must not make graph state or promoted records independent truth, hide provenance/freshness, or silently bypass canonical markdown authority. |
| No new surface after prompt or harness repair | Treat `oc-oowv` as reference pressure and keep all work on existing primitives after prompt or harness adjustment. | Avoids over-promoting from one failed natural row. | Leaves a real UX need unresolved: normal users should not need an 86-step ceremony for combined relationship and policy record lookup. |

## Selected Candidate

Select the narrow relationship-record lookup candidate for future targeted
evidence, not implementation.

The future candidate should evaluate a read-only helper or report surface that
accepts routine relationship-record lookup intent and returns evidence needed
to answer safely without requiring a separate scripted retrieval sequence. A
future response candidate should expose:

- canonical relationship evidence from markdown
- document links and incoming backlinks
- graph neighborhood evidence and graph projection freshness
- generic `records_lookup` and `record_entity` evidence
- source citations, entity provenance, and records projection freshness
- no-bypass, local-first runner-only, and approval-before-write boundaries
- authority limits explaining that canonical markdown remains source authority
  and graph/record projections are derived evidence

The candidate must not promote a semantic-label graph layer, policy-specific
record action, independent graph truth, independent record truth, storage-level
query path, browser/manual acquisition path, or durable write shortcut.

## Evidence Scorecard

| Evidence | Safety | Capability | UX quality |
| --- | --- | --- | --- |
| `oc-oowv` natural row | Passed. No bypass, direct storage, unsupported transport, generated-file inspection, or durable write risk was observed. | Passed. Runner-visible evidence existed and current primitives could express the workflow. | Failed with `ergonomics_gap`: 86 tools/commands, 13 assistant calls, 108.82s, and `answer_repair_needed`. |
| `oc-oowv` scripted control | Passed. Preserved canonical markdown authority, links/backlinks, graph freshness, record citations, provenance, records freshness, and no-bypass boundaries. | Passed with `none`: current `openclerk document` and `openclerk retrieval` expressed the workflow safely. | Still ceremonial: 34 tools/commands, 5 assistant calls, and 60.75s for a scripted control. |
| Prior graph and record decisions | Passed individually for graph semantics and promoted-record lookup reference pressure. | Passed individually through current primitives and generic records. | Prior natural evidence was high-touch and justified combining the lanes for `oc-oowv`. |

## Conclusion

Do not file an implementation bead from this comparison. File targeted
eval/promotion evidence for the selected narrow relationship-record lookup
candidate.

The future eval should compare the selected candidate against current
primitives and guidance-only repair. Promotion remains blocked until evidence
shows the candidate reduces ceremony while preserving canonical markdown
authority, citations, provenance, graph and records projection freshness,
local-first runner-only access, no-bypass controls, approval-before-write, and
validation controls.
