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
- post-v0.1.0 document history and review controls should make
  agent-authored durable edits inspectable, reviewable, and restorable
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
`synthesis/`, carry `type: synthesis`, `status: active`, `freshness:
fresh`, and single-line comma-separated `source_refs` frontmatter, and include
`## Sources` plus `## Freshness` sections. These pages are durable knowledge
artifacts, but they do not outrank the canonical source docs or promoted
records they cite.

The active synthesis lifecycle workflow is:

- search canonical sources before writing synthesis
- list `synthesis/` candidates before creating a new synthesis page
- retrieve an existing synthesis document before updating it
- inspect the `synthesis` projection state for existing synthesis documents
  before repairing stale claims
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

Post-v0.1.0 document lifecycle planning is recorded in
[`openclerk-document-post-v0.1.0.md`](openclerk-document-post-v0.1.0.md).
The direction is to add agent-visible document history, review, and rollback
semantics after the first release without replacing Git or adding a new public
runner action before eval evidence justifies it.

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

### Cognee alignment

Cognee is a useful external reference for graph/vector AI memory engines, but
it is not a markdown-canonical knowledge plane in the OpenClerk sense. Its
valuable lessons are retriever taxonomy, ontology grounding, temporal
retrieval, session memory, feedback weighting, and the operational cost of
coordinating graph, vector, relational, and cache stores.

OpenClerk should not adopt Cognee's `remember`/`recall` product surface,
memory-first canonical truth model, routine HTTP/MCP/Python bypasses, or graph
as an independent authority layer. Cognee-inspired ideas should enter
OpenClerk only as benchmark categories or internal implementation options that
preserve AgentOps, citations, provenance, freshness, and canonical
markdown/record authority.

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

The second typed promoted domain is decision records:

- ADRs and decision notes remain canonical markdown, independent of filename
  conventions
- `decisions_lookup` and `decision_record` expose stable decision IDs, title,
  status, scope, owner, date, supersession refs, source refs, and citations
- the `decisions` projection exposes freshness so superseded decisions are
  visible as stale while replacements remain fresh

The generic records projection remains backward compatible and should be
extended only where structured state clearly outperforms plain docs.

### Provenance and truth sync

The provenance layer now exposes:

- append-only event inspection through retrieval runner tasks
- current projection-state inspection through retrieval runner tasks

Current event and projection semantics are intentionally minimal:

- document create/update events
- source create/update events for canonical source docs
- projection invalidation events
- projection refresh events
- record extraction events
- decision extraction events
- fresh/stale projection state for current derived outputs
- synthesis projection state for source-linked synthesis pages, including
  current, superseded, missing, and stale source refs plus a freshness reason

## Agent evals

Production evals are regression gates for the AgentOps contract and
knowledge-model behavior. Reports track correctness, tool calls, assistant
calls, wall time, token use, stale surface inspection, module-cache inspection,
broad repo search, direct SQLite access, source-built runner usage, and raw log
references using `<run-root>` placeholders.

The current architecture direction is recorded in
[`eval-backed-knowledge-plane-adr.md`](eval-backed-knowledge-plane-adr.md).
The normative v1 knowledge configuration contract is recorded in
[`knowledge-configuration-v1-adr.md`](knowledge-configuration-v1-adr.md).
The final POC recommendation for the AgentOps knowledge-plane path is recorded
in
[`agentops-knowledge-plane-poc-decision.md`](agentops-knowledge-plane-poc-decision.md).
The promotion/defer/kill gate for memory, routing, semantic graph, broad
contradiction detection, and new public runner actions is recorded in
[`deferred-capability-promotion-gates.md`](deferred-capability-promotion-gates.md).
The `oc-jsg` decision to keep memory and routing as reference/deferred pressure
is recorded in
[`memory-routing-reference-decision.md`](memory-routing-reference-decision.md).
The post-v0.1.0 document lifecycle vision is recorded in
[`openclerk-document-post-v0.1.0.md`](openclerk-document-post-v0.1.0.md).
The deferred post-v1 design/eval spike for agent-chosen vault path selection is
recorded in
[`agent-chosen-vault-path-selection-adr.md`](agent-chosen-vault-path-selection-adr.md)
and
[`../evals/agent-chosen-path-selection-poc.md`](../evals/agent-chosen-path-selection-poc.md).

## Out of scope for this rewrite

- Mem0 or other long-term memory integration
- autonomous routing across docs, records, and memory
- treating the current generic records projection as the final structured model
- hiding derivation behind opaque heuristics
