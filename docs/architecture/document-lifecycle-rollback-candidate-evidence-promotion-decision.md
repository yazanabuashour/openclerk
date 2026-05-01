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
first targeted evidence run needs guidance or eval repair before promotion can
be trusted.

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

Capability pass: pass for the durable rollback operation and the eval-only
candidate response. The current-primitives control restored the target
document but failed answer/runner-shape verification after 20 tools/commands,
6 assistant calls, and 65.66 wall seconds because `list_documents` also used
`sources/history-review/`. The guidance-only natural row restored the target
document but failed the same shape verification after 36 tools/commands, 8
assistant calls, and 33.86 wall seconds. The eval-only response-candidate row
passed with classification `none` after 22 tools/commands, 6 assistant calls,
and 54.65 wall seconds.

UX quality: not acceptable as promotion evidence. A normal user would expect a
simpler lifecycle rollback surface than a 20-36 command ceremony, but the
current evidence cannot separate real ergonomics debt from an eval prompt or
verifier mismatch around source listing. The candidate JSON contract passed,
but promotion requires current primitives to pass cleanly first.

Non-promotion category: need exists, but the evaluated evidence shape needs
repair. The candidate remains viable enough for repaired targeted evidence,
but not for implementation authorization.

## Follow-Up

No implementation bead is authorized by this decision.

`bd search "document lifecycle rollback eval repair"`, `bd search "lifecycle
rollback candidate evidence"`, `bd search "review_lifecycle_rollback"`, and
`bd search "restore path prefix"` found no existing matching follow-up, so
follow-up `oc-cez4.1` was filed to repair the current-primitives and
guidance-only control rows before any future promotion decision.

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
