# Memory And Autonomous Router Revisit Comparison POC

## Status

Implemented POC framing for `oc-drb`. This document compares current
`openclerk document` and `openclerk retrieval` memory/router workflows with a
possible promoted memory or autonomous router surface. It does not add runner
actions, schemas, migrations, storage behavior, memory transports, autonomous
router APIs, public API behavior, or shipped skill behavior.

The governing ADR is
[`../architecture/memory-router-revisit-adr.md`](../architecture/memory-router-revisit-adr.md).
The targeted reduced report is
[`results/ockp-memory-router-revisit-pressure.md`](results/ockp-memory-router-revisit-pressure.md).

## Candidate Workflows

| Workflow | Existing primitives | Candidate promoted surface | Notes |
| --- | --- | --- | --- |
| Temporal recall | `search`, `list_documents`, `get_document` over canonical markdown | `memory_router_query` returning cited temporal status | Candidate must not let stale session evidence outrank current canonical docs. |
| Session promotion | Create or inspect canonical markdown and source-linked synthesis with source refs | `memory_router_query` proposing durable promotion candidates | Candidate must keep promotion as canonical markdown, not hidden memory state. |
| Feedback weighting | Retrieve feedback policy and session observation, then explain advisory weight | Candidate response includes feedback weight and source evidence | Candidate must not hide weaker, stale, or conflicting canonical evidence. |
| Routing choice | Use existing document/retrieval actions, provenance, and projection freshness | Candidate response packages route rationale | Candidate must expose why a route was chosen and preserve citations/freshness. |
| Decide promotion | Natural-intent and scripted-control eval rows | No public surface unless final decision promotes | Candidate must prove capability or repeated ergonomics gap through targeted evidence. |

## Ergonomics Scorecard

| Workflow | Candidate promoted surface | Tool or command count | Assistant calls | Wall time | Prompt specificity required | Failure classification | Authority/provenance/freshness risk |
| --- | --- | ---: | ---: | --- | --- | --- | --- |
| Prior reference POC, `memory-router-reference-poc` | None; current document/retrieval workflow | 28 | 9 | 60.93s | Scripted-control | `none` in prior selected pressure | Low: session material became canonical markdown with source refs; routing stayed on existing actions. |
| New natural-intent revisit | None; current document/retrieval workflow | 26 | 5 | 66.91s | Natural user intent | `none` | Low: current workflow preserved canonical memory/router authority, source refs, provenance, synthesis freshness, and bypass boundaries. |
| New scripted-control revisit | None; exact current primitive workflow | 26 | 7 | 45.43s | Scripted-control | `none` | Low: current workflow preserved canonical memory/router authority, source refs, provenance, synthesis freshness, and bypass boundaries. |

Prior measurements come from
[`results/ockp-memory-router-reference-poc.md`](results/ockp-memory-router-reference-poc.md).
New measurements come from
[`results/ockp-memory-router-revisit-pressure.md`](results/ockp-memory-router-revisit-pressure.md).

## Technical Expressibility

Current primitives are technically expressive for both the natural-intent and
scripted-control workflows:

- search memory/router evidence
- list the canonical memory/router documents by path prefix
- retrieve temporal, feedback, routing, and session observation documents
- inspect provenance for the session observation
- retrieve source-linked synthesis
- inspect synthesis projection freshness
- cite source paths and keep routing on existing AgentOps document/retrieval
  actions

Neither revisit row proved a `capability_gap`; both completed with
classification `none`.

## UX Acceptability

The natural-intent row is acceptable enough to keep as reference pressure. It
completed through the existing runner-visible document/retrieval workflow
without unsupported memory transports, `remember`/`recall`, autonomous router
APIs, vector DBs, embeddings, graph memory, new runner actions, writes, or
bypasses.

## Compatibility Expectations

Any future promoted surface must:

- keep canonical markdown as durable memory and routing authority
- return source refs or citations for memory/router claims
- expose provenance and projection freshness instead of hiding derived state
- keep feedback weighting advisory
- avoid direct SQLite, direct vault inspection, broad repo search,
  source-built runner paths, HTTP/MCP bypasses, backend variants,
  module-cache inspection, memory transports, remember/recall actions,
  autonomous router APIs, and unsupported transports
- preserve final-answer-only invalid-request behavior
- remain backward compatible with existing `openclerk document` and
  `openclerk retrieval` workflows

## POC Conclusion

The refreshed targeted pressure lane does not justify promotion. The natural
and scripted rows both completed with classification `none`, so the evidence
supports keeping the lane as reference pressure.

No memory API, remember/recall action, autonomous router surface, schema,
migration, storage behavior, or public API is authorized from this POC.

The final decision is recorded in
[`../architecture/memory-router-revisit-promotion-decision.md`](../architecture/memory-router-revisit-promotion-decision.md).
