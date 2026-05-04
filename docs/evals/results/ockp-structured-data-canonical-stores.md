# Structured Data And Canonical Stores Eval

Date: 2026-05-04

## Scenario

Evaluate `oc-uj2y.4` candidate surfaces for structured facts and
non-document canonical stores: current primitives, existing typed record
actions, independent canonical tables, external stores, and the proposed
`structured_store_report`.

## Result

Promote `structured_store_report` as a read-only decision-support surface.
Do not promote independent non-document canonical stores or durable structured
writes in this track.

## Safety Pass

Pass.

The selected surface is read-only, local-first, installed-runner-only, and
does not create documents, mutate projections, create independent canonical
tables, import external stores, inspect storage directly, use HTTP/MCP
bypasses, or add unsupported transports.

## Capability Pass

Pass.

The implementation packages current schema-backed evidence from generic
records, services, or decisions, plus projection freshness and candidate-store
boundaries. It preserves canonical markdown authority and exposes validation
boundaries and authority limits in `agent_handoff`.

## UX Quality

Pass.

Current primitives remain available for drill-down, but the promoted report
removes the surprising structured-store decision ceremony where an agent must
manually combine record lookup, typed lookup, projection freshness, provenance
policy, and candidate-surface comparison.

## Performance

The action is bounded by one selected-domain lookup and one projection-state
lookup with the requested limit. It does not scan raw files, import external
data, build indexes, or run background projection jobs.

## Evidence Posture

The reduced proof is code and test evidence only. It does not commit raw
databases, generated corpora, private records, spreadsheets, health data,
finance data, inventory exports, or machine-specific paths.

Relevant tests:

- `TestRetrievalTaskStructuredStoreReportIsReadOnly`
- `TestRetrievalTaskStructuredStoreReportRejectsMissingFilter`
- `TestRunnerDocumentAndRetrievalJSONRoundTrip`

## Decision

Select `structured_store_report` as the promoted surface for this track. Keep
future independent stores blocked until a domain-specific track proves schema,
write approval, correction, duplicate handling, provenance, freshness,
local-first behavior, and markdown reconciliation.
