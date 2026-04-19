# Agent Knowledge Plane

## Summary

OpenClerk is positioned as a single-surface agent-facing knowledge plane, not a domain-specific health application and not a menu of user-facing backend variants.

The production agent surface is the installed `openclerk` JSON runner. Agents
use task-shaped JSON for routine document and retrieval work; they do not need
to inspect implementation files, backend variants, or SQLite to operate the
knowledge plane.

It is also positioned as infrastructure for persistent agent-maintained knowledge: useful synthesis should become cited, inspectable markdown instead of being rediscovered from scratch on every query.

The product model is:

- canonical docs remain markdown in the vault
- source-linked synthesis can live in markdown when it preserves source refs and freshness
- graph traversal is a derived docs capability
- promoted records are selective structured domains, not the default storage shape
- provenance and projection-state APIs make derivation and freshness inspectable
- memory and routing are intentionally deferred until the docs and truth-sync layers are benchmarked

## Public contract

The public surface is:

- the installed `openclerk` runner for production agent workflows
- [`client/local`](../../client/local), including the code-first embedded facade

The public API is organized by capability, not implementation variant:

- docs and search are core
- graph is an optional derived-docs capability
- records are an optional promoted-domain capability
- provenance exposes truth-sync inspection

## Canonical and derived layers

### Docs

Canonical docs are markdown files under the vault with stable `doc_id`, `chunk_id`, vault-relative `path`, headings, and parsed frontmatter metadata.

The docs layer also supports source-linked synthesis: topic pages, entity pages, comparisons, overview notes, and filed answers that compile existing evidence into reusable markdown. These pages are durable knowledge artifacts, but they do not outrank the canonical source docs or promoted records they cite.

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

It should eventually be replaced or extended with explicit domain models where structured state clearly outperforms plain docs.

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

Production evals compare the runner-first `skills/openclerk` surface against an
archived SDK-oriented baseline. Reports track correctness, tool calls, assistant
calls, wall time, token use, stale surface inspection, module-cache inspection,
broad repo search, direct SQLite access, and raw log references using
`<run-root>` placeholders.

The provisional architecture decision and adoption gates are recorded in
[`eval-backed-knowledge-plane-adr.md`](eval-backed-knowledge-plane-adr.md).

## Out of scope for this rewrite

- Mem0 or other long-term memory integration
- autonomous routing across docs, records, and memory
- treating the current generic records projection as the final structured model
- hiding derivation behind opaque heuristics
