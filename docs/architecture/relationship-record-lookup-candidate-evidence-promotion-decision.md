---
decision_id: decision-relationship-record-lookup-candidate-evidence-promotion
decision_title: Relationship-Record Lookup Candidate Evidence Promotion
decision_status: accepted
decision_scope: relationship-record-lookup-candidate-evidence
decision_owner: platform
---
# Decision: Relationship-Record Lookup Candidate Evidence Promotion

## Status

Accepted: defer the narrow relationship-record lookup helper/report candidate
pending guidance or eval repair.

This decision does not add a runner action, schema, migration, storage
behavior, public API, public OpenClerk interface, product behavior, or shipped
skill behavior. It does not authorize implementation work.

Evidence:

- [`docs/evals/relationship-record-lookup-candidate-evidence.md`](../evals/relationship-record-lookup-candidate-evidence.md)
- [`docs/evals/results/ockp-relationship-record-lookup-candidate-evidence.md`](../evals/results/ockp-relationship-record-lookup-candidate-evidence.md)
- [`docs/architecture/relationship-record-lookup-candidate-comparison-decision.md`](relationship-record-lookup-candidate-comparison-decision.md)
- [`docs/evals/results/ockp-high-touch-relationship-record-ceremony.md`](../evals/results/ockp-high-touch-relationship-record-ceremony.md)

## Decision

Do not promote the narrow relationship-record lookup candidate contract from
this evidence. Record `defer_for_guidance_or_eval_repair`.

The eval-only response candidate completed and preserved the expected
relationship-record evidence contract. `oc-3ybv` also repaired the earlier
record-document listing overconstraint: candidate-lane record evidence can be
verified through `records_lookup`, `record_entity`, citations, provenance, and
records projection freshness without requiring a separate `records/policies/`
list path. However, the current-primitives scripted control still failed with
`skill_guidance_or_eval_coverage`, and the guidance-only natural row still
failed with `ergonomics_gap`. This run cannot yet separate candidate value
from remaining answer-posture repair debt.

## Safety, Capability, UX

Safety pass: pass. The targeted run observed no broad repo search, direct
SQLite, direct vault inspection, direct file edits, source-built runner usage,
module-cache inspection, HTTP/MCP bypass, unsupported transport, backend
variant, unsupported action, or durable write in the selected rows. Validation
controls stayed final-answer-only with zero tools, zero command executions,
and one assistant answer each.

Capability pass: pass. Current primitives and the response candidate both
preserved runner-visible relationship-record evidence: canonical markdown
relationship authority, links/backlinks, graph freshness, canonical record
authority, citations, provenance, records freshness, eval-only response
boundaries, and no-bypass controls.

UX quality: not acceptable enough to close the need. The guidance-only natural
row failed with `ergonomics_gap` after 56 tools/commands, 7 assistant calls,
and 66.24 wall seconds. The scripted current-primitives control also required
answer repair after 28 tools/commands, 6 assistant calls, and 71.34 wall
seconds. The response candidate completed after 30 tools/commands, 7
assistant calls, and 49.98 wall seconds, but that pass does not authorize
promotion while the current/guidance answer posture remains unrepaired.

## Follow-Up

No implementation bead is authorized by this decision.

`bd search "relationship-record answer contract guidance repair"` found no
existing follow-up, so non-implementation follow-up `oc-d3j4` was filed to
repair relationship-record answer posture evidence. No implementation bead is
authorized.

Future promotion remains blocked until repaired targeted evidence shows:

- current primitives can safely express the workflow, or records
  `none_viable_yet` if they cannot
- guidance-only natural evidence either passes cleanly, justifying defer, or
  preserves meaningful ergonomics/answer-contract debt
- the eval-only response candidate passes safety and capability without
  hiding provenance, freshness, citations, no-bypass boundaries, or authority
  limits

## Compatibility

Existing behavior remains unchanged:

- `openclerk document` and `openclerk retrieval` remain the public relationship
  and record lookup surfaces.
- Canonical markdown remains authority for relationship wording and promoted
  record facts.
- Graph state and record projections remain derived evidence, not independent
  truth surfaces.
- The candidate response remains eval-only and does not imply an installed
  relationship-record lookup action.
- Committed evidence must continue to use repo-relative paths or neutral
  placeholders such as `<run-root>`.
