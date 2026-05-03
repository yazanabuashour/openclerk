# Compile Synthesis Workflow Action Candidate Comparison

Bead: `oc-e8om`

## Decision

Select Candidate C: promote narrow document action `compile_synthesis` and keep
existing document/retrieval primitives for advanced or manual synthesis repair.

## Candidates

| Candidate | Shape | Safety | Capability | UX |
| --- | --- | --- | --- | --- |
| A | Keep current primitives with smaller skill. | Pass when agents search, list candidates, inspect existing target, check provenance/freshness, and update without duplicates. | Pass. Existing primitives can express the workflow. | Fails routine UX: repeated exact JSON, ordering, and response field choreography. |
| B | Promote one narrow `compile_synthesis` action. | Pass if the action preserves source refs, required sections, duplicate checks, provenance/freshness, and authority limits. | Pass. The runner can compose existing create/update, registry, provenance, projection, and duplicate checks. | Better routine UX, but advanced cases still need primitives. |
| C | Narrow action plus existing primitives. | Pass. Keeps guardrails in the action and preserves manual escape hatches. | Pass. No new storage or broad synthesis engine. | Selected. Natural prompts can route to one workflow action instead of a long skill recipe. |

## Selected Surface

`compile_synthesis` creates or updates exactly one `synthesis/` target with
required `path`, `title`, non-empty `source_refs`, `body`, and
`mode: "create_or_update"`. It preserves single-line `source_refs`, requires
`## Sources` and `## Freshness`, avoids duplicate synthesis paths, and returns
selected path, source evidence, duplicate status, provenance refs, projection
freshness, write status, validation boundaries, and authority limits.

## Evidence Requirements

The targeted lane is `ockp-compile-synthesis-workflow-action`. Candidate A/B/C
rows must report tool/command count, assistant turns, prompt specificity,
failure/retry rate, `safety_pass`, `capability_pass`, and `ux_quality`.
Scripted primitive success is capability evidence only; the natural workflow
action row must pass to claim UX.
