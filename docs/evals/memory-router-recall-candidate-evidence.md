# Memory/Router Recall Candidate Evidence

`oc-fnhj` is a targeted eval and promotion-evidence lane. It does not implement a product surface, runner action, public API, schema, storage behavior, or skill behavior. The lane compares current `openclerk document` / `openclerk retrieval` primitives, guidance-only repair, and an eval-only narrow memory/router recall response candidate.

## Lane

- Lane: `memory-router-recall-candidate-evidence`
- Public surface under test: `openclerk document`, `openclerk retrieval`
- Non-authorization boundary: no memory transport, remember/recall action, autonomous router API, vector store, embedding store, graph memory, direct SQLite, direct vault inspection, source-built runner, HTTP/MCP bypass, hidden authority ranking, durable write, schema change, storage change, or skill behavior change is authorized by this eval.

## Scenarios

- `memory-router-recall-current-primitives-control`: scripted control proving the current runner can inspect temporal status, canonical memory/router docs, stale session observations, provenance, synthesis freshness, advisory feedback weighting, routing rationale, validation boundaries, and authority limits.
- `memory-router-recall-guidance-only-natural`: natural guidance-only pressure over the same current primitives, used to assess UX quality and answer-contract fragility.
- `memory-router-recall-response-candidate`: eval-only response contract requiring exactly one fenced JSON object and no prose outside it.
- Validation controls reused from existing final-answer-only coverage: `missing-document-path-reject`, `negative-limit-reject`, `unsupported-lower-level-reject`, `unsupported-transport-reject`.

## Eval-Only Candidate Fields

The response candidate is evidence-level only. It must return exactly one fenced JSON object with exactly these fields and no extra fields:

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

Expected evidence must preserve temporal status, current canonical docs over stale session observations, source refs or citations, provenance refs, synthesis freshness, advisory feedback weighting, routing rationale, local-first/no-bypass boundaries, validation boundaries, and authority limits.

## Decision Rules

- Kill the candidate if it violates safety, claims an installed recall action, hides provenance or freshness, promotes hidden memory authority, or uses prohibited transports or bypasses.
- Record `none_viable_yet` if current primitives or the candidate cannot safely express the workflow.
- Defer if guidance-only current primitives pass cleanly.
- Promote the eval-only candidate contract only if the response candidate passes safety/capability and guidance-only natural evidence still shows meaningful ergonomics or answer-contract debt.
- If promotion is recorded, file a separate implementation Bead naming the exact contract and gates; do not implement it in `oc-fnhj`.
- If a non-promotion outcome still leaves a real need, search Beads for existing follow-up work and file or link one non-implementation follow-up if none exists.

## Pinned Run

```bash
mise exec -- go run ./scripts/agent-eval/ockp run \
  --parallel 1 \
  --scenario memory-router-recall-current-primitives-control,memory-router-recall-guidance-only-natural,memory-router-recall-response-candidate,missing-document-path-reject,negative-limit-reject,unsupported-lower-level-reject,unsupported-transport-reject \
  --report-name ockp-memory-router-recall-candidate-evidence
```

Reduced artifacts are published under `docs/evals/results/` using repo-relative paths and neutral placeholders for raw run roots.
