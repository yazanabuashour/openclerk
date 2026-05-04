# Memory Architecture And Recall Track Eval

Date: 2026-05-04

## Scenario

Evaluate `oc-uj2y.3` memory architecture candidates: no separate memory layer,
source-linked memory docs, internal memory projection, Mem0/external memory,
and the existing read-only `memory_router_recall_report`.

## Result

Keep `memory_router_recall_report` as the promoted surface for this track. Do
not add a memory write path or memory transport.

## Safety Pass

Pass.

The selected surface is read-only, installed-runner-only, local-first, and
does not create memory entries, write documents, call Mem0, use memory
transports, create vector or embedding stores, inspect SQLite directly, inspect
raw vault files, use HTTP/MCP bypasses, or hide stale authority.

## Capability Pass

Pass.

Existing implementation evidence under
`docs/evals/results/ockp-memory-router-recall-report-implementation.md`
shows the report returns canonical evidence refs, stale-session posture,
feedback weighting, routing rationale, provenance refs, synthesis freshness,
validation boundaries, and authority limits.

## UX Quality

Pass.

Earlier current-primitives evidence showed high ceremony for a normal memory
recall question. The report is already the lower-step promoted shape and keeps
ordinary knowledge questions on `search`.

## Performance

The action is bounded by runner-visible search, fixed canonical memory/router
document lookups, provenance inspection, and synthesis freshness inspection.
No import job, external memory call, vector build, or corpus scan is added by
this track.

## Evidence Posture

This reduced track report relies on committed implementation evidence and unit
coverage. It does not commit raw logs, private corpus examples, generated
corpora, SQLite databases, or machine-absolute paths.

## Decision

Select the existing read-only `memory_router_recall_report` implementation as
the promoted surface. Keep Mem0 and memory projections as reference/future
candidate tracks until evidence proves a write or projection surface can avoid
truth drift.
