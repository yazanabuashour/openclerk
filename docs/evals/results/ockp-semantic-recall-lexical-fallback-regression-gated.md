# OpenClerk Semantic Recall Report

- Lane: `semantic-recall`
- Mode: `lexical-fallback`
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

Description: Installed OpenClerk Search with SQLite FTS plus regression-gated zero-hit lexical fallback; no schema change.

Evidence posture: citation-bearing current lexical baseline; no vector evidence claimed

Validation boundary: uses embedded OpenClerk runtime only; zero-hit fallback is lexical and citation-bearing; no embedding store, provider default, or JSON schema change

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

### `lexical_token_overlap_fallback`

| Metric | Value |
| --- | ---: |
| status | `completed` |
| hit_at_3 | 7 |
| query_count | 8 |
| mrr | 0.833 |
| raw_duplicate_hits | 520 |
| total_seconds | 0.15 |

Description: Eval-only stopword-trimmed token-overlap fallback with title/path/heading weighting.

Evidence posture: eval-only no-vector lexical fallback; does not change production Search

Validation boundary: candidate scoring runs inside maintainer harness only; no openclerk retrieval JSON contract change

| Query | Kind | Expected | Rank | Hit | Top citations |
| --- | --- | --- | ---: | --- | --- |
| `wiki_synthesis` | `concept-recall` | `docs/architecture/agent-knowledge-plane.md` | 1 | `true` | docs/architecture/agent-knowledge-plane.md / Summary lines 3-38; docs/architecture/eval-backed-knowledge-plane-adr.md / Context lines 17-34; docs/architecture/knowledge-configuration-v1-adr.md / Context lines 20-39 |
| `semantic_retrieval_gap` | `paraphrase` | `docs/architecture/hybrid-retrieval-adr.md` | 6 | `false` | docs/architecture/harness-owned-web-search-fetch-adr.md / Context lines 27-39; docs/architecture/memory-architecture-recall-adr.md / Memory Architecture And Recall ADR lines 1-11; docs/architecture/knowledge-configuration-v1-adr.md / `oc-za6.5` POC Decision lines 438-473 |
| `structured_rows_vs_notes` | `synonym-drift` | `docs/architecture/structured-data-canonical-stores-adr.md` | 1 | `true` | docs/architecture/structured-data-canonical-stores-adr.md / Structured Data And Canonical Stores ADR lines 1-11; docs/architecture/agent-knowledge-plane.md / Canonical and derived layers lines 82-206; docs/architecture/git-lifecycle-version-control-adr.md / Context lines 27-41 |
| `checkpoint_not_restore` | `indirect-source` | `docs/architecture/git-lifecycle-version-control-adr.md` | 1 | `true` | docs/architecture/git-lifecycle-version-control-adr.md / Options lines 42-52; docs/architecture/knowledge-configuration-v1-adr.md / Runner Contract lines 145-195; docs/architecture/structured-data-canonical-stores-adr.md / Authority And Approval Boundaries lines 35-53 |
| `search_then_ingest` | `indirect-source` | `docs/architecture/harness-owned-web-search-fetch-adr.md` | 1 | `true` | docs/architecture/harness-owned-web-search-fetch-adr.md / Promoted Candidate lines 51-67; docs/architecture/generalized-artifact-ingestion-adr.md / Context lines 33-61; docs/architecture/knowledge-configuration-v1-adr.md / `oc-v1ed` Web URL Intake Decision lines 360-399 |
| `ocr_uncertain_artifact` | `concept-recall` | `docs/architecture/generalized-artifact-ingestion-adr.md` | 2 | `true` | docs/architecture/agent-knowledge-plane.md / Canonical and derived layers lines 82-206; docs/architecture/generalized-artifact-ingestion-adr.md / Promotion Gate lines 145-172; docs/architecture/artifact-intake-autofiling-tags-fields-adr.md / Non-Goals lines 128-139 |
| `memory_no_hidden_truth` | `paraphrase` | `docs/architecture/memory-architecture-recall-adr.md` | 1 | `true` | docs/architecture/memory-architecture-recall-adr.md / Context lines 12-39; docs/architecture/agent-knowledge-plane.md / Canonical and derived layers lines 82-206; docs/architecture/structured-data-canonical-stores-adr.md / Authority And Approval Boundaries lines 35-53 |
| `plan_filename_tags` | `synonym-drift` | `docs/architecture/artifact-intake-autofiling-tags-fields-adr.md` | 1 | `true` | docs/architecture/artifact-intake-autofiling-tags-fields-adr.md / ADR: Artifact Intake, Auto-Filing, Tags, and Fields lines 1-9; docs/architecture/generalized-artifact-ingestion-adr.md / ADR: Generalized Artifact Ingestion Direction lines 1-9; docs/architecture/harness-owned-web-search-fetch-adr.md / Options lines 40-50 |

### `lexical_alias_overlap_fallback`

| Metric | Value |
| --- | ---: |
| status | `completed` |
| hit_at_3 | 8 |
| query_count | 8 |
| mrr | 0.938 |
| raw_duplicate_hits | 573 |
| total_seconds | 0.15 |

Description: Eval-only token-overlap fallback plus documented domain aliases for each semantic-recall query row.

