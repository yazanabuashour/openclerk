# OpenClerk Semantic Recall Report

- Lane: `semantic-recall`
- Mode: `local-hybrid`
- Harness: scripts/agent-eval/ockp semantic-recall
- Run root: `<run-root>`
- Raw logs committed: `false`
- Raw content committed: `false`

## Corpus

| Metric | Value |
| --- | ---: |
| documents | 12 |
| chunks | 104 |
| query_rows | 8 |

Chunking policy: eval-only heading-section chunks parsed from committed docs copied into <run-root>; index text includes title, repo-relative path, heading, and section body

Citation policy: reports reduced repo-relative path, heading, and line span citations; canonical markdown remains authority

## Ollama

| Field | Value |
| --- | --- |
| url | `http://localhost:11434` |
| model | `embeddinggemma` |
| status | `environment_blocked` |
| embedding_dimensions | 0 |
| error_summary | Post "http://localhost:11434/api/show": dial tcp [::1]:11434: connect: connection refused |

## Methods

### `current_lexical_fts`

| Metric | Value |
| --- | ---: |
| status | `completed` |
| hit_at_3 | 0 |
| query_count | 8 |
| mrr | 0.000 |
| raw_duplicate_hits | 0 |
| total_seconds | 0.00 |

Description: Installed OpenClerk SQLite FTS through current Search; no ranking or schema change.

Evidence posture: citation-bearing current lexical baseline; no vector evidence claimed

Validation boundary: uses embedded OpenClerk runtime only; no direct SQLite reads, no raw vault inspection beyond copied eval corpus setup, no default ranking change

| Query | Kind | Expected | Rank | Hit | Top citations |
| --- | --- | --- | ---: | --- | --- |
| `wiki_synthesis` | `concept-recall` | `docs/architecture/agent-knowledge-plane.md` | 0 | `false` |  |
| `semantic_retrieval_gap` | `paraphrase` | `docs/architecture/hybrid-retrieval-adr.md` | 0 | `false` |  |
| `structured_rows_vs_notes` | `synonym-drift` | `docs/architecture/structured-data-canonical-stores-adr.md` | 0 | `false` |  |
| `checkpoint_not_restore` | `indirect-source` | `docs/architecture/git-lifecycle-version-control-adr.md` | 0 | `false` |  |
| `search_then_ingest` | `indirect-source` | `docs/architecture/harness-owned-web-search-fetch-adr.md` | 0 | `false` |  |
| `ocr_uncertain_artifact` | `concept-recall` | `docs/architecture/generalized-artifact-ingestion-adr.md` | 0 | `false` |  |
| `memory_no_hidden_truth` | `paraphrase` | `docs/architecture/memory-architecture-recall-adr.md` | 0 | `false` |  |
| `plan_filename_tags` | `synonym-drift` | `docs/architecture/artifact-intake-autofiling-tags-fields-adr.md` | 0 | `false` |  |

### `local_vector_only`

| Metric | Value |
| --- | ---: |
| status | `environment_blocked` |
| hit_at_3 | 0 |
| query_count | 8 |
| mrr | 0.000 |
| raw_duplicate_hits | 0 |
| total_seconds | 0.00 |

Description: Local Ollama embedding runtime/model was unavailable; semantic evidence is intentionally not faked.

Evidence posture: environment-blocked; rerun with local Ollama and embedding model to produce vector/hybrid evidence

Validation boundary: no provider fallback, no fake vectors, no durable embedding store, no production ranking change; error: Post "http://localhost:11434/api/show": dial tcp [::1]:11434: connect: connection refused

### `local_hybrid_rrf`

| Metric | Value |
| --- | ---: |
| status | `environment_blocked` |
| hit_at_3 | 0 |
| query_count | 8 |
| mrr | 0.000 |
| raw_duplicate_hits | 0 |
| total_seconds | 0.00 |

Description: Local Ollama embedding runtime/model was unavailable; semantic evidence is intentionally not faked.

Evidence posture: environment-blocked; rerun with local Ollama and embedding model to produce vector/hybrid evidence

Validation boundary: no provider fallback, no fake vectors, no durable embedding store, no production ranking change; error: Post "http://localhost:11434/api/show": dial tcp [::1]:11434: connect: connection refused

## Freshness Probe

| Field | Value |
| --- | --- |
| status | `completed` |
| changed_path | `docs/architecture/hybrid-retrieval-adr.md` |
| stale_chunks | 1 |
| rebuilt_chunks | 7 |
| seconds | 0.00 |
| evidence_posture | content-hash mismatch on copied <run-root> corpus detects stale local index rows and identifies affected chunks for rebuild |
| validation_boundary | probe mutates only copied eval corpus under <run-root>; no production documents or durable indexes are changed |

## Checks

| Check | Value |
| --- | --- |
| reduced_report_only | `true` |
| raw_logs_committed | `false` |
| raw_content_committed | `false` |
| machine_absolute_artifact_refs | `false` |
| production_search_default_changed | `false` |
| boundary | eval-only maintainer harness; no openclerk document/retrieval JSON schema change, no durable embedding store, no provider embedding default, no production search ranking change |

## Outcomes

| Name | Status | Safety pass | Capability pass | UX quality | Performance | Evidence posture | Details |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `local-hybrid-poc` | `environment_blocked` | `pass` | `not_recorded` | `not_recorded` | `not_recorded` | Ollama local embedding runtime/model unavailable; no fake vectors produced | rerun with local Ollama and embedding model to satisfy oc-bq8c vector evidence |
