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

## Embedding Provider

| Field | Value |
| --- | --- |
| provider | `ollama` |
| url | `http://localhost:11434` |
| model | `bge-m3` |
| status | `environment_blocked` |
| embedding_dimensions | 0 |
| error_summary | Post "http://localhost:11434/api/embed": context deadline exceeded (Client.Timeout exceeded while awaiting headers) |

## Methods

### `current_lexical_fts`

| Metric | Value |
| --- | ---: |
| status | `completed` |
| hit_at_3 | 7 |
| query_count | 8 |
| mrr | 0.900 |
| raw_duplicate_hits | 0 |
| total_seconds | 0.14 |

Description: Installed OpenClerk SQLite FTS through current Search; no ranking or schema change.

Evidence posture: citation-bearing current lexical baseline; no vector evidence claimed

Validation boundary: uses embedded OpenClerk runtime only; no direct SQLite reads, no raw vault inspection beyond copied eval corpus setup, no default ranking change

| Query | Kind | Expected | Rank | Hit | Top citations |
| --- | --- | --- | ---: | --- | --- |
| `wiki_synthesis` | `concept-recall` | `docs/architecture/agent-knowledge-plane.md` | 1 | `true` | docs/architecture/agent-knowledge-plane.md / Summary lines 3-38; docs/architecture/eval-backed-knowledge-plane-adr.md / Context lines 17-34; docs/architecture/hybrid-retrieval-adr.md / Context lines 12-47 |
| `semantic_retrieval_gap` | `paraphrase` | `docs/architecture/hybrid-retrieval-adr.md` | 5 | `false` | docs/architecture/harness-owned-web-search-fetch-adr.md / Context lines 27-39; docs/architecture/knowledge-configuration-v1-adr.md / `oc-za6.5` POC Decision lines 438-473; docs/architecture/memory-architecture-recall-adr.md / Memory Architecture And Recall ADR lines 10-11 |
| `structured_rows_vs_notes` | `synonym-drift` | `docs/architecture/structured-data-canonical-stores-adr.md` | 1 | `true` | docs/architecture/structured-data-canonical-stores-adr.md / Structured Data And Canonical Stores ADR lines 10-11; docs/architecture/agent-knowledge-plane.md / Records lines 160-187; docs/architecture/eval-backed-knowledge-plane-adr.md / Direction Considered lines 35-55 |
| `checkpoint_not_restore` | `indirect-source` | `docs/architecture/git-lifecycle-version-control-adr.md` | 1 | `true` | docs/architecture/git-lifecycle-version-control-adr.md / Options lines 42-52; docs/architecture/knowledge-configuration-v1-adr.md / Runner Contract lines 145-195; docs/architecture/structured-data-canonical-stores-adr.md / Authority And Approval Boundaries lines 35-53 |
| `search_then_ingest` | `indirect-source` | `docs/architecture/harness-owned-web-search-fetch-adr.md` | 1 | `true` | docs/architecture/harness-owned-web-search-fetch-adr.md / Promoted Candidate lines 51-67; docs/architecture/generalized-artifact-ingestion-adr.md / Decision lines 62-108; docs/architecture/knowledge-configuration-v1-adr.md / `oc-v1ed` Web URL Intake Decision lines 360-399 |
| `ocr_uncertain_artifact` | `concept-recall` | `docs/architecture/generalized-artifact-ingestion-adr.md` | 1 | `true` | docs/architecture/generalized-artifact-ingestion-adr.md / Promotion Gate lines 145-172; docs/architecture/agent-knowledge-plane.md / Agent Knowledge Plane lines 1-2; docs/architecture/artifact-intake-autofiling-tags-fields-adr.md / Decision lines 49-75 |
| `memory_no_hidden_truth` | `paraphrase` | `docs/architecture/memory-architecture-recall-adr.md` | 1 | `true` | docs/architecture/memory-architecture-recall-adr.md / Context lines 12-39; docs/architecture/structured-data-canonical-stores-adr.md / Authority And Approval Boundaries lines 35-53; docs/architecture/hybrid-retrieval-adr.md / Context lines 12-47 |
| `plan_filename_tags` | `synonym-drift` | `docs/architecture/artifact-intake-autofiling-tags-fields-adr.md` | 1 | `true` | docs/architecture/artifact-intake-autofiling-tags-fields-adr.md / ADR: Artifact Intake, Auto-Filing, Tags, and Fields lines 8-9; docs/architecture/generalized-artifact-ingestion-adr.md / ADR: Generalized Artifact Ingestion Direction lines 8-9; docs/architecture/harness-owned-web-search-fetch-adr.md / Options lines 40-50 |

### `local_vector_only`

| Metric | Value |
| --- | ---: |
| status | `environment_blocked` |
| hit_at_3 | 0 |
| query_count | 8 |
| mrr | 0.000 |
| raw_duplicate_hits | 0 |
| total_seconds | 30.15 |

Description: Local Ollama embedding runtime/model was unavailable; semantic evidence is intentionally not faked.

Evidence posture: environment-blocked; rerun with local Ollama and embedding model to produce vector/hybrid evidence

Validation boundary: no provider fallback, no fake vectors, no durable embedding store, no production ranking change; error: Post "http://localhost:11434/api/embed": context deadline exceeded (Client.Timeout exceeded while awaiting headers)

### `local_hybrid_rrf`

| Metric | Value |
| --- | ---: |
| status | `environment_blocked` |
| hit_at_3 | 0 |
| query_count | 8 |
| mrr | 0.000 |
| raw_duplicate_hits | 0 |
| total_seconds | 30.15 |

Description: Local Ollama embedding runtime/model was unavailable; semantic evidence is intentionally not faked.

Evidence posture: environment-blocked; rerun with local Ollama and embedding model to produce vector/hybrid evidence

Validation boundary: no provider fallback, no fake vectors, no durable embedding store, no production ranking change; error: Post "http://localhost:11434/api/embed": context deadline exceeded (Client.Timeout exceeded while awaiting headers)

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
