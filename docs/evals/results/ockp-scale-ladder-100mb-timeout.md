# OpenClerk 100 MB Scale-Ladder Timeout Report

- Lane: `scale-ladder-validation`
- Mode: `scale-ladder-timeout`
- Tier: `100mb`
- Seed: `53`
- Harness: maintainer-only OpenClerk embedded runtime maturity harness
- Run root: `<run-root>`
- Raw logs committed: `false`
- Raw content committed: `false`

## Attempts

| Attempt | Command reference | Elapsed before interrupt | Status | Observed state |
| --- | --- | ---: | --- | --- |
| `full-run-with-reopen` | `mise exec -- go run ./scripts/agent-eval/ockp maturity scale-ladder --tier 100mb --seed 53 --run-root <run-root> --report-name ockp-scale-ladder-100mb` | 600s+ | `interrupted` | CPU-bound harness process remained active after more than 10 minutes; no reduced report was produced. |
| `skip-reopen-rerun` | `mise exec -- go run ./scripts/agent-eval/ockp maturity scale-ladder --tier 100mb --seed 53 --run-root <run-root> --report-name ockp-scale-ladder-100mb --skip-reopen` | 360s+ | `interrupted` | CPU-bound harness process remained active after more than 6 minutes; no reduced report was produced. |

## Checks

| Check | Value |
| --- | --- |
| reduced_report_only | `true` |
| raw_logs_committed | `false` |
| raw_content_committed | `false` |
| machine_absolute_artifact_refs | `false` |
| routine_agent_bypass_events_available | `false` |
| boundary | This report records reduced timeout/stall evidence only. Generated corpus content, SQLite databases, process details, and raw run roots remain local-only. |

## Outcomes

| Name | Status | Safety pass | Capability pass | UX quality | Performance | Evidence posture | Details |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `reduced-report-boundary` | `completed` | `pass` | `pass` | `not_agent_ux_evidence` | `not_applicable` | neutral artifact references only; raw content and raw logs are not committed | The timeout report intentionally excludes generated document paths, titles, snippets, doc ids, chunk ids, raw run roots, and machine-absolute paths. |
| `runtime-scale-100mb` | `stalled` | `pass` | `incomplete` | `not_agent_ux_evidence` | `cliff_observed` | 10 MB completed; 100 MB did not produce a reduced report before manual interruption | The 100 MB tier needs performance investigation before it can be used as routine release-gate evidence or as justification for a 1 GB run. |

## Follow-Up

`oc-oa53.12` owns the investigation. It must determine whether the stall is
projection rebuild cost, FTS indexing overhead, synthetic corpus shape, or
harness behavior, then either produce a reduced 100 MB report or record a
decision-quality performance finding.
