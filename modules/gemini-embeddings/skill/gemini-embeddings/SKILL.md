---
name: gemini-embeddings
description: Use the optional Gemini embeddings module only after explicit opt-in, keeping remote semantic retrieval read-only, rate-limit-aware, redacted, and citation-bearing.
license: MIT
compatibility: Requires the separate semantic-retrieval-adapter command, the gemini-embeddings module registration, runtime_config:GEMINI_API_KEY, and the production OpenClerk skill plus installed openclerk retrieval runner.
---

# Gemini Embeddings

Use this optional module skill only when the host installed and enabled the
`gemini-embeddings` module and the user explicitly requests Gemini/provider
semantic retrieval. It does not replace the production OpenClerk skill or
change the supported runner surface:

```bash
openclerk retrieval
```

The module is registered through:

```bash
openclerk module
```

Gemini credentials are read from `runtime_config:GEMINI_API_KEY` in the target
database. Do not print, copy, infer, fetch, or write the key.

## No-Tools Boundary

Before using tools, answer once and do not run semantic retrieval when the
request:

- asks for direct SQLite edits, raw vault inspection, module-cache inspection,
  source-built runner paths, unsupported transports, hidden provider fallback,
  or default semantic ranking promotion
- asks the module to write documents, mutate vault files, store credentials,
  commit embedding caches, or replace lexical search
- lacks a complete explicit `semantic_search.query`
- asks for Gemini without explicit provider opt-in

For unsupported requests, reject and name the forbidden boundary. For missing
facts, ask for the missing public-safe fields.

## Module Contract

Use only explicit retrieval `semantic_search` requests with provider `gemini`.
The result must be `openclerk_semantic_retrieval.v1` compatible and every hit
used in an answer must include citations. Provider status may include redacted
credential refs, request counts, retry counts, and backoff seconds. Treat
semantic similarity as retrieval evidence only; use `get_document`,
`provenance_events`, and `projection_states` for authority drill-down before
any durable write.

For committed docs, fixtures, reduced results, and artifact references, use
repo-relative paths or neutral placeholders, never machine-absolute paths.
