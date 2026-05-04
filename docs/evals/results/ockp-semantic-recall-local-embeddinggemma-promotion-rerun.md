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
| model | `embeddinggemma` |
| status | `completed` |
| embedding_dimensions | 768 |

## Methods

### `current_lexical_fts`

| Metric | Value |
| --- | ---: |
| status | `completed` |
| hit_at_3 | 7 |
| query_count | 8 |
| mrr | 0.900 |
| raw_duplicate_hits | 0 |
| total_seconds | 0.15 |

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
| status | `completed` |
| hit_at_3 | 7 |
| query_count | 8 |
| mrr | 0.844 |
| raw_duplicate_hits | 736 |
| total_seconds | 11.20 |

Description: Ollama local embedding vector-only chunk ranking.

Evidence posture: real local/offline embedding evidence if Ollama is local and model metadata is recorded

Validation boundary: eval-only in-memory vectors; no durable embedding store, provider call, or production ranking change

| Query | Kind | Expected | Rank | Hit | Top citations |
| --- | --- | --- | ---: | --- | --- |
| `wiki_synthesis` | `concept-recall` | `docs/architecture/agent-knowledge-plane.md` | 1 | `true` | docs/architecture/agent-knowledge-plane.md / Agent Knowledge Plane lines 1-2; docs/architecture/eval-backed-knowledge-plane-adr.md / Invariants lines 91-112; docs/architecture/knowledge-configuration-v1-adr.md / Context lines 20-39 |
| `semantic_retrieval_gap` | `paraphrase` | `docs/architecture/hybrid-retrieval-adr.md` | 4 | `false` | docs/architecture/hybrid-retrieval-promotion-decision.md / UX Quality lines 60-65; docs/architecture/harness-owned-web-search-fetch-adr.md / Non-Goals lines 112-120; docs/architecture/local-first-hybrid-retrieval-implementation-candidate-decision.md / Follow-Up lines 78-97 |
| `structured_rows_vs_notes` | `synonym-drift` | `docs/architecture/structured-data-canonical-stores-adr.md` | 1 | `true` | docs/architecture/structured-data-canonical-stores-adr.md / Projection Versus Independent Store lines 54-68; docs/architecture/generalized-artifact-ingestion-adr.md / Decision lines 62-108; docs/architecture/agent-knowledge-plane.md / Canonical and derived layers lines 82-206 |
| `checkpoint_not_restore` | `indirect-source` | `docs/architecture/git-lifecycle-version-control-adr.md` | 2 | `true` | docs/architecture/structured-data-canonical-stores-adr.md / Authority And Approval Boundaries lines 35-53; docs/architecture/git-lifecycle-version-control-adr.md / Options lines 42-52; docs/architecture/memory-architecture-recall-adr.md / Decision lines 51-66 |
| `search_then_ingest` | `indirect-source` | `docs/architecture/harness-owned-web-search-fetch-adr.md` | 1 | `true` | docs/architecture/harness-owned-web-search-fetch-adr.md / Safety Constraints lines 68-81; docs/architecture/generalized-artifact-ingestion-adr.md / Decision lines 62-108; docs/architecture/hybrid-retrieval-adr.md / Authority And Approval Boundaries lines 48-66 |
| `ocr_uncertain_artifact` | `concept-recall` | `docs/architecture/generalized-artifact-ingestion-adr.md` | 1 | `true` | docs/architecture/generalized-artifact-ingestion-adr.md / Decision lines 62-108; docs/architecture/artifact-intake-autofiling-tags-fields-adr.md / Confidence Policy lines 110-119; docs/architecture/structured-data-canonical-stores-adr.md / Authority And Approval Boundaries lines 35-53 |
| `memory_no_hidden_truth` | `paraphrase` | `docs/architecture/memory-architecture-recall-adr.md` | 1 | `true` | docs/architecture/memory-architecture-recall-adr.md / Candidate Options lines 40-50; docs/architecture/structured-data-canonical-stores-adr.md / Promotion And Kill Criteria lines 98-127; docs/architecture/agent-knowledge-plane.md / Out of scope for this rewrite lines 239-245 |
| `plan_filename_tags` | `synonym-drift` | `docs/architecture/artifact-intake-autofiling-tags-fields-adr.md` | 1 | `true` | docs/architecture/artifact-intake-autofiling-tags-fields-adr.md / Decision lines 49-75; docs/architecture/generalized-artifact-ingestion-adr.md / Decision lines 62-108; docs/architecture/harness-owned-web-search-fetch-adr.md / Safety Constraints lines 68-81 |

