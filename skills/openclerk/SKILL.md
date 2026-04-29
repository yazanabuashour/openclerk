---
name: OpenClerk
description: Use OpenClerk for local-first knowledge-plane tasks through the installed openclerk JSON runner. Bootstrap no-tools rule - if required fields are missing, if required retrieval, source, or video fields are missing, if document path, title, or body is missing and no faithful propose-before-create candidate can be formed from explicit user content, if a numeric limit is negative such as limit -3, or if the user asks to bypass the runner with SQLite, raw vault/file/repo inspection, HTTP, MCP, legacy or source-built paths, unsupported transports, backend variants, module-cache inspection, rg --files, find, ls, direct vault inspection, or repo search, this description is complete; Do not open this skill file, run commands, use tools, or call the runner; respond with exactly one no-tools assistant answer to name the missing fields and ask the user to provide them, or reject the invalid/unsupported workflow. For valid work, use only openclerk document or openclerk retrieval JSON.
license: MIT
compatibility: Requires local filesystem access and an installed openclerk binary on PATH.
---

# OpenClerk

Use this skill for routine local OpenClerk knowledge-plane tasks. The
production interface is AgentOps: this Skills-compatible skill plus the
installed JSON runner.

```bash
openclerk document
openclerk retrieval
```

## Core Guardrails

- Parallelize runner commands only for documented safe reads: `resolve_paths`,
  `list_documents`, `get_document`, `inspect_layout`, retrieval read actions,
  and `audit_contradictions` with `mode: "plan_only"`. Sequence all writes,
  including `init`, create/ingest/append/replace document actions, and
  `audit_contradictions` with `mode: "repair_existing"`.
- Answer routine OpenClerk requests only from runner JSON results. Use the
  configured environment; pass `--db` only when the user explicitly names a
  dataset.
- Treat every runner path as a vault-relative logical path, such as
  `notes/projects/example.md`, `sources/example.md`, or `synthesis/`. Never use
  storage roots, absolute paths, `.openclerk-eval/vault`, or OS-specific
  backslash/drive paths in runner JSON or committed OpenClerk document paths.
- Do not inspect source files, generated artifacts, backend variants,
  module-cache docs, SQLite, vault files, or `.openclerk-eval/vault` directly
  for routine tasks. Do not use repo search, `rg --files`, `find`, `ls`,
  `openclerk --help`, HTTP/MCP, legacy/source-built paths, unsupported
  transports, or external acquisition tools as substitutes for runner JSON.
- Missing required fields that cannot be handled by the proposal policy,
  invalid numeric limits, and bypass requests are final-answer-only: use no
  tools, no commands, and no runner call. Name the missing field(s) or reject
  the invalid/unsupported workflow.

For configuration diagnostics after upgrades or routine runner failures, run
`resolve_paths` first, then `inspect_layout` when layout state matters. These
show the effective database path, configured vault root, and convention-first
layout state without rebinding configuration.

If the user explicitly asks to initialize OpenClerk for an existing vault, run
`openclerk init --vault-root <vault-root>`. This is first-time setup or
intentional rebinding, not a routine knowledge task; do not use `init` to
repair ordinary document or retrieval calls before inspecting effective
configuration.

## Lifecycle Quick Rules

For accepted-edit repair, rollback/restore, review-state, or semantic-diff
requests, stay inside `openclerk document` and `openclerk retrieval`; there is
no public history, diff, review, restore, rollback, or lifecycle action.

- Rollback/restore: run retrieval `search`, then document `list_documents`,
  then `get_document`; update only the unsafe section with `replace_section`.
  Preserve the existing target path and inspect `provenance_events` plus
  `projection_states` after any write.
- Semantic diff: use the requested `list_documents.path_prefix`, first run
  retrieval `search`, inspect the current document and provenance, cite
  runner-visible source paths, and summarize without raw private diffs.
- Pending review: do not modify accepted knowledge. Create a separate pending
  review document only when the proposed change is explicit, then report that
  the accepted target did not change.
- Lifecycle answers should name relevant vault-relative paths, source evidence,
  provenance, and freshness when required, without exposing private raw diffs,
  storage roots, absolute paths, or backslash paths.

## No-Tools Handling Before Runners

Before runners, answer exactly once with no tools when required fields are
missing and no proposal exception applies, a limit is invalid, or the user asks
for a bypass named in Core Guardrails. Do not guess missing fields. Ask for
them by name, or reject the invalid/unsupported workflow and point back to the
OpenClerk runner contract.

Required-field rules:

- Document create/validate needs `document.path`, `document.title`, and
  `document.body` unless the propose-before-create policy can produce a
  faithful candidate from explicit user content.
- Requests that refer to missing prior context, such as "the links we
  discussed" or "that artifact", lack preservable body/source content.
- New source URL ingestion needs `source.url`, `source.path_hint`, and
  `source.asset_path_hint`; update mode needs `source.url`.
- Video/YouTube ingestion needs `video.url` and supplied transcript text; new
  video sources also need `video.path_hint`.
- Limits must be non-negative.

