---
decision_id: adr-knowledge-configuration-v1
decision_title: Knowledge Configuration v1
decision_status: accepted
decision_scope: knowledge-configuration
decision_owner: platform
---
# ADR: Knowledge Configuration v1

## Status

Accepted as the v1 production contract for OpenClerk-compatible knowledge
vaults on the current AgentOps surface.

This ADR defines the contract agents can rely on through the installed
`openclerk` JSON runner and `skills/openclerk/SKILL.md`. It does not require a
committed manifest file. `oc-za6.2` promoted runner-derived layout inspection
through `inspect_layout` and kept the underlying v1 layout convention-first.

## Context

OpenClerk is a local-first, agent-facing knowledge plane. Routine agents need
to create, inspect, retrieve, synthesize, and explain local knowledge without
reading implementation files, backend variants, module caches, SQLite directly,
HTTP/MCP internals, or source-built command paths.

The v1 configuration contract therefore has to be visible through AgentOps:
the installed `openclerk document` and `openclerk retrieval` JSON runners plus
the shipped skill instructions. A production agent should be able to explain
where the vault is, which markdown files are canonical, which derived records
or projections are fresh, and which source evidence supports a synthesis page
using only runner JSON results.

The existing architecture direction already accepts canonical markdown docs,
source-linked synthesis, promoted records, provenance events, projection
freshness, and eval evidence as the path toward durable agent-maintained
knowledge. This ADR turns those ideas into the v1 knowledge configuration
contract.

## Decision

OpenClerk knowledge configuration v1 is runner-visible and convention-first.

An OpenClerk-compatible vault is defined by:

- runner-resolved storage paths for the effective data directory, database
  path, and vault root
- vault-relative markdown paths under the vault root
- document frontmatter and section conventions for first-class document kinds
- runner-maintained document registry entries with stable ids, chunk ids,
  headings, metadata, and citations
- derived SQLite projections for graph, records, services, provenance, and
  projection freshness that remain explainable through runner JSON

The conventional layout is part of the product contract, but not every path is
hard-coded:

- `vault/` is the conventional markdown root under the effective data
  directory.
- `notes/sources/` is the conventional home for canonical source docs.
- `notes/synthesis/` is the conventional home for source-linked synthesis.
- record-shaped and service-shaped markdown conventions feed promoted record
  projections.
- `source_refs`, `## Sources`, and `## Freshness` are conventional synthesis
  evidence and freshness fields.
- the effective data directory, database path, and vault root are configurable
  through supported runner path resolution and environment/config inputs.

No committed manifest, schema file, or separate configuration document is
required for v1. `inspect_layout` is the runner-visible configuration model:
it explains the effective conventions and reports invalid or incomplete
layouts through JSON checks without making a committed artifact authoritative.

## V1 Concepts

Canonical docs are markdown files under the vault root. They have stable
`doc_id`, `chunk_id`, vault-relative `path`, headings, parsed frontmatter
metadata, timestamps, and citations exposed by runner results. Canonical docs
are the default source of truth for local knowledge.

Canonical source docs are canonical docs used as source authority for later
answers, synthesis pages, or promoted records. They conventionally live under
`notes/sources/`, but source authority comes from runner-visible citations,
paths, chunk ids, metadata, provenance, and freshness, not from folder naming
alone.

Synthesis docs are durable compiled knowledge pages that summarize or reconcile
canonical evidence. They conventionally live under `notes/synthesis/` with
frontmatter containing `type: synthesis`, `status: active`, `freshness: fresh`,
and single-line comma-separated `source_refs`. They include `## Sources` and
`## Freshness` sections. Synthesis docs do not outrank the canonical source
docs or promoted records they cite.

Promoted records are selective structured projections derived from canonical
markdown docs when a domain benefits from typed lookup. Generic record-shaped
docs use record frontmatter and facts. The service registry is the first typed
promoted-domain prototype, exposed through service-specific lookup and record
actions. Promoted records remain derived from canonical markdown and must keep
citations and freshness inspectable.

