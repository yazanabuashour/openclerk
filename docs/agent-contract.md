# OpenClerk Agent Contract

This is the compact contract for coding agents using OpenClerk. It names the
safe surface; runner help and architecture docs carry the details.

## Product Surface

- Installed `openclerk` runner.
- Thin `skills/openclerk/SKILL.md` router.
- JSON in / JSON out actions under `config`, `document`, `retrieval`,
  `clerk`, `module`, and `capabilities`.
- Local vault and local storage only.

## Authority Model

- Canonical Markdown is authority.
- Indexes, projections, graph state, records, semantic recall, and OCR output
  are derived or optional recall layers.
- Citations, source refs, provenance, duplicate handling, and projection
  freshness are evidence, not decoration.
- Optional modules do not become truth and must be installed, enabled, and
  verified before use.

## Agent Loop

1. Before work, run `openclerk inspect`.
2. Request task context with `openclerk clerk context_pack`.
3. Read cited docs, decisions, and stale-context warnings.
4. Do the work outside OpenClerk.
5. After work, pass an explicit session note, handoff, or inbox artifact to the
   read-only `openclerk clerk` planning surface.
6. Apply durable writes only through approved `openclerk document` lifecycle
   APIs.

## Allowed Reads

- `openclerk inspect`, runner help, and `openclerk capabilities`.
- Citation-bearing `openclerk retrieval` results.
- Read-only `openclerk document` planning, validation, layout, duplicate, and
  lifecycle reports.
- Read-only `openclerk clerk` context and session-planning reports.

## Allowed Writes

- Proposed paths, titles, body previews, source refs, and next runner requests
  in an agent response.
- Durable Markdown only after approval, through `openclerk document`.
- Configuration only through explicit `openclerk config` or `openclerk module`
  actions.

## Approval Boundary

Approval is required before creating, appending, replacing, moving, renaming,
promoting, fetching into, or configuring durable local state. Read, fetch,
inspect, and plan permission is not durable-write approval.

## Unsupported

- Direct SQLite access, raw vault inspection, source-built runner paths,
  module-cache inspection, hidden provider fallback, or unverified modules as
  agent workflows.
- Hosted services, daemons, background repair, cloud sync, remote HTTP APIs, or
  multi-user server contracts.
- Treating OCR, semantic search, graph state, records, indexes, or projections
  as independent authority.

## Minimal Commands

```bash
openclerk demo init --template codebase-decisions
openclerk inspect
openclerk clerk context_pack --task "change the auth callback behavior" --limit 5
openclerk clerk run --once \
  --inbox-path examples/knowledge-packs/agent-session-to-docs/handoffs/session.md \
  --task "summarize completed auth callback work into repo knowledge" \
  --limit 5
```

For pack layout, see [Knowledge Packs](knowledge-packs.md). For product
boundaries, see [Agent Knowledge Plane](architecture/agent-knowledge-plane.md)
and [Chronicler Lite Boundary](architecture/chronicler-boundary.md).
