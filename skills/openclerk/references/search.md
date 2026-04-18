# Search Task Recipes

Use `go run ./cmd/openclerk-agentops retrieval` for source-grounded search. It
returns `search.hits` with `doc_id`, `chunk_id`, snippets, and citations that
point back to canonical Markdown paths and line ranges.

## Search

```bash
printf '%s\n' '{"action":"search","search":{"text":"architecture","limit":10}}' |
  go run ./cmd/openclerk-agentops retrieval
```

For scoped questions, add path or metadata filters:

```bash
printf '%s\n' '{
  "action": "search",
  "search": {
    "text": "roadmap",
    "path_prefix": "notes/projects/",
    "metadata_key": "status",
    "metadata_value": "active",
    "limit": 10
  }
}' | go run ./cmd/openclerk-agentops retrieval
```

Search before creating source-linked synthesis. If a relevant topic, entity,
comparison, or overview page already exists, update it with a document task
instead of creating a duplicate. When answering source-sensitive questions,
report from the returned citations rather than unsupported memory.
