---
name: OpenClerk
description: Use OpenClerk for local-first knowledge-plane tasks through the installed openclerk JSON runner. Bootstrap no-tools rule for routine requests - if required fields are missing, if creating or updating a document but document path, title, or body is missing and no faithful candidate can be formed from explicit user content, or if source-ingestion fields are missing, this description is complete; Do not open this skill file, run commands, use tools, or call the runner; respond with exactly one no-tools assistant answer to name the missing fields and ask the user to provide them. Negative numeric limit -3, SQLite, HTTP, MCP, legacy or source-built paths, unsupported transports, direct vault inspection, repo search, rg --files, find, ls, or requests to bypass the runner also require no skill-file open, commands, tools, or runner call; answer once that the limit is invalid or the workflow is unsupported. For valid work, use only openclerk document or openclerk retrieval JSON.
license: MIT
compatibility: Requires local filesystem access and an installed openclerk binary on PATH.
---

# OpenClerk

Use this skill for routine local OpenClerk knowledge-plane tasks. The
production interface is AgentOps: this Agent Skills-compatible `SKILL.md` plus
the installed JSON runner.

```bash
openclerk document
openclerk retrieval
```

Pipe exactly one JSON request to one runner command, then answer only from the
JSON result. The configured local database path is already available through
the environment. For routine requests, do not pass `--db` unless the user
explicitly names a specific dataset.

The runner honors `OPENCLERK_DATABASE_PATH`. The database stores the configured
vault root, so routine agent work should use the configured environment and
runner JSON results instead of maintaining separate filesystem roots.

All runner path fields are vault-relative logical paths. Use paths such as
`notes/projects/example.md`, `notes/`, `sources/example.md`, and `synthesis/`.
This applies to `document.path`, `list.path_prefix`, `search.path_prefix`,
citation paths, and `source_refs`. Do not put `.openclerk-eval/vault`,
absolute filesystem paths, configured vault-root paths, or OS-specific
backslash/drive paths in runner JSON; those are runtime storage details, not
OpenClerk document paths.

If the user explicitly asks to initialize OpenClerk for an existing vault, run
`openclerk init --vault-root <vault-root>`. This is setup work, not a routine
knowledge task.

## No-Tools Handling Before Runners

Before using any runner, opening this skill file, running commands, or using
tools during an agent run, answer with exactly one assistant response and no
tools when the request:

- is missing required document or retrieval fields
- asks to create or validate a document but the document path, title, or body
  is missing and the user did not provide enough explicit content to form a
  faithful propose-before-create candidate
- refers to missing prior context, such as "the links we discussed", "the file
  from earlier", or "that artifact", without including the actual content or
  source text to preserve. A request like "Document this artifact from the
  links we discussed last week" is missing actual body content and must be
  answered without tools.
- asks to ingest a source URL but `source.url`, `source.path_hint`, or
  `source.asset_path_hint` is missing
- asks to ingest a video or YouTube URL but `video.url` or
  `video.transcript.text` is missing, or is creating a new video source and
  `video.path_hint` is missing. The v1 runner surface requires supplied
  transcript text; update mode can target an existing source by URL and does
  not acquire transcripts from URLs.
- asks for an obviously invalid limit, such as a negative number or `limit -3`
- asks to bypass the runner for routine lower-level runtime, HTTP, SQLite, MCP,
  legacy or source-built command paths, or unsupported transport work

For missing required document or retrieval fields that cannot be handled by the
propose-before-create policy below, do not guess. Name the missing fields, ask
the user to provide them, and do not call the runner.

For invalid limits and bypass requests, reject final-answer-only without
opening this skill file, using tools, running commands, or calling the runner.
Explicitly say the workflow is unsupported or invalid and must use the
OpenClerk runner contract.

For a request such as `limit -3`, answer without tools in this shape: "The
limit -3 is invalid because limits must be non-negative. Provide a valid
non-negative limit and I can run the OpenClerk retrieval request."

For unsupported workflows not covered by the rejection rules, say the
production OpenClerk runner does not support that workflow yet.