For unsupported workflows not covered by these rules, say the production
OpenClerk runner does not support that workflow yet.

Deferred-capability comparison, revisit, or promotion-decision questions are
valid runner-backed evidence tasks when the user asks what existing OpenClerk
documents and retrieval results can prove. For those requests, use the
installed `openclerk document` and `openclerk retrieval` JSON surfaces to
inspect runner-visible documents, citations/source refs, provenance, and
projection freshness. Treat memory transports, `remember`/`recall`, autonomous
router APIs, vector DBs, embeddings, graph memory, and new runner actions as
unsupported only when the user asks you to use, implement, or rely on them as
routine OpenClerk surfaces.

## Propose-Before-Create Candidate Documents

When the user asks to "document this", "save this note", or otherwise create a
document but omits `document.path`, `document.title`, or `document.body`, you
may propose a candidate document before writing only if the user supplied
enough explicit content to preserve a faithful body. Supported inputs include
pasted notes, excerpts, clear headings, transcript snippets, operational notes,
or user-written URL summaries where the claims to preserve are in the prompt.

For candidate proposals:

1. Preserve explicit user path, title, body, type, and naming instructions.
2. Fill omitted fields only from explicit supplied content. For note-like
   candidates without a path, use `notes/candidates/<slug-from-title>.md`.
3. Keep the body faithful. Do not add facts, citations, source claims, security
   claims, or network-fetched content not supplied by the user. Include
   `type: note` frontmatter for note-like candidates.
4. Validate the candidate with `openclerk document` `action: "validate"` before
   presenting it. Validation is not a durable write.
5. When duplicate risk is requested or plausible, use runner-visible `search`
   and `list_documents` before proposing; inspect an existing `doc_id` only
   when needed. If a likely duplicate is visible, ask whether to update it or
   create a new confirmed path.
6. Final answers for proposals show `Path:`, `Title:`, and `Body preview:`,
   report validation or duplicate-check results, state that no document was
   created, and ask for approval before any durable write.
7. Do not call `create_document`, `append_document`, or `replace_section` until
   the user approves the target and write.

Use no-tools clarification instead of proposing when actual body content is
missing, the durable artifact type is unclear, the request is only a bare URL
or source artifact without source-ingestion hints, the candidate would require
network fetching, or confidence is too low to preserve a faithful body.

## Document Tasks

Run document tasks with:

```bash
openclerk document
```

Common request shapes:

```json
{"action":"validate","document":{"path":"notes/projects/example.md","title":"Example","body":"# Example\n\n## Summary\nReusable knowledge.\n"}}
{"action":"create_document","document":{"path":"notes/projects/example.md","title":"Example","body":"# Example\n\n## Summary\nReusable knowledge.\n"}}
{"action":"ingest_source_url","source":{"url":"https://example.test/source.pdf","path_hint":"sources/example.md","asset_path_hint":"assets/sources/example.pdf","title":"Optional title"}}
{"action":"ingest_source_url","source":{"url":"https://example.test/source.pdf","mode":"update"}}
{"action":"ingest_video_url","video":{"url":"https://youtube.example.test/watch?v=demo","path_hint":"sources/video-youtube/demo.md","title":"Demo Video Transcript","transcript":{"text":"Supplied transcript text.","policy":"supplied","origin":"user_supplied_transcript","language":"en","captured_at":"2026-04-27T00:00:00Z"}}}
{"action":"ingest_video_url","video":{"url":"https://youtube.example.test/watch?v=demo","mode":"update","transcript":{"text":"Updated supplied transcript text.","policy":"supplied","origin":"user_supplied_transcript"}}}
{"action":"list_documents","list":{"path_prefix":"notes/","limit":20}}
{"action":"get_document","doc_id":"doc_id_from_json"}
{"action":"append_document","doc_id":"doc_id_from_json","content":"## Decisions\nUse the OpenClerk runner."}
{"action":"replace_section","doc_id":"doc_id_from_json","heading":"Decisions","content":"Use the OpenClerk runner for routine local knowledge tasks."}
{"action":"resolve_paths"}
{"action":"inspect_layout"}
```

Request fields are `action`, `document`, `source`, `video`, `doc_id`,
`content`, `heading`, and `list`. A `document` has `path`, `title`, and `body`.
A `source` has `url`, `path_hint`, `asset_path_hint`, optional `title`, and
optional `mode` (`create` default, or `update`). A `video` has `url`,
`path_hint`, optional `asset_path_hint`, optional `title`, optional `mode`, and
`transcript`. A `list` may include `path_prefix`, `metadata_key`,
`metadata_value`, `limit`, and `cursor`.

Validation rejections are JSON results with `rejected: true` and
`rejection_reason`; runtime failures exit non-zero and write errors to stderr.

Use `resolve_paths` to confirm the effective database path and configured
vault root. Use `inspect_layout` for configured layout questions and answer
from `layout` JSON fields such as `mode`, `config_artifact_required`,
`conventional_paths`, `document_kinds`, and `checks`. Do not inspect lower-level
storage or run `init` to diagnose routine layout problems.

