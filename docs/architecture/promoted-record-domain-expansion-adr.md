---
decision_id: adr-promoted-record-domain-expansion
decision_title: Promoted Record Domain Expansion
decision_status: accepted
decision_scope: promoted-records
decision_owner: platform
---
# ADR: Promoted Record Domain Expansion

## Status

Accepted as a deferred-capability evidence track.

This ADR frames promoted record domain expansion as evidence gathering only. It
does not add runner actions, typed record domains, schemas, storage behavior,
migrations, skill behavior, or public APIs.

## Context

OpenClerk already exposes generic promoted records through `records_lookup` and
`record_entity`, and typed service and decision records through
`services_lookup`, `service_record`, `decisions_lookup`, and
`decision_record`. Canonical markdown remains authoritative; promoted records
are derived projections with citations, provenance, and freshness.

The question for this track is whether domains beyond services and decisions
need their own promoted runner surface, or whether generic records plus current
document/retrieval primitives are enough. The track tests both:

- `capability_gap`: whether existing `openclerk document` and
  `openclerk retrieval` actions are structurally insufficient for a
  domain-shaped record workflow.
- `ergonomics_gap`: whether existing actions can express the workflow but are
  too slow, too scripted, too brittle, too guidance-dependent, or too costly for
  routine AgentOps use.

## Decision

Use targeted ADR, POC, eval, and decision artifacts before any implementation
work. The default outcome is to keep expanded promoted record domains
deferred/reference unless targeted evidence proves a capability gap or repeated
ergonomics gap.

The current public surface remains:

- `openclerk document`
- `openclerk retrieval`

Canonical markdown remains the authority for record identity and domain facts.
Generic records, services, and decisions remain derived projections that must
preserve citations, provenance, freshness, and local-first AgentOps operation.

## Options

| Option | Description | Promotion posture |
| --- | --- | --- |
| Keep current primitives | Use search, `list_documents`, `get_document`, `records_lookup`, `record_entity`, provenance, and records projection freshness. | Default/reference if natural and scripted pressure pass with acceptable ergonomics. |
| Add a narrow domain surface | Add a promoted typed action for one domain, such as policy records, that packages lookup, detail, citations, provenance, and freshness. | Consider only if repeated evidence shows generic records are structurally insufficient or too costly for routine use. |
| Add independent domain truth | Store or infer typed domain facts outside canonical markdown authority. | Kill unless it preserves canonical authority, citations, provenance, freshness, local-first operation, and no-bypass invariants. |

## Invariants

Any future promoted surface must preserve:

- AgentOps-only routine operation through installed runner JSON.
- Canonical markdown authority for record identity, facts, and domain meaning.
- Citations, source refs, or stable source identifiers for source-sensitive
  claims.
- Inspectable provenance and records projection freshness.
- Local-first operation.
- No broad repo search, direct vault inspection, direct SQLite, source-built
  runner paths, HTTP/MCP bypasses, backend variants, module-cache inspection,
  or ad hoc lower-level transports for routine tasks.
- Final-answer-only handling for invalid no-tools requests.

## Non-Goals

This ADR does not:

- define a promoted policy, contact, asset, project, or other typed record
  action
- add storage tables, migrations, parser pipelines, background jobs, indexes,
  or public APIs
- make generic or typed records more authoritative than canonical markdown
- relax citation, provenance, freshness, duplicate-prevention, validation, or
  bypass requirements

## Promotion Gates

Promotion via `capability_gap` requires repeated scripted-control failures
showing current document/retrieval primitives cannot safely express a
domain-shaped record workflow while preserving authority, citations,
provenance, freshness, and bypass boundaries.

Promotion via `ergonomics_gap` requires repeated natural-intent failures or
unacceptable UX cost where scripted controls still pass. Evidence must show
high step count, latency, prompt brittleness, retries, wrong domain target
selection, missing citations, skipped provenance/freshness checks, or
workflow-specific guidance dependence that a proposed surface would reduce
without weakening the invariants above.

Defer when failures are guidance, answer-contract, eval coverage, data hygiene,
partial evidence, one-off ergonomics pressure, or insufficient scripted-control
evidence.

Keep as reference when current primitives pass with acceptable ergonomics and a
domain-specific surface would mostly repackage generic records and document
retrieval.

Kill the candidate if it creates independent record truth, hides citations,
hides provenance or freshness, weakens canonical markdown authority, silently
routes around generic records, or encourages bypasses.

## Evidence Plan

The POC comparison is
[`../evals/promoted-record-domain-expansion-comparison-poc.md`](../evals/promoted-record-domain-expansion-comparison-poc.md).
The targeted reduced report is
[`../evals/results/ockp-promoted-record-domain-expansion-pressure.md`](../evals/results/ockp-promoted-record-domain-expansion-pressure.md).
The final promotion decision is
[`promoted-record-domain-expansion-promotion-decision.md`](promoted-record-domain-expansion-promotion-decision.md).
