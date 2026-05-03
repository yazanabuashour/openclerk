---
name: OpenClerk
description: Use openclerk runner. Bootstrap no-tools rule - if required fields are missing; if document path, title, or body is missing without a faithful propose-before-create candidate, duplicate-risk check, or public-link placement proposal from explicit user content; if limit -3; if opaque image/screenshot, slide deck/PPTX, email archive, exported chat, form, or bundle lacks pasted/supplied content; or if asked to bypass the runner with OCR, PPTX parsing, email/chat/form/bundle parsing, local file reads, browser automation, SQLite, HTTP, MCP, legacy or source-built paths, unsupported transports, backend variants, module-cache inspection, rg --files, find, ls, direct vault inspection, or repo search, this description is complete. Do not open this skill file, run commands, use tools, or call the runner; respond with exactly one no-tools assistant answer to name the missing fields, ask the user to provide them, or name the blocked parser/bypass. Valid work uses openclerk document or openclerk retrieval JSON.
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
  transports, browser automation, OCR, PPTX parsing, email/chat/form/bundle
  parsing, local file reads, or external acquisition tools as substitutes for
  runner JSON.
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
- A request with explicit note body content and missing path/title is not a
  no-tools missing-field case when the propose-before-create policy can derive
  a faithful candidate.
- A request with explicit note body content and unresolved duplicate
  update-versus-new intent is not a no-tools missing-field case. Use
  runner-visible duplicate checks before validating or writing.
- A request with supplied local-file-derived source content and unresolved
  duplicate source intent is not a no-tools local-file-read case when the task
  only asks for runner-visible duplicate/provenance inspection. Use runner
  evidence before validating or writing.
- A "document these links" request with explicit public web URLs and omitted
  `source.path_hint` or synthesis placement is not a no-tools missing-field
  case when the document-these-links placement policy can propose safe
  source and synthesis paths before any fetch or write.
- Requests that refer to missing prior context, such as "the links we
  discussed" or "that artifact", lack preservable body/source content.
- New PDF source URL ingestion needs `source.url`, `source.path_hint`, and
  `source.asset_path_hint`; new web source URL ingestion needs `source.url`
  and `source.path_hint`. Update mode needs `source.url`.
- Video/YouTube ingestion needs `video.url` and supplied transcript text; new
  video sources also need `video.path_hint`.
- Limits must be non-negative.

Unsupported opaque artifact rules:

- Opaque images or screenshots, slide decks or PPTX files, email archives,
  exported chats, forms, and mixed bundles are unsupported when the user has
  not pasted or explicitly supplied preservable text/body content.
- For those opaque artifact requests, use one no-tools answer. Say the
  artifact kind is unsupported as opaque input, ask for pasted or explicitly
  supplied content, or ask the user to approve a candidate-document workflow
  only when a faithful candidate can be formed from supplied content.
- Public read or inspect permission is not durable-write approval. Keep
  durable writes gated on explicit approval, and do not treat permission to
  view a referenced artifact as permission to create or update a document.
- Do not claim parser truth, OCR results, hidden file inspection, attachment
  contents, or bundle contents that the user did not paste or explicitly
  supply.

Parser and acquisition bypass rules:

- Reject requests to use OCR, PPTX parsing, email import or parsing,
  chat/form/bundle parsing or extraction, local file reads, browser automation,
  direct vault inspection, direct SQLite, HTTP/MCP bypasses, source-built
  runners, legacy paths, unsupported transports, backend variants,
  module-cache inspection, repo search, `rg --files`, `find`, or `ls` as
  substitutes for installed `openclerk document` or `openclerk retrieval` JSON.
- The rejection is final-answer-only: no tools, no commands, no runner call,
  no lower-level file inspection, and no attempt to acquire or parse the
  unsupported artifact outside the runner contract.

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

For routine read-only memory/router recall report requests, use the promoted
retrieval action `memory_router_recall_report` with `memory_router_recall.query`
and a non-negative `memory_router_recall.limit`. It returns query summary,
temporal status, canonical evidence refs, stale session status, advisory
feedback weighting, routing rationale, provenance refs, synthesis freshness,
validation boundaries, and authority limits. This action is read-only; it is
not a memory transport, `remember`/`recall` action, autonomous router API,
vector/embedding/graph memory surface, or write approval.