Evidence posture: eval-only no-vector lexical fallback; does not change production Search

Validation boundary: candidate scoring runs inside maintainer harness only; no openclerk retrieval JSON contract change

| Query | Kind | Expected | Rank | Hit | Top citations |
| --- | --- | --- | ---: | --- | --- |
| `wiki_synthesis` | `concept-recall` | `docs/architecture/agent-knowledge-plane.md` | 1 | `true` | docs/architecture/agent-knowledge-plane.md / Summary lines 3-38; docs/architecture/eval-backed-knowledge-plane-adr.md / Context lines 17-34; docs/architecture/knowledge-configuration-v1-adr.md / Context lines 20-39 |
| `semantic_retrieval_gap` | `paraphrase` | `docs/architecture/hybrid-retrieval-adr.md` | 2 | `true` | docs/architecture/local-first-hybrid-retrieval-implementation-candidate-decision.md / Decision: Local-First Hybrid Retrieval Implementation Candidates lines 1-11; docs/architecture/hybrid-retrieval-adr.md / Candidates lines 67-78; docs/architecture/hybrid-retrieval-promotion-decision.md / Hybrid Retrieval Promotion Decision lines 1-11 |
| `structured_rows_vs_notes` | `synonym-drift` | `docs/architecture/structured-data-canonical-stores-adr.md` | 1 | `true` | docs/architecture/structured-data-canonical-stores-adr.md / Structured Data And Canonical Stores ADR lines 1-11; docs/architecture/agent-knowledge-plane.md / Canonical and derived layers lines 82-206; docs/architecture/hybrid-retrieval-adr.md / Context lines 12-47 |
| `checkpoint_not_restore` | `indirect-source` | `docs/architecture/git-lifecycle-version-control-adr.md` | 1 | `true` | docs/architecture/git-lifecycle-version-control-adr.md / ADR: Local Git Lifecycle Version Control lines 1-9; docs/architecture/artifact-intake-autofiling-tags-fields-adr.md / Context lines 21-39; docs/architecture/structured-data-canonical-stores-adr.md / Authority And Approval Boundaries lines 35-53 |
| `search_then_ingest` | `indirect-source` | `docs/architecture/harness-owned-web-search-fetch-adr.md` | 1 | `true` | docs/architecture/harness-owned-web-search-fetch-adr.md / Promoted Candidate lines 51-67; docs/architecture/knowledge-configuration-v1-adr.md / `oc-v1ed` Web URL Intake Decision lines 360-399; docs/architecture/generalized-artifact-ingestion-adr.md / Decision lines 62-108 |
| `ocr_uncertain_artifact` | `concept-recall` | `docs/architecture/generalized-artifact-ingestion-adr.md` | 1 | `true` | docs/architecture/generalized-artifact-ingestion-adr.md / Decision lines 62-108; docs/architecture/artifact-intake-autofiling-tags-fields-adr.md / Confidence Policy lines 110-119; docs/architecture/agent-knowledge-plane.md / Canonical and derived layers lines 82-206 |
| `memory_no_hidden_truth` | `paraphrase` | `docs/architecture/memory-architecture-recall-adr.md` | 1 | `true` | docs/architecture/memory-architecture-recall-adr.md / Memory Architecture And Recall ADR lines 1-11; docs/architecture/structured-data-canonical-stores-adr.md / Authority And Approval Boundaries lines 35-53; docs/architecture/agent-knowledge-plane.md / Canonical and derived layers lines 82-206 |
| `plan_filename_tags` | `synonym-drift` | `docs/architecture/artifact-intake-autofiling-tags-fields-adr.md` | 1 | `true` | docs/architecture/artifact-intake-autofiling-tags-fields-adr.md / ADR: Artifact Intake, Auto-Filing, Tags, and Fields lines 1-9; docs/architecture/generalized-artifact-ingestion-adr.md / ADR: Generalized Artifact Ingestion Direction lines 1-9; docs/architecture/harness-owned-web-search-fetch-adr.md / Options lines 40-50 |

## Freshness Probe

| Field | Value |
| --- | --- |
| status | `not_run` |
| changed_path | `` |
| stale_chunks | 0 |
| rebuilt_chunks | 0 |
| seconds | 0.00 |
| evidence_posture | freshness probe belongs to local-hybrid mode |
| validation_boundary | no document mutation outside <run-root> |

## Checks

| Check | Value |
| --- | --- |
| reduced_report_only | `true` |
| raw_logs_committed | `false` |
| raw_content_committed | `false` |
| machine_absolute_artifact_refs | `false` |
| production_search_default_changed | `true` |
| boundary | maintainer harness; current openclerk retrieval search now includes the regression-gated zero-hit lexical fallback, no embedding store, provider embedding default, or JSON schema change |

## Outcomes

| Name | Status | Safety pass | Capability pass | UX quality | Performance | Evidence posture | Details |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `lexical-fallback-eval` | `recorded` | `pass` | `pass_for_zero_hit_fallback` | `pass_if_invisible_in_search` | `low_cost_production_fallback` | reduced query-row metrics; no vector or provider calls | current production Search includes the regression-gated zero-hit lexical fallback; alias expansion remains eval-only |
