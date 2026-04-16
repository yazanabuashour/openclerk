# Agent Knowledge Plane

## Summary

OpenClerk is positioned as a single-surface agent-facing knowledge plane, not a domain-specific health application and not a menu of user-facing backend variants.

The product model is:

- canonical docs remain markdown in the vault
- graph traversal is a derived docs capability
- promoted records are selective structured domains, not the default storage shape
- provenance and projection-state APIs make derivation and freshness inspectable
- memory and routing are intentionally deferred until the docs and truth-sync layers are benchmarked

## Public contract

The public SDK surface is:

- [`client/local`](../../client/local), including the code-first embedded facade
- [`client/openclerk`](../../client/openclerk), for generated OpenAPI fallback work
- [`openapi/v1/openclerk.yaml`](../../openapi/v1/openclerk.yaml)

The public API is organized by capability, not implementation variant:

- docs and search are core
- graph is an optional derived-docs capability
- records are an optional promoted-domain capability
- provenance exposes truth-sync inspection

## Canonical and derived layers

### Docs

Canonical docs are markdown files under the vault with stable `doc_id`, `chunk_id`, vault-relative `path`, headings, and parsed frontmatter metadata.

The docs layer now exposes:

- document registry listing over stable ids and paths
- metadata-aware listing and search filters
- safe write operations for canonical docs
- docs-centric link expansion
- citation-bearing retrieval results

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

- append-only event inspection through `GET /v1/provenance/events`
- current projection-state inspection through `GET /v1/provenance/projections`

Current event and projection semantics are intentionally minimal:

- document create/update events
- projection invalidation events
- projection refresh events
- record extraction events
- fresh/stale projection state for current derived outputs

## Implementation variants

The repo still keeps `fts`, `hybrid`, `graph`, and `records` clients and examples for eval-driven implementation comparison.

Those variants are implementation fixtures, not the preferred application-facing SDK.

## Out of scope for this rewrite

- Mem0 or other long-term memory integration
- autonomous routing across docs, records, and memory
- treating the current generic records projection as the final structured model
- hiding derivation behind opaque heuristics
