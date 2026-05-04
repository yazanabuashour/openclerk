---
decision_id: decision-openclerk-next-phase-maturity-validation
decision_title: OpenClerk Next-Phase Maturity Validation
decision_status: accepted
decision_scope: maturity-validation
decision_owner: platform
decision_date: 2026-05-04
source_refs: docs/architecture/openclerk-next-phase-maturity-evidence-inventory.md, docs/evals/real-vault-dogfood.md, docs/evals/scale-ladder-validation.md, docs/evals/results/ockp-real-vault-agentops-trial.md, docs/evals/results/ockp-real-vault-dogfood.md, docs/evals/results/ockp-scale-ladder-10mb.md, docs/evals/results/ockp-scale-ladder-100mb-timeout.md, docs/evals/results/ockp-scale-ladder-100mb.md, docs/evals/results/ockp-scale-ladder-1gb.md, docs/evals/results/ockp-scale-ladder-100mb-fts-write-tuned.md, docs/evals/results/ockp-scale-ladder-1gb-fts-write-tuned.md
---
# Decision: OpenClerk Next-Phase Maturity Validation

## Status

Accepted for `oc-oa53`.

This decision completes the first next-phase maturity pass after repo-docs
dogfooding. It does not add a public runner action, schema, storage migration,
skill behavior, hybrid/vector retrieval, memory transport, or release gate.

## Evidence Summary

| Lane | Evidence | Safety pass | Capability pass | UX quality | Performance | Evidence posture |
| --- | --- | --- | --- | --- | --- | --- |
| Current production gate | [`docs/evals/results/ockp-agentops-production.md`](../evals/results/ockp-agentops-production.md) | Pass: no direct SQLite, broad repo search, source-built runner, module-cache inspection, or invalid validation tooling. | Pass: 30/30 production scenarios passed. | Pass for the existing release gate. | Harness elapsed 176.53s in the recorded report. | Release-blocking AgentOps evidence, not real-vault or scale evidence. |
| Repo-docs dogfood | [`docs/evals/results/ockp-repo-docs-dogfood.md`](../evals/results/ockp-repo-docs-dogfood.md) | Pass: all rows report `none_observed` safety risks. | Pass: 7/7 selected repo-docs rows completed. | Completed for targeted repo-docs rows. | Scenario wall times ranged from 11.12s to 24.03s. | Public repo markdown only; no private vault evidence. |
| Sanitized real-vault trial | [`docs/evals/results/ockp-real-vault-agentops-trial.md`](../evals/results/ockp-real-vault-agentops-trial.md) | Pass: workflow used installed runner JSON and no direct SQLite, direct vault inspection, broad repo search, HTTP/MCP, source-built runner path during workflow execution, copied vault files, screenshots, or raw logs as evidence. | Pass for tested workflows: source discovery, synthesis create/update, freshness/provenance inspection, decision-record lookup, stale or duplicate synthesis detection. | Acceptable for this sanitized evidence pass; no new routine surface is justified. | Reduced report does not include numeric latency; future real-vault refreshes should use the new maturity harness for timing. | Sanitized aggregate real-vault evidence; private paths, titles, snippets, citations, document IDs, and raw JSON remain local-only. |
| Reduced real-vault maturity timing | [`docs/evals/results/ockp-real-vault-dogfood.md`](../evals/results/ockp-real-vault-dogfood.md) | Pass: reduced report only, no raw vault content, logs, private query text, or machine-local artifact refs. | Pass: local representative vault synced and read probes completed. | Not agent UX evidence; command count is represented as 7 read probes, not Codex command telemetry. | Import/sync 5.90s; reopen/rebuild 11.05s; FTS probes 0.04s total. | Local representative reduced report with `<private-vault>` placeholder; routine-agent bypass events are not available from this maintainer harness. |
| 10 MB scale ladder | [`docs/evals/results/ockp-scale-ladder-10mb.md`](../evals/results/ockp-scale-ladder-10mb.md) | Pass: reduced report only, no raw generated corpus, logs, or machine-local artifact refs. | Pass: deterministic generated corpus synced and read probes completed. | Not agent UX evidence. | Import/sync 5.99s; reopen/rebuild 11.14s; FTS probes 0.04s total. | Synthetic scale evidence over 80 generated docs and about 10 MB corpus. |
| Initial 100 MB timeout | [`docs/evals/results/ockp-scale-ladder-100mb-timeout.md`](../evals/results/ockp-scale-ladder-100mb-timeout.md) | Pass for reduced-report boundary. | Incomplete: no completed reduced runtime report. | Not agent UX evidence. | Performance cliff observed: full run exceeded 10 minutes; `--skip-reopen` rerun exceeded 6 minutes. | Timeout/stall evidence only; superseded by the `oc-oa53.12` diagnostic run. |
| Tuned 100 MB scale ladder | [`docs/evals/results/ockp-scale-ladder-100mb.md`](../evals/results/ockp-scale-ladder-100mb.md) | Pass: reduced report only, no raw generated corpus, logs, or machine-local artifact refs. | Pass: deterministic generated corpus synced and read probes completed. | Not agent UX evidence. | Import/sync 19.38s; reopen/no-op sync 0.39s; FTS probes 0.28s total. | Synthetic scale evidence over 800 generated docs and about 100 MB corpus, with reduced sync diagnostics. |
| 1 GB scale ladder | [`docs/evals/results/ockp-scale-ladder-1gb.md`](../evals/results/ockp-scale-ladder-1gb.md) | Pass: reduced report only, no raw generated corpus, logs, or machine-local artifact refs. | Pass: deterministic generated corpus synced and read probes completed. | Not agent UX evidence. | Import/sync 1657.81s; reopen/no-op sync 4.91s; FTS probes 8.88s total. | Synthetic scale evidence over 8,183 generated docs and about 1 GB corpus, with reduced sync diagnostics. |
| FTS-write-tuned 100 MB scale ladder | [`docs/evals/results/ockp-scale-ladder-100mb-fts-write-tuned.md`](../evals/results/ockp-scale-ladder-100mb-fts-write-tuned.md) | Pass: reduced report only, no raw generated corpus, logs, or machine-local artifact refs. | Pass: deterministic generated corpus synced and read probes completed. | Not agent UX evidence. | Import/sync 4.06s; reopen/no-op sync 0.30s; FTS probes 0.25s total. | Synthetic scale evidence with bulk FTS rebuild diagnostics; incremental FTS write time was 0.00s during full import. |
| FTS-write-tuned 1 GB scale ladder | [`docs/evals/results/ockp-scale-ladder-1gb-fts-write-tuned.md`](../evals/results/ockp-scale-ladder-1gb-fts-write-tuned.md) | Pass: reduced report only, no raw generated corpus, logs, or machine-local artifact refs. | Pass: deterministic generated corpus synced and read probes completed. | Not direct agent UX evidence; below the 600s guardrail used for routine usability pressure. | Import/sync 68.06s; reopen/no-op sync 5.97s; FTS probes 4.40s total. | Synthetic scale evidence over 8,183 generated docs and about 1 GB corpus; 1 GB import is about 1.6x byte-linear relative to tuned 100 MB. |

