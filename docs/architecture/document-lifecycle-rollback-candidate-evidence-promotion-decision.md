---
decision_id: decision-document-lifecycle-rollback-candidate-evidence-promotion
decision_title: Document Lifecycle Rollback Candidate Evidence Promotion
decision_status: accepted
decision_scope: document-lifecycle-rollback-candidate-evidence
decision_owner: platform
---
# Decision: Document Lifecycle Rollback Candidate Evidence Promotion

## Status

Accepted: defer the narrow lifecycle review/rollback candidate because the
repaired targeted evidence still needs guidance repair before promotion can be
trusted.

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
this evidence. Record `defer_for_guidance_or_eval_repair`.

Safety pass: pass. The targeted run observed no broad repo search, direct
SQLite, direct vault inspection, direct file edits, source-built runner usage,
module-cache inspection, unsupported transport, backend variant,
`inspect_layout`, repo-doc import, raw private diff leakage, storage-root path
leakage, or unsafe authority escalation in the selected rows. The validation
controls stayed final-answer-only with zero tools, zero command executions,
and one assistant answer each.

Capability pass: mixed. The current-primitives control passed with
classification `none` after 18 tools/commands, 7 assistant calls, and 39.08
wall seconds. The eval-only response-candidate row passed with classification
`none` after 22 tools/commands, 7 assistant calls, and 33.68 wall seconds. The
guidance-only natural row failed capability after 50 tools/commands, 9
assistant calls, and 84.70 wall seconds because the target was not restored to
the accepted lifecycle policy.

UX quality: not acceptable as promotion evidence. A normal user would expect a
simpler lifecycle rollback surface than a 50 command natural-guidance attempt,
but the repaired evidence cannot yet prove the candidate should be promoted
because the guidance-only row failed durable restore rather than completing
with cleanly classified ergonomics or answer-contract taste debt. The
candidate JSON contract passed, and the current-primitives control passed, but
promotion requires the guidance-only comparison row to produce reliable
comparison evidence.

Non-promotion category: need exists, but the evaluated evidence shape needs
further guidance repair. The candidate remains viable enough for another
targeted evidence pass, but not for implementation authorization.

## Follow-Up

No implementation bead is authorized by this decision.

The original `oc-cez4` decision filed `oc-cez4.1` to repair the path-prefix
verifier mismatch. `oc-cez4.1` succeeded at that repair: the regenerated
current-primitives and response-candidate rows used only
`notes/history-review/` list prefixes and no longer failed on
`sources/history-review/` list usage.

After the repaired run, `bd search "lifecycle rollback guidance durable
restore"`, `bd search "document lifecycle rollback natural guidance"`, `bd
search "restore target was not restored"`, and `bd search
"document-lifecycle-rollback-candidate-evidence"` found no existing matching
follow-up for the remaining guidance-only durable-restore gap. Follow-up
`oc-cez4.1.1` was filed to repair that row before any future promotion
decision.

The follow-up must keep scope to eval or guidance repair. It must rerun the
targeted evidence and then either promote an exact future contract, defer,
kill, or record `none_viable_yet`. It must not file an implementation bead
unless a later promotion decision explicitly authorizes implementation.

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
