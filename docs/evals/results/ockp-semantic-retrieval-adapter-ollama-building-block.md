# OpenClerk Semantic Retrieval Adapter Building Block

## Summary

`modules/semantic-retrieval-adapter` was exercised as an optional module over a
small runner-created fixture. It returned `semantic_retrieval_adapter.v1`
citation-bearing results using local Ollama `embeddinggemma`.

## Evidence

| Field | Value |
| --- | --- |
| command | `semantic-retrieval-adapter search` |
| provider | `ollama` |
| model | `embeddinggemma` |
| embedding dimensions | 768 |
| ranking | `hybrid_rrf_vector_lexical` |
| cache status | `rebuilt` |
| cache ref | `user_cache:semantic-retrieval-adapter/<cache-key>.json` |
| cache committed | `false` |
| fixture docs | 2 |
| fixture chunks | 4 |
| hit count | 2 |
| duplicate chunks | 2 |
| top citations | `docs/architecture/lexical.md`; `docs/architecture/hybrid.md` |

## Safety, Capability, UX

Safety pass: yes. The module used read-only OpenClerk runner access, kept the
embedding cache outside the repository, committed only reduced evidence, and
did not change core `openclerk retrieval search`.

Capability pass: yes for optional building-block status. The command produced
local embedding-backed hybrid RRF results with citations and cache rebuild
metadata.

UX quality: acceptable as an optional module. The shape is useful for agentic
composition and future promotion evidence, but normal users should not need to
manage provider/cache mechanics in routine search.

## Boundary

This result does not promote a core runner schema change, durable OpenClerk
vector store, provider config write, committed embedding cache, or default
ranking change.
