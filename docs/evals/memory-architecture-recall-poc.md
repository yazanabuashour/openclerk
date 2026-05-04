# Memory Architecture And Recall POC

## Scope

This POC compares memory recall candidate surfaces for `oc-uj2y.3`.

The POC does not add a new memory store. It uses existing deterministic runner
fixtures and the installed `memory_router_recall_report` behavior as the
promoted read-only baseline.

Required references:

- docs/architecture/agent-knowledge-plane.md
- https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md
- https://mitchellh.com/writing/building-block-economy
- https://developers.openai.com/api/docs/guides/prompt-guidance
- https://openai.com/index/harness-engineering/
- https://developers.openai.com/api/docs/guides/embeddings
- https://developers.openai.com/api/docs/guides/retrieval
- https://docs.mem0.ai/open-source/overview

## Candidate Shapes

| Shape | What It Proves | What It Does Not Prove |
| --- | --- | --- |
| Current primitives only | Search, list, get, provenance, and projection checks can express memory evidence. | Acceptable routine UX; previous evidence showed high step count and prompt choreography. |
| Source-linked memory docs | Canonical markdown can hold durable memory policy and recall evidence. | Fast repeated recall without a report action. |
| Derived memory projection | Potential read-side acceleration if fully rebuilt from canonical docs. | Correction/delete lifecycle or durable memory authority. |
| Explicit memory write action | Possible future write UX for approved memory notes. | Safety until correction/delete lifecycle, citations, freshness, duplicate handling, privacy, and canonical-conflict behavior are specified. |
| `memory_router_recall_report` | One read-only action returns approved recall fields and no-bypass boundaries. | Autonomous memory writes or a general memory API. |
| Mem0 or external memory | Useful recall architecture comparison. | OpenClerk authority, privacy, freshness, and local-first boundaries. |

## Selected POC Surface

`memory_router_recall_report` remains selected:

```json
{"action":"memory_router_recall_report","memory_router_recall":{"query":"memory router temporal recall session promotion feedback weighting routing canonical docs","limit":10}}
```

The report returns:

- `query_summary`
- `temporal_status`
- `canonical_evidence_refs`
- `stale_session_status`
- `feedback_weighting`
- `routing_rationale`
- `provenance_refs`
- `synthesis_freshness`
- `validation_boundaries`
- `authority_limits`

## Taste Review

A normal user should not have to understand Mem0, router classification,
projection freshness, stale session observations, and provenance calls just to
ask a routine memory/router recall question. A single read-only runner report
is the simpler surface while OpenClerk keeps memory writes out of scope.

## Closure

Safety pass, capability pass, and UX quality are recorded separately in
`docs/evals/results/ockp-memory-architecture-recall-track.md`. Remaining work
is represented by linked beads:

- `oc-tnnw.6.3` eval for safety, capability, and UX quality.
- `oc-tnnw.6.4` promotion decision.
- `oc-tnnw.6.5` conditional implementation only if promoted.
- `oc-tnnw.6.6` iteration and follow-up bead creation.