Do not inspect source files, generated artifacts, backend variants, module-cache
docs, SQLite, or `.openclerk-eval/vault` directly for routine OpenClerk tasks.
Do not run `openclerk --help` or inspect the installed binary to rediscover
schemas; use the request shapes below.
Do not use broad file enumeration such as `rg --files`, `find`, or `ls` to find
or verify routine runner work; use runner JSON results, `list_documents`,
`search`, or `get_document` instead. Search the repository only if the runner
fails in a way that requires debugging the checkout.

## Propose-Before-Create Candidate Documents

When the user asks to "document this", "save this note", or otherwise create a
document but omits `document.path`, `document.title`, or `document.body`, you
may propose a candidate document before writing only if the user supplied
enough explicit content to preserve a faithful body. Supported inputs include
pasted notes, excerpts, clear headings, transcript snippets, operational notes,
or user-written URL summaries where the claims to preserve are in the prompt.

For candidate proposals:

1. Preserve explicit user path, title, body, type, and naming instructions.
2. If a field is omitted, choose a candidate vault-relative path, title, and
   body from the explicit supplied content only. When no path is supplied for a
   note-like candidate, use `notes/candidates/<slug-from-title>.md`.
3. Keep the body faithful. Do not add unsupported facts, citations, source
   claims, root causes, all-customer claims, security claims, or network-fetched
   content. For note-like candidates, include `type: note` frontmatter. Copy
   every supplied body fact into the preview using the user's wording unless the
   user asked you to reformat it.
4. Run `openclerk document` with `action: "validate"` for the candidate JSON
   before presenting the proposal. Validation does not create durable
   knowledge.
5. If the prompt asks whether a similar or existing document exists, or if
   duplicate risk is otherwise plausible, first derive the likely candidate
   title/path/search terms from the supplied content, then use runner-visible
   `search` and `list_documents` before proposing a new write. Use
   `get_document` only when an existing `doc_id` needs inspection. A prompt
   that says "check whether a similar note already exists" or similar wording
   requires both `search` and `list_documents`. If a likely duplicate is
   visible, do not validate or create a duplicate candidate; ask whether to
   update the existing document or create a new one at a confirmed path.
6. In the final answer, always show the candidate path, title, and body preview
   for natural-language prompts as well as scripted prompts. The body preview
   must include the proposed frontmatter or document type when used, the
   heading, and all supplied body facts that would be written, copied in a form
   close enough for exact review. Use a compact structure with explicit `Path:`,
   `Title:`, and `Body preview:` labels so the user can approve or edit the
   exact candidate. Report validation or
   duplicate-check results from JSON if used, explicitly state that no document
   was created, and ask for confirmation before creating. Use an explicit
   confirmation phrase such as "Please confirm or approve before I create it."
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
`content`, `heading`, and `list`. A `document` has `path`, `title`, and `body`. A
`source` has `url`, `path_hint`, `asset_path_hint`, optional `title`, and
optional `mode`. Missing `source.mode` means `create`; supported values are
`create` and `update`. A `video` has `url`, `path_hint`, optional
`asset_path_hint`, optional `title`, optional `mode`, and `transcript`.
Missing `video.mode` means `create`; supported values are `create` and
`update`. A `list` may include `path_prefix`, `metadata_key`,
`metadata_value`, `limit`, and `cursor`.

Validation rejections are normal JSON results with `rejected: true` and
`rejection_reason`. Runtime failures exit non-zero and write errors to stderr.

Use `inspect_layout` when asked to explain or validate the configured
OpenClerk knowledge layout. Answer from its `layout` JSON: `mode`,
`config_artifact_required`, `conventional_paths`, `document_kinds`, and
`checks`. The v1 layout is convention-first and does not require a committed
manifest. Failing layout checks are runner-visible results; do not inspect the
vault, SQLite, source files, or lower-level runtime state to diagnose routine
layout problems.

