# Synthesis Compile Revisit Comparison POC

## Status

Implemented POC framing for `oc-ayo`. This document compares the current
source-linked synthesis workflow with a possible `compile_synthesis` surface.
It does not add runner actions, schemas, migrations, storage behavior, public
API behavior, or shipped skill behavior.

The governing ADR is
[`../architecture/synthesis-compile-revisit-adr.md`](../architecture/synthesis-compile-revisit-adr.md).
The targeted reduced report is
[`results/ockp-synthesis-compile-revisit-pressure.md`](results/ockp-synthesis-compile-revisit-pressure.md).

## Candidate Workflows

| Workflow | Existing primitives | Candidate promoted surface | Notes |
| --- | --- | --- | --- |
| Create new source-linked synthesis | `search`, `list_documents`, `create_document` | `compile_synthesis` with `mode: create_or_update` | Candidate could reduce call count, but must still prove source evidence and duplicate checks. |
| Update existing synthesis | `search`, `list_documents`, `get_document`, `projection_states`, `replace_section` or `append_document` | `compile_synthesis` targeting the existing path | Candidate could combine candidate discovery, freshness inspection, and update. |
| Repair stale synthesis | `search`, `list_documents`, `get_document`, `projection_states`, `provenance_events`, `replace_section` or `append_document` | `compile_synthesis` with freshness-aware response | Candidate must expose freshness and provenance rather than hide them behind a write result. |
| Mixed records and synthesis | `records_lookup`, `provenance_events`, `projection_states`, document writes | `compile_synthesis` plus optional promoted-record evidence fields | Candidate must keep canonical docs and promoted records higher authority than synthesis. |

## Ergonomics Scorecard

| Workflow | Candidate promoted surface | Tool or command count | Assistant calls | Wall time | Prompt specificity required | Failure classification | Authority/provenance/freshness risk |
| --- | --- | ---: | ---: | --- | --- | --- | --- |
| Prior scripted source-linked create, `search-synthesis` | None; current document/retrieval workflow | 16 | 6 | 37.91s | Scenario-specific | `none` in selected prior pressure | Low when source refs and freshness text are preserved; high call count remains visible. |
| Prior scripted stale update, `stale-synthesis-update` | None; current document/retrieval workflow | 12 | 5 | 42.33s | Scripted-control | `none` in selected prior pressure | Low after existing synthesis is retrieved and updated rather than duplicated. |
| Prior scripted candidate pressure, `synthesis-candidate-pressure` | `compile_synthesis` could combine candidate selection and update | 30 | 10 | 61.47s | High scenario-specific choreography | `ergonomics_gap` pressure if repeated under natural intent | Medium: candidate selection and freshness inspection are easy to skip if hidden. |
| Prior scripted multi-turn drift repair, `mt-synthesis-drift-pressure` | `compile_synthesis` could combine drift repair and freshness response | 40 | 11 | 111.63s | High scenario-specific choreography | `ergonomics_gap` pressure if natural intent is brittle | Medium-high: provenance and final freshness must stay inspectable after repair. |
| New natural-intent revisit | `compile_synthesis` as one narrow create-or-update call | Measured by targeted eval | Measured by targeted eval | Measured by targeted eval | Natural user intent | Classified by targeted eval | Must not create duplicate synthesis, drop `source_refs`, or hide freshness. |
| New scripted-control revisit | None; exact current primitive workflow | Measured by targeted eval | Measured by targeted eval | Measured by targeted eval | Scripted-control | Classified by targeted eval | Low if the control preserves current invariant checks. |

Prior measurements come from
[`results/ockp-synthesis-compiler-pressure.md`](results/ockp-synthesis-compiler-pressure.md).
The prior maintenance decision in
[`results/ockp-synthesis-maintenance-ergonomics.md`](results/ockp-synthesis-maintenance-ergonomics.md)
deferred product/API promotion because existing pressure did not show repeated
runner insufficiency.

## Technical Expressibility

Current primitives can express source-linked synthesis safely when the agent
follows the documented workflow. The prior pressure lane created, updated, and
repaired synthesis pages using installed runner JSON while preserving
single-line `source_refs`, `## Sources`, `## Freshness`, candidate discovery,
projection freshness, and no-bypass invariants.

The technical risk is not the inability to write markdown. The risk is whether
correct synthesis requires enough separate discovery and inspection steps that
routine agents skip one of them. A promoted `compile_synthesis` would need to
make those checks explicit in its request or response; otherwise it would
reduce visible workflow cost by hiding the evidence that makes the write safe.

## UX Acceptability

The current scripted workflow is acceptable for precise control prompts and
maintenance tasks. It becomes questionable when routine natural user intent
requires the user or scenario to specify every runner step: search source
evidence, list candidates, retrieve before editing, inspect freshness, inspect
provenance, then choose replace or append.

The targeted eval must therefore compare natural intent against a scripted
control. Promotion should not follow from high command counts alone. It should
follow only if natural intent repeatedly produces duplicate synthesis, missed
candidate discovery, skipped freshness, dropped source refs, missing citations,
or excessive retry/guidance dependence while the scripted control proves the
current primitives remain technically sufficient.

## POC Conclusion

Do not promote `compile_synthesis` from the POC alone. Run the targeted revisit
pressure lane, then decide in
[`../architecture/synthesis-compile-revisit-promotion-decision.md`](../architecture/synthesis-compile-revisit-promotion-decision.md)
whether to promote, defer, kill, or keep the candidate as reference.
