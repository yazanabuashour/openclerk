---
decision_id: decision-gemini-provider-mimic-hybrid-retrieval-evidence
decision_title: Gemini Provider-Mimic Hybrid Retrieval Evidence
decision_status: deferred
decision_scope: semantic-recall-provider-mimic-hybrid
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/evals/results/ockp-semantic-recall-gemini-provider-mimic.md, docs/evals/results/ockp-semantic-recall-local-hybrid.md, docs/evals/results/ockp-semantic-recall-hybrid-vector-prototype.md, docs/architecture/local-offline-hybrid-retrieval-environment-blocked-decision.md
---
# Decision: Gemini Provider-Mimic Hybrid Retrieval Evidence

## Status

Deferred for `oc-bq8c` local/offline completion.

The `ockp semantic-recall` harness now supports an eval-only Gemini
provider-mimic embedding path and recorded reduced evidence in
[`docs/evals/results/ockp-semantic-recall-gemini-provider-mimic.md`](../evals/results/ockp-semantic-recall-gemini-provider-mimic.md).

This is real provider-backed vector evidence, but it is not local/offline
embedding evidence. It does not close `oc-bq8c`.

## Decision

Keep `oc-bq8c` deferred until a real local/offline embedding run succeeds.

The Gemini provider-mimic run proves the vector and hybrid mechanics can
recover semantic-recall rows with citation-preserving chunks:

- current lexical FTS: 0/8 hit@3, 0.000 MRR
- Gemini provider-mimic vector-only: 8/8 hit@3, 0.938 MRR
- Gemini provider-mimic hybrid RRF: 8/8 hit@3, 0.938 MRR
- embedding dimensions: 3072
- request count: 34
- retry count: 6
- backoff seconds: 51.55

Do not promote a durable embedding store, default hybrid ranking, provider
configuration, background index, public runner schema, or production search
behavior from this evidence.

Gemini implementation references:

- https://ai.google.dev/gemini-api/docs/embeddings
- https://ai.google.dev/api/embeddings
- https://ai.google.dev/gemini-api/docs/rate-limits

## Safety, Capability, UX

Safety pass: partial. The run used copied committed docs under `<run-root>`,
redacted the credential as `runtime_config:GEMINI_API_KEY`, committed only
reduced reports, and changed no production behavior. It still sent copied
chunk/query text to a remote provider, so it is not offline/local-first.

Capability pass: partial. Provider-backed vector and hybrid ranking recovered
the reduced semantic-recall set and preserved repo-relative citations, but the
acceptance criterion for local/offline embeddings remains unsatisfied.

UX quality: pass for the desired future shape and fail for direct exposure.
Normal users should not manage provider keys, rate limits, retries, or hidden
embedding imports for routine `search`.

## Follow-Up

Search performed before this decision:

- `bd search "local offline hybrid retrieval Ollama rerun" --status all`: no
  separate issue found.

Remaining local/offline work stays on `oc-bq8c`: rerun with local Ollama or
another approved local embedding model, then record model/version/dimensions,
vector-only and hybrid hit@3/MRR, citation correctness, duplicate pressure,
freshness, import/rebuild timing, privacy/offline fit, and final outcome.

## Compatibility

Existing behavior remains unchanged:

- `openclerk retrieval search` remains lexical and citation-bearing.
- `hybrid_retrieval_report` remains read-only decision support.
- No provider API key or raw content is committed.
- The Gemini key is read only from runtime config and is never written back.
