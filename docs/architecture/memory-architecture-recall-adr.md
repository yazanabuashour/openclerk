---
decision_id: adr-memory-architecture-recall
decision_status: accepted
decision_scope: memory-architecture-recall
decision_owner: agentops
decision_date: 2026-05-04
source_refs: docs/architecture/agent-knowledge-plane.md, docs/architecture/memory-router-recall-report-implementation-decision.md, docs/evals/memory-architecture-recall-poc.md, docs/evals/results/ockp-memory-architecture-recall-track.md
---

# Memory Architecture And Recall ADR

## Context

The `oc-uj2y.3` track evaluates memory as source-linked recall, not canonical
truth. OpenClerk already has a read-only `memory_router_recall_report` action
under `openclerk retrieval`. This ADR reconciles the broader memory
architecture candidates against that existing implementation.

OpenClerk's memory rules are:

- canonical markdown and promoted records remain durable authority
- memory is recall support, not a source of truth
- source refs, provenance, and freshness must be visible
- stale session observations cannot outrank current canonical evidence
- feedback weighting is advisory
- durable memory writes, `remember`/`recall` actions, autonomous routers, Mem0
  transports, vector stores, and hidden ranking need separate promotion

Reference URLs:

- https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md
- https://mitchellh.com/writing/building-block-economy
- https://developers.openai.com/api/docs/guides/prompt-guidance
- https://openai.com/index/harness-engineering/
- https://developers.openai.com/api/docs/guides/embeddings
- https://developers.openai.com/api/docs/guides/retrieval
- https://docs.mem0.ai/open-source/overview

## Candidate Options

| Candidate | Safety | Capability | UX quality | Decision |
| --- | --- | --- | --- | --- |
| No separate memory layer | Pass. Avoids truth drift. | Weak repeated recall and personalization story. | Poor for repeated memory/router questions. | Not enough. |
| Source-linked memory docs | Pass if canonical markdown owns memory facts. | Good for inspected durable recall. | Acceptable but can require several primitive calls. | Keep as authority pattern. |
| Internal SQLite memory projection | Possible if fully derived from canonical docs. | Useful later for typed recall. | Too early without stale/supersession lifecycle evidence. | Not promoted. |
| Mem0 integration | Useful external memory reference. | Could help cross-session recall. | Too much new transport and authority surface now. | Reference only. |
| Existing `memory_router_recall_report` | Pass. Read-only, source-linked, no memory writes. | Returns temporal status, canonical refs, stale-session posture, provenance, freshness, and boundaries. | Pass. One runner action replaces high-step choreography. | Promote/keep. |

## Decision

Use the existing `memory_router_recall_report` as the promoted memory recall
surface for this track.

Do not add autonomous durable memory writes, `remember`/`recall` APIs, Mem0
transport, vector memory, graph memory, hidden authority ranking, or a memory
router. The report is the correct current shape because it gives agents
source-linked recall evidence without creating a second truth system.

## Non-Goals

- No memory-first canonical truth model.
- No automatic session-to-memory promotion.
- No write action for memory.
- No Mem0 dependency or transport.
- No hidden ranking that can suppress stale or conflicting canonical evidence.

## Promotion And Kill Criteria

Future memory writes or memory projections require evidence that they improve
repeated recall without increasing truth drift, leaking private evidence,
hiding stale state, or weakening approval-before-durable-write. Kill any memory
surface that cannot show source refs, freshness, provenance, and canonical
override behavior.
