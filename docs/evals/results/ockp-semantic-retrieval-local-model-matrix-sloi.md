# Semantic Retrieval Local Model Matrix For `oc-sloi`

## Summary

`oc-sloi` re-ran the semantic retrieval promotion evidence across a fixed local
Ollama model matrix. Two local models cleared the recall threshold, but default
hybrid search is still not promoted because exact/source lookup expectations
and normal-user ranking risk remain better served by keeping default
`openclerk retrieval search` lexical.

Outcome: implement an explicit core `semantic_search` mode backed by local
Ollama and a rebuildable local cache. Keep the optional
`modules/semantic-retrieval-adapter` as a module surface and Gemini as
benchmark-only evidence.

## Matrix

| Model | Status | Dims | Hybrid hit@3 | Hybrid MRR | Time | Outcome |
| --- | --- | ---: | ---: | ---: | ---: | --- |
| `embeddinggemma` | completed | 768 | 7/8 | 0.906 | 11.44s | Below recall threshold |
| `mxbai-embed-large` | completed | 1024 | 8/8 | 0.938 | 23.31s | Clears explicit-mode threshold |
| `bge-m3` | environment_blocked | n/a | 0/8 | 0.000 | 30.15s | Timed out on local `/api/embed` |
| `nomic-embed-text` | completed | 768 | 8/8 | 0.938 | 4.92s | Selected local model |

All completed local runs preserved reduced repo-relative citations and used the
same eight query rows from the semantic-recall pressure set. The freshness probe
detected a copied-corpus content-hash mismatch and rebuilt affected chunks.

## Source-Sensitive Checks

| Check | Result |
| --- | --- |
| citation fields | pass |
| document collapse | pass; raw chunk duplicate pressure remains high before collapse |
| stale rebuild | pass |
| path-prefix filter | pass |
| tag filter | pass after adapter/core hardening |
| metadata filter | pass after adapter/core hardening |
| provider blocked state | pass |
| empty corpus state | implemented in core `semantic_search` |
| default search regression | pass; default `search` remains lexical |

## Safety, Capability, UX

Safety pass: pass for explicit local mode. The implemented core mode uses only
loopback Ollama, writes no provider configuration, commits no embedding cache,
and does not change default search ranking.

Capability pass: pass for explicit mode. `mxbai-embed-large` and
`nomic-embed-text` both reached 8/8 hit@3 and 0.938 hybrid MRR. `bge-m3` is
blocked on this machine, and `embeddinggemma` remains below the threshold.

UX quality: explicit mode is the best fit. A normal user should not have default
exact/source lookup behavior silently mixed with semantic ranking, but an
explicit `semantic_search` action is simpler than managing an optional module
command for routine semantic recall.

## Evidence References

- `docs/evals/results/ockp-semantic-recall-local-embeddinggemma-sloi.md`
- `docs/evals/results/ockp-semantic-recall-local-mxbai-embed-large-sloi.md`
- `docs/evals/results/ockp-semantic-recall-local-bge-m3-sloi.md`
- `docs/evals/results/ockp-semantic-recall-local-nomic-embed-text-sloi.md`
- `docs/evals/results/ockp-semantic-recall-gemini-promotion-benchmark.md`
