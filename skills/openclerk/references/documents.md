# Document Task Recipes

Use `go run ./cmd/openclerk-agentops document` for routine document work. It
reads an `agentops.DocumentTaskRequest` as JSON and returns a
`DocumentTaskResult` with `rejected`, `rejection_reason`, `document`,
`documents`, `paths`, `page_info`, and `summary`.

## Create A Canonical Document

```bash
printf '%s\n' '{
  "action": "create_document",
  "document": {
    "path": "notes/projects/openclerk-roadmap.md",
    "title": "Roadmap",
    "body": "---\ntype: project\nstatus: active\n---\n# Roadmap\n\n## Summary\nCanonical project note.\n"
  }
}' | go run ./cmd/openclerk-agentops document
```

`create_document` rejects missing `path`, `title`, or `body` before opening the
runtime. Duplicate paths fail as runtime errors; do not overwrite a whole
document unless the user explicitly asks.

## List, Read, Append, And Replace

```bash
printf '%s\n' '{"action":"list_documents","list":{"path_prefix":"notes/","metadata_key":"status","metadata_value":"active","limit":20}}' |
  go run ./cmd/openclerk-agentops document

printf '%s\n' '{"action":"get_document","doc_id":"doc_id_from_json"}' |
  go run ./cmd/openclerk-agentops document

printf '%s\n' '{"action":"append_document","doc_id":"doc_id_from_json","content":"## Decisions\nUse the AgentOps runner."}' |
  go run ./cmd/openclerk-agentops document

printf '%s\n' '{"action":"replace_section","doc_id":"doc_id_from_json","heading":"Decisions","content":"Use `cmd/openclerk-agentops` for routine agent workflows."}' |
  go run ./cmd/openclerk-agentops document
```

Use `append_document` or `replace_section` for incremental updates. Preserve
unrelated content and existing source refs.

## Resolve Paths

```bash
printf '%s\n' '{"action":"resolve_paths"}' |
  go run ./cmd/openclerk-agentops document
```

Use this when the user asks which local OpenClerk dataset was checked. Report
the returned `data_dir`, `database_path`, and `vault_root` only when useful.
