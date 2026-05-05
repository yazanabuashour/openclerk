---
name: ollama-embeddings
description: Use the optional Ollama embeddings module only when it is installed and enabled, keeping semantic retrieval explicit, local-first, read-only, and citation-bearing.
license: MIT
compatibility: Requires the separate semantic-retrieval-adapter command, the ollama-embeddings module registration, and the production OpenClerk skill plus installed openclerk retrieval runner.
---

# Ollama Embeddings

Use this optional module skill only when the host installed and enabled the
`ollama-embeddings` module. It does not replace the production OpenClerk skill
or change the supported runner surface:

```bash
openclerk retrieval
```

The module is registered through:

```bash
openclerk module
```

## No-Tools Boundary

Before using tools, answer once and do not run semantic retrieval when the
request:

- asks for direct SQLite edits, raw vault inspection, module-cache inspection,
  source-built runner paths, unsupported transports, remote provider fallback,
  or default semantic ranking promotion
- asks the module to write documents, mutate vault files, store credentials,
  commit embedding caches, or replace lexical search
- lacks a complete explicit `semantic_search.query`
- asks for Gemini or remote-provider use through this Ollama module

For unsupported requests, reject and name the forbidden boundary. For missing
facts, ask for the missing public-safe fields.

## Module Contract

Use only explicit retrieval `semantic_search` requests with provider `ollama`.
The result must be `openclerk_semantic_retrieval.v1` compatible and every hit
used in an answer must include citations. Treat semantic similarity as
retrieval evidence only; use `get_document`, `provenance_events`, and
`projection_states` for authority drill-down before any durable write.

For committed docs, fixtures, reduced results, and artifact references, use
repo-relative paths or neutral placeholders, never machine-absolute paths.