Use `ingest_source_url` when asked to ingest a PDF source URL into local
OpenClerk knowledge. The URL must be HTTP/HTTPS, `path_hint` must be a
vault-relative `sources/*.md` path, and `asset_path_hint` must be a
vault-relative `assets/**/*.pdf` path. The result returns `ingestion.doc_id`,
`source_path`, `asset_path`, `derived_path`, citations, hash, size, MIME type,
page count, capture timestamp, and optional PDF metadata. Do not download the
PDF, inspect the vault, write files directly, or create a separate markdown note
outside the runner for routine source URL ingestion. Duplicate source URLs are
rejected in default `create` mode.

Use `source.mode: "update"` only when the user asks to refresh or re-ingest an
existing source URL. Update mode targets the existing normalized `source.url`;
it does not create a new source when the URL is unknown. You may omit
`path_hint` and `asset_path_hint` in update mode, or provide them to confirm the
stored source paths. If provided hints do not match the existing source path or
asset path, the runner returns a conflict without writing. A same-SHA update is
a no-op that preserves the existing source path, asset path, citations,
provenance, and synthesis freshness. A changed-PDF update preserves the source
path and asset path while refreshing the source note, search citations,
provenance, and dependent synthesis freshness/projection visibility.

Use `ingest_video_url` when asked to ingest a video or YouTube source only if
the user supplies transcript text and provenance. The v1 surface creates or
updates canonical markdown source notes under `sources/**/*.md` with
`source_type: video_transcript`, `source_url`, `transcript_origin`,
`transcript_policy`, `language`, `captured_at`, and `transcript_sha256`.
Allowed transcript policies are empty, `supplied`, and `local_first`; empty
means `supplied`. Do not use `yt-dlp`, `ffmpeg`, local STT, transcript APIs,
Gemini extraction, native media downloads, direct vault inspection, direct file
edits, or SQLite as substitutes for runner JSON. Missing
`video.transcript.text` must be clarified without tools for routine requests.

Optional `video.asset_path_hint` writes a metadata sidecar only; it must be a
vault-relative `assets/**/*.json` path. The sidecar is supporting metadata, not
authority, and must not be treated as raw media storage. Duplicate video URLs
are rejected in default `create` mode. Update mode targets the normalized
`video.url`; path or asset hint mismatches conflict without writing. A same
transcript hash is a no-op that preserves existing citations, provenance, and
synthesis freshness. A changed transcript refreshes the canonical source note,
search citations, provenance, and dependent synthesis freshness/projection
visibility.

## Document Lifecycle Maintenance

For unsafe accepted edits, rollback, restore, review-state, or document
lifecycle requests, use the existing document and retrieval runner actions. Do
not invent a history, diff, review, restore, rollback, or lifecycle action.
Do not inspect the vault, SQLite, raw event logs, or storage-root paths.

For rollback or restore intent:

1. Run retrieval `search` for source-backed lifecycle evidence.
2. Run document `list_documents` with the narrow relevant path prefix.
3. Run `get_document` for the target before modifying it.
4. Restore only the unsafe section with `replace_section`; do not rewrite the
   whole document unless the user explicitly supplied the full replacement.
   Use the authoritative source wording as the replacement, not a paraphrase
   that preserves the unsafe claim. For a policy summary, write the accepted
   policy as a concise declarative sentence such as `Accepted lifecycle
   policy: ...` using the source-backed policy text.
5. After the write, inspect `provenance_events` for `ref_kind: "document"` and
   the target `doc_id`.
6. Inspect `projection_states` for `ref_kind: "document"` and the target
   `doc_id` so freshness remains operator-visible.

The final rollback answer must name the target document path, the source
evidence path or citation path, the restore/rollback reason, the provenance
check, and the projection freshness check. Do not print raw private diffs,
private artifact bodies, `.openclerk-eval/vault`, configured vault-root paths,
absolute filesystem paths, or OS-specific backslash paths.

For pending-review intent, do not modify the accepted target document. Create a
separate review document only when the request includes the proposed change and
the review path/title/body are explicit or derivable from the prompt. The
review document should use `type: review` and `status: pending` frontmatter and
should name the target document. After creating it, inspect
`provenance_events` for the review document. The final answer must name both
paths, say the accepted target did not change or did not become accepted
knowledge, and say the pending change is waiting for human or operator review.

