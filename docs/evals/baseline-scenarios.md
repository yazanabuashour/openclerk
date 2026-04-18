# Baseline Scenarios

These scenarios define the initial eval task set for the unified OpenClerk
AgentOps runner and the archived SDK-oriented baseline.

## Source-Grounded Retrieval

- Create canonical notes with stable headings and exact terms.
- Verify search returns the correct `doc_id`, `chunk_id`, and citations.
- Verify path-prefix and metadata filters reduce scope correctly.

## Source Ingest And Synthesis

- Create a source-shaped canonical doc and a source-linked synthesis page that
  cites it.
- Add a second source that updates or challenges the synthesis.
- Verify the synthesis is updated rather than duplicated.
- Verify source-sensitive claims preserve citation paths, chunk ids, or explicit
  source refs.

## Docs Navigation

- Create linked notes.
- Verify document link expansion returns outgoing and incoming relationships.
- Verify graph neighborhood expansion stays source-linked to canonical docs.

## Promoted-Domain Lookup

- Create a canonical record-shaped doc with `entity_*` frontmatter and `Facts`.
- Verify promoted lookup returns the expected entity and citations.
- Update the canonical source and verify the derived projection refreshes.

## Provenance And Freshness

- Verify document create/update events are emitted.
- Verify projection invalidation and projection refresh events are visible.
- Verify projection-state reads expose current freshness and version markers.
- Verify synthesis pages can be traced back to the canonical docs or records
  they summarize.
