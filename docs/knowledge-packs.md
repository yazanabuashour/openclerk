# OpenClerk Knowledge Packs

An OpenClerk knowledge pack is a portable, repo-relative vault layout plus
small walkthrough artifacts that are ready for the `openclerk` runner. It is
the example content shape; `openclerk clerk context_pack` is the runner report
that retrieves task context from a bound vault.

A knowledge pack is not a package manager, hosted service, daemon, server, sync
format, module registry, or second truth system. Canonical Markdown remains the
human-readable authority. OpenClerk storage, indexes, projections, graph state,
semantic recall, OCR output, and records remain derived or optional recall
layers.

## Folder Conventions

These folders are conventions, not required schema. The core vocabulary is:

- `sources/` for source notes, reviewed references, or imported material.
- `synthesis/` for durable source-linked summaries and runbooks.
- `decisions/` for ADRs and decision notes.

Common optional folders include `handoffs/` for session notes, `inbox/` for
reviewed candidate material, `runbooks/` for procedures, and `incidents/` for
timelines or postmortems. Use `tasks/` only when a consuming repo already
treats task-like notes as canonical knowledge.

## Optional Frontmatter

Do not add required pack frontmatter. Prefer fields that current OpenClerk
projections already understand when they fit naturally:

- `type: source`, `type: synthesis`, or `type: decision`
- `status: active`, `status: draft`, or `status: superseded`
- `source_refs: sources/example.md`
- `supersedes:` or `superseded_by:`
- decision fields such as `decision_id`, `decision_title`,
  `decision_status`, `decision_scope`, `decision_owner`, and `decision_date`

Frontmatter supports projection and retrieval ergonomics. It does not make the
database the pack.

## Agent Use

Run `openclerk inspect` before guessing, then request cited task context with
`openclerk clerk context_pack`. After work, use an explicit session note,
handoff, or inbox artifact with `openclerk clerk session_record_report`;
durable writes still go through approved `openclerk document` lifecycle APIs.

Agents should treat citations and source refs as evidence, avoid SQLite/raw
vault/module-cache/source-built bypasses as the workflow, treat semantic or OCR
output as candidate recall/extraction, and update existing durable documents
when duplicate evidence points to one.

## Reviewable Candidate Knowledge

Useful candidate outputs are concrete and approval-ready: repo-relative path,
title, type, body preview or section replacement, source refs, citations,
duplicate/update target, stale context notes, and exact next approved runner
request.

If evidence is ambiguous, keep it as an open question or `needs_review`
candidate. Unclear shorthand, OCR output, copied notes, or session summaries are
not canonical truth until reviewed and written through the document lifecycle.
