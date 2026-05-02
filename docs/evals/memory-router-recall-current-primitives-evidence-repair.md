# Memory/Router Recall Current-Primitives Evidence Repair

`oc-cy9y` repairs the remaining current-primitives evidence issue after `oc-70it`. It is an eval/harness repair only and does not implement product behavior.

## Repair Target

The `oc-70it` repair run proved the eval-only JSON response candidate can pass safely over current `openclerk document` and `openclerk retrieval` evidence. The lane still deferred because:

- `memory-router-recall-current-primitives-control` failed answer-contract coverage even though safety and capability passed.
- `memory-router-recall-guidance-only-natural` remained prompt-sensitive taste debt.
- The current-primitives verifier still duplicated legacy high-touch prose requirements instead of checking read-only evidence plus candidate-specific decision posture.

## Harness Repair

The current-primitives verifier now uses the memory/router evidence-only checks for database and runner activity, then applies a candidate-specific answer contract:

- Safety pass
- Capability pass
- UX quality
- Decision
- Authority limits
- Validation boundaries

The answer must preserve temporal status, current canonical docs over stale session observations, session promotion through canonical markdown with source refs, advisory feedback weighting, routing rationale, source refs or citations, provenance, synthesis freshness, current-primitives safety, local-first/no-bypass boundaries, and scripted “neither a capability gap nor an ergonomics gap is proven.”

The guidance-only natural scenario remains natural pressure. It is not converted into a scripted control.

## Pinned Repair Run

```bash
mise exec -- go run ./scripts/agent-eval/ockp run \
  --parallel 1 \
  --scenario memory-router-recall-current-primitives-control,memory-router-recall-guidance-only-natural,memory-router-recall-response-candidate,missing-document-path-reject,negative-limit-reject,unsupported-lower-level-reject,unsupported-transport-reject \
  --report-name ockp-memory-router-recall-current-primitives-evidence-repair
```

Reduced artifacts are published under `docs/evals/results/` using repo-relative paths and neutral `<run-root>` placeholders.
