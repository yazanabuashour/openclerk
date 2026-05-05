---
decision_id: decision-semantic-retrieval-building-blocks
decision_title: Semantic Retrieval Building Blocks
decision_status: accepted
decision_scope: semantic-retrieval-modules
decision_owner: agentops
decision_date: 2026-05-05
source_refs: docs/architecture/agent-knowledge-plane.md, docs/architecture/semantic-retrieval-optional-module-decision.md, modules/ollama-embeddings/module.json, modules/gemini-embeddings/module.json, modules/semantic-retrieval-adapter/README.md, https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md, https://mitchellh.com/writing/building-block-economy, https://developers.openai.com/api/docs/guides/prompt-guidance, https://openai.com/index/harness-engineering/, https://developers.openai.com/api/docs/guides/embeddings, https://developers.openai.com/api/docs/guides/retrieval, https://docs.mem0.ai/open-source/overview
---
# Decision: Semantic Retrieval Building Blocks

## Status

Accepted: semantic retrieval is a supported optional building-block family,
not a default search replacement.

## Decision

OpenClerk core search remains lexical and authoritative. Explicit
`semantic_search` is available only through installed, enabled, and manifest-
verified optional modules. Core stores module enabled state, manifest digest,
module command, command args, and redacted provider config in SQLite
`runtime_config`; it does not store provider secrets, committed embedding
caches, durable vector indexes, or hidden provider fallback state.

Supported initial modules:

| Module | Provider | Default posture | Credential posture |
| --- | --- | --- | --- |
| `modules/ollama-embeddings/module.json` | `ollama` | local-first default for explicit semantic_search | no credential |
| `modules/gemini-embeddings/module.json` | `gemini` | explicit opt-in provider mode | `runtime_config:GEMINI_API_KEY` only |

Both modules use the shared `openclerk_semantic_retrieval.v1` contract through
`semantic-retrieval-adapter search`. Results must carry citations. Core rejects
module output that lacks citations before returning it as semantic evidence.

## Install, Remove, Configure

Hosts manage modules with `openclerk module`:

- `install_module` verifies an `openclerk-module.v1` manifest, stores the
  manifest SHA-256, command, args, enabled state, and redacted provider config.
- `configure_module` updates enabled state or redacted provider defaults such
  as `embedding_model`, `ollama_url`, `gemini_api_base`, or
  `embedding_output_dimensions`.
- `remove_module` removes OpenClerk's module registration and disables routing
  for that provider without deleting unrelated runtime credentials.
- `list_modules` returns only redacted module/provider state.

Gemini remains explicit opt-in. The Gemini module reads
`runtime_config:GEMINI_API_KEY` from the target database at execution time and
reports only `credential_ref`, request count, retry count, and backoff seconds.
Ollama remains the local-first default provider for explicit semantic_search,
but only after the Ollama module is installed and verified.

## Rationale

The Agent Knowledge Plane keeps durable authority in canonical markdown,
provenance, and projections. Karpathy's LLM Wiki pattern supports the same
direction: sources are curated, durable synthesis compounds over time, and
search tools are optional accelerators rather than the truth layer. Mitchell
Hashimoto's building-block framing fits the module boundary: small composable
parts should be installable and removable without turning core OpenClerk into a
provider-specific bundle.

OpenAI's prompt guidance and harness engineering references support keeping
the agent-facing surface explicit and testable: the runner contract should make
the expected behavior clear, while the harness/module boundary absorbs
provider-specific execution details. OpenAI's embeddings and retrieval docs
support embeddings as a retrieval representation, not authority by themselves.
Mem0 is useful as an external memory reference, but OpenClerk should not adopt
a memory-first truth model or opaque recall layer before citations, provenance,
freshness, and local-first behavior are preserved.

## Safety, Capability, UX

Safety pass: pass. Core runs semantic_search only through a verified module
registration, preserves lexical default search, rejects citation-free module
hits, stores redacted module/provider config, and has no hidden provider
fallback.

Capability pass: pass for building-block status. Ollama and Gemini provider
modules are supported, share one citation-bearing contract, and reuse the
adapter's cache, retry/backoff, and provider-status behavior.

UX quality: acceptable for maintainers and agentic workflows, not promoted as
the default user search experience. Normal users should not be surprised by
provider configuration, model pulls, cache lifecycle, or remote embedding
egress. Future promotion must compare the optional-module shape against simpler
candidate surfaces before default semantic ranking changes.

## Non-Promotion Boundary

This decision does not promote:

- default semantic or hybrid ranking
- hidden remote fallback from Ollama to Gemini
- provider-specific code bundled into core search
- direct SQLite mutation by agents
- committed embedding caches or durable core vector indexes
- citation-free semantic answers

Default `openclerk retrieval search` remains lexical and citation-bearing.
