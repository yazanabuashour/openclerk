---
decision_id: decision-thin-skill-workflow-surface-comparison
decision_title: Thin Skill Workflow Surface Comparison
decision_status: accepted
decision_scope: thin-skill-workflow-surfaces
decision_owner: platform
decision_date: 2026-05-03
source_refs: docs/evals/results/ockp-capture-save-this-note-candidate.md, docs/evals/results/ockp-capture-duplicate-candidate-update.md, docs/evals/results/ockp-capture-duplicate-candidate-guidance-hardening.md, docs/evals/results/ockp-capture-document-these-links-placement.md, docs/evals/results/ockp-capture-document-these-links-placement-skill-policy.md, docs/evals/results/ockp-document-lifecycle-rollback-candidate-evidence.md, docs/evals/results/ockp-populated-vault-targeted.md, docs/evals/results/ockp-populated-vault-guidance-hardening.md
---
# Decision: Thin Skill Workflow Surface Comparison

## Status

Accepted: resolve the thin-skill workflow surface comparison backlog and
promote the two selected runner-owned surfaces without expanding
`skills/openclerk/SKILL.md`.

This decision closes:

- `oc-phvu`: propose-before-create, save-this-note, and low-risk capture
- `oc-ond7`: duplicate candidate update versus new document
- `oc-80l1`: document-these-links placement
- `oc-sow6`: document lifecycle review and rollback
- `oc-z55m`: populated-vault polluted or decoy evidence handling

This decision records which surface should own each workflow so routine UX is
not repaired by rebuilding long `SKILL.md` recipes. It promotes
`duplicate_candidate_report` and `ingest_source_url` `mode: "plan"` as
runner-owned interfaces. It does not add storage behavior, eval harness
behavior, lifecycle actions, populated-vault-specific actions, or long skill
recipes.

## Comparison

Candidate surfaces:

- A: existing runner primitives plus caller autonomy and compact safety
  constraints
- B: extend an existing natural runner action
- C: add one narrow workflow action with `agent_handoff`

| Workflow | Selected surface | Safety pass | Capability pass | UX quality | Follow-up |
| --- | --- | --- | --- | --- | --- |
| Propose-before-create, save-this-note, and low-risk capture | A: existing primitives plus caller autonomy | Pass. Existing evidence preserved runner-only access, local-first behavior, candidate faithfulness, validation-before-create, duplicate handling, and approval-before-write. | Pass. `validate`, search/list/get, and runner rejections can express faithful candidate proposal and no-write boundaries without a new action. | Acceptable after the thin-skill reset. Candidate path/title/body generation belongs to the agent caller once the runner and safety boundaries are clear. Exact wording expectations beyond safety and faithfulness are eval-contract debt, not durable skill content. | None for a runner surface. File an eval-contract cleanup only if future evals require exact title/body wording instead of safety, faithfulness, and no-write behavior. |
| Duplicate candidate update versus new document | C: new read-only duplicate candidate report action | Pass. Existing runs preserved runner-visible evidence, target accuracy, no-write status, approval-before-write, and no-bypass boundaries. | Pass for current primitives. Search/list/get can identify the likely existing target safely. | Not acceptable as routine primitive choreography. Natural and scripted evidence repeatedly required many runner calls and exact sequencing for a normal update-versus-new clarification. | Implemented by `oc-aw2d` as `duplicate_candidate_report`. |
| Document-these-links placement | B: extend `ingest_source_url` with a planning mode | Pass. Existing evidence preserved public fetch only through the runner, durable-write approval, source refs, citations, duplicate handling, and no-bypass boundaries. | Pass for current primitives. Approved source fetch, source-linked synthesis validation, and duplicate placement inspection can be done today. | Not acceptable as durable skill policy. The request naturally belongs with source intake, and repeated source-path/synthesis-placement choreography is runner UX debt. | Implemented by `oc-7bjj` as `ingest_source_url` `mode: "plan"`. |
| Document lifecycle review and rollback | A/defer: current document and retrieval primitives | Pass. Repaired lifecycle evidence preserved canonical markdown authority, source refs, provenance, projection freshness, rollback target accuracy, privacy boundaries, write status, and no-bypass controls. | Pass. Current primitives completed the guidance-only lifecycle workflow in the repaired evidence. | Defer. A simpler lifecycle surface remains plausible, but the post-guidance evidence did not justify promotion. | None. Keep `review_lifecycle_rollback` reference-only until stronger repeated evidence appears. |
| Populated-vault polluted or decoy evidence handling | A/defer: retrieval primitives plus compact authority policy | Pass. Focused evidence preserved runner-visible authority, metadata filters, citations, `doc_id`, `chunk_id`, local-first operation, and no-bypass boundaries. | Pass. Current retrieval actions can reject polluted and decoy evidence after compact guidance. | Acceptable for this pressure. The original failure was answer handling, then a focused rerun passed without a new action. | None. Do not promote a polluted-vault-specific runner surface from this evidence. |

