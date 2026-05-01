---
decision_id: decision-document-lifecycle-rollback-candidate-evidence-promotion
decision_title: Document Lifecycle Rollback Candidate Evidence Promotion
decision_status: accepted
decision_scope: document-lifecycle-rollback-candidate-evidence
decision_owner: platform
---
# Decision: Document Lifecycle Rollback Candidate Evidence Promotion

## Status

Accepted: defer the narrow lifecycle review/rollback candidate because
guidance-only current primitives satisfied the repaired targeted evidence.

This decision does not add a runner action, schema, migration, storage
behavior, public API, public OpenClerk interface, product behavior, or shipped
skill behavior. It does not authorize implementation work.

Evidence:

- [`docs/evals/document-lifecycle-rollback-candidate-evidence.md`](../evals/document-lifecycle-rollback-candidate-evidence.md)
- [`docs/evals/results/ockp-document-lifecycle-rollback-candidate-evidence.md`](../evals/results/ockp-document-lifecycle-rollback-candidate-evidence.md)
- [`docs/architecture/document-lifecycle-rollback-candidate-comparison-decision.md`](document-lifecycle-rollback-candidate-comparison-decision.md)
- [`docs/architecture/document-lifecycle-ceremony-promotion-decision.md`](document-lifecycle-ceremony-promotion-decision.md)

## Decision

Do not promote the narrow lifecycle review/rollback candidate contract from
this evidence. Record `defer_guidance_only_current_primitives_sufficient`.

Safety pass: pass. The targeted run observed no broad repo search, direct
SQLite, direct vault inspection, direct file edits, source-built runner usage,
module-cache inspection, unsupported transport, backend variant,
`inspect_layout`, repo-doc import, raw private diff leakage, storage-root path
leakage, or unsafe authority escalation in the selected rows. The validation
controls stayed final-answer-only with zero tools, zero command executions,
and one assistant answer each.

Capability pass: pass. The current-primitives control passed with
classification `none` after 16 tools/commands, 5 assistant calls, and 45.47
wall seconds. The guidance-only natural row passed with classification `none`
after 26 tools/commands, 7 assistant calls, and 35.00 wall seconds. The
eval-only response-candidate row passed with classification `none` after 16
tools/commands, 6 assistant calls, and 27.68 wall seconds.

UX quality: acceptable for this targeted pressure. A normal user would still
prefer a simpler lifecycle rollback surface, but this repaired evidence does
not prove candidate promotion because the guidance-only natural row completed
the safe durable restore with current `openclerk document` and `openclerk
retrieval` primitives. The candidate JSON contract also passed, but it does
not justify promotion while guidance-only current primitives are sufficient in
the same evidence run.

Non-promotion category: no need under this repaired targeted evidence. Keep
the candidate deferred pending stronger repeated ergonomics evidence.

## Follow-Up

No implementation bead is authorized by this decision.

The original `oc-cez4` decision filed `oc-cez4.1` to repair the path-prefix
verifier mismatch. `oc-cez4.1` succeeded at that repair and filed
`oc-cez4.1.1` for the remaining guidance-only durable-restore gap. The
`oc-cez4.1.1` repaired run resolved that gap: all three lifecycle candidate
rows passed, and each used only `notes/history-review/` list prefixes.

No implementation bead is authorized. `bd search "lifecycle rollback repeated
ergonomics"`, `bd search "document lifecycle rollback guidance sufficient"`,
`bd search "review_lifecycle_rollback promotion"`, and `bd search "lifecycle
rollback candidate deferred"` found no existing matching follow-up.
Follow-up `oc-v5k6` was filed to compare lifecycle rollback candidate surfaces
after the guidance-only pass. It must compare current `openclerk document` and
`openclerk retrieval` guidance, an eval-only structured response contract, and
a narrow `review_lifecycle_rollback` runner action if still plausible, then
choose the best shape, combine useful behaviors if appropriate, defer, kill,
or record `none viable yet`. It must not authorize implementation unless a
later promotion decision explicitly does so.

## Compatibility

Existing behavior remains unchanged:

- `openclerk document` and `openclerk retrieval` remain the public document
  lifecycle surfaces.
- `review_lifecycle_rollback` is not implemented and remains only an
  eval-framed candidate name.
- Markdown remains canonical authority; storage-level Git or sync history
  remains outside OpenClerk semantic lifecycle state.
- Source-sensitive lifecycle answers must preserve citations/source refs,
  provenance, freshness, rollback target accuracy, privacy boundaries, and
  no-bypass invariants.
- Committed evidence must continue to use repo-relative paths or neutral
  placeholders such as `<run-root>`.
