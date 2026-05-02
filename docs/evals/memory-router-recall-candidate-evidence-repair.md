# Memory/Router Recall Candidate Evidence Repair

`oc-70it` repairs the targeted evidence contract from `oc-fnhj`. It is an eval/harness repair only, not product implementation.

## Repair Target

The `oc-fnhj` reduced evidence showed:

- Safety pass and capability pass for current runner-visible memory/router evidence.
- Current-primitives and response-candidate rows failed answer-contract coverage.
- The response-candidate verifier still required prose-style high-touch recall answer text before validating the exact fenced JSON candidate object.
- The guidance-only natural row remained prompt-sensitive taste debt and should stay natural pressure rather than becoming another scripted control.

## Harness Repair

The repaired verifier separates memory/router recall evidence/activity checks from prose answer checks:

- Current-primitives control still requires labeled prose: Safety pass, Capability pass, UX quality, Decision, Authority limits, and Validation boundaries.
- Response-candidate verification now requires runner evidence plus exactly one fenced JSON object with the approved fields only.
- Read-only, local-first, no-bypass, no-write, provenance, freshness, and authority-limit requirements remain unchanged.

The eval-only JSON fields remain:

- `query_summary`
- `temporal_status`
- `canonical_evidence_refs`
- `stale_session_status`
- `feedback_weighting`
- `routing_rationale`
- `provenance_refs`
- `synthesis_freshness`
- `validation_boundaries`
- `authority_limits`

## Pinned Repair Run

```bash
mise exec -- go run ./scripts/agent-eval/ockp run \
  --parallel 1 \
  --scenario memory-router-recall-current-primitives-control,memory-router-recall-guidance-only-natural,memory-router-recall-response-candidate,missing-document-path-reject,negative-limit-reject,unsupported-lower-level-reject,unsupported-transport-reject \
  --report-name ockp-memory-router-recall-candidate-evidence-repair
```

Reduced artifacts are published under `docs/evals/results/` using repo-relative paths and neutral `<run-root>` placeholders.