## Decisions

Keep `skills/openclerk/SKILL.md` as a thin activation, routing, and safety
contract. Do not add long workflow recipes for these lanes.

Use caller autonomy for candidate capture. The agent should derive a faithful
candidate path, title, and body from explicit user content, validate it through
`openclerk document`, inspect runner-visible duplicate evidence when needed,
state no durable write occurred, and ask for approval. The runner does not need
to own ordinary path/title/body creativity for this pass.

Promote only the duplicate-candidate routine to a read-only retrieval workflow
action. `duplicate_candidate_report` returns the likely target, evidence
inspected, no-write status, approval boundary, validation boundaries,
authority limits, and `agent_handoff`. It must not create, update, validate a
new candidate while update-versus-new is unresolved, or hide runner-visible
evidence.

Extend `ingest_source_url` for document-these-links placement planning. The
`plan` mode performs no fetch and no durable write, proposes candidate source
path hints from explicit public URLs, reports duplicate source status, includes
synthesis-placement guidance, preserves public-read versus durable-write
approval, and returns `agent_handoff`.

Defer lifecycle and populated-vault actions. Current `openclerk document` and
`openclerk retrieval` primitives remain the supported surfaces. Future
promotion needs stronger repeated ergonomics, answer-contract, auditability,
or capability evidence while preserving authority, citations/source refs,
provenance, freshness, local-first runner-only access, approval boundaries,
and no-bypass controls.

## Evidence

Primary evidence:

- [`docs/evals/results/ockp-capture-save-this-note-candidate.md`](../evals/results/ockp-capture-save-this-note-candidate.md)
- [`docs/evals/results/ockp-capture-duplicate-candidate-update.md`](../evals/results/ockp-capture-duplicate-candidate-update.md)
- [`docs/evals/results/ockp-capture-duplicate-candidate-guidance-hardening.md`](../evals/results/ockp-capture-duplicate-candidate-guidance-hardening.md)
- [`docs/evals/results/ockp-capture-document-these-links-placement.md`](../evals/results/ockp-capture-document-these-links-placement.md)
- [`docs/evals/results/ockp-capture-document-these-links-placement-skill-policy.md`](../evals/results/ockp-capture-document-these-links-placement-skill-policy.md)
- [`docs/evals/results/ockp-document-lifecycle-rollback-candidate-evidence.md`](../evals/results/ockp-document-lifecycle-rollback-candidate-evidence.md)
- [`docs/evals/results/ockp-populated-vault-targeted.md`](../evals/results/ockp-populated-vault-targeted.md)
- [`docs/evals/results/ockp-populated-vault-guidance-hardening.md`](../evals/results/ockp-populated-vault-guidance-hardening.md)

Decision context:

- [`capture-save-this-note-candidate-promotion-decision.md`](capture-save-this-note-candidate-promotion-decision.md)
- [`capture-duplicate-candidate-update-promotion-decision.md`](capture-duplicate-candidate-update-promotion-decision.md)
- [`capture-document-these-links-placement-promotion-decision.md`](capture-document-these-links-placement-promotion-decision.md)
- [`document-lifecycle-rollback-post-guidance-surface-comparison-decision.md`](document-lifecycle-rollback-post-guidance-surface-comparison-decision.md)

## Compatibility

Existing behavior remains unchanged:

- `openclerk document` and `openclerk retrieval` remain the public routine
  AgentOps surfaces.
- `duplicate_candidate_report` is read-only.
- `ingest_source_url` `mode: "plan"` is read-only; create/update modes remain
  the only source URL modes that fetch or write.
- `skills/openclerk/SKILL.md` remains a thin router; routine workflow detail
  belongs in runner actions, runner help, eval/maintainer docs, or caller
  autonomy.
- Public read/fetch/inspect permission remains separate from durable-write
  approval.
- Committed docs, reports, and artifact references must stay repo-relative or
  use neutral placeholders such as `<run-root>`.
