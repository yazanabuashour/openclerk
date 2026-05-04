---
decision_id: decision-local-offline-hybrid-retrieval-environment-blocked
decision_title: Local Offline Hybrid Retrieval Environment-Blocked Evidence
decision_status: deferred
decision_scope: semantic-recall-local-hybrid
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/evals/results/ockp-semantic-recall-local-hybrid.md, docs/evals/results/ockp-semantic-recall-hybrid-vector-prototype.md, docs/architecture/local-first-hybrid-retrieval-implementation-candidate-decision.md
---
# Decision: Local Offline Hybrid Retrieval Environment-Blocked Evidence

## Status

Deferred: the eval harness exists, but `oc-bq8c` does not yet have real
local/offline embedding evidence because local Ollama was not reachable in
this run.

The `oc-bq8c` implementation pass added `ockp semantic-recall --mode
local-hybrid` and recorded a reduced environment-blocked report in
[`docs/evals/results/ockp-semantic-recall-local-hybrid.md`](../evals/results/ockp-semantic-recall-local-hybrid.md).

No fake vectors, provider fallback, durable embedding store, provider
configuration, background index, or default ranking change was introduced.

## Decision

Keep `oc-bq8c` open/deferred until it can be rerun with a real local Ollama
embedding model. The harness is ready to call Ollama `/api/show` for model
metadata and `/api/embed` for vectors, then compute local vector-only and
hybrid RRF rankings over the same semantic-recall query set.

The current run records only:

- environment-blocked Ollama status for `embeddinggemma`
- no vector-only or hybrid recall metrics
- no embedding dimensions
- copied-corpus stale-index detection for `docs/architecture/hybrid-retrieval-adr.md`
- reduced reports with no raw logs, raw content, machine-absolute paths, or
  production changes

## Safety, Capability, UX

Safety pass: pass for the blocked run. It did not call a provider, did not
fake vectors, did not create a durable index, did not change production
ranking, and mutated only copied eval corpus files under `<run-root>` for the
freshness probe.

Capability pass: not recorded. The acceptance criteria require comparing at
least one local/offline embedding model against current lexical FTS. This
environment could not provide that evidence.

UX quality: not recorded for local hybrid. The intended UX remains plain
`search` with hidden model/index mechanics, but no local model quality,
latency, or rebuild-cost evidence exists yet.

## Freshness Evidence

The copied-corpus stale-index probe completed. It appended an eval-only marker
to `docs/architecture/hybrid-retrieval-adr.md` under `<run-root>`, detected a
content-hash mismatch, and identified affected chunks for rebuild. This proves
the POC harness can record stale-index invalidation mechanics without touching
production docs or a durable index.

## Follow-Up

Searches performed before this deferral:

- `bd search "Ollama semantic recall local hybrid" --status all`: no existing
  issue found.
- `bd search "local offline hybrid retrieval rerun" --status all`: no
  existing issue found.

Remaining work stays on `oc-bq8c`: rerun the harness with local Ollama and an
available embedding model, then record model/version/dimensions, vector-only
and hybrid hit@3/MRR, citation correctness, duplicate pressure, freshness,
import/rebuild timing, privacy/offline fit, and final promote/defer/kill
outcome.

## Compatibility

Existing behavior remains unchanged:

- `openclerk retrieval search` remains lexical and citation-bearing.
- `hybrid_retrieval_report` remains read-only decision support.
- The new harness is maintainer/eval tooling only.
- Provider embeddings remain reference evidence only unless a later decision
  explicitly approves an opt-in provider surface.
