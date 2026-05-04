---
decision_id: adr-structured-data-canonical-stores
decision_status: accepted
decision_scope: structured-data-canonical-stores
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/architecture/agent-knowledge-plane.md, docs/architecture/knowledge-configuration-v1-adr.md, docs/architecture/promoted-record-domain-expansion-promotion-decision.md, docs/evals/structured-data-canonical-stores-poc.md, docs/evals/results/ockp-structured-data-canonical-stores.md
---

# Structured Data And Canonical Stores ADR

## Context

The `oc-uj2y.4` track evaluates structured facts that should not be answered
only through prose search: records, metrics, measurements, finance, inventory,
health-like observations, contacts, assets, structured preferences, and
time-series facts.

OpenClerk already has schema-backed derived projections for generic records,
services, decisions, provenance, and projection freshness. Canonical markdown
remains the current durable authority for record identity, facts, citations,
source refs, and human review.

Required reference URLs:

- docs/architecture/agent-knowledge-plane.md
- https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md
- https://mitchellh.com/writing/building-block-economy
- https://developers.openai.com/api/docs/guides/prompt-guidance
- https://openai.com/index/harness-engineering/
- https://developers.openai.com/api/docs/guides/embeddings
- https://developers.openai.com/api/docs/guides/retrieval
- https://docs.mem0.ai/open-source/overview

## Authority And Approval Boundaries

Read/fetch/inspect permission is enough for a runner report to inspect current
records, typed projections, provenance, and projection freshness. It is not
approval to create a new canonical table, migrate user data into a hidden
store, add a connector, or let derived rows outrank visible markdown-backed
records.

Durable structured writes require an exact promoted surface with:

- approval before write
- domain schema and validation
- correction and delete lifecycle
- duplicate handling
- provenance and freshness inspection
- local-first storage behavior
- citation or source-ref authority back to canonical markdown or approved
  records

## Projection Versus Independent Store

A report or projection is enough when the domain can be rebuilt from visible
canonical markdown, such as service records, decision records, generic facts,
and source-linked summaries. The projection may add query ergonomics, typed
filters, and freshness inspection, but it remains derived and repairable from
the docs layer.

An independent canonical store is justified only when markdown-backed records
are structurally insufficient for a domain: dense measurements, high-volume
time-series facts, correction-heavy observations, or imported records whose
identity and conflict semantics cannot be faithfully represented as prose plus
frontmatter. This ADR does not find enough domain-specific evidence to promote
one of those stores.

## Candidate Options

| Candidate | Safety | Capability | UX quality | Decision |
| --- | --- | --- | --- | --- |
| Current primitives only | Pass. Keeps current authority model. | Can inspect records, services, decisions, provenance, and projections. | Too ceremonial for structured-store decisions. | Keep as drill-down. |
| Domain-specific typed actions | Pass for existing services and decisions. | Strong for mature schema domains. | Good when a domain is repeated and stable. | Use selectively; do not add new domains here. |
| Independent SQLite canonical tables | Not proven. Requires write approval, correction, provenance, freshness, and markdown reconciliation. | Could help dense measurements or time series later. | Surprising if hidden tables outrank visible records. | Not promoted. |
| External domain stores/connectors | Not proven for routine local-first operation. | Useful import/reference candidate. | Adds provider, sync, and approval ceremony. | Not promoted. |
| `structured_store_report` | Pass. Read-only, runner-only, packages existing projections. | Exposes current structured evidence and candidate-store boundaries. | One action replaces records/projection/candidate-policy choreography. | Promote. |

## Decision

Promote `structured_store_report` as the read-only structured-store decision
surface.

This does not promote independent non-document canonical stores. The report
keeps current canonical authority in markdown-backed records, services, and
decisions while exposing schema-backed projections, projection freshness,
candidate-store guidance, validation boundaries, and authority limits.

## Non-Goals

- No durable structured write action.
- No independent metrics, measurements, health, finance, inventory, contacts,
  assets, preferences, or time-series canonical store.
- No direct SQLite or raw file workflow for routine agents.
- No external store connector or sync protocol.
- No hidden ranking that lets derived records outrank canonical markdown.

## Promotion And Kill Criteria

Future non-document canonical stores require repeated domain-specific evidence
that markdown-backed records are structurally insufficient. Promotion must name
an exact schema, JSON contract, correction lifecycle, duplicate handling,
provenance, freshness, local-first behavior, approval-before-durable-write, and
source authority model.

Kill any candidate that makes hidden tables or external stores outrank visible
canonical evidence, hides citations/provenance/freshness, or requires routine
direct storage access.

Safety, capability, and UX quality remain separate gates:

- Safety pass requires runner-only access, markdown/approved-record authority,
  provenance and freshness inspection, duplicate handling, and approval before
  durable writes.
- Capability pass requires a domain-specific reason that projections or
  reports cannot faithfully answer the workflow.
- UX quality pass requires the normal user surface to stay simpler than a menu
  of backend storage choices.

Remaining work is represented by linked beads:

- `oc-tnnw.2.2` POC for structured-store candidate evidence.
- `oc-tnnw.2.3` eval for safety, capability, and UX quality.
- `oc-tnnw.2.4` promotion decision.
- `oc-tnnw.2.5` conditional implementation only if promoted.
- `oc-tnnw.2.6` iteration and follow-up bead creation.
