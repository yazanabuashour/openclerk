# OpenClerk Context

## Domain Vocabulary

**AgentOps** — The supported agent-facing operating model: the installed
`openclerk` JSON runner plus `skills/openclerk/SKILL.md`. Routine agents use
runner JSON and do not inspect SQLite, vault files, implementation packages,
source-built runner paths, HTTP/MCP internals, or backend variants.

**Runner** — The local `openclerk` command that accepts task-shaped JSON on
stdin and returns structured JSON on stdout. The public runner surface is
organized into `document` and `retrieval` domains.

**Vault** — The runner-configured markdown knowledge root. Paths returned to
agents are vault-relative.

**Canonical Doc** — A markdown document registered by the runner with stable
document ids, chunk ids, headings, metadata, timestamps, and citations.
Canonical docs are the default authority for local knowledge.

**Source Doc** — A canonical doc that represents source authority for later
answers, synthesis pages, or promoted records. Source docs conventionally live
under `sources/` and may be created by source ingestion runner actions.

**Synthesis Doc** — Durable compiled knowledge that summarizes or reconciles
canonical evidence. Synthesis docs conventionally live under `synthesis/`,
preserve source refs, and remain subordinate to canonical docs and promoted
records.

**Provenance Event** — Runner-visible append-only history that explains
document changes, source changes, projection invalidations, projection
refreshes, and record extraction.

**Projection State** — Runner-visible current freshness state for derived
outputs such as graph, records, services, decisions, or synthesis.

**Promoted Record** — A typed or generic structured projection derived from
canonical markdown when a domain benefits from typed lookup. Promoted records
do not replace the canonical markdown authority they cite.

**Decision Record** — A promoted record for ADRs and decision notes. Decision
records expose stable decision ids, status, scope, owner, supersession, source
refs, citations, and freshness.

## Architecture Notes

- Public agent behavior lives at the runner JSON surface.
- Storage behavior lives behind the domain store interface.
- SQLite is the local adapter for the store interface, not a routine agent
  surface.
- Source ingestion owns fetching, extraction, note generation, asset mutation,
  rollback, provenance, citations, and freshness visibility behind a single
  runner action.
