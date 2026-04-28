# Graph Semantics Revisit Comparison POC

## Status

Implemented POC framing for `oc-9jn`. This document compares current
`openclerk document` and `openclerk retrieval` graph workflows with a possible
promoted graph semantics surface. It does not add runner actions, schemas,
migrations, storage behavior, public API behavior, or shipped skill behavior.

The governing ADR is
[`../architecture/graph-semantics-revisit-adr.md`](../architecture/graph-semantics-revisit-adr.md).
The targeted reduced report is
[`results/ockp-graph-semantics-revisit-pressure.md`](results/ockp-graph-semantics-revisit-pressure.md).

## Candidate Workflows

| Workflow | Existing primitives | Candidate promoted surface | Notes |
| --- | --- | --- | --- |
| Explain relationship meaning | `search`, `list_documents`, `get_document` | `graph_semantics_query` returning cited relationship snippets | Candidate could reduce choreography, but markdown remains semantic authority. |
| Navigate graph context | `document_links`, incoming backlinks, `graph_neighborhood` | `graph_semantics_query` combining links, backlinks, and neighborhood | Candidate must keep structural edges derived and cited. |
| Verify freshness | `projection_states` for projection `graph` | Candidate response includes graph projection freshness | Candidate must expose freshness instead of hiding stale derived graph state. |
| Decide promotion | Natural-intent and scripted-control eval rows | No public surface unless final decision promotes | Candidate must prove capability or ergonomics gap through repeated targeted evidence. |

## Ergonomics Scorecard

| Workflow | Candidate promoted surface | Tool or command count | Assistant calls | Wall time | Prompt specificity required | Failure classification | Authority/provenance/freshness risk |
| --- | --- | ---: | ---: | --- | --- | --- | --- |
| Prior reference POC, `graph-semantics-reference-poc` | None; current document/retrieval workflow | 14 | 4 | 27.44s | Scripted-control | `none` in prior selected pressure | Low: canonical markdown carried relationship meaning; graph output stayed structural and cited. |
| Baseline navigation, `canonical-docs-navigation-baseline` | None; current document/retrieval workflow | 16 | 5 | 35.64s | Scripted-control | `none` in prior selected pressure | Low: links, backlinks, graph neighborhood, and projection freshness were inspectable. |
| New natural-intent revisit | Possible `graph_semantics_query` if repeated natural UX fails | 26 | 7 | 80.88s | Natural user intent | `ergonomics_gap` | Medium if a promoted surface hides markdown evidence, citations, or graph freshness. |
| New scripted-control revisit | None; exact current primitive workflow | 28 | 6 | 86.21s | Scripted-control | `skill_guidance_or_eval_coverage` | Low if search, document retrieval, links, graph neighborhood, and freshness all remain visible. |

Prior measurements come from
[`results/ockp-graph-semantics-reference-poc.md`](results/ockp-graph-semantics-reference-poc.md).
New measurements come from
[`results/ockp-graph-semantics-revisit-pressure.md`](results/ockp-graph-semantics-revisit-pressure.md).

## Technical Expressibility

Current primitives can express relationship-shaped graph tasks when the agent
uses the documented workflow:

- search canonical markdown for relationship wording such as requires,
  supersedes, related to, and operationalizes
- retrieve the canonical document before interpreting relationship meaning
- inspect outgoing links and incoming backlinks
- inspect graph neighborhood context
- inspect graph projection freshness
- cite canonical markdown and keep graph state derived

This means a promotion decision should not treat high command count alone as a
structural insufficiency. A `capability_gap` requires scripted-control failure:
the current primitives must be unable to express the workflow safely even with
exact instructions.

## UX Acceptability

The open question is whether the current workflow is acceptable under natural
routine intent. Natural prompts should not have to prescribe every request
shape, but they may name the evidence the answer must preserve: relationship
text, citations, links/backlinks, graph neighborhood, and projection freshness.

An `ergonomics_gap` requires repeated natural-intent evidence showing the
workflow is too brittle, too slow, too many steps, too retry-prone, or too
dependent on graph-specific prompt choreography. The scripted control must pass
to prove the pressure is UX/reliability cost rather than structural
insufficiency.

## Compatibility Expectations

Any future promoted surface must:

- keep canonical markdown as the source of semantic relationship authority
- return citation/source evidence for relationship claims
- expose graph projection freshness and provenance-relevant identifiers
- avoid direct SQLite, direct vault inspection, broad repo search,
  source-built runner paths, HTTP/MCP bypasses, backend variants, and
  module-cache inspection
- preserve final-answer-only invalid-request behavior
- remain backward compatible with existing `openclerk document` and
  `openclerk retrieval` workflows

## POC Conclusion

The targeted pressure lane found one natural-intent `ergonomics_gap` and one
scripted-control `skill_guidance_or_eval_coverage` row. The scripted-control
database evidence passed, so the failure does not prove structural
insufficiency. The current evidence supports deferring promotion for guidance or
eval repair rather than creating a graph semantics implementation follow-up.

The final decision is recorded in
[`../architecture/graph-semantics-revisit-promotion-decision.md`](../architecture/graph-semantics-revisit-promotion-decision.md).
