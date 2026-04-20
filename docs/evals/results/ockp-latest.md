# OpenClerk Agent Eval

- Model: `gpt-5.4-mini`
- Reasoning effort: `medium`
- Configured parallelism: `4`
- Cache mode: `shared`
- Cache prewarm seconds: `20.39`
- Harness elapsed seconds: `126.65`
- Effective parallel speedup: `1.90x`
- Parallel efficiency: `0.48`
- Raw logs: `<run-root>/<variant>/<scenario>/turn-N/events.jsonl`

## Production Gate

Variant: `production`

Passes gate: `false`

Recommendation: `fix_production_agentops_before_release`

| Criterion | Status | Details |
| --- | --- | --- |
| `production_passes_all_scenarios` | `pass` | 5/5 production scenarios passed |
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
| copy_repo | 0.10 |
| install_variant | 5.41 |
| warm_cache | 0.00 |
| seed_data | 0.10 |
| agent_run | 241.02 |
| parse_metrics | 0.00 |
| verify | 0.16 |
| total | 246.81 |

## Results

| Variant | Scenario | Status | Tools | Commands | Assistant Calls | Non-Cached Input | Wall Seconds | Raw Log |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `production` | `search-synthesis` | `completed` | 14 | 14 | 5 | 7060 | 38.62 | `<run-root>/production/search-synthesis/turn-1/events.jsonl` |
| `production` | `answer-filing` | `completed` | 10 | 10 | 4 | 23260 | 26.48 | `<run-root>/production/answer-filing/turn-1/events.jsonl` |
| `production` | `stale-synthesis-update` | `completed` | 10 | 10 | 3 | 11482 | 28.53 | `<run-root>/production/stale-synthesis-update/turn-1/events.jsonl` |
| `production` | `mixed-synthesis-records` | `completed` | 46 | 46 | 7 | 30173 | 105.03 | `<run-root>/production/mixed-synthesis-records/turn-1/events.jsonl` |
| `production` | `mt-source-then-synthesis` | `completed` | 16 | 16 | 6 | 19456 | 42.36 | `<run-root>/production/mt-source-then-synthesis/turn-2/events.jsonl` |
