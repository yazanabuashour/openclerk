---
name: openclerk
description: Use OpenClerk for local-first knowledge-plane tasks over canonical notes, source-linked synthesis, promoted records, provenance-backed retrieval, and projection freshness through the installed openclerk JSON runner.
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
JSON result. The configured local data path is already available through the
environment. For routine requests, do not pass `--data-dir`, `--db`,
`--vault-root`, or `--embedding-provider` unless the user explicitly names a
specific dataset.

The runner honors `OPENCLERK_DATA_DIR`, `OPENCLERK_DATABASE_PATH`, and
`OPENCLERK_VAULT_ROOT`. Keep those paths together by relying on the configured
environment.

## Reject Before Tools

Before using any runner, reject final-answer-only, with exactly one assistant
answer and no tools, when the request:

- is missing required document or retrieval fields
- asks for an obviously invalid limit, such as a negative number
- asks to bypass the runner for routine lower-level runtime, HTTP, SQLite,
  legacy source-built command paths, or unevaluated MCP-style work

For bypass requests, explicitly say the workflow is unsupported and must use
the OpenClerk runner.

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

## Document Tasks

Run document tasks with:

```bash
openclerk document
```

Common request shapes:

```json
{"action":"validate","document":{"path":"notes/projects/example.md","title":"Example","body":"# Example\n\n## Summary\nReusable knowledge.\n"}}
{"action":"create_document","document":{"path":"notes/projects/example.md","title":"Example","body":"# Example\n\n## Summary\nReusable knowledge.\n"}}
{"action":"list_documents","list":{"path_prefix":"notes/","limit":20}}
{"action":"get_document","doc_id":"doc_id_from_json"}
{"action":"append_document","doc_id":"doc_id_from_json","content":"## Decisions\nUse the OpenClerk runner."}
{"action":"replace_section","doc_id":"doc_id_from_json","heading":"Decisions","content":"Use the OpenClerk runner for routine local knowledge tasks."}
{"action":"resolve_paths"}
```

Request fields are `action`, `document`, `doc_id`, `content`, `heading`, and
`list`. A `document` has `path`, `title`, and `body`. A `list` may include
`path_prefix`, `metadata_key`, `metadata_value`, `limit`, and `cursor`.

Validation rejections are normal JSON results with `rejected: true` and
`rejection_reason`. Runtime failures exit non-zero and write errors to stderr.

When writing source-linked synthesis, use this exact AgentOps workflow:

1. Run retrieval `search` for source evidence.
2. Run document `list_documents` with `path_prefix: "notes/synthesis/"` to
   find existing synthesis candidates.
3. Run `get_document` before modifying an existing synthesis page.
4. Prefer `replace_section` or `append_document` over creating duplicates.
5. Inspect `provenance_events` and `projection_states` when the synthesis
   depends on promoted records, services, derivation history, or freshness.

Prototype synthesis pages live under `notes/synthesis/`. Include frontmatter
with `type: synthesis`, `status: active`, `freshness: fresh`, and `source_refs`
set to a single-line comma-separated source path list. Do not use YAML list
syntax for `source_refs`. Include a `## Sources` section with source paths or
citation paths from runner JSON, and a `## Freshness` section that states which
runner retrieval results were checked. Use only documented runner actions, not
`upsert_document` or direct file edits. Synthesis is durable compiled knowledge,
not a higher authority than the canonical sources it cites.

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
{"action":"provenance_events","provenance":{"ref_kind":"document","ref_id":"doc_id_from_json","limit":20}}
{"action":"projection_states","projection":{"ref_kind":"document","ref_id":"doc_id_from_json","limit":20}}
```

Request fields are `action`, `search`, `doc_id`, `chunk_id`, `node_id`,
`entity_id`, `service_id`, `records`, `services`, `provenance`, `projection`,
and `limit`.

Use search for source-grounded answers, document links for explicit markdown
relationships, graph neighborhoods for nearby derived context, records lookup
for promoted record-shaped documents, provenance events for derivation history,
and projection states for freshness. Use services lookup for service-centric
questions before falling back to plain docs search; canonical markdown remains
the source of truth and service records are a derived promoted-domain
projection.

## Answering From Results

Answer from JSON fields such as `document`, `documents`, `search`, `links`,
`graph`, `records`, `entity`, `provenance`, `projections`, `paths`, or
`rejection_reason`.

Preserve citation paths, source refs, chunk ids, and provenance details for
source-sensitive claims. For filtered list answers, mention only returned rows
unless the user explicitly asks about omitted data.
