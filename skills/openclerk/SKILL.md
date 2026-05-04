---
name: OpenClerk
description: Use OpenClerk through installed openclerk document and openclerk retrieval JSON runner. With explicit user content, validate a faithful candidate through the runner but do not write before approval. Bootstrap no-tools rule - if required fields are missing; if document path, title, or body is missing without explicit user content for a faithful candidate, duplicate-risk check, or public-link placement proposal; if numeric limit is negative; or if asked to bypass the runner with SQLite, raw vault/file/repo inspection, HTTP/MCP, legacy/source-built paths, unsupported transports, backend variants, module-cache inspection, rg, find, ls, OCR, browser automation, local file reads, or opaque artifact parsing, this description is complete. For those invalid cases only, Do not open this skill file, run commands, use tools, or call the runner; respond with exactly one no-tools assistant answer naming the missing/invalid fields or unsupported workflow, and for missing fields ask the user to provide them.
license: MIT
compatibility: Requires local filesystem access and an installed openclerk binary on PATH.
---

# OpenClerk

Use this skill for routine local OpenClerk knowledge-plane tasks. The
production interface is AgentOps: this Skills-compatible router plus the
installed JSON runner.

```bash
openclerk document
openclerk retrieval
```

`SKILL.md` is an activation, routing, and safety contract. It is not the
durable home for long workflow recipes. Once a request is safely inside the
runner contract, use agent autonomy with runner JSON results, runner
rejections, returned `agent_handoff`, and compact `openclerk document --help`
or `openclerk retrieval --help` output. Repeated exact JSON, command ordering,
or scenario-specific choreography is product UX evidence for a runner
workflow-action comparison, not a reason to expand this file.

## Action Index

Prefer a promoted workflow action over manual primitive choreography when it
matches the request:

- Source-linked synthesis create/update: document `compile_synthesis`, then
  answer from `compile_synthesis.agent_handoff`.
- Source URL placement before durable fetch/write: document
  `ingest_source_url` with `mode: "plan"`, then answer from
  `source_placement_plan.agent_handoff`.
- Source-sensitive audit explain/repair: retrieval `source_audit_report`, then
  answer from `source_audit.agent_handoff`.
- Records, decisions, provenance, and projection evidence bundles: retrieval
  `evidence_bundle_report`, then answer from
  `evidence_bundle.agent_handoff`.
- Duplicate update-versus-new clarification: retrieval
  `duplicate_candidate_report`, then answer from
  `duplicate_candidate.agent_handoff`.
- Routine read-only memory/router recall reports: retrieval
  `memory_router_recall_report`, then answer from
  `memory_router_recall.agent_handoff` or returned evidence.
- Hybrid/vector retrieval decision support: retrieval
  `hybrid_retrieval_report`, then answer from
  `hybrid_retrieval.agent_handoff`. Do not claim vector-ranked retrieval or
  embedding-store evidence from this report.

Use lower-level primitives for explicit primitive requests, advanced/manual
cases, unsupported workflow-action inputs, and follow-up inspection after a
runner rejection. For promoted workflow actions, answer from the returned
handoff before doing follow-up primitive inspection.

## Core Guardrails

- Answer routine OpenClerk requests only from runner JSON results. Use the
  configured environment; pass `--db` only when the user explicitly names a
  dataset.
- Treat every runner path as vault-relative, such as
  `notes/projects/example.md`, `sources/example.md`, or `synthesis/`. Never
  use storage roots, machine-absolute paths, `.openclerk-eval/vault`, or
  OS-specific backslash/drive paths in runner JSON or committed OpenClerk
  document paths.
- Do not inspect source files, generated artifacts, backend variants,
  module-cache docs, SQLite, vault files, or `.openclerk-eval/vault` directly
  for routine tasks. Do not use repo search, `rg --files`, `find`, `ls`,
  `openclerk --help`, HTTP/MCP, legacy/source-built paths, unsupported
  transports, browser automation, OCR, PPTX parsing, email/chat/form/bundle
  parsing, local file reads, or external acquisition tools as substitutes for
  runner JSON.
