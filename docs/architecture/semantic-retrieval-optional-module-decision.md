---
decision_id: decision-semantic-retrieval-optional-module
decision_title: Semantic Retrieval Optional Module Decision
decision_status: accepted
decision_scope: semantic-retrieval-module
decision_owner: agentops
decision_date: 2026-05-04
source_refs: modules/semantic-retrieval-adapter/module.json, modules/semantic-retrieval-adapter/README.md, modules/ollama-embeddings/module.json, modules/gemini-embeddings/module.json, docs/architecture/semantic-retrieval-building-blocks.md, docs/evals/results/ockp-semantic-retrieval-adapter-ollama-building-block.md, docs/evals/results/ockp-semantic-recall-local-embeddinggemma-m1.md, docs/evals/results/ockp-semantic-recall-gemini-provider-mimic.md
---
# Decision: Semantic Retrieval Optional Module

## Status

Accepted: add semantic retrieval as optional OpenClerk building blocks. Keep
default search lexical, and route explicit core `semantic_search` only through
installed, enabled, manifest-verified modules.

## Decision

Implement the module shape as the production-quality building block for
semantic retrieval evidence:

- command: `semantic-retrieval-adapter search < request.json`
- provider modules: `modules/ollama-embeddings/module.json` and
  `modules/gemini-embeddings/module.json`
- providers: local-first `ollama` and explicit opt-in `gemini`
- default local model: `embeddinggemma`
- Gemini credential reference: `runtime_config:GEMINI_API_KEY`
- cache: rebuildable user-cache data outside committed artifacts
- output: shared `openclerk_semantic_retrieval.v1` citation-bearing
  semantic/hybrid RRF search JSON with provider status, cache status, privacy
  disclosure, model/dimensions, retry/backoff counts, and source citations

Core `openclerk retrieval semantic_search` is explicit and refuses to rank
unless a provider module is installed, enabled, and verified against its
manifest SHA-256 in `runtime_config`. It gives maintainers and agents an
auditable semantic retrieval surface without silently changing default search.

## Safety, Capability, UX

Safety pass: pass. The modules are read-only, redact provider credentials in
core runtime config, keep caches outside committed artifacts, report privacy
posture, and forbid core document writes, direct SQLite mutation, default
ranking changes, and provider config secret writes. Gemini is never an
implicit fallback from an omitted request field.

Capability pass: pass for building-block status. Local Ollama evidence shows
the embedding path can recover 7/8 semantic-recall rows with citations; Gemini
provider evidence remains a fallback/benchmark path, not local/offline proof.

UX quality: acceptable for agents and maintainers, not yet for normal users.
The optional-module shape is explicit and auditable, but a routine user should
eventually get a simpler retrieval surface after promotion evidence justifies
it.

## Compatibility

Core `openclerk retrieval search` remains lexical. `semantic_search` is
explicit and module-gated. No durable core vector index, committed embedding
cache, or hidden provider call is introduced.

## Follow-Up

Created `oc-by5n` to compare whether the optional adapter should later graduate
into default search, an explicit core semantic mode, or stay separate.
