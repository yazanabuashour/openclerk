# Baseline Scenarios

These scenarios define the eval task set for the installed OpenClerk AgentOps
runner/skill surface. They also define the proof obligations for the
eval-backed knowledge-plane ADR in
`docs/architecture/eval-backed-knowledge-plane-adr.md`.

## Source-Grounded Retrieval

- Create canonical notes with stable headings and exact terms.
- Verify search returns the correct `doc_id`, `chunk_id`, and citations.
- Verify path-prefix and metadata filters reduce scope correctly.

## Source Ingest And Synthesis

- Create a source-shaped canonical doc and a source-linked synthesis page that
  cites it.
- Add a second source that updates or challenges the synthesis.
- Verify the synthesis is updated rather than duplicated.
- Verify synthesis pages live under `notes/synthesis/`, include `type:
  synthesis`, `status: active`, `freshness: fresh`, `source_refs`, a
  `Sources` section, and a `Freshness` section.
- Verify source-sensitive claims preserve citation paths, chunk ids, or explicit
  source refs.
- Verify useful answer material can be filed back into durable markdown instead
  of remaining only in chat history.
- Verify LLM Wiki-style synthesis stays subordinate to canonical sources and
  does not create a second authority layer.

## Contradiction And Stale Synthesis

- Create an initial source and synthesis page with a cited claim.
- Add a later source that supersedes or contradicts the claim.
- Verify the agent finds the existing synthesis page before writing a new one.
- Verify the synthesis is updated with the newer evidence, contradiction note,
  or explicit stale-state language.
- Verify the agent retrieves the existing synthesis document before replacing
  or appending sections.
- Verify the final answer identifies which source is current when the prompt is
  source-sensitive.

## Docs Navigation

- Create linked notes.
- Verify document link expansion returns outgoing and incoming relationships.
- Verify graph neighborhood expansion stays source-linked to canonical docs.

## Promoted-Domain Lookup

- Create a canonical record-shaped doc with `entity_*` frontmatter and `Facts`.
- Verify promoted lookup returns the expected entity and citations.
- Create a canonical service-shaped doc and verify typed `services_lookup`
  returns service id, owner, status, interface, facts, and citations.
- Update the canonical source and verify the derived projection refreshes.
- Compare the service registry path against plain docs retrieval for the same
  service-centric task.
- Accept the promoted-domain path only when it improves precision, update
  safety, or structured lookup behavior without weakening citation correctness.

## Provenance And Freshness

- Verify document create/update events are emitted.
- Verify projection invalidation and projection refresh events are visible.
- Verify projection-state reads expose current freshness and version markers.
- Verify synthesis pages can be traced back to the canonical docs or records
  they summarize.
- Verify promoted-record synthesis inspects `records_lookup`,
  `provenance_events`, and `projection_states` before writing durable
  synthesis.

## Agent Surface Comparison

- Verify production tasks use `openclerk` rather than direct SQLite, backend
  variants, stale API paths, or ad hoc runtime programs.
- Verify routine attempts to bypass the OpenClerk runner through legacy
  source-built command paths or an unevaluated MCP-style path are rejected
  final-answer-only without tools.
- Compare the runner against CLI-style or MCP-style alternatives only when the
  alternative exposes equivalent task-shaped document and retrieval semantics.
- Accept a CLI or MCP adapter only if it matches runner correctness and improves
  a measured agent-behavior metric without increasing forbidden access patterns.
