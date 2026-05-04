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
| provider | `gemini` |
| url | `https://generativelanguage.googleapis.com/v1beta` |
| model | `gemini-embedding-001` |
| status | `completed` |
| credential_ref | `runtime_config:GEMINI_API_KEY` |
| embedding_dimensions | 3072 |
| output_dimensions | 3072 |
| request_count | 34 |
| retry_count | 6 |
| backoff_seconds | 51.55 |

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

### `provider_mimic_vector_only`

| Metric | Value |
| --- | ---: |
| status | `completed` |
| hit_at_3 | 8 |
| query_count | 8 |
| mrr | 0.938 |
| raw_duplicate_hits | 736 |
| total_seconds | 60.15 |

Description: Gemini provider-mimic vector-only chunk ranking.

Evidence posture: provider-backed Gemini embedding mimic evidence; useful for vector/hybrid mechanics, not local/offline completion

Validation boundary: eval-only Gemini provider call using redacted runtime_config credential; no durable embedding store, provider config write, or production ranking change

| Query | Kind | Expected | Rank | Hit | Top citations |
| --- | --- | --- | ---: | --- | --- |
| `wiki_synthesis` | `concept-recall` | `docs/architecture/agent-knowledge-plane.md` | 1 | `true` | docs/architecture/agent-knowledge-plane.md / Canonical and derived layers lines 82-206; docs/architecture/eval-backed-knowledge-plane-adr.md / Context lines 17-34; docs/architecture/knowledge-configuration-v1-adr.md / Context lines 20-39 |
| `semantic_retrieval_gap` | `paraphrase` | `docs/architecture/hybrid-retrieval-adr.md` | 1 | `true` | docs/architecture/hybrid-retrieval-adr.md / Candidates lines 67-78; docs/architecture/local-first-hybrid-retrieval-implementation-candidate-decision.md / Decision lines 31-48; docs/architecture/hybrid-retrieval-promotion-decision.md / UX Quality lines 60-65 |
| `structured_rows_vs_notes` | `synonym-drift` | `docs/architecture/structured-data-canonical-stores-adr.md` | 1 | `true` | docs/architecture/structured-data-canonical-stores-adr.md / Projection Versus Independent Store lines 54-68; docs/architecture/knowledge-configuration-v1-adr.md / V1 Concepts lines 72-114; docs/architecture/artifact-intake-autofiling-tags-fields-adr.md / Non-Goals lines 128-139 |
| `checkpoint_not_restore` | `indirect-source` | `docs/architecture/git-lifecycle-version-control-adr.md` | 1 | `true` | docs/architecture/git-lifecycle-version-control-adr.md / Context lines 27-41; docs/architecture/structured-data-canonical-stores-adr.md / Authority And Approval Boundaries lines 35-53; docs/architecture/hybrid-retrieval-adr.md / Authority And Approval Boundaries lines 48-66 |
| `search_then_ingest` | `indirect-source` | `docs/architecture/harness-owned-web-search-fetch-adr.md` | 1 | `true` | docs/architecture/harness-owned-web-search-fetch-adr.md / Promoted Candidate lines 51-67; docs/architecture/knowledge-configuration-v1-adr.md / `oc-v1ed` Web URL Intake Decision lines 360-399; docs/architecture/artifact-intake-autofiling-tags-fields-adr.md / Decision lines 49-75 |
| `ocr_uncertain_artifact` | `concept-recall` | `docs/architecture/generalized-artifact-ingestion-adr.md` | 2 | `true` | docs/architecture/artifact-intake-autofiling-tags-fields-adr.md / Confidence Policy lines 110-119; docs/architecture/generalized-artifact-ingestion-adr.md / Decision lines 62-108; docs/architecture/knowledge-configuration-v1-adr.md / `oc-n31` Contract Decision lines 321-359 |
| `memory_no_hidden_truth` | `paraphrase` | `docs/architecture/memory-architecture-recall-adr.md` | 1 | `true` | docs/architecture/memory-architecture-recall-adr.md / Decision lines 51-66; docs/architecture/eval-backed-knowledge-plane-adr.md / Invariants lines 91-112; docs/architecture/hybrid-retrieval-adr.md / Authority And Approval Boundaries lines 48-66 |
| `plan_filename_tags` | `synonym-drift` | `docs/architecture/artifact-intake-autofiling-tags-fields-adr.md` | 1 | `true` | docs/architecture/artifact-intake-autofiling-tags-fields-adr.md / Override Precedence lines 95-109; docs/architecture/generalized-artifact-ingestion-adr.md / Decision lines 62-108; docs/architecture/harness-owned-web-search-fetch-adr.md / Promotion And Kill Criteria lines 82-111 |

