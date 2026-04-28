# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `video-youtube-canonical-source-note`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `19.58`
- Harness elapsed seconds: `121.97`
- Effective parallel speedup: `0.69x`
- Parallel efficiency: `0.69`
- Targeted acceptance: video/YouTube rows report natural URL-only intent, scripted transcript control, synthesis freshness, bypass rejection, ergonomics scorecard fields, and final capability or ergonomics classification
- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`

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
| copy_repo | 0.13 |
| install_variant | 17.83 |
| warm_cache | 0.00 |
| seed_data | 0.03 |
| agent_run | 84.28 |
| parse_metrics | 0.00 |
| verify | 0.12 |
| total | 102.40 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `video-youtube-natural-intent` | `completed` | 0 | 0 | 1 | 2493 | 6.90 | `<run-root>/production/video-youtube-natural-intent/turn-1/events.jsonl` |
| `production` | `video-youtube-scripted-transcript-control` | `completed` | 8 | 8 | 5 | 7719 | 22.20 | `<run-root>/production/video-youtube-scripted-transcript-control/turn-1/events.jsonl` |
| `production` | `video-youtube-synthesis-freshness` | `completed` | 28 | 28 | 10 | 9862 | 50.62 | `<run-root>/production/video-youtube-synthesis-freshness/turn-1/events.jsonl` |
| `production` | `video-youtube-bypass-reject` | `completed` | 0 | 0 | 1 | 2469 | 4.56 | `<run-root>/production/video-youtube-bypass-reject/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `promote_video_ingest_surface_design`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: promote follow-up design for ingest_video_url only; no runner action, parser, dependency, schema, storage migration, or public API is implemented by this eval.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- | --- |
| `production` | `video-youtube-natural-intent` | `completed` | `ergonomics_gap` | 0 | 0 | 1 | 6.90 | `natural-user-intent` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | current primitives safely reject URL-only native video ingestion, but the natural user intent still cannot produce a canonical source note without manual transcript acquisition |
| `production` | `video-youtube-scripted-transcript-control` | `completed` | `none` | 8 | 8 | 5 | 22.20 | `scripted-control` | `completed` | `normal` | 0 | 8 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | current document/retrieval runner evidence preserved video transcript authority, citations, provenance, freshness, and bypass boundaries when transcript text was supplied |
| `production` | `video-youtube-synthesis-freshness` | `completed` | `none` | 28 | 28 | 10 | 50.62 | `scenario-specific` | `completed` | `normal` | 0 | 28 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | current document/retrieval runner evidence preserved video transcript authority, citations, provenance, freshness, and bypass boundaries when transcript text was supplied |
| `production` | `video-youtube-bypass-reject` | `completed` | `none` | 0 | 0 | 1 | 4.56 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | current document/retrieval runner evidence preserved video transcript authority, citations, provenance, freshness, and bypass boundaries when transcript text was supplied |
