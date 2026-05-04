# OpenClerk Semantic Retrieval Promotion Comparison

## Summary

`oc-by5n` compared the optional semantic retrieval adapter against promotion
thresholds for default `openclerk retrieval search`. The outcome is to keep
`modules/semantic-retrieval-adapter` as an optional module and not add a core
semantic mode yet.

## Evidence

| Candidate | Evidence | hit@3 | MRR | Time | Provider | Outcome |
| --- | --- | ---: | ---: | ---: | --- | --- |
| Current lexical `search` | `docs/evals/results/ockp-semantic-recall-lexical-fallback-promotion-rerun.md` | 7/8 | 0.900 | 0.14s | none | keep as default |
| Local Ollama hybrid | `docs/evals/results/ockp-semantic-recall-local-embeddinggemma-promotion-rerun.md` | 7/8 | 0.906 | 11.20s | none | do not promote |
| Gemini benchmark hybrid | `docs/evals/results/ockp-semantic-recall-gemini-promotion-benchmark.md` | 8/8 | 0.938 | 59.88s | `runtime_config:GEMINI_API_KEY` | benchmark only |

Local Ollama `embeddinggemma` recorded 768-dimensional vectors and completed
the stale-index probe, but it still missed the known semantic retrieval gap and
kept high raw duplicate pressure before document collapse.

Gemini reached 8/8, but it required an explicit provider call, 34 requests, 6
retries, and 51.55s of backoff. It is useful benchmark evidence, not a default
local/offline search basis.

## Source-Sensitive Checks

| Check | Result |
| --- | --- |
| path-prefix filtering | pass |
| citation fields | pass |
| document collapse | pass |
| cache hit | pass |
| stale cache rebuild | pass |
| provider blocked without implicit Gemini | pass |
| Gemini credential disclosure | pass |
| empty corpus reporting | pass |
| tag filtering | not supported by adapter request |
| metadata filtering | not supported by adapter request |

## Safety, Capability, UX

Safety pass: pass for keeping the optional module. No core runner schema,
provider config, committed cache, durable vector store, or default ranking
changed.

Capability pass: partial. Local/offline semantic retrieval is useful and
citation-bearing, but the promotion threshold requires 8/8 hit@3 and the local
run remains at 7/8.

UX quality: pass for agent/maintainer usage, not for normal-user default
promotion. A normal user should not manage model pulls, cache mechanics,
provider fallback, or filter differences during routine search.

## Outcome

Keep `modules/semantic-retrieval-adapter` optional. Do not promote local hybrid
ranking into default search, and do not add an explicit core semantic mode in
`oc-by5n`.

Deferred follow-up: `oc-sloi`.