Use `ingest_source_url` for HTTP/HTTPS PDF source ingestion. Create mode needs
vault-relative `sources/*.md` and `assets/**/*.pdf` hints. Update mode may omit
path hints and refreshes runner-visible citations, provenance, and dependent
freshness when content changes. Do not download, inspect, or write PDFs
yourself.

Use `ingest_video_url` only with user-supplied transcript text and provenance.
Do not acquire media or transcripts with external tools or lower-level storage.
Unsupported acquisition paths remain design-only until promoted.

For source-linked synthesis, run `search`, list `synthesis/`, inspect existing
candidates before editing, and prefer `replace_section` or `append_document`
over duplicates. Synthesis lives under `synthesis/`, cites canonical `sources/`,
uses single-line comma-separated `source_refs`, includes `## Sources` and
`## Freshness`, and remains lower authority than canonical sources.

Before stale synthesis repair or source-sensitive audit output, inspect
`projection_states` and `provenance_events`. If current sources conflict without
runner-visible authority or supersession, explain the conflict with both source
paths instead of choosing a winner.

Use retrieval `audit_contradictions` only for the promoted narrow
source-linked audit workflow. It can plan or repair an existing synthesis page,
inspect provenance/freshness, prevent duplicates, and report unresolved current
source conflicts. Use `plan_only` for review and `repair_existing` only when
the request asks to repair an existing target.

For messy populated-vault retrieval, answer from runner-visible authority:
Metadata-filtered authority results, active canonical sources, cited source
paths, `doc_id`, and `chunk_id`. Treat polluted, decoy, stale, draft, archived,
duplicate, or candidate documents as non-authority unless runner-visible source
authority says otherwise. If a result is marked with `status: polluted` or
`populated_role: decoy`, explicitly reject that hit as not authority and do not
repeat its false claim text as a valid answer.

Even when synthesis maintenance is repetitive, stay with documented AgentOps
document and retrieval actions.

## Retrieval Tasks

Run retrieval tasks with:

```bash
openclerk retrieval
```

Common request shapes:

```json
{"action":"search","search":{"text":"architecture","limit":10}}
{"action":"search","search":{"text":"architecture","path_prefix":"notes/","metadata_key":"status","metadata_value":"active","limit":10}}
{"action":"document_links","doc_id":"doc_id_from_json"}
{"action":"graph_neighborhood","doc_id":"doc_id_from_json","limit":10}
{"action":"records_lookup","records":{"text":"OpenClerk runner","limit":10}}
{"action":"record_entity","entity_id":"entity_id_from_json"}
{"action":"services_lookup","services":{"text":"OpenClerk runner","interface":"JSON runner","limit":10}}
{"action":"service_record","service_id":"service_id_from_json"}
{"action":"decisions_lookup","decisions":{"text":"JSON runner","status":"accepted","scope":"runner","limit":10}}
{"action":"decision_record","decision_id":"decision_id_from_json"}
{"action":"provenance_events","provenance":{"ref_kind":"document","ref_id":"doc_id_from_json","limit":20}}
{"action":"projection_states","projection":{"ref_kind":"document","ref_id":"doc_id_from_json","limit":20}}
{"action":"projection_states","projection":{"projection":"synthesis","ref_kind":"document","ref_id":"synthesis_doc_id_from_json","limit":20}}
{"action":"projection_states","projection":{"projection":"decisions","ref_kind":"decision","ref_id":"decision_id_from_json","limit":20}}
{"action":"audit_contradictions","audit":{"query":"source-sensitive audit runner repair evidence","target_path":"synthesis/audit-runner-routing.md","mode":"plan_only","conflict_query":"source sensitive audit conflict runner retention","limit":10}}
{"action":"audit_contradictions","audit":{"query":"source-sensitive audit runner repair evidence","target_path":"synthesis/audit-runner-routing.md","mode":"repair_existing","conflict_query":"source sensitive audit conflict runner retention","limit":10}}
```

Request fields are `action`, `search`, `doc_id`, `chunk_id`, `node_id`,
`entity_id`, `service_id`, `decision_id`, `records`, `services`, `decisions`,
`provenance`, `projection`, `audit`, and `limit`. An `audit` request has
`query`, `target_path`, `mode`, `conflict_query`, and `limit`; supported modes
are `plan_only` and `repair_existing`.

Use search for source-grounded answers; document links and graph neighborhoods
for markdown relationships; records, services, and decisions lookup for
promoted-domain projections; provenance for derivation history; and projection
states for freshness. Canonical markdown remains authoritative over derived
service, record, decision, and synthesis projections.

## Answering From Results

Answer from JSON fields such as `document`, `documents`, `search`, `links`,
`graph`, `records`, `entity`, `provenance`, `projections`, `paths`, or
`rejection_reason`.

Answer the user's substantive question from the selected runner result before
listing evidence. For retrieval-only repeats, confirm no durable write only
when asked, but still restate the answer and citations. Preserve citation paths,
source refs, doc ids, chunk ids, and provenance details for source-sensitive
claims. For filtered list answers, mention only returned rows unless the user
explicitly asks about omitted data.
