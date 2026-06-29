---
decision_id: decision-compile-synthesis-workflow-action-surface-comparison
decision_title: Compile synthesis workflow-action surface comparison
decision_status: accepted
decision_scope: compile-synthesis-workflow-action
decision_owner: platform
decision_date: 2026-06-29
source_refs: docs/evals/results/ockp-compile-synthesis-workflow-action.md, docs/architecture/synthesis-compile-revisit-adr.md
---
# Compile synthesis workflow-action surface comparison

## Context

Issue #30 tracks the remaining taste debt from the `compile_synthesis`
workflow-action lane: the runner action safely updates the existing synthesis
target, but the natural row previously required extra workflow ceremony before
the action was invoked.

The need remains valid. Routine source-linked synthesis should be one approved
runner call followed by a final answer from runner-returned evidence. It should
not move a long recipe into `skills/openclerk/SKILL.md`, broaden source
authority, or create a general synthesis engine.

## Candidate Comparison

| Candidate | Safety | Capability | UX quality | Decision |
| --- | --- | --- | --- | --- |
| Keep current `compile_synthesis` action plus compact help | Pass: preserves approval-before-write, canonical markdown/source authority, provenance, freshness, duplicate prevention, and runner-only access. | Pass: existing action updates the target and returns evidence. | Taste debt: prior natural row still had pre-action ceremony. | Reject as incomplete. |
| Improve runner handoff/help so natural requests route to one `compile_synthesis` call | Pass: changes only help text, handoff placement, and response evidence; write behavior is unchanged. | Pass: keeps the existing request shape and durable update path. | Selected: makes the stdin JSON route and final answer handoff explicit without skill bloat. | Promote. |
| Expand request/response schema around source evidence, duplicate status, provenance, freshness, and write status | Pass if the added response fields remain derived from the runner write. | Pass: useful as response sugar over existing evidence. | Partially useful: `final_answer` reduces reporting ceremony, but a broader schema expansion is not needed now. | Combine compact `final_answer` into selected shape. |

## Decision

Promote the compact handoff and route-clarity shape:

- keep `compile_synthesis` as the single workflow action for approved
  source-linked synthesis create/update;
- expose the compile-synthesis handoff at the document result top level;
- return `compile_synthesis.final_answer` as a compact one-string final-answer
  contract derived from the same handoff evidence;
- document that document actions are selected through stdin JSON, not workflow
  subcommands or `--action` flags.

## Boundaries

Safety pass: approval-before-write, canonical markdown/source authority,
runner-only access, provenance, projection freshness, and duplicate prevention
remain unchanged.

Capability pass: the existing action still performs exactly one create/update
target selection through the runner and still refuses duplicate target paths.

UX quality: the promoted shape targets `workflow_action_acceptable` by making
the one-call route and final answer evidence directly visible.

No schema migration, source authority expansion, broad synthesis engine,
direct vault behavior, HTTP/MCP path, direct SQLite path, or skill recipe bloat
is authorized by this decision.
