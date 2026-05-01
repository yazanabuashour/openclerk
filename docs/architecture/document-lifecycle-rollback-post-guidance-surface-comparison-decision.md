---
decision_id: decision-document-lifecycle-rollback-post-guidance-surface-comparison
decision_title: Document Lifecycle Rollback Post-Guidance Surface Comparison
decision_status: accepted
decision_scope: document-lifecycle-rollback-candidate-evidence
decision_owner: platform
---
# Decision: Document Lifecycle Rollback Post-Guidance Surface Comparison

## Status

Accepted: defer lifecycle rollback surface promotion after comparing the
post-guidance candidate surfaces.

This decision does not add a runner action, schema, migration, storage
behavior, public API, public OpenClerk interface, product behavior, shipped
skill behavior, eval harness behavior, or implementation Bead. It does not
authorize implementation work.

Evidence:

- [`docs/architecture/document-lifecycle-rollback-candidate-evidence-promotion-decision.md`](document-lifecycle-rollback-candidate-evidence-promotion-decision.md)
- [`docs/evals/results/ockp-document-lifecycle-rollback-candidate-evidence.md`](../evals/results/ockp-document-lifecycle-rollback-candidate-evidence.md)
- [`docs/evals/document-lifecycle-rollback-candidate-evidence.md`](../evals/document-lifecycle-rollback-candidate-evidence.md)
- [`docs/architecture/document-lifecycle-rollback-candidate-comparison-decision.md`](document-lifecycle-rollback-candidate-comparison-decision.md)

## Comparison

The repaired targeted evidence changed the decision posture. Earlier lifecycle
rollback comparison selected a narrow future helper for targeted evidence
because natural lifecycle rollback looked too ceremonial. The repaired
candidate evidence then showed all three targeted rows passing, including the
guidance-only current-primitives row. That makes implementation premature.

Candidate surfaces:

| Surface | Safety pass | Capability pass | UX quality | Decision |
| --- | --- | --- | --- | --- |
| Current `openclerk document` and `openclerk retrieval` guidance | Pass: repaired evidence observed no broad repo search, direct SQLite, direct vault inspection, direct file edits, source-built runner usage, unsupported transport, backend variant, module-cache inspection, raw private diff leakage, storage-root path leakage, or authority escalation. | Pass: guidance-only natural intent completed with classification `none`, preserved source authority, provenance/freshness checks, rollback target accuracy, privacy boundaries, write status, and no-bypass boundaries. | Acceptable for this pressure: it still required 26 tools/commands and 7 assistant calls, but it completed safely without proving repeated ergonomics or answer-contract debt. | Keep as the supported surface for now. |
| Eval-only structured response contract | Pass: the candidate row preserved the same safety boundaries and explicitly stated that it did not implement `review_lifecycle_rollback`. | Pass: the assembled object supplied target identity, source evidence, before/after summary, restore reason, provenance refs, projection freshness, write status, privacy boundaries, validation boundaries, and authority limits. | Useful as evidence and a report shape, but too artificial to promote while current primitives plus guidance passed in the same repaired run. | Keep as reference evidence only. |
| Narrow future `review_lifecycle_rollback` runner action | Conditionally plausible only if later evidence proves it can preserve canonical markdown authority, source refs, provenance, freshness, local-first runner-only access, privacy/no-diff boundaries, and no-bypass controls. | Not proven beyond the eval-only assembled response. No installed runner action exists. | A normal user would reasonably expect a simpler lifecycle rollback surface, but the repaired guidance-only pass does not show enough repeated taste debt to select or promote this action now. | Defer; do not select for implementation. |

## Decision

Record `defer_guidance_only_current_primitives_sufficient`.

Current `openclerk document` and `openclerk retrieval` guidance remains the
best supported lifecycle rollback surface. The eval-only structured response
contract remains useful as reference evidence for future report design, but it
does not justify promotion while guidance-only current primitives completed
the safe workflow. The narrow `review_lifecycle_rollback` runner action remains
unimplemented and unselected for promotion.

Taste check: a normal user would prefer a simpler OpenClerk surface for
reviewing and restoring unsafe accepted lifecycle content. That preference is
real, but this repaired evidence shows acceptable guidance-only behavior for
the targeted pressure. Future action needs stronger repeated natural-intent
evidence showing meaningful ergonomics or answer-contract debt while the
candidate still preserves safety and capability.

## Follow-Up

No implementation Bead is authorized by this decision.

Required Beads searches before closure found no newer matching follow-up:

- `bd search "lifecycle rollback ergonomics" --status all` found no issues
- `bd search "review_lifecycle_rollback" --status all` found no issues
- `bd search "document lifecycle rollback guidance" --status all` found no
  issues

The future trigger remains stronger repeated natural-intent evidence that
current `openclerk document` and `openclerk retrieval` guidance leaves
meaningful ergonomics or answer-contract debt while a candidate surface still
preserves authority, citations or source refs, provenance, freshness,
local-first runner-only access, privacy boundaries, approval-before-write, and
no-bypass invariants. If that trigger appears, open evaluation/design follow-up
first; do not file implementation work until a later accepted promotion
decision names the exact surface and gates.

## Compatibility

Existing behavior remains unchanged:

- `openclerk document` and `openclerk retrieval` remain the public document
  lifecycle surfaces.
- `review_lifecycle_rollback` remains unimplemented and reference-only.
- Markdown remains canonical authority; storage-level Git or sync history
  remains outside OpenClerk semantic lifecycle state.
- Source-sensitive lifecycle answers must preserve citations or source refs,
  provenance, freshness, rollback target accuracy, privacy boundaries, and
  no-bypass invariants.
- Committed evidence must continue to use repo-relative paths or neutral
  placeholders such as `<run-root>`.
