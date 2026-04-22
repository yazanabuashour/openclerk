# ADR: Knowledge Configuration v1

## Status

Accepted as the v1 production contract for OpenClerk-compatible knowledge
vaults on the current AgentOps surface.

This ADR defines the contract agents can rely on through the installed
`openclerk` JSON runner and `skills/openclerk/SKILL.md`. It does not require a
committed manifest file. `oc-za6.2` will decide whether convention-first layout
remains sufficient, whether runner-derived configuration inspection is enough,
or whether a small runner-visible configuration artifact should be promoted.

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
required for v1. If future evals show that convention-first layout cannot be
validated or explained well enough through runner JSON, `oc-za6.2` should
promote a small runner-visible configuration model or artifact.

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
- `provenance_events` exposes derivation, update, invalidation, and refresh
  history.
- `projection_states` exposes current derived freshness and version state.

These actions are the production agent interface for explaining and validating
knowledge layout. Unsupported lower-level workflows should be rejected rather
than routed around the runner.

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
