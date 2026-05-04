---
decision_id: adr-hybrid-retrieval
decision_status: accepted
decision_scope: hybrid-retrieval
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/architecture/agent-knowledge-plane.md, docs/evals/hybrid-retrieval-candidate-comparison-poc.md, docs/evals/results/ockp-hybrid-retrieval-candidate-comparison.md
---

# Hybrid Retrieval ADR

## Context

OpenClerk currently uses local SQLite FTS for source-grounded retrieval.
Search results carry stable document and chunk identifiers plus citations, and
canonical markdown or promoted records remain the authority.

The `oc-uj2y.2` track evaluates whether embeddings, vectors, hybrid fusion, or
hosted vector stores should become retrieval infrastructure. The architecture
constraint is strict: vector search may improve recall, but it must not become
canonical truth or weaken citations, provenance, freshness, duplicate handling,
local-first behavior, or runner-only access.

Relevant public references:

- Karpathy's LLM Wiki pattern values durable wiki synthesis over repeated
  query-time RAG and treats search tools as optional helpers.
- Mitchell Hashimoto's building-block framing favors small composable tools
  over monolithic product surfaces.
- OpenAI embedding and retrieval docs describe vector representations and
  hosted retrieval options as useful retrieval infrastructure, not source
  authority.
- OpenAI prompt and harness guidance reinforces explicit tasks, evidence, and
  harness-owned evaluation.
- Mem0 remains a memory/reference comparison for later recall work, not the
  docs retrieval store for this track.

Reference URLs:

- https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md
- https://mitchellh.com/writing/building-block-economy
- https://developers.openai.com/api/docs/guides/prompt-guidance
- https://openai.com/index/harness-engineering/
- https://developers.openai.com/api/docs/guides/embeddings
- https://developers.openai.com/api/docs/guides/retrieval
- https://docs.mem0.ai/open-source/overview

## Candidates

| Candidate | Safety | Capability | UX quality | Decision |
| --- | --- | --- | --- | --- |
| Keep lexical FTS as default | Pass. Local, citation-bearing, no new store. | Pass for exact/source-sensitive lookup and existing scale evidence. | Acceptable for routine search. | Keep as default. |
| Add only `search.mode` with `lexical`/`hybrid` | Pass only if `hybrid` is not misleading. | Weak without a real second signal. | Poor taste: a normal user expects the mode to change retrieval quality. | Do not promote. |
| Durable local vector index | Potentially viable. Needs embedding provenance, stale-index invalidation, rebuild cost, compression, and citation regression evidence. | Could improve semantic recall. | Useful only after index operations are hidden behind runner behavior. | Not promoted yet. |
| External or hosted vector store | Local-first risk unless opt-in with a stronger privacy and authority model. | Useful as benchmark/reference. | Too much provider ceremony for routine local users. | Reference only. |
| Read-only `hybrid_retrieval_report` | Pass. Uses current runner evidence only and declares boundaries. | Packages baseline evidence and candidate-surface posture without claiming vector ranking. | Improves deferred-capability decision work by avoiding repeated policy choreography. | Promote. |

## Decision

Promote `openclerk retrieval` action `hybrid_retrieval_report`.

The promoted surface is read-only. It runs the current citation-bearing lexical
search baseline, returns candidate-surface comparison guidance, and includes
`agent_handoff`. It does not create embeddings, call embedding APIs, build a
vector store, scan raw vault files, read SQLite directly, or change default
ranking.

## Non-Goals

- No default retrieval ranking change.
- No durable embedding table or vector index.
- No hosted vector-store integration.
- No memory or Mem0 write path.
- No claim that report output is vector-ranked evidence.

## Promotion And Kill Criteria

Future durable hybrid/vector retrieval can be promoted only if targeted evals
show material recall or UX gains while preserving citation correctness,
freshness, provenance, local-first operation, import/reopen performance, and
100 MB/1 GB scale behavior.

Kill or defer any vector path that needs direct SQLite/vault access, obscures
source authority, loses citations, requires routine provider setup, or makes a
normal user manage retrieval infrastructure before a clear quality win exists.