Provenance events are runner-visible append-only facts about document changes,
source changes, projection invalidations, projection refreshes, and record or
service extraction. They explain how a document, record, service, or projection
was created or refreshed.

Projection freshness is the runner-visible current state for derived outputs.
Projection states report freshness, projection version, observed time, and
details for derived graph, record, service, or synthesis outputs. The
`synthesis` projection is the v1 freshness contract for source-linked synthesis
and can explain current, stale, missing, or superseded source refs.

Eval evidence is the committed reduced proof that the production AgentOps
surface satisfies the contract. Reports under `docs/evals/results/` use
repo-relative paths and neutral artifact placeholders such as `<run-root>`.
They gate runner use, source authority, citations, freshness, final-answer-only
rejections, and bypass prevention.

## OpenClerk-Compatible Vault

A vault is OpenClerk-compatible for v1 when an agent can use runner JSON to
resolve paths, validate or create markdown documents, list and retrieve
canonical docs, find source evidence, inspect derived records, inspect
provenance, and inspect projection freshness.

Compatibility requires these observable properties:

- `openclerk document` can resolve the effective data directory, database path,
  and vault root.
- documents use vault-relative paths and markdown bodies accepted by
  `validate` or `create_document`.
- canonical docs are discoverable through `list_documents`, `get_document`,
  and `search` without direct vault inspection.
- source-sensitive answers can cite runner-visible paths, `doc_id`, `chunk_id`,
  headings, line ranges, source refs, or provenance.
- synthesis pages under `notes/synthesis/` preserve `type: synthesis`,
  `status: active`, `freshness: fresh`, `source_refs`, `## Sources`, and
  `## Freshness`.
- promoted records and services remain derived from canonical markdown and are
  inspectable through records, services, provenance, and projection-state
  runner actions.
- stale, missing, superseded, or refreshed derived knowledge is explainable
  through projection states and provenance events.

The database is an implementation detail for the runner. It may contain the
document registry, chunks, search indexes, links, projections, provenance, and
promoted records, but routine agents must not query it directly.

## Runner Contract

The v1 document runner actions are:

- `validate` checks document request shape without writing a document.
- `resolve_paths` exposes the effective OpenClerk data, database, and vault
  paths.
- `inspect_layout` exposes the convention-first layout contract, conventional
  prefixes, first-class document kinds, and pass/warn/fail validation checks.
- `create_document` writes a new canonical markdown document and registers it.
- `list_documents` exposes document registry entries by path prefix or metadata.
- `get_document` returns a canonical document by stable `doc_id`.
- `append_document` appends durable markdown content to an existing document.
- `replace_section` replaces one named markdown section while preserving the
  rest of the document.

The v1 retrieval runner actions are:

- `search` finds source-grounded document chunks with citations.
- `document_links` exposes outgoing markdown links and incoming backlinks.
- `graph_neighborhood` exposes derived docs graph context.
- `records_lookup` finds promoted generic record entities.
- `record_entity` returns one promoted generic record entity.
- `services_lookup` finds promoted service records.
- `service_record` returns one promoted service record.
- `decisions_lookup` finds promoted decision records with status, scope, owner,
  and text filters.
- `decision_record` returns one promoted decision record by stable decision ID.
- `provenance_events` exposes derivation, update, invalidation, and refresh
  history.
- `projection_states` exposes current derived freshness and version state.

These actions are the production agent interface for explaining and validating
knowledge layout. Unsupported lower-level workflows should be rejected rather
than routed around the runner.

## `oc-za6.2` POC Decision

Decision: promote runner-derived layout inspection and keep v1
convention-first. A committed manifest or separate config artifact is not
needed for v1.

The POC compared three options:

- convention-first only: paths, frontmatter, and discovery are enough for
  routine operation, but agents needed too many separate reads to explain why a
  layout is valid or invalid.