## Propose-Before-Create Candidate Documents

When the user asks to "document this", "save this note", or otherwise create a
document but omits `document.path`, `document.title`, or `document.body`, you
may propose a candidate document before writing only if the user supplied
enough explicit content to preserve a faithful body. Supported inputs include
pasted notes, excerpts, clear headings, transcript snippets, operational notes,
or user-written URL summaries where the claims to preserve are in the prompt.
If the user provides a public web URL plus the required source path hint, use
`ingest_source_url` instead of proposing a candidate; no separate approval is
needed before the runner fetches the URL.

Opaque artifact references are not explicit supplied content. Do not propose a
candidate from a screenshot, slide deck, PPTX, email archive, exported chat
file, form, or bundle unless the user pasted or explicitly supplied the text to
preserve. When the user supplies that text, keep the candidate faithful, run
`validate`, show `Path:`, `Title:`, and `Body preview:`, state no document was
created, and ask for approval before any durable write.

For "save this note" requests with explicit note content but no path or title,
derive a faithful note candidate from the supplied content, validate it, show
the candidate, state that no document was created, and ask for approval before
creating anything. For bare prior-context requests such as "save this note from
what we discussed last week", use the no-tools rule: ask for the actual note
content plus any path, title, or placement preferences, and do not invent a
path, title, or body.

For routine low-risk note capture, such as "save this low-risk note" with
explicit note body but no path or title, treat the request as valid
runner-backed propose-before-create work. Derive the path and title from the
supplied content, using `notes/candidates/<slug-from-title>.md` when no path is
given. Use the content subject for the title and slug, not request-framing
words such as "save", "capture", or "note". For example, "Support handoff
should note the owner..." becomes title `Support Handoff` and path
`notes/candidates/support-handoff.md`, not a sentence-length slug or a title
ending in `Note`. Validate the candidate through `openclerk document`
`action: "validate"`; do not rely on reasoning-only validation. Then answer
with `Path:`, `Title:`, and `Body preview:` before the validation result,
no-write statement, and approval request. The body preview for note-like
captures includes the faithful `type: note` frontmatter, `# <Title>` heading,
and supplied body text. Do not answer with only validation status or only an
approval prompt; the candidate path, title, and body preview must be visible
before any durable write is approved. If the body preview is missing from the
final answer, the workflow is incomplete even when validation passed.

For low-risk duplicate checks, do not treat the missing update-versus-new
choice as a no-tools missing-field rejection. When explicit note content exists
and duplicate risk is plausible or requested, use retrieval `search`, document
`list_documents`, and `get_document` for the likely target before answering.
Then report the likely target path and title, say no document was created or
updated, and ask whether to update the existing target or create a new document
at a confirmed path.

For supplied local-file-derived source duplicate checks, do not read local
files, parse artifacts, run OCR, inspect vault files directly, or treat the
duplicate check as a no-tools local-file-read request. Use only runner-visible
retrieval `search`, document `list_documents`, document `get_document`, and
retrieval `provenance_events` for the likely existing source before answering.
When the request supplies the duplicate search text, path prefix, existing
source path, or enough equivalent runner-visible target evidence, a no-tools
answer is incomplete; run the read-only runner checks before answering.
Then name the existing source path, the candidate path that was not created,
the duplicate/provenance evidence, the no-local-file-read/parser/OCR boundary,
and approval-before-write. Include the exact ideas `duplicate` or `existing`,
`provenance`, `was not created` or `no document was created`, and
`approval-before-write` or `approval before write`. Do not call `validate`,
`create_document`, `append_document`, `replace_section`, `ingest_source_url`,
or `ingest_video_url` while duplicate update-versus-new source intent remains
unresolved.

For "document these links" requests with explicit public web URLs but missing
`source.path_hint` values or synthesis placement, treat the request as valid
placement-proposal work. Do not fetch, validate, create, append, or replace
while placement is unapproved. Propose one source path per public web URL using
`sources/candidates/<slug-from-label-or-url>.md`, plus a synthesis path using
`synthesis/<shared-topic-or-url-set>.md` when the user asks for a combined
document. State that no source or synthesis document was created and ask for
approval before any durable source fetch or synthesis write.

