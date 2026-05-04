# Semantic Search Core Hardening For `oc-sloi`

## Summary

`oc-sloi` promoted the evidence-cleared shape into an explicit core
`semantic_search` retrieval action, not a default-search ranking change.

The optional `modules/semantic-retrieval-adapter` was also hardened with
path/tag/metadata filtering so future promotion comparisons are fair.

## Implemented Surface

`openclerk retrieval semantic_search` is explicit and local/offline:

```json
{
  "action": "semantic_search",
  "semantic_search": {
    "query": "semantic recall citation quality",
    "path_prefix": "docs/",
    "limit": 10,
    "embedding_model": "nomic-embed-text"
  }
}
```

It uses loopback Ollama `/api/embed`, stores a rebuildable cache in the user cache
directory, detects stale corpus hashes, collapses duplicate chunks to one hit
per document, and returns citation-bearing `semantic_search` JSON with provider,
cache, privacy, validation-boundary, and authority-limit fields.

## Checks

| Check | Result |
| --- | --- |
| adapter tag/metadata validation | pass |
| adapter filtered chunk loading | pass |
| adapter cache filter isolation | pass |
| core `semantic_search` tag/metadata validation | pass |
| core filtered chunk loading | pass |
| core cache hit | pass |
| core stale cache rebuild | pass |
| core provider blocked state | pass |
| non-loopback Ollama URL rejection | pass |
| core citations | pass |
| hidden Gemini fallback | absent |

Test command:

```bash
mise exec -- go test ./cmd/openclerk ./internal/runner ./modules/semantic-retrieval-adapter
```

## Boundaries

Default `openclerk retrieval search` remains lexical plus its existing zero-hit
lexical fallback. `semantic_search` is explicit because semantic ranking can be
better for recall while still being surprising for exact/source lookup.

Gemini remains explicit benchmark evidence only. The core mode has no provider
config writes, no remote fallback, no committed embedding cache, and no durable
document writes.