- runner-derived inspection: `inspect_layout` can explain the configured
  layout and report incomplete synthesis, source refs, record, and service
  identity conventions through one JSON result.
- committed config artifact: deferred or killed for v1 because it would create
  a second source of truth without improving canonical markdown authority.

The promoted model is therefore not a manifest. It is a runner-visible
inspection result derived from canonical markdown registry state. Targeted
AgentOps evidence is recorded in
`docs/evals/results/ockp-layout-configuration.md`: both
`configured-layout-explain` and `invalid-layout-visible` passed while
preserving the no broad repo search, no direct SQLite, and no source-built
runner invariants.

## `oc-za6.3` POC Decision

Decision: keep the existing document and retrieval actions for source-linked
synthesis maintenance. Do not promote a dedicated synthesis/compiler action for
v1.

The POC pressure-tested the current workflow against candidate-selection,
multi-source creation, stale repair, mixed records/synthesis, and resumed
multi-turn drift repair. The selected pressure scenarios required agents to
search source evidence, list `notes/synthesis/` candidates, retrieve existing
synthesis before editing, inspect synthesis projection freshness where
relevant, preserve single-line `source_refs`, keep `## Sources` and
`## Freshness`, and update without duplicate synthesis pages.

Targeted AgentOps evidence is recorded in
`docs/evals/results/ockp-synthesis-compiler-pressure.md`. This was a targeted
run, not a full production-gate run: all 12 selected synthesis, pressure, and
contract-enforcement scenarios passed; the production gate remained false only
because unrelated scenarios were intentionally not selected. The selected run
also preserved the no broad repo search, no direct SQLite, no source-built
runner usage, final-answer-only invalid-request rejection, and citation/source
freshness invariants.

The deferred candidate action shape remains:

```json
{
  "action": "compile_synthesis",
  "synthesis": {
    "path": "notes/synthesis/example.md",
    "title": "Example",
    "source_refs": ["notes/sources/source-a.md", "notes/sources/source-b.md"],
    "body": "# Example\n\n## Summary\n...\n\n## Sources\n...\n\n## Freshness\n...",
    "mode": "create_or_update"
  }
}
```

Future work should revisit that shape only if repeated eval failures show the
document/retrieval workflow is structurally too many steps and directly causes
missed candidate discovery, missed freshness inspection, duplicate synthesis,
or dropped source refs.

## `oc-za6.4` POC Decision

Decision: promote decision and architecture records as the second typed
promoted domain after services.

Canonical markdown remains authoritative. Decision projection is driven by
frontmatter fields such as `decision_id`, `decision_title`, `decision_status`,
`decision_scope`, `decision_owner`, `decision_date`, `supersedes`,
`superseded_by`, and `source_refs`; ADR-like filenames are useful conventions
but are not required. `records/decisions/` is a conventional home, while ADRs
or decision notes under other paths are valid when the metadata is present.

The POC adds `decisions_lookup` and `decision_record` because decision-centric
tasks benefit from typed status/scope/owner filters, stable repeatable lookup,
update safety, citations, and supersession freshness. Plain docs search remains
useful for broad discovery, but it is weaker for questions such as "what is the
current accepted decision?" when old and current decisions coexist.

Decision projection freshness treats current decisions as fresh. Superseded
decisions or decisions with `superseded_by` are stale, with projection details
that expose the replacement IDs and freshness reason; replacement decisions
with `supersedes` remain fresh when their canonical markdown source is current.

Targeted AgentOps evidence is recorded in
`docs/evals/results/ockp-decision-records-poc.md`. The targeted run covers
decision-vs-doc precision, supersession freshness, no broad repo search, no
direct SQLite, no source-built runner usage, final-answer-only invalid-request
rejection, and preserved citations/source refs/freshness.

Follow-up hardening evidence for `oc-j0a` is recorded in
`docs/evals/results/ockp-decision-records-hardening.md`. The targeted partial
run covers migrated real ADR markdown under `docs/architecture/`, widened
decision text lookup, explicit `decision_record` supersession checks, fresh
decision projection states, provenance, and repo-relative citation paths. All
3 selected decision scenarios passed; the production gate remained false only
because unrelated scenarios and final-answer-only validation scenarios were not
selected in that partial run.