When writing source-linked synthesis, use this exact AgentOps workflow:

1. Run retrieval `search` for source evidence.
2. Run document `list_documents` with `path_prefix: "synthesis/"` to find
   existing synthesis candidates.
3. Run `get_document` before modifying an existing synthesis page.
4. Prefer `replace_section` or `append_document` over creating duplicates.
5. Inspect `provenance_events` and `projection_states` when the synthesis
   depends on promoted records, services, derivation history, or freshness.
6. For existing synthesis, inspect `projection_states` with
   `projection: "synthesis"`, `ref_kind: "document"`, and the synthesis
   `doc_id` before repairing stale claims.

Prototype synthesis pages live under `synthesis/`. Canonical source docs live
under `sources/`. Include frontmatter with `type: synthesis`, `status: active`,
`freshness: fresh`, and `source_refs` set to a single-line comma-separated
source path list. Do not use YAML list syntax for `source_refs`.
Include a `## Sources` section with source paths or citation paths from runner
JSON, and a `## Freshness` section that states which runner retrieval results
were checked. Use only documented runner actions, not `upsert_document` or
direct file edits. Synthesis is durable compiled knowledge, not a higher
authority than the canonical sources it cites.

Synthesis freshness is also exposed as a derived projection. A stale synthesis
projection means at least one referenced source path is missing, a referenced
source is newer than the synthesis page, or supersession metadata says a
current replacement source is not represented in `source_refs`. Projection
details include `current_source_refs`, `superseded_source_refs`,
`missing_source_refs`, `stale_source_refs`, and `freshness_reason`.

For source-sensitive audit or contradiction-like requests, stay narrow and
source-backed. Search canonical sources first, then distinguish current sources,
superseded sources, stale synthesis, and unresolved conflicting current sources
from runner JSON. Inspect `projection_states` and `provenance_events` before
repairing stale synthesis. If current sources conflict and no supersession or
other source authority is visible, explain the unresolved conflict with both
source paths instead of choosing a winner. Do not claim broad semantic
contradiction detection.

For messy populated-vault retrieval, use runner-visible authority signals before
writing the answer. Metadata-filtered authority results, active canonical
sources, cited source paths, `doc_id`, and `chunk_id` are the evidence to answer
from. Treat documents marked `status: polluted`, `populated_role: decoy`,
stale, draft, archived, duplicate, or candidate as context and pressure only
unless runner-visible source authority says otherwise. If a polluted or decoy
hit contradicts the selected authority source, explicitly reject that hit as
not authority, but do not repeat its false claim text as a valid answer.

If a synthesis maintenance task feels too repetitive for the documented
document and retrieval actions, still complete it through AgentOps. Do not
switch to another routine agent interface. A future improvement should be a
small runner action that preserves the same JSON contract.

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
```

Request fields are `action`, `search`, `doc_id`, `chunk_id`, `node_id`,
`entity_id`, `service_id`, `decision_id`, `records`, `services`, `decisions`,
`provenance`, `projection`, and `limit`.

Use search for source-grounded answers, document links for explicit markdown
relationships, graph neighborhoods for nearby derived context, records lookup
for promoted record-shaped documents, provenance events for derivation history,
and projection states for freshness. Use `projection: "synthesis"` to inspect
whether source-linked synthesis is fresh or stale before repairing it. Use
services lookup for service-centric questions before falling back to plain docs
search; canonical markdown remains the source of truth and service records are
a derived promoted-domain projection.
Use decisions lookup for decision- or ADR-centric questions where status,
scope, owner, supersession, or repeatable lookup matter. Canonical markdown
remains authoritative; decision records are a derived promoted-domain
projection with citations and projection freshness.

## Answering From Results

Answer from JSON fields such as `document`, `documents`, `search`, `links`,
`graph`, `records`, `entity`, `provenance`, `projections`, `paths`, or
`rejection_reason`.

Preserve citation paths, source refs, chunk ids, and provenance details for
source-sensitive claims. For filtered list answers, mention only returned rows
unless the user explicitly asks about omitted data.