### `provider_mimic_hybrid_rrf`

| Metric | Value |
| --- | ---: |
| status | `completed` |
| hit_at_3 | 8 |
| query_count | 8 |
| mrr | 0.938 |
| raw_duplicate_hits | 736 |
| total_seconds | 60.15 |

Description: RRF fusion over eval current lexical-token score and Gemini provider-mimic vector chunk ranks.

Evidence posture: provider-backed Gemini embedding mimic evidence; useful for vector/hybrid mechanics, not local/offline completion

Validation boundary: eval-only Gemini provider call using redacted runtime_config credential; no durable embedding store, provider config write, or production ranking change

| Query | Kind | Expected | Rank | Hit | Top citations |
| --- | --- | --- | ---: | --- | --- |
| `wiki_synthesis` | `concept-recall` | `docs/architecture/agent-knowledge-plane.md` | 1 | `true` | docs/architecture/agent-knowledge-plane.md / Summary lines 3-38; docs/architecture/eval-backed-knowledge-plane-adr.md / Context lines 17-34; docs/architecture/knowledge-configuration-v1-adr.md / Context lines 20-39 |
| `semantic_retrieval_gap` | `paraphrase` | `docs/architecture/hybrid-retrieval-adr.md` | 2 | `true` | docs/architecture/harness-owned-web-search-fetch-adr.md / Context lines 27-39; docs/architecture/hybrid-retrieval-adr.md / Candidates lines 67-78; docs/architecture/local-first-hybrid-retrieval-implementation-candidate-decision.md / Candidate Comparison lines 49-56 |
| `structured_rows_vs_notes` | `synonym-drift` | `docs/architecture/structured-data-canonical-stores-adr.md` | 1 | `true` | docs/architecture/structured-data-canonical-stores-adr.md / Authority And Approval Boundaries lines 35-53; docs/architecture/agent-knowledge-plane.md / Canonical and derived layers lines 82-206; docs/architecture/knowledge-configuration-v1-adr.md / V1 Concepts lines 72-114 |
| `checkpoint_not_restore` | `indirect-source` | `docs/architecture/git-lifecycle-version-control-adr.md` | 1 | `true` | docs/architecture/git-lifecycle-version-control-adr.md / Context lines 27-41; docs/architecture/structured-data-canonical-stores-adr.md / Authority And Approval Boundaries lines 35-53; docs/architecture/hybrid-retrieval-adr.md / Authority And Approval Boundaries lines 48-66 |
| `search_then_ingest` | `indirect-source` | `docs/architecture/harness-owned-web-search-fetch-adr.md` | 1 | `true` | docs/architecture/harness-owned-web-search-fetch-adr.md / Promoted Candidate lines 51-67; docs/architecture/knowledge-configuration-v1-adr.md / `oc-v1ed` Web URL Intake Decision lines 360-399; docs/architecture/generalized-artifact-ingestion-adr.md / Decision lines 62-108 |
| `ocr_uncertain_artifact` | `concept-recall` | `docs/architecture/generalized-artifact-ingestion-adr.md` | 1 | `true` | docs/architecture/generalized-artifact-ingestion-adr.md / Decision lines 62-108; docs/architecture/artifact-intake-autofiling-tags-fields-adr.md / Non-Goals lines 128-139; docs/architecture/knowledge-configuration-v1-adr.md / `oc-n31` Contract Decision lines 321-359 |
| `memory_no_hidden_truth` | `paraphrase` | `docs/architecture/memory-architecture-recall-adr.md` | 1 | `true` | docs/architecture/memory-architecture-recall-adr.md / Decision lines 51-66; docs/architecture/agent-knowledge-plane.md / Canonical and derived layers lines 82-206; docs/architecture/hybrid-retrieval-adr.md / Authority And Approval Boundaries lines 48-66 |
| `plan_filename_tags` | `synonym-drift` | `docs/architecture/artifact-intake-autofiling-tags-fields-adr.md` | 1 | `true` | docs/architecture/artifact-intake-autofiling-tags-fields-adr.md / Override Precedence lines 95-109; docs/architecture/generalized-artifact-ingestion-adr.md / Invariants lines 109-128; docs/architecture/harness-owned-web-search-fetch-adr.md / Options lines 40-50 |

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
| `provider-mimic-hybrid-poc` | `recorded` | `partial` | `recorded_for_vector_mechanics` | `pass_if_hidden_behind_search` | `recorded_provider_latency_and_retries` | real Gemini provider-mimic embedding evidence with citations, duplicate counts, and freshness probe | provider evidence does not satisfy local/offline oc-bq8c acceptance; rerun with local Ollama remains required |
