---
decision_id: decision-semantic-retrieval-optional-module
decision_title: Semantic Retrieval Optional Module Decision
decision_status: accepted
decision_scope: semantic-retrieval-module
decision_owner: agentops
decision_date: 2026-05-04
source_refs: modules/semantic-retrieval-adapter/module.json, modules/semantic-retrieval-adapter/README.md, docs/evals/results/ockp-semantic-retrieval-adapter-ollama-building-block.md, docs/evals/results/ockp-semantic-recall-local-embeddinggemma-m1.md, docs/evals/results/ockp-semantic-recall-gemini-provider-mimic.md
---
# Decision: Semantic Retrieval Optional Module

## Status

Accepted: add `modules/semantic-retrieval-adapter` as an optional OpenClerk
module and keep it separate from the core retrieval runner.

## Decision

Implement the module shape as the production-quality building block for
semantic retrieval evidence:

- command: `semantic-retrieval-adapter search < request.json`
- providers: `ollama` and explicit `gemini`
- default local model: `embeddinggemma`
- Gemini credential reference: `runtime_config:GEMINI_API_KEY`
- cache: rebuildable user-cache data outside committed artifacts
- output: citation-bearing semantic/hybrid RRF search JSON with provider status,
  cache status, privacy disclosure, model/dimensions, retry/backoff counts, and
  source citations

The module is not automatically loaded by OpenClerk core and is not a new
`openclerk retrieval` subcommand. It gives maintainers and agents an auditable
semantic retrieval surface without silently changing default search.

## Safety, Capability, UX

Safety pass: pass. The module is read-only, redacts provider credentials,
keeps caches outside committed artifacts, reports privacy posture, and forbids
core document writes, direct SQLite mutation, default ranking changes, and
provider config writes. Gemini is never an implicit fallback from an omitted
request field.

Capability pass: pass for building-block status. Local Ollama evidence shows
the embedding path can recover 7/8 semantic-recall rows with citations; Gemini
provider evidence remains a fallback/benchmark path, not local/offline proof.

UX quality: acceptable for agents and maintainers, not yet for normal users.
The optional-module shape is explicit and auditable, but a routine user should
eventually get a simpler retrieval surface after promotion evidence justifies
it.

## Compatibility

Core `openclerk retrieval search` remains lexical. No public core JSON schema,
storage migration, durable core vector index, committed embedding cache, or
hidden provider call is introduced.

## Follow-Up

Created `oc-by5n` to compare whether the optional adapter should later graduate
into default search, an explicit core semantic mode, or stay separate.
