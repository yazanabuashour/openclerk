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
| status | `provider_blocked` |
| credential_ref | `runtime_config:GEMINI_API_KEY` |
| embedding_dimensions | 3072 |
| output_dimensions | 3072 |
| request_count | 33 |
| retry_count | 6 |
| backoff_seconds | 51.55 |
| error_summary | gemini returned HTTP 429: {   "error": {     "code": 429,     "message": "You exceeded your current quota, please check your plan and billing details. For more information on this error, head to: https://ai.google.dev/gemini-api/docs/rate-l |

## Methods

### `current_lexical_fts`

| Metric | Value |
| --- | ---: |
| status | `completed` |
| hit_at_3 | 7 |
| query_count | 8 |
| mrr | 0.833 |
| raw_duplicate_hits | 0 |
| total_seconds | 0.15 |

Description: Installed OpenClerk SQLite FTS through current Search; no ranking or schema change.

Evidence posture: citation-bearing current lexical baseline; no vector evidence claimed

Validation boundary: uses embedded OpenClerk runtime only; no direct SQLite reads, no raw vault inspection beyond copied eval corpus setup, no default ranking change

| Query | Kind | Expected | Rank | Hit | Top citations |
| --- | --- | --- | ---: | --- | --- |
| `wiki_synthesis` | `concept-recall` | `docs/architecture/agent-knowledge-plane.md` | 1 | `true` | docs/architecture/agent-knowledge-plane.md / Summary lines 3-40; docs/architecture/eval-backed-knowledge-plane-adr.md / Context lines 17-34; docs/architecture/hybrid-retrieval-adr.md / Context lines 12-47 |
| `semantic_retrieval_gap` | `paraphrase` | `docs/architecture/hybrid-retrieval-adr.md` | 6 | `false` | docs/architecture/harness-owned-web-search-fetch-adr.md / Context lines 27-39; docs/architecture/knowledge-configuration-v1-adr.md / `oc-za6.5` POC Decision lines 438-473; docs/architecture/memory-architecture-recall-adr.md / Memory Architecture And Recall ADR lines 10-11 |
| `structured_rows_vs_notes` | `synonym-drift` | `docs/architecture/structured-data-canonical-stores-adr.md` | 1 | `true` | docs/architecture/structured-data-canonical-stores-adr.md / Structured Data And Canonical Stores ADR lines 10-11; docs/architecture/agent-knowledge-plane.md / Summary lines 3-40; docs/architecture/eval-backed-knowledge-plane-adr.md / Direction Considered lines 35-55 |
| `checkpoint_not_restore` | `indirect-source` | `docs/architecture/git-lifecycle-version-control-adr.md` | 1 | `true` | docs/architecture/git-lifecycle-version-control-adr.md / Options lines 42-52; docs/architecture/agent-knowledge-plane.md / Semantic retrieval building blocks lines 141-167; docs/architecture/knowledge-configuration-v1-adr.md / Runner Contract lines 145-195 |
| `search_then_ingest` | `indirect-source` | `docs/architecture/harness-owned-web-search-fetch-adr.md` | 1 | `true` | docs/architecture/harness-owned-web-search-fetch-adr.md / Promoted Candidate lines 51-67; docs/architecture/generalized-artifact-ingestion-adr.md / Decision lines 62-108; docs/architecture/knowledge-configuration-v1-adr.md / `oc-v1ed` Web URL Intake Decision lines 360-399 |
| `ocr_uncertain_artifact` | `concept-recall` | `docs/architecture/generalized-artifact-ingestion-adr.md` | 2 | `true` | docs/architecture/agent-knowledge-plane.md / OCR and artifact extraction lines 168-188; docs/architecture/generalized-artifact-ingestion-adr.md / Promotion Gate lines 145-172; docs/architecture/artifact-intake-autofiling-tags-fields-adr.md / Decision lines 49-75 |
| `memory_no_hidden_truth` | `paraphrase` | `docs/architecture/memory-architecture-recall-adr.md` | 1 | `true` | docs/architecture/memory-architecture-recall-adr.md / Context lines 12-39; docs/architecture/structured-data-canonical-stores-adr.md / Authority And Approval Boundaries lines 35-53; docs/architecture/hybrid-retrieval-adr.md / Context lines 12-47 |
| `plan_filename_tags` | `synonym-drift` | `docs/architecture/artifact-intake-autofiling-tags-fields-adr.md` | 1 | `true` | docs/architecture/artifact-intake-autofiling-tags-fields-adr.md / ADR: Artifact Intake, Auto-Filing, Tags, and Fields lines 8-9; docs/architecture/generalized-artifact-ingestion-adr.md / ADR: Generalized Artifact Ingestion Direction lines 8-9; docs/architecture/harness-owned-web-search-fetch-adr.md / Options lines 40-50 |

### `provider_mimic_vector_only`

| Metric | Value |
| --- | ---: |
| status | `provider_blocked` |
| hit_at_3 | 0 |
| query_count | 8 |
| mrr | 0.000 |
| raw_duplicate_hits | 0 |
| total_seconds | 59.20 |

Description: Gemini provider-mimic embedding call was unavailable or rate-limited; semantic evidence is intentionally not faked.

Evidence posture: provider-blocked; rerun with configured Gemini API key/quota or local Ollama for local/offline evidence

Validation boundary: no fake vectors, no durable embedding store, no provider config write, no production ranking change; error: gemini returned HTTP 429: {
  "error": {
    "code": 429,
    "message": "You exceeded your current quota, please check your plan and billing details. For more information on this error, head to: https://ai.google.dev/gemini-api/docs/rate-l

### `provider_mimic_hybrid_rrf`

| Metric | Value |
| --- | ---: |
| status | `provider_blocked` |
| hit_at_3 | 0 |
| query_count | 8 |
| mrr | 0.000 |
| raw_duplicate_hits | 0 |
| total_seconds | 59.20 |

Description: Gemini provider-mimic embedding call was unavailable or rate-limited; semantic evidence is intentionally not faked.

Evidence posture: provider-blocked; rerun with configured Gemini API key/quota or local Ollama for local/offline evidence

Validation boundary: no fake vectors, no durable embedding store, no provider config write, no production ranking change; error: gemini returned HTTP 429: {
  "error": {
    "code": 429,
    "message": "You exceeded your current quota, please check your plan and billing details. For more information on this error, head to: https://ai.google.dev/gemini-api/docs/rate-l

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
| `provider-mimic-hybrid-poc` | `provider_blocked` | `pass` | `not_recorded` | `not_recorded` | `not_recorded` | Gemini provider-mimic embedding call unavailable or rate-limited; no fake vectors produced | rerun with a valid runtime_config Gemini key/quota or with local Ollama for oc-bq8c local evidence |
