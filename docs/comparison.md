# OpenClerk Comparison

OpenClerk is best understood as the governance layer for LLM-maintained
Markdown wikis: canonical Markdown stays human-readable, while retrieval,
provenance, freshness, duplicate checks, and write approval stay explicit.

| Approach | Main idea | Canonical store | Agent write boundary | Citations | Staleness | Duplicate handling | Local-first |
|---|---|---|---|---|---|---|---|
| OpenClerk | Runner-governed agent memory over Markdown | Markdown vault | Strict JSON runner; durable writes require approval | Always returned on retrieval | Explicit projection freshness and stale synthesis reports | First-class duplicate candidate reports | Yes |
| Basic Memory | Shared Markdown knowledge base for humans and AI assistants | Markdown files | MCP/tool-mediated assistant writes | Tool/result dependent | Project/index dependent | Project/workflow dependent | Local-first core; cloud option exists |
| Obsidian plugin | Human notes with plugin automation | Obsidian vault | Plugin-specific | Plugin-specific | Usually manual or plugin-specific | Usually manual or plugin-specific | Yes |
| MCP memory server | Persistent memory exposed through MCP tools | Server-specific | Tool/server policy | Server-specific | Server-specific | Server-specific | Depends on server |
| RAG pipeline | Retrieve chunks from an index | Index or upstream corpus | Usually no durable document-write contract | Varies | Usually index-refresh dependent | Usually ad hoc | Depends on stack |
| NotebookLM | Hosted notebook over uploaded sources | Hosted service | No local agent-write contract | Yes | Hosted source state | Hosted product behavior | No |

Use OpenClerk when the hard part is not search alone, but keeping agent-readable
memory auditable: source refs, citations, stale projections, duplicate risk,
and approved writes.

Use a lighter memory tool when you mainly need quick note capture or assistant
recall without a strict provenance and write-approval contract.

Sources to verify before relying on this page as a buyer matrix:

- OpenClerk: `README.md`
- Basic Memory: <https://docs.basicmemory.com/start-here/what-is-basic-memory>
- MCP servers: <https://github.com/modelcontextprotocol/servers>

