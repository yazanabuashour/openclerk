---
decision_id: decision-openclerk-next-phase-maturity-validation
decision_title: OpenClerk Next-Phase Maturity Validation
decision_status: accepted
decision_scope: maturity-validation
decision_owner: platform
decision_date: 2026-05-04
source_refs: docs/architecture/openclerk-next-phase-maturity-evidence-inventory.md, docs/evals/real-vault-dogfood.md, docs/evals/scale-ladder-validation.md, docs/evals/results/ockp-real-vault-agentops-trial.md, docs/evals/results/ockp-real-vault-dogfood.md, docs/evals/results/ockp-scale-ladder-10mb.md, docs/evals/results/ockp-scale-ladder-100mb-timeout.md
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
| 100 MB scale ladder | [`docs/evals/results/ockp-scale-ladder-100mb-timeout.md`](../evals/results/ockp-scale-ladder-100mb-timeout.md) | Pass for reduced-report boundary. | Incomplete: no completed reduced runtime report. | Not agent UX evidence. | Performance cliff observed: full run exceeded 10 minutes; `--skip-reopen` rerun exceeded 6 minutes. | Timeout/stall evidence only; follow-up `oc-oa53.12` owns diagnosis. |

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

Retrieval/indexing decision: tune or diagnose the current SQLite FTS and
projection path before considering hybrid/vector retrieval. The 10 MB tier
passed, but the 100 MB tier stalled before producing a completed reduced
report. This is performance evidence, not evidence that semantic/vector
retrieval would solve the problem. `oc-oa53.12` remains the required follow-up
to determine whether the issue is projection rebuild cost, FTS indexing
overhead, synthetic corpus shape, or harness behavior.

The 1 GB tier is not justified now. Running it before `oc-oa53.12` resolves
the 100 MB cliff would likely consume time without producing new decision
quality. The correct outcome for `oc-oa53.8` is `not_run_blocked_by_100mb`.

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
evidence until the real-vault timing refresh and 100 MB diagnosis are complete.
No release-doc update is warranted by this evidence.

## Taste Review

A normal user could eventually expect smoother high-level health or context
pack surfaces, but the current evidence does not show repeated real-vault
workflow ceremony after the promoted workflow actions already added
`compile_synthesis`, `source_audit_report`, `evidence_bundle_report`,
`duplicate_candidate_report`, `memory_router_recall_report`, and
`ingest_source_url` plan mode.

The real taste debt is performance and observability at larger corpus sizes:
the 100 MB scale ladder stalled without a completed report. The evaluated
hybrid/vector shape is not selected because the failing evidence is import,
projection, or harness cost rather than retrieval relevance. The valid
follow-up category is: need exists, evaluated shape is incomplete, existing
follow-up `oc-oa53.12` required.

Beads searches before closing this decision found no existing hybrid/vector,
LLM-wiki, or release-gate candidate work matching the non-promotion outcomes.
The scale-ladder follow-up is linked as `oc-oa53.12`.

## Compatibility

- Existing public runner surfaces remain `openclerk document` and
  `openclerk retrieval`.
- Existing release gates remain separate from maturity-validation lanes.
- Committed reports continue to use repo-relative paths or neutral placeholders
  such as `<run-root>` and `<private-vault>`.
- Raw logs, generated scale corpora, SQLite databases, raw private vault
  content, machine-absolute paths, private document paths, private titles,
  snippets, document ids, and chunk ids remain out of committed artifacts.
