# Agent Knowledge Plane

## Summary

OpenClerk is positioned as a single-surface agent-facing knowledge plane, not a domain-specific health application and not a menu of user-facing backend variants.

The first principle is AgentOps: the installed `openclerk` JSON runner plus
`skills/openclerk/SKILL.md`. Agents use task-shaped JSON for routine document
and retrieval work. They do not inspect implementation files, backend variants,
HTTP server internals, source-built command paths, module caches, or SQLite to
operate the knowledge plane.

It is also positioned as infrastructure for persistent agent-maintained knowledge: useful synthesis should become cited, inspectable markdown instead of being rediscovered from scratch on every query.

The product model is:

- canonical docs remain markdown in the vault
- source-linked synthesis is the active next build slice for durable
  agent-maintained wiki pages
- graph traversal is a derived docs capability behind AgentOps
- promoted records are selective structured domains behind AgentOps, not the
  default storage shape
- provenance and projection-state runner actions make derivation and freshness
  inspectable
- memory and routing remain deferred until the docs, synthesis, and truth-sync
  layers are reliable through AgentOps

## Public contract

The public product surface is:

- the installed `openclerk` runner for production agent workflows
- the Agent Skills-compatible `skills/openclerk/SKILL.md` guidance

The public runner contract is organized by capability, not implementation
variant:

- docs and search are core
- graph is an optional derived-docs capability
- records are an optional promoted-domain capability
- provenance exposes truth-sync inspection

## Canonical and derived layers

### Docs

Canonical docs are markdown files under the vault with stable `doc_id`, `chunk_id`, vault-relative `path`, headings, and parsed frontmatter metadata.

The docs layer also supports source-linked synthesis: topic pages, entity
pages, comparisons, overview notes, and filed answers that compile existing
evidence into reusable markdown. Prototype synthesis pages live under
`notes/synthesis/`, carry `type: synthesis`, `status: active`, `freshness:
fresh`, and single-line comma-separated `source_refs` frontmatter, and include
`## Sources` plus `## Freshness` sections. These pages are durable knowledge
artifacts, but they do not outrank the canonical source docs or promoted
records they cite.

The active synthesis lifecycle workflow is:

- search canonical sources before writing synthesis
- list `notes/synthesis/` candidates before creating a new synthesis page
- retrieve an existing synthesis document before updating it
- prefer section replacement or append over duplicate creation
- preserve source refs, citations, `## Sources`, and `## Freshness`
- repair stale or contradictory claims by naming current sources and superseded
  sources
- inspect provenance and projection freshness when synthesis depends on
  promoted records or services

The docs layer now exposes:

- document registry listing over stable ids and paths
- metadata-aware listing and search filters
- safe write operations for canonical docs
- docs-centric link expansion
- citation-bearing retrieval results

### LLM Wiki alignment

Karpathy's LLM Wiki pattern maps cleanly onto OpenClerk, but OpenClerk should implement it as a provenance-backed docs workflow rather than a literal clone.

| LLM Wiki concept | OpenClerk mapping |
| --- | --- |
| Raw sources | canonical source docs and assets |
| Wiki | source-linked synthesis and accepted canonical notes |
| Schema | repo docs plus `skills/openclerk` guidance |
| `index.md` | search, metadata filters, graph neighborhoods, and optional index notes |
| `log.md` | provenance events, projection states, and optional human-readable activity notes |

The shared idea is that agents should maintain summaries, links, contradiction notes, and filed answers so knowledge compounds over time. The OpenClerk-specific constraint is that synthesis must stay inspectable through stable ids, citations, provenance events, and projection freshness. It should not become an opaque second truth system.

### Graph

Graph state is derived from markdown links and chunk/document relationships. It remains source-linked and refreshable from canonical docs.

The graph layer must not become a second truth system.

### Records

The current records projection is still a baseline prototype:

- it uses `entity_*` frontmatter plus a `Facts` section
- it rebuilds on canonical updates
- it is suitable for evals and promoted-domain experiments

The first explicit promoted-domain prototype is the service registry:

- it uses dedicated service projection tables rather than generic entity rows
- canonical markdown service docs remain the source of truth
- `services_lookup` and `service_record` expose typed runner retrieval behavior
- service provenance and projection states make derivation and freshness
  inspectable

The generic records projection remains backward compatible and should be
extended only where structured state clearly outperforms plain docs.

### Provenance and truth sync

The provenance layer now exposes:

- append-only event inspection through retrieval runner tasks
- current projection-state inspection through retrieval runner tasks

Current event and projection semantics are intentionally minimal:

- document create/update events
- projection invalidation events
- projection refresh events
- record extraction events
- fresh/stale projection state for current derived outputs

## Agent evals

Production evals are regression gates for the AgentOps contract and
knowledge-model behavior. Reports track correctness, tool calls, assistant
calls, wall time, token use, stale surface inspection, module-cache inspection,
broad repo search, direct SQLite access, source-built runner usage, and raw log
references using `<run-root>` placeholders.

The current architecture direction is recorded in
[`eval-backed-knowledge-plane-adr.md`](eval-backed-knowledge-plane-adr.md).

## Out of scope for this rewrite

- Mem0 or other long-term memory integration
- autonomous routing across docs, records, and memory
- treating the current generic records projection as the final structured model
- hiding derivation behind opaque heuristics