After source paths are approved for public web URLs, use `openclerk document`
`ingest_source_url` with `source_type: "web"` and the approved
`source.path_hint`; report citation evidence such as `doc_id`, `chunk_id`, or
returned citations. PDF and other artifact URLs still require the existing
source and asset path hints before ingestion. Do not fetch URLs with browser,
HTTP, filesystem, or other non-runner tools.

After source intent is clear but synthesis creation is not approved, use
retrieval `search`, document `list_documents`, and `get_document` to inspect
source evidence and existing synthesis candidates. Validate a source-linked
synthesis candidate with single-line `source_refs`, `## Sources`, and
`## Freshness`; state that no synthesis document was created and ask for
approval before creating it.

For document-these-links duplicate checks, do not treat the missing
update-versus-new choice as a no-tools rejection. When duplicate source or
synthesis placement is plausible or requested, use retrieval `search`, document
`list_documents`, and `get_document` for likely source and synthesis targets
before answering. Then report the existing source or synthesis paths, summarize
the search/list/get evidence, say no source or synthesis document was created
or updated, and ask whether to update the existing placement or create new
confirmed paths. Do not call `validate`, `ingest_source_url`,
`create_document`, `append_document`, or `replace_section` while duplicate
update-versus-new placement is unresolved.

For candidate proposals:

1. Preserve explicit user path, title, body, type, and naming instructions.
2. Fill omitted fields only from explicit supplied content. For note-like
   candidates without a path, use `notes/candidates/<slug-from-title>.md`.
3. Keep the body faithful. Do not add facts, citations, source claims, security
   claims, or network-fetched content not supplied by the user. Include
   `type: note` frontmatter for note-like candidates.
   If the user explicitly supplies frontmatter tags, preserve those tags
   exactly. If tags are not explicit and sensible tags would help retrieval,
   the agent may propose `tag: <value>` in the visible body preview before
   approval. The runner does not infer or add tags itself.
4. Validate the candidate with `openclerk document` `action: "validate"` before
   presenting it. Validation is not a durable write.
5. When duplicate risk is requested or plausible, treat it as valid
   runner-backed capture work. Before validating or proposing a new candidate
   path, run runner-visible retrieval `search` and document `list_documents`.
   If the user or candidate context gives a likely collection or path prefix,
   include that `path_prefix` in the retrieval search and use the same prefix
   for `list_documents`. When a likely duplicate is visible, run
   `get_document` for that target. For supplied local-file-derived source
   duplicate checks, also run retrieval `provenance_events` for the same
   target. Present the likely target path and title, briefly summarize the
   search/list/get evidence and any required provenance evidence, state that no
   document was created or updated, and ask whether to update the existing
   target or create a new document at a confirmed path.
6. Do not call `validate`, `create_document`, `append_document`, or
   `replace_section` while duplicate update-versus-new-path intent is
   unresolved.
7. Final answers for proposals show `Path:`, `Title:`, and `Body preview:`,
   report validation or duplicate-check results, state that no document was
   created, and ask for approval before any durable write.
8. Do not call `create_document`, `append_document`, or `replace_section` until
   the user approves the target and write.