### `local_hybrid_rrf`

| Metric | Value |
| --- | ---: |
| status | `completed` |
| hit_at_3 | 7 |
| query_count | 8 |
| mrr | 0.906 |
| raw_duplicate_hits | 736 |
| total_seconds | 11.20 |

Description: RRF fusion over eval current lexical-token score and Ollama local vector chunk ranks.

Evidence posture: real local/offline embedding evidence if Ollama is local and model metadata is recorded

Validation boundary: eval-only in-memory vectors; no durable embedding store, provider call, or production ranking change

| Query | Kind | Expected | Rank | Hit | Top citations |
| --- | --- | --- | ---: | --- | --- |
| `wiki_synthesis` | `concept-recall` | `docs/architecture/agent-knowledge-plane.md` | 1 | `true` | docs/architecture/agent-knowledge-plane.md / Agent Knowledge Plane lines 1-2; docs/architecture/eval-backed-knowledge-plane-adr.md / Context lines 17-34; docs/architecture/knowledge-configuration-v1-adr.md / Context lines 20-39 |
| `semantic_retrieval_gap` | `paraphrase` | `docs/architecture/hybrid-retrieval-adr.md` | 4 | `false` | docs/architecture/harness-owned-web-search-fetch-adr.md / Context lines 27-39; docs/architecture/hybrid-retrieval-promotion-decision.md / UX Quality lines 60-65; docs/architecture/local-first-hybrid-retrieval-implementation-candidate-decision.md / Safety, Capability, UX lines 57-77 |
| `structured_rows_vs_notes` | `synonym-drift` | `docs/architecture/structured-data-canonical-stores-adr.md` | 1 | `true` | docs/architecture/structured-data-canonical-stores-adr.md / Authority And Approval Boundaries lines 35-53; docs/architecture/agent-knowledge-plane.md / Canonical and derived layers lines 82-206; docs/architecture/git-lifecycle-version-control-adr.md / Context lines 27-41 |
| `checkpoint_not_restore` | `indirect-source` | `docs/architecture/git-lifecycle-version-control-adr.md` | 1 | `true` | docs/architecture/git-lifecycle-version-control-adr.md / Options lines 42-52; docs/architecture/structured-data-canonical-stores-adr.md / Authority And Approval Boundaries lines 35-53; docs/architecture/hybrid-retrieval-adr.md / Authority And Approval Boundaries lines 48-66 |
| `search_then_ingest` | `indirect-source` | `docs/architecture/harness-owned-web-search-fetch-adr.md` | 1 | `true` | docs/architecture/harness-owned-web-search-fetch-adr.md / Promoted Candidate lines 51-67; docs/architecture/generalized-artifact-ingestion-adr.md / Decision lines 62-108; docs/architecture/knowledge-configuration-v1-adr.md / `oc-v1ed` Web URL Intake Decision lines 360-399 |
| `ocr_uncertain_artifact` | `concept-recall` | `docs/architecture/generalized-artifact-ingestion-adr.md` | 1 | `true` | docs/architecture/generalized-artifact-ingestion-adr.md / Decision lines 62-108; docs/architecture/artifact-intake-autofiling-tags-fields-adr.md / Non-Goals lines 128-139; docs/architecture/structured-data-canonical-stores-adr.md / Structured Data And Canonical Stores ADR lines 1-11 |
| `memory_no_hidden_truth` | `paraphrase` | `docs/architecture/memory-architecture-recall-adr.md` | 1 | `true` | docs/architecture/memory-architecture-recall-adr.md / Decision lines 51-66; docs/architecture/structured-data-canonical-stores-adr.md / Promotion And Kill Criteria lines 98-127; docs/architecture/agent-knowledge-plane.md / Canonical and derived layers lines 82-206 |
| `plan_filename_tags` | `synonym-drift` | `docs/architecture/artifact-intake-autofiling-tags-fields-adr.md` | 1 | `true` | docs/architecture/artifact-intake-autofiling-tags-fields-adr.md / Decision lines 49-75; docs/architecture/generalized-artifact-ingestion-adr.md / Invariants lines 109-128; docs/architecture/harness-owned-web-search-fetch-adr.md / Options lines 40-50 |

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
| `local-hybrid-poc` | `recorded` | `partial` | `recorded` | `pass_if_hidden_behind_search` | `recorded` | real local Ollama embedding evidence with citations, duplicate counts, and freshness probe | local/offline hybrid evidence is available for promotion/defer decision |
