# Semantic Retrieval Adapter

`semantic-retrieval-adapter` is an optional OpenClerk building block. It is not
loaded by OpenClerk core and does not change `openclerk retrieval search`.

## Command

```bash
semantic-retrieval-adapter search < request.json
```

Request:

```json
{
  "query": "semantic recall citation quality",
  "path_prefix": "docs/architecture/",
  "tag": "semantic-retrieval",
  "limit": 10,
  "provider": "ollama",
  "fallback_provider": "gemini"
}
```

The adapter reads OpenClerk documents through the embedded read-only runner
client, applies path-prefix, tag, or metadata filters before chunking, builds
citation-preserving chunks, embeds them with Ollama or Gemini, stores a
rebuildable cache under the user cache directory, and returns
`semantic_retrieval_adapter.v1` JSON with hybrid RRF ranking and citations.

Filter rules match core search validation: `tag` must be non-empty when
provided, `metadata_key` and `metadata_value` must be provided together, and
`tag` cannot be combined with metadata filters.

## Boundaries

- Ollama keeps corpus/query text local when the local service and model are
  available.
- Gemini is explicit provider-backed mode or fallback only and reads
  `runtime_config:GEMINI_API_KEY`; the key is never printed or written back.
- The cache is outside the committed repository and can be deleted/rebuilt.
- Results are retrieval evidence only. Canonical markdown citations and
  approved OpenClerk runner writes remain authority.
- The module performs no durable OpenClerk writes, schema migrations, provider
  config writes, or default search ranking changes.
