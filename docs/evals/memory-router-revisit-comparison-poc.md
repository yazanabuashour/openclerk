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
| New natural-intent revisit | Possible `memory_router_query` if repeated natural UX fails | 0 | 2 | 19.53s | Natural user intent | `ergonomics_gap` | Medium: agent did not run the current-primitives workflow, so provenance/freshness were not inspected. |
| New scripted-control revisit | None; exact current primitive workflow | 26 | 5 | 60.49s | Scripted-control | `skill_guidance_or_eval_coverage` | Low to medium: runner-visible evidence existed, but the final answer did not satisfy the full decision contract. |

Prior measurements come from
[`results/ockp-memory-router-reference-poc.md`](results/ockp-memory-router-reference-poc.md).
New measurements come from
[`results/ockp-memory-router-revisit-pressure.md`](results/ockp-memory-router-revisit-pressure.md).

## Technical Expressibility

Current primitives appear technically expressive for the scripted workflow:

- search memory/router evidence
- list the canonical memory/router documents by path prefix
- retrieve temporal, feedback, routing, and session observation documents
- inspect provenance for the session observation
- retrieve source-linked synthesis
- inspect synthesis projection freshness
- cite source paths and keep routing on existing AgentOps document/retrieval
  actions

The scripted-control row did not prove a `capability_gap`; its durable evidence
was present, and the failure was classified as
`skill_guidance_or_eval_coverage` because the assistant answer did not satisfy
the full comparison contract.

## UX Acceptability

The natural-intent row is not acceptable enough to keep as a clean reference
decision. It failed before using runner tools, producing an `ergonomics_gap`
classification for the natural workflow.

This is not sufficient promotion evidence by itself because promotion via
ergonomics requires repeated natural failures and a passing scripted control.
The current evidence instead supports guidance or eval repair followed by a
rerun.

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

The refreshed targeted pressure lane does not justify promotion. It also does
not justify a clean keep-reference decision: natural intent failed as
`ergonomics_gap`, while the scripted control reached runner-visible evidence
but failed the answer contract as `skill_guidance_or_eval_coverage`.

The evidence supports deferring promotion for guidance/eval repair and rerun,
with no memory API, remember/recall action, autonomous router surface, schema,
migration, storage behavior, or public API authorized from this POC.

The final decision is recorded in
[`../architecture/memory-router-revisit-promotion-decision.md`](../architecture/memory-router-revisit-promotion-decision.md).
