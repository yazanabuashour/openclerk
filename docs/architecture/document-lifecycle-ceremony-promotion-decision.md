---
decision_id: decision-document-lifecycle-ceremony-promotion
decision_title: Document Lifecycle Ceremony Promotion
decision_status: accepted
decision_scope: high-touch-document-lifecycle-ceremony
decision_owner: platform
---
# Decision: Document Lifecycle Ceremony Promotion

## Status

Accepted: keep high-touch document lifecycle ceremony as reference evidence.

This decision does not add a runner action, schema, migration, storage
behavior, public API, public OpenClerk interface, or shipped skill behavior.

Evidence:

- [`../evals/high-touch-document-lifecycle-ceremony.md`](../evals/high-touch-document-lifecycle-ceremony.md)
- [`../evals/results/ockp-high-touch-document-lifecycle-ceremony.md`](../evals/results/ockp-high-touch-document-lifecycle-ceremony.md)
- [`../evals/high-touch-successful-workflows-ceremony-eval-design.md`](../evals/high-touch-successful-workflows-ceremony-eval-design.md)
- [`document-lifecycle-promotion-decision.md`](document-lifecycle-promotion-decision.md)

## Decision

Keep document lifecycle ceremony as reference pressure and keep the current
public lifecycle path on:

- `openclerk document`
- `openclerk retrieval`

Safety pass: pass. The targeted run observed no broad repo search, direct
SQLite, direct vault inspection, direct file edits, source-built runner usage,
HTTP/MCP bypass, unsupported transport, backend variant, module-cache
inspection, raw private diff leakage, or storage-root path leakage in the
selected rows. The four validation controls stayed final-answer-only with zero
tools, zero command executions, and one assistant answer each.

Capability pass: pass for current primitives. The scripted control completed
with classification `none` using 22 tools/commands, 7 assistant calls, and
59.79 wall seconds. It preserved restore target accuracy,
`notes/history-review/restore-target.md`, source evidence from
`sources/history-review/restore-authority.md`, provenance inspection,
projection freshness, privacy-safe summaries, and no-bypass boundaries.

UX quality: completed but high-touch. The natural-intent row completed with
classification `none` using 40 tools/commands, 10 assistant calls, and 42.29
wall seconds. A normal user would expect a simpler lifecycle review and
rollback surface than this ceremony, but this targeted run did not show a
capability gap or repeated ergonomics failure. Promotion remains unjustified
without candidate-surface comparison evidence that a simpler shape can preserve
authority, source refs or citations, provenance, projection freshness, rollback
target accuracy, privacy-safe summaries, local-first runner-only access, and
approval-before-write.

Non-promotion category: need exists, candidate comparison required.

## Follow-Up

No implementation bead is authorized by this decision.

The remaining need is real: lifecycle review and rollback completed safely,
but the natural path still required 40 runner/tool steps and 10 assistant
calls. `bd search "document lifecycle"`, `bd search "lifecycle ceremony"`,
`bd search "document history candidate"`, and `bd search "rollback surface"`
found no existing candidate-surface follow-up outside `oc-k8ba`, so follow-up
`oc-awo6` was filed to compare:

- repaired guidance over existing document/retrieval primitives
- a narrow lifecycle review/rollback helper that exposes target, source
  evidence, before/after summary, provenance, projection freshness, and
  privacy/no-diff boundaries
- no new surface after prompt or harness repair

Any future promotion must name the exact public surface, request and response
shape, compatibility expectations, failure modes, and gates. It must preserve
canonical markdown authority, source refs or citations, provenance, projection
freshness, local-first runner-only access, rollback target accuracy,
privacy-safe summaries, no raw private diff leakage, no-bypass controls, and
approval-before-write.

## Compatibility

Existing behavior remains unchanged:

- `openclerk document` and `openclerk retrieval` remain the public routine
  document lifecycle surfaces.
- Markdown remains canonical authority; storage-level Git or sync history
  remains outside OpenClerk semantic lifecycle state.
- Source-sensitive lifecycle answers must preserve citations/source refs,
  provenance, freshness, privacy boundaries, and no-bypass invariants.
- Public reports must use repo-relative paths or neutral placeholders such as
  `<run-root>` and must not expose raw private diffs or private artifact bodies.
- Missing-field and invalid-request handling must continue to preserve the
  final-answer-only validation contract.
