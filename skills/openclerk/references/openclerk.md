# Source-Linked Synthesis Workflow

Use source-linked synthesis when the user wants durable knowledge that should
compound beyond the current chat. Compose the workflow from the document and
retrieval runner tasks.

## Routine Flow

1. Search first with `cmd/openclerk-agentops retrieval`.
2. If a relevant synthesis page exists, update it with `append_document` or
   `replace_section`.
3. If no suitable page exists, create one with `create_document`.
4. Preserve source-sensitive claims with citation paths, source refs, or
   provenance references in the Markdown body/frontmatter.
5. Inspect provenance events and projection states when the user asks where
   knowledge came from or whether a synthesis page is stale.

Synthesis pages are durable knowledge artifacts, but they do not outrank the
canonical source docs or promoted records they cite. Treat promoted records as
selective structured domains, not the default wiki mechanism.

## Example Synthesis Create

```bash
printf '%s\n' '{
  "action": "create_document",
  "document": {
    "path": "notes/synthesis/openclerk-knowledge-plane.md",
    "title": "OpenClerk knowledge plane synthesis",
    "body": "---\ntype: synthesis\nstatus: active\nfreshness: fresh\nsource_refs:\n  - notes/architecture/knowledge-plane.md\n---\n# OpenClerk knowledge plane synthesis\n\n## Summary\nSource-linked synthesis of the current architecture.\n\n## Sources\n- notes/architecture/knowledge-plane.md\n"
  }
}' | go run ./cmd/openclerk-agentops document
```

File an answer back into OpenClerk only when it is reusable beyond the current
chat and can point back to source evidence.
