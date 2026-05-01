---
decision_id: decision-relationship-record-repeated-ergonomics-checkpoint
decision_title: Relationship-Record Repeated Ergonomics Checkpoint
decision_status: accepted
decision_scope: relationship-record-lookup-candidate-evidence
decision_owner: platform
---
# Decision: Relationship-Record Repeated Ergonomics Checkpoint

## Status

Accepted: close the conditional repeated-ergonomics evidence follow-up with no
new eval run and no implementation work.

This checkpoint does not add a runner action, schema, migration, storage
behavior, public API, public OpenClerk interface, product behavior, shipped
skill behavior, OCKP scenario, or eval harness behavior.

Evidence:

- [`docs/architecture/relationship-record-lookup-candidate-evidence-promotion-decision.md`](relationship-record-lookup-candidate-evidence-promotion-decision.md)
- [`docs/evals/relationship-record-lookup-candidate-evidence.md`](../evals/relationship-record-lookup-candidate-evidence.md)
- [`docs/evals/results/ockp-relationship-record-lookup-candidate-evidence.md`](../evals/results/ockp-relationship-record-lookup-candidate-evidence.md)

## Checkpoint

`oc-hp3m` was conditional on a repeated natural-intent ergonomics or
answer-contract debt signal after the repaired `oc-d3j4` evidence. No such new
trigger is present.

The latest accepted result remains
`defer_guidance_only_current_primitives_sufficient`: current primitives,
guidance-only natural intent, the eval-only response candidate, and validation
controls all passed the repaired targeted lane. Guidance-only current
primitives were sufficient for that pressure, so the candidate contract remains
deferred rather than promoted.

The UX watch item remains visible: the guidance-only natural row used 56
tools/commands, 8 assistant calls, and 68.10 wall seconds. That is taste debt
to monitor, but this checkpoint found no repeated post-repair trigger that
would justify reopening promotion evidence or filing implementation work.

Beads searches before closure, using `--status all` where closed follow-ups
must remain visible, found no newer trigger:

- `bd search "relationship-record repeated ergonomics" --status all` returned
  no issues
- `bd search "relationship record ergonomics" --status all` returned no issues
- `bd search "relationship-record natural intent" --status all` returned no
  issues
- `bd search "relationship-record candidate promotion evidence" --status all`
  returned no issues
- `bd search "relationship-record answer-contract debt" --status all` returned
  no issues

## Decision

Do not run another targeted eval from this checkpoint. Do not file an
implementation bead. Do not promote a relationship-record helper or report
surface.

The valid future trigger remains the one already recorded in the promotion
decision: stronger repeated evidence that natural guidance over current
`openclerk document` and `openclerk retrieval` primitives leaves meaningful
ergonomics or answer-contract debt while a candidate contract preserves safety
and capability.

## Compatibility

Existing behavior remains unchanged:

- `openclerk document` and `openclerk retrieval` remain the public relationship
  and record lookup surfaces.
- The relationship-record lookup candidate remains deferred/reference only.
- Canonical markdown remains authority for relationship wording and promoted
  record facts.
- Graph state and record projections remain derived evidence, not independent
  truth surfaces.
- Committed evidence must continue to use repo-relative paths or neutral
  placeholders such as `<run-root>`.
