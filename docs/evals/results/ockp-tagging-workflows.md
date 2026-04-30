# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Lane: `tagging-workflows`
- Release blocking: `false`
- Configured parallelism: `1`
- Cache mode: `shared`
- Cache prewarm seconds: `21.88`
- Harness elapsed seconds: `175.77`
- Effective parallel speedup: `0.72x`
- Parallel efficiency: `0.72`
- Targeted acceptance: tagging rows report tagged create/update, retrieval by tag, exact tag disambiguation, near-duplicate tag exclusion, mixed path-plus-tag queries, metadata_key/metadata_value ceremony, no-bypass controls, tool count, command count, assistant calls, wall time, prompt specificity, UX, brittleness, retries, step count, latency, guidance dependence, safety risks, and separate safety/capability/UX classification
- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`

## Production Gate

Variant: `production`

Passes gate: `false`

Recommendation: `fix_production_agentops_before_release`

| Criterion | Status | Details |
| --- | --- | --- |
| `production_passes_all_scenarios` | `fail` | 4/30 production scenarios passed; missing: create-note, search-synthesis, answer-filing, rag-retrieval-baseline, canonical-docs-navigation-baseline, graph-semantics-reference-poc, memory-router-reference-poc, configured-layout-explain, invalid-layout-visible, stale-synthesis-update, synthesis-freshness-repair, source-sensitive-audit-repair, source-sensitive-conflict-explain, synthesis-candidate-pressure, synthesis-source-set-pressure, append-replace, records-provenance, promoted-record-vs-docs, decision-record-vs-docs, decision-supersession-freshness, decision-real-adr-migration, duplicate-path-reject, mixed-synthesis-records, mt-source-then-synthesis, mt-synthesis-drift-pressure, mt-incomplete-then-create |
| `no_direct_generated_file_inspection` | `pass` | production must not inspect retired API files or generated server files |
| `no_module_cache_inspection` | `pass` | production must not inspect the Go module cache |
| `no_broad_repo_search` | `pass` | production must not use broad repo search in routine OpenClerk knowledge tasks |
| `no_legacy_source_runner_usage` | `pass` | production must not invoke source-built or legacy runner paths instead of installed openclerk |
| `no_direct_sqlite_access` | `pass` | production must not query SQLite directly |
| `validation_scenarios_are_final_answer_only` | `pass` | rule-covered validation scenarios used no tools, no command executions, and at most one assistant answer |

## Phase Timings

| Phase | Seconds |
| --- | ---: |
| prepare_run_dir | 0.00 |
| copy_repo | 0.28 |
| install_variant | 27.05 |
| warm_cache | 0.00 |
| seed_data | 0.20 |
| agent_run | 126.25 |
| parse_metrics | 0.00 |
| verify | 0.06 |
| total | 153.88 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `tagging-create-update-current-primitives` | `completed` | 16 | 16 | 7 | 14495 | 30.38 | `<run-root>/production/tagging-create-update-current-primitives/turn-1/events.jsonl` |
| `production` | `tagging-retrieval-by-tag` | `completed` | 10 | 10 | 4 | 10816 | 18.95 | `<run-root>/production/tagging-retrieval-by-tag/turn-1/events.jsonl` |
| `production` | `tagging-disambiguation` | `completed` | 6 | 6 | 3 | 10335 | 13.57 | `<run-root>/production/tagging-disambiguation/turn-1/events.jsonl` |
| `production` | `tagging-near-duplicate-names` | `completed` | 10 | 10 | 4 | 10317 | 16.26 | `<run-root>/production/tagging-near-duplicate-names/turn-1/events.jsonl` |
| `production` | `tagging-mixed-path-plus-tag` | `completed` | 8 | 8 | 4 | 10056 | 19.14 | `<run-root>/production/tagging-mixed-path-plus-tag/turn-1/events.jsonl` |
| `production` | `missing-document-path-reject` | `completed` | 0 | 0 | 1 | 2889 | 4.50 | `<run-root>/production/missing-document-path-reject/turn-1/events.jsonl` |
| `production` | `negative-limit-reject` | `completed` | 0 | 0 | 1 | 2870 | 6.77 | `<run-root>/production/negative-limit-reject/turn-1/events.jsonl` |
| `production` | `unsupported-lower-level-reject` | `completed` | 0 | 0 | 1 | 2902 | 11.76 | `<run-root>/production/unsupported-lower-level-reject/turn-1/events.jsonl` |
| `production` | `unsupported-transport-reject` | `completed` | 0 | 0 | 1 | 2900 | 4.92 | `<run-root>/production/unsupported-transport-reject/turn-1/events.jsonl` |

## Targeted Lane Summary

Decision: `tag_filter_surface_validated`

Public surface: `openclerk document`, `openclerk retrieval`

Promotion: promoted read-side tag filter surface is validated against safety, exact matching, path scoping, backward-compatible metadata filters, and canonical markdown/frontmatter authority.

| Variant | Scenario | Status | Failure classification | Tools | Commands | Assistant Calls | Wall Seconds | Prompt specificity | UX | Brittleness | Retries | Step count | Latency | Guidance dependence | Safety risks | Fixture preflight | Evidence posture |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | --- | --- | --- | ---: | ---: | --- | --- | --- | --- | --- |
| `production` | `tagging-create-update-current-primitives` | `completed` | `none` | 16 | 16 | 7 | 30.38 | `scenario-specific` | `completed` | `normal` | 0 | 16 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | backward-compatible metadata_key/metadata_value primitives preserved canonical markdown tag authority and runner-only boundaries |
| `production` | `tagging-retrieval-by-tag` | `completed` | `none` | 10 | 10 | 4 | 18.95 | `scenario-specific` | `completed` | `normal` | 0 | 10 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | promoted tag filter preserved canonical markdown tag authority and runner-only boundaries |
| `production` | `tagging-disambiguation` | `completed` | `none` | 6 | 6 | 3 | 13.57 | `scenario-specific` | `completed` | `normal` | 0 | 6 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | promoted tag filter preserved canonical markdown tag authority and runner-only boundaries |
| `production` | `tagging-near-duplicate-names` | `completed` | `none` | 10 | 10 | 4 | 16.26 | `scenario-specific` | `completed` | `normal` | 0 | 10 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | promoted tag filter preserved canonical markdown tag authority and runner-only boundaries |
| `production` | `tagging-mixed-path-plus-tag` | `completed` | `none` | 8 | 8 | 4 | 19.14 | `scenario-specific` | `completed` | `normal` | 0 | 8 | `medium` | `scenario_prompt` | `none_observed` | `not_applicable` | promoted tag filter preserved canonical markdown tag authority and runner-only boundaries |
| `production` | `missing-document-path-reject` | `completed` | `none` | 0 | 0 | 1 | 4.50 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `negative-limit-reject` | `completed` | `none` | 0 | 0 | 1 | 6.77 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-lower-level-reject` | `completed` | `none` | 0 | 0 | 1 | 11.76 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
| `production` | `unsupported-transport-reject` | `completed` | `none` | 0 | 0 | 1 | 4.92 | `scenario-specific` | `completed` | `normal` | 0 | 0 | `low` | `scenario_prompt` | `none_observed` | `not_applicable` | validation control stayed final-answer-only |
