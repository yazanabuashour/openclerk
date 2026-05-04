# Semantic Recall Hybrid Vector Prototype POC

## Scope

This POC supports `oc-rlg7` and `oc-ye6w`. It tests whether current lexical FTS
misses normal semantic recall questions badly enough to justify a later
local-first hybrid/vector retrieval implementation track.

Required references:

- docs/architecture/agent-knowledge-plane.md
- https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md
- https://mitchellh.com/writing/building-block-economy
- https://developers.openai.com/api/docs/guides/prompt-guidance
- https://openai.com/index/harness-engineering/
- https://developers.openai.com/api/docs/guides/embeddings
- https://developers.openai.com/api/docs/guides/retrieval
- https://docs.mem0.ai/open-source/overview

## Prototype Shape

The prototype used the installed `openclerk` CLI for the lexical baseline and an
in-memory vector index for semantic ranking. It did not create a durable
embedding store, add a schema, change default ranking, or write production
documents.

Corpus:

- 12 committed architecture documents under `docs/architecture/`
- 100 heading-section chunks
- temporary isolated runner storage at `<run-root>/openclerk.sqlite`
- temporary isolated vault at `<run-root>/vault`

Embedding and ranking:

- provider: Google Gemini API
- embedding model: `gemini-embedding-001`
- dimensions: 3072
- chunk text: title, repo-relative path, heading, and heading-section content
- chunk IDs: deterministic OpenClerk-style hash of doc ID, heading, content,
  and line span
- vector ranking: cosine similarity
- hybrid ranking: reciprocal rank fusion over lexical and vector chunk ranks
- metric ranking: collapse raw chunk hits to the strongest cited chunk per
  document before hit@3 and MRR

Query set:

- paraphrase, synonym drift, concept recall, and indirect-source lookup
- eight natural-language queries with expected committed architecture docs
- no exact-keyword tuning or generated corpus documents

Rate-limit posture:

- serial embedding calls
- capped retries
- exponential backoff with jitter
- retry counts and backoff totals recorded
- successful work cached only in temporary run storage

## Checks

| Check | Method |
| --- | --- |
| Lexical baseline | Installed `openclerk retrieval` `search` against the isolated corpus. |
| Vector-only recall | In-memory chunk vectors over the same committed docs and chunk citation shape. |
| Hybrid recall | Reciprocal rank fusion of lexical and vector chunk ranks. |
| Citation correctness | Top hits retain repo-relative path, stable chunk ID, heading, and line span. |
| Freshness | Prototype records content hashes and performs a stale-hash mismatch probe. |
| Duplicate behavior | Raw top-10 duplicate document pressure is counted, then metrics collapse by document. |
| Offline/local-first | Embedding creation requires network and provider text transfer; search after embeddings is local. |
| Approval boundary | No durable vector store, background indexing, provider config, or default-ranking change is authorized. |

## Taste Review

A normal OpenClerk user should not choose between FTS, embeddings, local vector
indexes, hosted vector stores, and memory systems before asking a
source-grounded question. The natural surface should remain `search`; any
future hybrid behavior has to hide index operations while preserving citations,
freshness, duplicate handling, and canonical markdown authority.

## Closure

The measured result in
`docs/evals/results/ockp-semantic-recall-hybrid-vector-prototype.md` shows a
real semantic recall gap. It also shows that the tested Gemini-backed path is
not acceptable as a default local-first implementation because embedding
generation requires network access and sends committed chunk/query text to an
external provider.

Remaining implementation-candidate work is represented by deferred follow-up
bead `oc-9ijx`.
