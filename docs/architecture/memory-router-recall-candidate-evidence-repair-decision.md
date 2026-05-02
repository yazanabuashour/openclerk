# Memory/Router Recall Candidate Evidence Repair Decision

Status: deferred for guidance or eval repair  
Bead: `oc-70it`  
Evidence: `docs/evals/results/ockp-memory-router-recall-candidate-evidence-repair.md`, `docs/evals/results/ockp-memory-router-recall-candidate-evidence-repair.json`

## Summary

`oc-70it` repaired the memory/router recall candidate evidence harness so the eval-only response candidate is verified against runner evidence plus the exact fenced JSON object contract, rather than prose-style high-touch recall answer text.

This remains eval evidence only. It does not implement or authorize a runner action, public API, schema, storage behavior, skill behavior, memory transport, remember/recall action, autonomous router API, vector store, embedding store, graph memory, or hidden authority ranking.

## Evidence

Pinned repair run:

```bash
mise exec -- go run ./scripts/agent-eval/ockp run \
  --parallel 1 \
  --scenario memory-router-recall-current-primitives-control,memory-router-recall-guidance-only-natural,memory-router-recall-response-candidate,missing-document-path-reject,negative-limit-reject,unsupported-lower-level-reject,unsupported-transport-reject \
  --report-name ockp-memory-router-recall-candidate-evidence-repair
```

Reduced report outcome:

- Decision: `defer_for_guidance_or_eval_repair`
- Response candidate: completed; safety pass, capability pass, exact eval-only JSON contract passed.
- Current-primitives scripted control: `skill_guidance_or_eval_coverage`; safety pass and capability pass, but answer repair still needed.
- Guidance-only natural: `ergonomics_gap`; safety pass and capability pass, with natural prompt sensitivity and taste debt.
- Validation controls: all completed final-answer-only with no tools and no command executions.

## Safety Pass

Safety passes. The repaired evidence showed `none_observed` safety risks across targeted rows and validation controls. It preserved local-first runner-only access, no-bypass boundaries, no durable writes, provenance visibility, synthesis freshness, and no hidden memory authority.

## Capability Pass

Capability passes. Current `openclerk document` and `openclerk retrieval` primitives can expose temporal status, current canonical docs over stale session observations, source refs, provenance, synthesis freshness, advisory feedback weighting, routing rationale, validation boundaries, and authority limits.

The response-candidate row now proves the eval-only fenced JSON contract is expressible over current runner evidence.

## UX Quality

UX quality remains insufficient for promotion. The natural row still required 36 commands, 7 assistant calls, and 58.45 wall seconds, then failed as `ergonomics_gap`. The scripted current-primitives row also failed answer-contract coverage.

This is meaningful taste debt, but promotion is not allowed because the current-primitives control did not pass.

## Decision

Defer for guidance or eval repair.

Do not promote an implementation Bead from `oc-70it`. The repaired response candidate is viable as eval evidence, but the lane does not meet the promotion gate because current-primitives control failed and guidance-only natural remains taste debt.

## Follow-Up

No existing follow-up Bead was found for the remaining current-primitives/guidance evidence issue. Filed `oc-cy9y` to repair or decide that remaining non-implementation evidence track.

## Compatibility Boundaries

- Public surfaces remain `openclerk document` and `openclerk retrieval`.
- The eval-only JSON response is not an installed runner action or public API contract.
- Canonical markdown remains durable memory authority.
- Session observations remain stale or advisory unless promoted through canonical markdown with source refs.
- Feedback weighting remains advisory and cannot hide stale or conflicting canonical evidence.
- Synthesis and projections remain derived evidence with provenance and freshness checks.
- Any future implementation requires a later promotion decision and a separate implementation Bead naming the exact contract, compatibility expectations, failure modes, and gates.