- Parallelize runner commands only for documented safe reads:
  `resolve_paths`, `list_documents`, `get_document`, `inspect_layout`,
  retrieval read actions, `source_audit_report` with `mode: "explain"`, and
  `audit_contradictions` with `mode: "plan_only"`. Sequence all writes,
  including `init`, create/ingest/append/replace document actions,
  `compile_synthesis`, `source_audit_report` with `mode: "repair_existing"`,
  and `audit_contradictions` with `mode: "repair_existing"`.
- Durable writes require explicit approval when the agent is proposing a
  candidate path, title, body, source placement, synthesis placement, or
  update-versus-new choice. Public read, fetch, or inspect permission is not
  durable-write approval.

## No-Tools Before Runners

Before runners, answer exactly once with no tools when required fields are
missing and no proposal exception applies, a numeric limit is negative, or the
user asks for a bypass named in Core Guardrails. Do not guess missing fields.
Ask for them by name, or reject the invalid/unsupported workflow and point
back to the OpenClerk runner contract.

Proposal exceptions are valid runner-backed work only when explicit user
content is present:

- document create/validate without a path, title, or body may propose a
  faithful candidate from pasted or explicitly supplied content; explicit
  note/body content with missing path or title is valid runner-backed
  propose-before-create work, not a no-tools missing-field case
- duplicate-risk checks may use runner-visible retrieval/list/get/provenance
  evidence before choosing update versus new
- public-link placement may propose source and synthesis paths before durable
  fetch or write

Opaque screenshots, images, slide decks or PPTX files, email archives,
exported chats, forms, and bundles are unsupported as opaque input unless the
user pasted or explicitly supplied preservable text/body content. Do not claim
parser truth, OCR results, hidden file inspection, attachment contents, or
bundle contents that the user did not supply.

## Workflow Policies

Keep workflow-specific procedure out of this skill. Apply these compact
policies and let runner results drive the answer:

- Candidate documents: preserve explicit user path/title/body/type/naming
  instructions; fill omitted fields only from supplied content; validate with
  `openclerk document` before presenting a candidate; show `Path:`, `Title:`,
  and `Body preview:`; state no document was created; ask for approval before
  any durable write. For note-like candidates without an explicit path, use
  `notes/candidates/<slug-from-title>.md`; derive the title from the content
  subject as a concise singular noun phrase, not request-framing words such as
  save, capture, or note. Include `type: note` frontmatter in note-like body
  previews and a `# <Title>` heading before the supplied body.
- Duplicate checks: when duplicate risk is requested or plausible, use
  runner-visible evidence before validating or writing. Report the likely
  existing target, evidence inspected, and that no document was created or
  updated; ask whether to update the existing target or create a confirmed new
  path.
- Public URL/source intake: use `ingest_source_url` for HTTP/HTTPS PDF and
  public web source ingestion. Do not fetch URLs with browser, HTTP,
  filesystem, or other non-runner tools. When explicit public web URLs lack
  source or synthesis placement, propose source paths and synthesis placement,
  state that no source or synthesis document was created, and ask for approval
  before any durable source fetch or synthesis write. Update mode may omit path
  hints.
- Video/YouTube source intake: use `ingest_video_url` only with user-supplied
  transcript text and provenance. Do not acquire media or transcripts with
  external tools or lower-level storage.
- Document lifecycle review, rollback, restore, and semantic diff: stay inside
  `openclerk document` and `openclerk retrieval`. There is no public history,
  raw diff, review, restore, rollback, or lifecycle action. Use runner-visible
  search/list/get evidence plus `provenance_events` and `projection_states`
  when current state, repair, or freshness matters; preserve the accepted
  target unless the user approves a durable edit. Before lifecycle repair, list
  the likely target collection, get the target document, then inspect
  provenance and projection freshness after any write.
