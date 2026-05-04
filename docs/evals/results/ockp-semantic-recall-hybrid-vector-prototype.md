# Semantic Recall Hybrid Vector Prototype Eval

Date: 2026-05-04

## Scenario

Compare current OpenClerk lexical FTS against a real embedding-backed
chunk-level vector prototype and a hybrid reciprocal-rank-fusion prototype for
paraphrase, synonym-drift, concept-recall, and indirect-source questions.

Required references:

- docs/architecture/agent-knowledge-plane.md
- https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md
- https://mitchellh.com/writing/building-block-economy
- https://developers.openai.com/api/docs/guides/prompt-guidance
- https://openai.com/index/harness-engineering/
- https://developers.openai.com/api/docs/guides/embeddings
- https://developers.openai.com/api/docs/guides/retrieval
- https://docs.mem0.ai/open-source/overview

## Setup

| Field | Value |
| --- | --- |
| Lexical surface | Installed `openclerk retrieval` `search` |
| Corpus storage | `<run-root>/openclerk.sqlite` |
| Vault storage | `<run-root>/vault` |
| Corpus docs | 12 committed docs |
| Chunks | 100 heading-section chunks |
| Queries | 8 |
| Embedding provider | Google Gemini API |
| Embedding model | `gemini-embedding-001` |
| Dimensions | 3072 |
| Vector index | in-memory only |
| Hybrid fusion | reciprocal rank fusion over lexical and vector chunk ranks |

## Summary

| Method | Hit@3 | MRR |
| --- | ---: | ---: |
| Lexical FTS | 0/8 | 0.000 |
| Vector-only | 8/8 | 0.938 |
| Hybrid | 8/8 | 0.938 |

The lexical baseline produced no cited hits for these long natural paraphrase
queries. The vector-only and hybrid prototypes recovered every expected
document in the top three after collapsing raw chunk hits to one best cited
chunk per document.

## Rows

| Query | Kind | Expected | Lexical Rank | Vector Rank | Hybrid Rank | Hybrid Top Citation |
| --- | --- | --- | ---: | ---: | ---: | --- |
| `wiki_synthesis` | concept-recall | `docs/architecture/agent-knowledge-plane.md` | none | 1 | 1 | `docs/architecture/agent-knowledge-plane.md`, `LLM Wiki alignment`, lines 125-138 |
| `semantic_retrieval_gap` | paraphrase | `docs/architecture/hybrid-retrieval-adr.md` | none | 1 | 1 | `docs/architecture/hybrid-retrieval-adr.md`, `Candidates`, lines 67-78 |
| `structured_rows_vs_notes` | synonym-drift | `docs/architecture/structured-data-canonical-stores-adr.md` | none | 1 | 1 | `docs/architecture/structured-data-canonical-stores-adr.md`, `Non-Goals`, lines 89-97 |
| `checkpoint_not_restore` | indirect-source | `docs/architecture/git-lifecycle-version-control-adr.md` | none | 1 | 1 | `docs/architecture/git-lifecycle-version-control-adr.md`, `Options`, lines 42-52 |
| `search_then_ingest` | indirect-source | `docs/architecture/harness-owned-web-search-fetch-adr.md` | none | 1 | 1 | `docs/architecture/harness-owned-web-search-fetch-adr.md`, `Promoted Candidate`, lines 51-67 |
| `ocr_uncertain_artifact` | concept-recall | `docs/architecture/generalized-artifact-ingestion-adr.md` | none | 2 | 2 | `docs/architecture/artifact-intake-autofiling-tags-fields-adr.md`, `Confidence Policy`, lines 110-119 |
| `memory_no_hidden_truth` | paraphrase | `docs/architecture/memory-architecture-recall-adr.md` | none | 1 | 1 | `docs/architecture/memory-architecture-recall-adr.md`, `Decision`, lines 51-66 |
| `plan_filename_tags` | synonym-drift | `docs/architecture/artifact-intake-autofiling-tags-fields-adr.md` | none | 1 | 1 | `docs/architecture/artifact-intake-autofiling-tags-fields-adr.md`, `Context`, lines 21-39 |

## Timing And Rate Limits

| Measurement | Value |
| --- | ---: |
| Runner init | 0.05s |
| Document import | 0.22s |
| Lexical search total | 0.20s |
| Embedding elapsed | 49.62s |
| Vector query search | 7.23s |
| Total measured | 57.31s |
| Embedding requests | 117 |
| Cache hits | 0 |
| Retries | 9 |
| Backoff total | 22.22s |
| Rate-limit failures | 0 |

## Freshness, Duplicate, And Citation Notes

- Freshness: content hashes were recorded; a one-document stale-hash probe was
  detected and would require re-embedding 12 chunks.
- Duplicate behavior: raw top-10 duplicate document hits across all queries
  were `0` for lexical, `59` for vector-only, and `59` for hybrid. The metric
  policy collapses to the strongest cited chunk per document before hit@3 and
  MRR.
- Citation integrity: vector and hybrid hits carried repo-relative path,
  stable chunk ID, heading, and line span. No lexical overlap existed in this
  query set because lexical returned no hits.

## Safety Pass

Pass for evidence only. The run used an isolated temp runner database and vault,
committed repo docs, temporary embedding cache, no production document writes,
no durable embedding store, no provider configuration write, and no default
ranking change.

## Capability Pass

Pass for proving a real semantic recall gap and for validating that
citation-preserving chunk vectors can recover expected docs on this reduced
corpus. Partial pass for product readiness: the prototype does not prove
offline local-first embedding generation, durable index storage, large-corpus
rebuild cost, or production stale-index behavior.

## UX Quality

Pass for the desired user shape and fail for the exposed prototype shape. The
result supports a simpler future `search` surface, but the current tested path
requires provider credentials, network latency, retries, and hidden embedding
operations that should not be exposed directly to normal users.

## Privacy And Offline Fit

The vector prototype is not offline during embedding generation. It sends
committed corpus chunk text and query text to the Gemini embedding provider.
After embeddings exist, vector and hybrid scoring are local and in-memory.

## Decision

The result justifies continuing hybrid/vector retrieval design, but not
promoting a durable embedding store or default hybrid ranking. The next work is
deferred follow-up bead `oc-9ijx`, which must compare local/offline embeddings,
explicit opt-in provider embeddings, and lexical-tuning fallback candidates
before implementation.
