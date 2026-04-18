# Baseline Scenarios

These scenarios define the initial eval task set used to compare the internal `fts`, `hybrid`, `graph`, `records`, and unified `openclerk` implementations.

Production agent eval guidance for the code-first SDK surface lives in
[`agent-production.md`](agent-production.md).

## Source-grounded retrieval

- Create canonical notes with stable headings and exact terms.
- Verify search returns the correct `doc_id`, `chunk_id`, and citations.
- Verify path-prefix and metadata filters reduce scope correctly.

## Source ingest and synthesis

- Create a source-shaped canonical doc and a source-linked synthesis page that cites it.
- Add a second source that updates or challenges the synthesis.
- Verify the synthesis is updated rather than duplicated.
- Verify source-sensitive claims preserve citation paths, chunk ids, or explicit source refs.
- Verify stale or contradicted synthesis is visible through freshness/provenance checks instead of silently winning over newer source evidence.

## Docs navigation

- Create linked notes.
- Verify docs-centric link expansion returns outgoing and incoming relationships.
- Verify graph neighborhood expansion stays source-linked to canonical docs.

## Promoted-domain lookup

- Create a canonical record-shaped doc with `entity_*` frontmatter and `Facts`.
- Verify promoted lookup returns the expected entity and citations.
- Update the canonical source and verify the derived projection refreshes.

## Provenance and freshness

- Verify document create/update events are emitted.
- Verify projection invalidation and projection refresh events are visible.
- Verify projection-state reads expose the current freshness and version markers.
- Verify synthesis pages can be traced back to the canonical docs or records they summarize.

## Write correctness

- Verify create, append, and replace-section operations preserve stable chunk identity for unaffected sections.
- Verify promoted-domain refresh behavior tracks canonical changes instead of silent dual-write drift.
- Verify useful query answers can be filed back as source-linked synthesis when they are reusable.
- Verify duplicate synthesis pages are avoided when an existing topic/entity/comparison page should be updated.

## Wiki health

- Verify contradiction linting surfaces conflicts between source docs and synthesis pages.
- Verify orphan or missing-cross-link checks find important synthesis pages with no inbound or outbound links.
- Verify missing-source-ref checks flag synthesis claims that lack evidence.
- Verify stale-source detection flags synthesis built from superseded canonical docs.
