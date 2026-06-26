---
decision_id: decision-local-embeddinggemma-hybrid-retrieval-promotion
decision_title: Local EmbeddingGemma Hybrid Retrieval Promotion Decision
decision_status: accepted
decision_scope: local-offline-hybrid-retrieval
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/evals/results/ockp-semantic-recall-local-embeddinggemma-m1.md, docs/evals/results/ockp-semantic-retrieval-adapter-ollama-building-block.md, docs/architecture/local-first-hybrid-retrieval-implementation-candidate-decision.md
---
# Decision: Local EmbeddingGemma Hybrid Retrieval Promotion

## Status

Accepted for `oc-bq8c`: local/offline Ollama `embeddinggemma` is viable as the
next retrieval building block. Do not promote vector or hybrid ranking into
default `openclerk retrieval search` yet.

Evidence:

- [`docs/evals/results/ockp-semantic-recall-local-embeddinggemma-m1.md`](../evals/results/ockp-semantic-recall-local-embeddinggemma-m1.md)
- [`docs/evals/results/ockp-semantic-retrieval-adapter-ollama-building-block.md`](../evals/results/ockp-semantic-retrieval-adapter-ollama-building-block.md)

## Decision

Select local/offline embeddings as the near-term semantic retrieval path and
expose that path through an optional module first. The local run used Ollama
`embeddinggemma`, recorded 768-dimensional vectors, and completed on the M1
machine without provider calls.

The semantic-recall result meets the local evidence threshold:

| Method | hit@3 | MRR | Duplicate pressure |
| --- | ---: | ---: | ---: |
| current lexical baseline before fallback | 0/8 | 0.000 | 0 |
| local vector-only | 7/8 | 0.844 | 736 |
| local hybrid RRF | 7/8 | 0.906 | 736 |

Freshness evidence passed: a content-hash mismatch on the copied eval corpus
identified stale chunks and rebuilt affected chunks. The run committed only
reduced reports with repo-relative citations and `<run-root>` placeholders.

## Safety, Capability, UX

Safety pass: pass for optional-module promotion, partial for default search
promotion. The evidence is local/offline, citation-bearing, and cache/index
state can be rebuilt. Default ranking still needs broader regression evidence
because duplicate pressure is high and vector ranking changes authority order.

Capability pass: pass. Local `embeddinggemma` cleared the requested threshold
of at least 7/8 hit@3 and MRR >= 0.75, with stale-index behavior recorded.

UX quality: pass only if hidden behind a simple search surface later. Users
should not manage model pulls, cache keys, rebuild policy, or provider
fallbacks during routine retrieval. The optional module is acceptable as a
maintainer/agent building block while evidence accumulates.

## Compatibility

This decision does not add a core runner schema change, provider credential
write, durable core vector store, committed embedding cache, or default hybrid
ranking. `openclerk retrieval search` remains lexical and citation-bearing,
with only the separately approved zero-hit lexical fallback.

## Follow-Up

Default semantic or hybrid ranking remains a valid need, but it is not ready to
promote from this work item. Follow-up work must compare optional-module evidence,
local cache/index lifecycle, duplicate collapse, performance, and user-facing
promotion before changing core default search.

Searches performed before closing `oc-bq8c`:

- `follow-up search "semantic hybrid default search promotion" --status all`: no
  existing issue found.
- `follow-up search "semantic retrieval adapter default search" --status all`: no
  existing issue found.

Created follow-up:

- `oc-by5n`: compare semantic retrieval adapter promotion into default search.
