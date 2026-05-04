---
decision_id: decision-structured-data-canonical-stores
decision_status: accepted
decision_scope: structured-data-canonical-stores
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/architecture/structured-data-canonical-stores-adr.md, docs/evals/structured-data-canonical-stores-poc.md, docs/evals/results/ockp-structured-data-canonical-stores.md, docs/architecture/knowledge-configuration-v1-adr.md
---

# Structured Data And Canonical Stores Promotion Decision

## Decision

Accept `structured_store_report` as the promoted read-only surface for
`oc-uj2y.4`.

Do not promote independent non-document canonical stores, durable structured
writes, external connectors, or domain-specific tables from this track.

## Safety Pass

Pass. The selected report is read-only and forbids durable writes, independent
canonical tables, direct SQLite reads, raw storage inspection, HTTP/MCP
bypasses, source-built runners, unsupported transports, external connectors,
hidden authority ranking, and projection mutation.

## Capability Pass

Pass. The report exposes current structured evidence for the selected domain:

- generic promoted records
- service records
- decision records
- projection freshness
- candidate-store comparison
- validation boundaries
- authority limits
- `agent_handoff`

## UX Quality

Pass. A normal user can ask for structured-store decision support through one
runner action instead of manually choreographing record lookup, typed lookup,
projection freshness, provenance policy, candidate comparison, and final-answer
boundaries.

## Conditional Implementation

Implemented in this epic:

- runner JSON action `structured_store_report`
- request object `structured_store`
- response object `structured_store`
- candidate-surface comparison fields
- read-only records, services, decisions, and projection-state evidence
- validation for missing filters, invalid domains, and limits
- runner help and CLI JSON round-trip coverage
- README promoted-action guidance
- OpenClerk skill action guidance
- unit tests for read-only behavior and validation

No storage schema, projection lifecycle, migration, background job, durable
write path, or external connector is added.

## Iteration Gate

Future domain-specific stores should compare at least three candidate shapes:

- markdown-backed promoted records plus report guidance
- a narrow typed runner action for one stable domain
- an independent canonical store or import adapter

Promotion requires a domain-specific schema, correction lifecycle, duplicate
handling, provenance, freshness, local-first behavior,
approval-before-durable-write, and explicit authority reconciliation with
canonical markdown. If none of the candidate shapes preserve these boundaries,
record `none viable yet`.
