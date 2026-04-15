# Baseline Scenarios

These scenarios define the initial eval task set used to compare the internal `fts`, `hybrid`, `graph`, `records`, and unified `openclerk` implementations.

## Source-grounded retrieval

- Create canonical notes with stable headings and exact terms.
- Verify search returns the correct `doc_id`, `chunk_id`, and citations.
- Verify path-prefix and metadata filters reduce scope correctly.

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

## Write correctness

- Verify create, append, and replace-section operations preserve stable chunk identity for unaffected sections.
- Verify promoted-domain refresh behavior tracks canonical changes instead of silent dual-write drift.