Use no-tools clarification instead of proposing when actual body content is
missing, the durable artifact type is unclear, the request is only a bare URL
or source artifact without source-ingestion hints and the document-these-links
policy cannot form a safe placement proposal, the candidate would require
network fetching outside `ingest_source_url`, or confidence is too low to
preserve a faithful body.

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
{"action":"ingest_source_url","source":{"url":"https://example.test/page.html","path_hint":"sources/web/example.md","source_type":"web","title":"Optional title"}}
{"action":"ingest_source_url","source":{"url":"https://example.test/page.html","mode":"update","source_type":"web"}}
{"action":"ingest_video_url","video":{"url":"https://youtube.example.test/watch?v=demo","path_hint":"sources/video-youtube/demo.md","title":"Demo Video Transcript","transcript":{"text":"Supplied transcript text.","policy":"supplied","origin":"user_supplied_transcript","language":"en","captured_at":"2026-04-27T00:00:00Z"}}}
{"action":"ingest_video_url","video":{"url":"https://youtube.example.test/watch?v=demo","mode":"update","transcript":{"text":"Updated supplied transcript text.","policy":"supplied","origin":"user_supplied_transcript"}}}
{"action":"list_documents","list":{"path_prefix":"notes/","limit":20}}
{"action":"list_documents","list":{"path_prefix":"notes/","tag":"account-renewal","limit":20}}
{"action":"get_document","doc_id":"doc_id_from_json"}
{"action":"append_document","doc_id":"doc_id_from_json","content":"## Decisions\nUse the OpenClerk runner."}
{"action":"replace_section","doc_id":"doc_id_from_json","heading":"Decisions","content":"Use the OpenClerk runner for routine local knowledge tasks."}
{"action":"resolve_paths"}
{"action":"inspect_layout"}
```

Request fields are `action`, `document`, `source`, `video`, `doc_id`,
`content`, `heading`, and `list`. A `document` has `path`, `title`, and `body`.
A `source` has `url`, `path_hint`, optional `asset_path_hint`, optional
`title`, optional `mode` (`create` default, or `update`), and optional
`source_type` (`pdf` or `web`). A `video` has `url`,
`path_hint`, optional `asset_path_hint`, optional `title`, optional `mode`, and
`transcript`. A `list` may include `path_prefix`, `tag`, `metadata_key`,
`metadata_value`, `limit`, and `cursor`.

Validation rejections are JSON results with `rejected: true` and
`rejection_reason`; runtime failures exit non-zero and write errors to stderr.

Use `resolve_paths` to confirm the effective database path and configured
vault root. Use `inspect_layout` for configured layout questions and answer
from `layout` JSON fields such as `mode`, `config_artifact_required`,
`conventional_paths`, `document_kinds`, and `checks`. Do not inspect lower-level
storage or run `init` to diagnose routine layout problems.

Use `ingest_source_url` for HTTP/HTTPS PDF and public web source ingestion.
PDF create mode needs vault-relative `sources/*.md` and `assets/**/*.pdf`
hints. Web create mode needs a vault-relative `sources/*.md` hint and may set
`source_type: "web"`; it must not set `source.asset_path_hint`. If
`source_type` is omitted, the runner detects PDF versus HTML from the URL and
response. Update mode may omit path hints and refreshes runner-visible
citations, provenance, and dependent freshness when content changes. Do not
download, inspect, or write source URLs yourself with external HTTP, browser,
or file tools.

For product pages, preserve only runner-visible public page text and metadata.
Do not automate carts, purchases, login, account state, captcha, paywall, or
private-network acquisition. If a page cannot be fetched as public HTML, report
the runner rejection and ask for pasted content or a supported source.

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
{"action":"search","search":{"text":"renewal","tag":"account-renewal","limit":10}}
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
`provenance`, `projection`, `audit`, `memory_router_recall`, and `limit`. A
`search` request may include `text`, `path_prefix`, `metadata_key`,
`metadata_value`, `tag`, `limit`, and `cursor`. A document `list_documents`
request may include `path_prefix`, `metadata_key`, `metadata_value`, `tag`,
`limit`, and `cursor`. `records`, `services`, and `decisions` lookup requests
may include their documented text/filter fields plus `limit` and `cursor`; for
decisions those fields include `text`, `status`, `scope`, and `owner`. A
`memory_router_recall` request may include `query` and `limit`. An `audit`
request has `query`, `target_path`, `mode`, `conflict_query`, and `limit`;
supported modes are `plan_only` and `repair_existing`.

Use search for source-grounded answers; document links and graph neighborhoods
for markdown relationships; records, services, and decisions lookup for
promoted-domain projections; provenance for derivation history; and projection
states for freshness. Canonical markdown remains authoritative over derived
service, record, decision, and synthesis projections.

For tag-shaped retrieval, prefer `search.tag` or `list.tag` over spelling the
same lookup as `metadata_key: "tag"` plus `metadata_value`. The `tag` field is
a single exact scalar frontmatter filter over canonical Markdown authority; it
does not imply stemming, fuzzy matching, aliases, taxonomy lookup, or multi-tag
intersection. Existing metadata filters remain valid for compatibility and
non-tag metadata, but do not combine `tag` with `metadata_key` or
`metadata_value` in one request.

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