## `oc-za6.5` POC Decision

Decision: keep source-sensitive audit and contradiction-like workflows as a
reference pattern on existing provenance and freshness primitives. Do not
promote a broad semantic contradiction engine or a new audit runner action for
v1.

The POC stays narrow: agents search canonical sources, list existing synthesis
candidates, retrieve the target synthesis before editing, inspect synthesis
projection freshness and provenance, then repair stale synthesis without
creating duplicates. When current sources conflict and no supersession metadata
or other runner-visible source authority chooses a winner, agents should
explain the unresolved conflict with both source paths instead of asserting a
general contradiction result.

Targeted AgentOps evidence is recorded in
`docs/evals/results/ockp-source-sensitive-audit-poc.md`. The targeted run
covers stale audit repair, unresolved conflicting current sources, existing
stale-synthesis repair workflows, no broad repo search, no direct SQLite, no
source-built runner usage, final-answer-only invalid-request rejection, and
preserved citations/source refs/freshness. All 8 selected scenarios passed; the
production gate remained false only because unrelated scenarios were
intentionally not selected. No follow-up implementation issue is filed unless
future evals show repeated failures that the existing provenance/freshness
workflow cannot address.

## `oc-za6.6` POC Decision

Decision: keep richer graph semantics as a reference pattern and do not promote
a semantic-label graph layer for v1.

The POC keeps relationship meaning in canonical markdown. Agents can search for
relationship words such as requires, supersedes, related to, and
operationalizes; inspect the source document; expand outgoing links and
incoming backlinks; inspect the derived graph neighborhood; and verify graph
projection freshness. The derived graph remains structural and cited: edge
kinds such as `links_to` and `mentions` explain navigation, while canonical
markdown remains the source of semantic relationship authority.

Targeted AgentOps evidence is recorded in
`docs/evals/results/ockp-graph-semantics-reference-poc.md`. The targeted run
covers graph semantic-label pressure, canonical docs navigation, no broad repo
search, no direct SQLite, no source-built runner usage, final-answer-only
invalid-request rejection, and preserved citations/source refs/freshness. No
follow-up implementation issue is filed because the POC decision is reference
and deferred rather than promoted.

## Production-Valid for AgentOps

A vault is valid enough for production AgentOps use when:

- routine create, list, retrieve, search, synthesis, record, provenance, and
  freshness workflows are expressible through documented runner actions
- required request fields and invalid limits reject with understandable runner
  or final-answer-only behavior
- source-sensitive answers preserve citations, source refs, chunk ids, paths,
  provenance, or projection-state details
- synthesis is updated rather than duplicated when an existing page already
  covers the topic
- canonical docs and promoted records outrank synthesis
- derived graph, record, service, and synthesis outputs expose freshness or
  provenance sufficient to explain their relationship to canonical docs
- routine agents do not inspect implementation files, backend variants, module
  caches, generated server files, SQLite directly, HTTP/MCP transports, or
  source-built runner paths
- selected production AgentOps evals pass and their reduced reports are
  committed under `docs/evals/results/`

Production validity is a product and eval contract, not a promise that every
possible markdown folder is semantically complete. Missing conventions,
unsupported document kinds, stale projections, or weak citation evidence should
surface as runner-visible validation gaps, stale freshness, missing results, or
follow-up work.

## Non-Goals

Knowledge configuration v1 does not include:

- autonomous routing across docs, records, graph, and future memory
- memory-first `remember` or `recall` semantics
- a rich semantic graph as an independent truth system
- an HTTP or MCP daemon as a routine production agent surface
- multi-user corporate memory
- a required committed vault manifest
- a dedicated synthesis runner action
- direct SQLite, backend variant, module-cache, source-built runner, or ad hoc
  runtime workflows for routine agents