## Representative Real-Vault Detail

The sanitized real-vault AgentOps workflow covered five workflow rows and
recorded 11 runner action usages across nine unique action classes:
`list_documents`, `get_document`, `search`, `create_document`,
`replace_section`, `projection_states`, `provenance_events`,
`decisions_lookup`, and `decision_record`. Every workflow row recorded failure
classification `none`.

The reduced real-vault timing harness covered seven read probes:
`list-documents`, `get-document`, three `fts-search` probes,
`projection-synthesis-sample`, and `provenance-sample`. This records timing and
runtime capability, but it is not Codex command telemetry and does not include
routine-agent bypass events.

## Decisions

Representative real-vault workflows stay on current v1 AgentOps surfaces for
now. The sanitized real-vault trial and reduced timing report did not show a
capability gap or UX gap that justifies new runner-owned surfaces. No
candidate-comparison Beads are needed from the tested real-vault workflows.

Retrieval/indexing decision: continue with lexical SQLite FTS for now. `oc-oa53.12`
showed that the original 100 MB cliff was caused by full-sync document import
repeatedly rebuilding projections and lacking interruption-surviving progress
diagnostics. `oc-oa53.12.1` then showed that the 1 GB import-bound result was
primarily the current FTS write shape, not a retrieval-mode failure: deferring
full-import FTS rows and bulk-rebuilding `chunk_fts` reduced 1 GB import/sync
from 1657.81s to 68.06s.

The 1 GB tier remains maturity evidence rather than a release gate. A 10-minute
cutoff is not a correctness threshold, but the tuned 1 GB import is below that
guardrail and scales about 1.6x worse than byte-linear relative to the tuned
100 MB run. This supports continued current-path SQLite FTS tuning before any
hybrid/vector candidate comparison.

LLM-wiki next surfaces: keep the current mapping to existing source intake,
source-linked synthesis, search/list/get, graph/document links, provenance,
projection states, records/decisions, and promoted workflow actions. Do not
promote wiki health check, context pack, document lifecycle review, or
source-linked answer filing from this evidence. The user need is already
covered well enough by existing surfaces for this pass, and no unresolved
candidate-comparison need remains.

Release-gate policy: keep the full production gate and repo-docs dogfood as
the mandatory pre-release evidence. Do not make real-vault dogfood or scale
ladder mandatory yet. Real-vault and scale reports should remain maturity
evidence. The 100 MB and 1 GB scale reports are useful maturity inputs, but
their runtime profile is too expensive for a mandatory release gate. No
release-doc update is warranted by this evidence.

## Taste Review

A normal user could eventually expect smoother high-level health or context
pack surfaces, but the current evidence does not show repeated real-vault
workflow ceremony after the promoted workflow actions already added
`compile_synthesis`, `source_audit_report`, `evidence_bundle_report`,
`duplicate_candidate_report`, `memory_router_recall_report`, and
`ingest_source_url` plan mode.

The real taste debt was performance and observability at larger corpus sizes.
`oc-oa53.12` improved observability with reduced sync diagnostics and moved
100 MB from timeout to completion. `oc-oa53.12.1` reduced the 1 GB import/write
cost enough that the remaining larger cost is projection rebuild plus bulk FTS
maintenance, not retrieval relevance.

Beads searches before closing this decision found no existing hybrid/vector,
LLM-wiki, or release-gate candidate work matching the non-promotion outcomes.
The scale-ladder follow-ups are linked as `oc-oa53.12` and `oc-oa53.12.1`.

## Compatibility

- Existing public runner surfaces remain `openclerk document` and
  `openclerk retrieval`.
- Existing release gates remain separate from maturity-validation lanes.
- Committed reports continue to use repo-relative paths or neutral placeholders
  such as `<run-root>` and `<private-vault>`.
- Raw logs, generated scale corpora, SQLite databases, raw private vault
  content, machine-absolute paths, private document paths, private titles,
  snippets, document ids, and chunk ids remain out of committed artifacts.
