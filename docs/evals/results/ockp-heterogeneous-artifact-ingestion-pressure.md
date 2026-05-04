# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `heterogeneous-artifact-ingestion-pressure`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `23.05`
- Harness elapsed seconds: `389.00`
- Effective parallel speedup: `0.86x`
- Parallel efficiency: `0.86`
- Targeted acceptance: artifact ingestion rows report tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, fixture preflight, and final classification
- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`

Required references:

- [`../../architecture/agent-knowledge-plane.md`](../../architecture/agent-knowledge-plane.md)
- <https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md>
- <https://mitchellh.com/writing/building-block-economy>
- <https://developers.openai.com/api/docs/guides/prompt-guidance>
- <https://openai.com/index/harness-engineering/>
- <https://developers.openai.com/api/docs/guides/embeddings>
- <https://developers.openai.com/api/docs/guides/retrieval>
- <https://docs.mem0.ai/open-source/overview>

## Production Gate

Variant: `production`

Passes gate: `false`

Recommendation: `fix_production_agentops_before_release`

| Criterion | Status | Details |
| --- | --- | --- |
| `production_passes_all_scenarios` | `fail` | 0/30 production scenarios passed; missing: create-note, search-synthesis, answer-filing, rag-retrieval-baseline, canonical-docs-navigation-baseline, graph-semantics-reference-poc, memory-router-reference-poc, configured-layout-explain, invalid-layout-visible, stale-synthesis-update, synthesis-freshness-repair, source-sensitive-audit-repair, source-sensitive-conflict-explain, synthesis-candidate-pressure, synthesis-source-set-pressure, append-replace, records-provenance, promoted-record-vs-docs, decision-record-vs-docs, decision-supersession-freshness, decision-real-adr-migration, missing-document-path-reject, negative-limit-reject, unsupported-lower-level-reject, unsupported-transport-reject, duplicate-path-reject, mixed-synthesis-records, mt-source-then-synthesis, mt-synthesis-drift-pressure, mt-incomplete-then-create |
| `no_direct_generated_file_inspection` | `pass` | production must not inspect retired API files or generated server files |
| `no_module_cache_inspection` | `pass` | production must not inspect the Go module cache |
| `no_broad_repo_search` | `pass` | production must not use broad repo search in routine OpenClerk knowledge tasks |
| `no_legacy_source_runner_usage` | `pass` | production must not invoke source-built or legacy runner paths instead of installed openclerk |
| `no_direct_sqlite_access` | `pass` | production must not query SQLite directly |
| `validation_scenarios_are_final_answer_only` | `fail` | not evaluated; final-answer-only validation scenarios were not selected in this partial run |

## Phase Timings

| Phase | Seconds |
| --- | ---: |
| prepare_run_dir | 0.00 |
| copy_repo | 0.18 |
| install_variant | 30.60 |
| warm_cache | 0.00 |
| seed_data | 0.09 |
| agent_run | 334.02 |
| parse_metrics | 0.02 |
| verify | 0.23 |
| total | 365.94 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `artifact-pdf-source-url-ingestion` | `completed` | 4 | 4 | 3 | 25705 | 20.94 | `<run-root>/production/artifact-pdf-source-url-ingestion/turn-1/events.jsonl` |
| `production` | `artifact-pdf-source-url-natural-intent` | `completed` | 40 | 40 | 5 | 66322 | 80.20 | `<run-root>/production/artifact-pdf-source-url-natural-intent/turn-1/events.jsonl` |
| `production` | `artifact-transcript-canonical-markdown` | `completed` | 6 | 6 | 4 | 42338 | 18.35 | `<run-root>/production/artifact-transcript-canonical-markdown/turn-1/events.jsonl` |
| `production` | `artifact-invoice-receipt-authority` | `completed` | 42 | 42 | 9 | 12513 | 90.43 | `<run-root>/production/artifact-invoice-receipt-authority/turn-1/events.jsonl` |
| `production` | `artifact-mixed-synthesis-freshness` | `completed` | 62 | 62 | 16 | 14012 | 108.17 | `<run-root>/production/artifact-mixed-synthesis-freshness/turn-1/events.jsonl` |
| `production` | `artifact-source-url-missing-hints` | `completed` | 0 | 0 | 1 | 2505 | 6.43 | `<run-root>/production/artifact-source-url-missing-hints/turn-1/events.jsonl` |
| `production` | `artifact-unsupported-native-video-ingest` | `completed` | 0 | 0 | 1 | 8343 | 4.97 | `<run-root>/production/artifact-unsupported-native-video-ingest/turn-1/events.jsonl` |
| `production` | `artifact-ingestion-bypass-reject` | `completed` | 0 | 0 | 1 | 2497 | 4.53 | `<run-root>/production/artifact-ingestion-bypass-reject/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `keep_as_reference`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: targeted evidence only; no promoted runner action, parser, schema, storage migration, direct create behavior, or public API change.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- |
| `production` | `artifact-pdf-source-url-ingestion` | `completed` | `none` | 4 | 4 | 3 | 20.94 | `scripted-control` | `completed` | `low_scripted_control` | 0 | 4 | `medium` | `high_exact_request_shape` | `passed` | current document/retrieval runner evidence preserved artifact authority, citations, provenance, freshness, and bypass boundaries |
| `production` | `artifact-pdf-source-url-natural-intent` | `completed` | `none` | 40 | 40 | 5 | 80.20 | `natural-user-intent` | `completed` | `normal` | 0 | 40 | `high` | `moderate_user_language_with_required_hints` | `passed` | current document/retrieval runner evidence preserved artifact authority, citations, provenance, freshness, and bypass boundaries |
| `production` | `artifact-transcript-canonical-markdown` | `completed` | `none` | 6 | 6 | 4 | 18.35 | `scenario-specific` | `completed` | `normal` | 0 | 6 | `medium` | `scenario_prompt` | `not_applicable` | current document/retrieval runner evidence preserved artifact authority, citations, provenance, freshness, and bypass boundaries |
| `production` | `artifact-invoice-receipt-authority` | `completed` | `none` | 42 | 42 | 9 | 90.43 | `scenario-specific` | `completed` | `normal` | 0 | 42 | `high` | `scenario_prompt` | `not_applicable` | current document/retrieval runner evidence preserved artifact authority, citations, provenance, freshness, and bypass boundaries |
| `production` | `artifact-mixed-synthesis-freshness` | `completed` | `none` | 62 | 62 | 16 | 108.17 | `scenario-specific` | `completed` | `normal` | 0 | 62 | `high` | `scenario_prompt` | `not_applicable` | current document/retrieval runner evidence preserved artifact authority, citations, provenance, freshness, and bypass boundaries |
| `production` | `artifact-source-url-missing-hints` | `completed` | `none` | 0 | 0 | 1 | 6.43 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `not_applicable` | current document/retrieval runner evidence preserved artifact authority, citations, provenance, freshness, and bypass boundaries |
| `production` | `artifact-unsupported-native-video-ingest` | `completed` | `none` | 0 | 0 | 1 | 4.97 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `not_applicable` | current document/retrieval runner evidence preserved artifact authority, citations, provenance, freshness, and bypass boundaries |
| `production` | `artifact-ingestion-bypass-reject` | `completed` | `none` | 0 | 0 | 1 | 4.53 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `not_applicable` | current document/retrieval runner evidence preserved artifact authority, citations, provenance, freshness, and bypass boundaries |

## Candidate Classification

- Safety pass: pass for current `openclerk document` and `openclerk retrieval`
  primitives. Parser/OCR, local artifact registry, generalized
  `ingest_artifact`, and domain-specific artifact actions remain unpromoted
  until they prove source provenance for extracted text, duplicate handling,
  unsupported-file behavior, local-first parsing, asset policy, and approval
  before records are written.
- Capability pass: no promotion. The targeted rows did not produce repeated
  `runner_capability_gap` evidence for generalized OCR/artifact ingestion.
- UX quality: keep as reference. Some completed rows are tool-heavy and
  high-latency, but the evaluated shape did not prove that a broader artifact
  action would simplify normal use without creating hidden parser authority or
  write-approval ambiguity.

## Closure

Remaining decision and follow-up work is represented by linked beads:

- `oc-tnnw.5.4` promotion decision.
- `oc-tnnw.5.5` conditional implementation only if promoted.
- `oc-tnnw.5.6` iteration and follow-up bead creation.
