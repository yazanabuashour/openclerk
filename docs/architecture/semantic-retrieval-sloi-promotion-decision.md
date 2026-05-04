---
decision_id: decision-semantic-retrieval-sloi-promotion
status: accepted
date: 2026-05-04
bead: oc-sloi
source_refs: docs/evals/results/ockp-semantic-retrieval-local-model-matrix-sloi.md, docs/evals/results/ockp-semantic-search-core-hardening-sloi.md, docs/evals/results/ockp-semantic-recall-local-nomic-embed-text-sloi.md, docs/evals/results/ockp-semantic-recall-local-mxbai-embed-large-sloi.md
---

# Semantic Retrieval `oc-sloi` Promotion Decision

## Decision

Accept explicit core `semantic_search` and keep default
`openclerk retrieval search` lexical.

`nomic-embed-text` is the selected local default model for the explicit mode
because it reached 8/8 hit@3 and 0.938 hybrid MRR with the lowest completed
local runtime in the matrix. `mxbai-embed-large` also clears the recall
threshold and remains a valid explicit model choice. `embeddinggemma` stays
below the recall threshold, and `bge-m3` timed out on this 16GB M1 run.

## Candidate Comparison

| Candidate | Safety | Capability | UX quality | Outcome |
| --- | --- | --- | --- | --- |
| Keep optional module only | Pass | Partial; adapter works, but normal users still manage module/cache/model ceremony | Weak for normal users | Superseded by explicit core mode |
| Explicit core semantic mode | Pass; local Ollama only, citation-bearing, no default ranking change | Pass; `nomic-embed-text` and `mxbai-embed-large` clear threshold | Best balance; explicit recall surface without hidden ranking change | Accepted |
| Default hybrid search | Partial; exact/source lookup risk remains | Pass on recall for two models | Risky because normal `search` users expect exact/source behavior first | Not promoted |
| None viable yet | Not applicable | Not applicable | Not applicable | Rejected; local explicit mode is viable |

## Safety Pass

Pass. The core implementation uses loopback Ollama only, records provider and
cache status, returns citations, writes only rebuildable user-cache artifacts,
and does not commit embedding data. It does not write provider configuration or
fall back to Gemini.

## Capability Pass

Pass for explicit semantic retrieval. The local matrix recorded:

- `nomic-embed-text`: 8/8 hit@3, 0.938 hybrid MRR, 4.92s.
- `mxbai-embed-large`: 8/8 hit@3, 0.938 hybrid MRR, 23.31s.
- `embeddinggemma`: 7/8 hit@3, 0.906 hybrid MRR.
- `bge-m3`: environment blocked by local Ollama timeout.

The freshness probe passed, citation collapse passed, and path/tag/metadata
filters now pass in both the optional adapter and the core explicit mode.

## UX Quality

Accepted for explicit mode, not for default search. A normal user would expect a
simpler surface than an optional module command, so preserving only the adapter
would be taste debt after the local model matrix cleared recall thresholds.

Default search still carries exact/source lookup expectations. Silent hybrid
ranking would be surprising when a user expects lexical matches, source paths,
or metadata-scoped lookups to dominate. The explicit `semantic_search` action
keeps semantic recall available without changing the stable lexical contract.

## Follow-Ups

`oc-sloi.1`, `oc-sloi.2`, and `oc-sloi.3` were created from this decision and
implemented in the same slice:

- `oc-sloi.1`: explicit core semantic retrieval mode.
- `oc-sloi.2`: local semantic cache/index lifecycle.
- `oc-sloi.3`: docs and eval rollout evidence.

No default-search promotion follow-up is ready. A later decision can compare
default hybrid ranking only after repeated exact/source lookup regression
evidence supports changing normal `search`.