- Messy populated-vault retrieval: answer from runner-visible authority:
  metadata-filtered authority results, active canonical sources, cited source
  paths, `doc_id`, and `chunk_id`. Treat polluted, decoy, stale, draft,
  archived, duplicate, or candidate documents as non-authority unless
  runner-visible source authority says otherwise. If a result is marked
  `status: polluted` or `populated_role: decoy`, explicitly reject that hit as
  not authority and do not repeat its false claim text as a valid answer.
- Synthesis maintenance: prefer `compile_synthesis`; use lower-level document
  and retrieval actions only for explicit primitive or manual cases. Preserve
  `source_refs`, `## Sources`, `## Freshness`, provenance, and projection
  freshness.

Detailed versions of these workflows belong in runner actions, compact runner
help, maintainer/eval docs, or follow-up candidate-surface comparisons, not in
this file.

## Document Tasks

Run document tasks with:

```bash
openclerk document
```

Common actions are `validate`, `create_document`, `ingest_source_url`,
`ingest_video_url`, `list_documents`, `get_document`, `append_document`,
`replace_section`, `resolve_paths`, `inspect_layout`, and `compile_synthesis`.
Use `openclerk document --help` for primitive and promoted workflow-action
request shapes, including source placement, source ingestion, and video fields.

Minimal request shapes:

```json
{"action":"validate","document":{"path":"notes/example.md","title":"Example","body":"# Example\n\nBody."}}
{"action":"list_documents","list":{"path_prefix":"notes/","limit":20}}
{"action":"get_document","doc_id":"doc_id_from_json"}
{"action":"replace_section","doc_id":"doc_id_from_json","heading":"Summary","content":"Updated summary."}
```

Validation rejections are JSON results with `rejected: true` and
`rejection_reason`; runtime failures exit non-zero and write errors to stderr.

## Retrieval Tasks

Run retrieval tasks with:

```bash
openclerk retrieval
```

Common actions are `search`, `document_links`, `graph_neighborhood`,
`records_lookup`, `record_entity`, `services_lookup`, `service_record`,
`decisions_lookup`, `decision_record`, `provenance_events`,
`projection_states`, `audit_contradictions`, `source_audit_report`,
`evidence_bundle_report`, `duplicate_candidate_report`, and
`memory_router_recall_report`, and `hybrid_retrieval_report`. Use
`openclerk retrieval --help` for promoted workflow-action request shape.

Use search for source-grounded answers; document links and graph neighborhoods
for markdown relationships; records, services, and decisions lookup for
promoted-domain projections; provenance for derivation history; and projection
states for freshness. Canonical markdown remains authoritative over derived
service, record, decision, and synthesis projections.

Minimal request shapes:

```json
{"action":"search","search":{"text":"authority evidence","metadata_key":"status","metadata_value":"active","limit":10}}
{"action":"provenance_events","provenance":{"ref_kind":"document","ref_id":"doc_id_from_json","limit":20}}
{"action":"projection_states","projection":{"ref_kind":"document","ref_id":"doc_id_from_json","limit":20}}
```

## Deferred Capability Evidence

Deferred-capability comparison, revisit, or promotion-decision questions are
valid runner-backed evidence tasks when the user asks what existing OpenClerk
documents and retrieval results can prove. Use installed `openclerk document`
and `openclerk retrieval` JSON to inspect runner-visible documents,
citations/source refs, provenance, and projection freshness.

Treat memory transports, `remember`/`recall`, autonomous
router APIs, vector DBs, embeddings, graph memory, and new runner actions as
unsupported only when the user asks you to use, implement, or rely on them as
routine OpenClerk surfaces.

## Answering From Results

Answer the user's substantive question from selected runner JSON fields before
listing evidence. Preserve citation paths, source refs, doc ids, chunk ids,
provenance, projection freshness, validation boundaries, and authority limits
for source-sensitive claims. For retrieval-only repeats, confirm no durable
write only when asked, but still restate the answer and citations.

For unsupported workflows not covered above, say the production OpenClerk
runner does not support that workflow yet.
